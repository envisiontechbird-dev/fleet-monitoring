package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"net/http"
	"sync"
	"time"
	"github.com/gorilla/mux"
	"encoding/json"
)

type Device struct {
	ID string
	Heartbeats []time.Time
	UploadTimes []int
	mutex sync.RWMutex
}

var devices = make(map[string]*Device)
var globalMutex sync.RWMutex

func main() {
	port := flag.String("port", "8080", "HTTP port")
	csvPath := flag.String("csv", "devices.csv", "Path to devices CSV")
	flag.Parse()

	if err := loadDevices(*csvPath); err != nil {
		log.Fatalf("Failed to load devices: %v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/devices/{id}/heartbeat", handleHeartbeat).Methods("POST")
	r.HandleFunc("/devices/{id}/stats", handleStats).Methods("POST")
	r.HandleFunc("/devices/{id}/stats", getStats).Methods("GET")
	fmt.Printf("Starting server on port %s...\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, r))
}

func loadDevices(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil {
		return err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	globalMutex.Lock()
	defer globalMutex.Unlock()

	for _, record := range records {
		deviceID := record[0]
		devices[deviceID] = &Device{
			ID: deviceID,
			Heartbeats: make([]time.Time, 0),
			UploadTimes: make([]int, 0),
		}
		fmt.Printf("Loaded device: %s\n", deviceID)
	}
	return nil
}

func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["id"]

	var req struct {
		SentAt time.Time `json:"sent_at"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request format", http.StatusBadRequest)
		return
	}

	globalMutex.RLock()
	device, exists := devices[deviceID]
	globalMutex.RUnlock()
	
	if !exists {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	device.mutex.Lock()
	device.Heartbeats = append(device.Heartbeats, req.SentAt)
	device.mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["id"]

	var req struct {
		SentAt time.Time `json:"sent_at"`
		UploadTime int `json:"upload_time"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request format", http.StatusBadRequest)
		return
	}

	globalMutex.RLock()
	device, exists := devices[deviceID]
	globalMutex.RUnlock()
	
	if !exists {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	device.mutex.Lock()
	device.UploadTimes = append(device.UploadTimes, req.UploadTime)
	device.mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
}

func getStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["id"]
	globalMutex.RLock()
	device, exists := devices[deviceID]
	globalMutex.RUnlock()
	
	if !exists {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	device.mutex.RLock()
	defer device.mutex.RUnlock()

	var uptime float64 = 0
	if len(device.Heartbeats) > 0 {
		first := device.Heartbeats[0]
		last := device.Heartbeats[0]
		
		for _, hb := range device.Heartbeats {
			if hb.Before(first) {
				first = hb
			}
			if hb.After(last) {
				last = hb
			}
		}
		
		minutesBetween := last.Sub(first).Minutes()
		
		if minutesBetween > 0 {
			uptime = float64(len(device.Heartbeats)) / minutesBetween * 100
		}
	}

	var avgUpload int = 0
	if len(device.UploadTimes) > 0 {
		sum := 0
		for _, t := range device.UploadTimes {
			sum += t
		}
		avgUpload = sum / len(device.UploadTimes)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Uptime float64 `json:"uptime"`
		AvgUploadTime string `json:"avg_upload_time"`
	}{
		Uptime: uptime,
		AvgUploadTime: fmt.Sprintf("%d", avgUpload),
	})
}