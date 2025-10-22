--Fleet Monitoring Service ;
A monitoring service for SafelyYou's fleet of edge devices that tracks device heartbeats and upload statistics.

--Overview ;
This service provides an API for monitoring device health and performance metrics. It tracks heartbeat signals to calculate uptime percentages and records upload times to determine average performance.

--Running Locally ;
1. Clone the repository
2. Install dependencies: 
    go mod download
3. Run the service
    go run main.go
4. http://localhost:8080

--Running the Docker ;
1. Build and run the container:
    docker-compose up --build
2. http://localhost:8080

--Challenges and Solutions ;
During development, I faced a few interesting challenges:

Docker Build Issues: Ran into errors with the Go version format in go.mod (had to use major.minor format) and Dockerfile keyword casing (changed "as" to "AS").

Concurrency: Implemented mutex locks to safely handle concurrent API requests accessing shared device data.