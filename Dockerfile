FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o fleet-monitoring .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/fleet-monitoring /app/
COPY devices.csv /app/

WORKDIR /app

EXPOSE 8080

CMD ["./fleet-monitoring"]