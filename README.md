# Radoise Server

A simple Go REST API server that provides endpoints for controlling an MPD (Music Player Daemon) instance. The server allows you to play music files, pause playback, and control volume through HTTP requests.

## Prerequisites

- Go 1.25.0 or later
- MPD (Music Player Daemon) running on localhost:6600
- Audio files available in MPD's music directory

## Development

### Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd radoise-server
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Ensure MPD is running and accessible on localhost:6600:
   ```bash
   # Check if MPD is running
   systemctl status mpd

   # Or start MPD if not running
   systemctl start mpd
   ```

### Running the Server

Start the development server:

```bash
go run main.go
```

The server will start on port 3000 and log "API on :3000" when ready.

### API Endpoints

The server provides the following endpoints:

- **GET/POST `/play`** - Clears the current playlist, adds `wn.mp3`, and starts playback
- **GET/POST `/pause`** - Pauses the current playback
- **GET/POST `/volume`** - Sets the volume to 50%

All endpoints support CORS and return plain text responses.

## Configuration

Current configuration is hardcoded in `main.go`:

- **Port**: 3000
- **MPD Address**: localhost:6600
- **Audio File**: wn.mp3
- **Volume Level**: 50%

Modify these values in the source code before building as needed for your environment.

## Deployment

### 1. Building for Production

Create a production binary:

```bash
# Build for current platform
go build -o radoise-server main.go

# Build for Linux (if cross-compiling)
GOOS=linux GOARCH=amd64 go build -o radoise-server-linux main.go
```

### 2a. Running in Production

1. Copy the binary to your production server
2. Ensure MPD is installed and configured
3. Run the server:

```bash
./radoise-server
```

### 2b. Systemd Service (Optional)

Create a systemd service file at `/etc/systemd/system/radoise-server.service`:

```ini
[Unit]
Description=Radoise Music Server
After=network.target mpd.service

[Service]
Type=simple
ExecStart=/path/to/radoise-server/radoise-server

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable radoise-server
sudo systemctl start radoise-server
```

### 2c. Docker Deployment (Alternative)

Create a `Dockerfile`:

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o radoise-server main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/radoise-server .
EXPOSE 3000
CMD ["./radoise-server"]
```

Build and run:

```bash
docker build -t radoise-server .
docker run -p 3000:3000 --network host radoise-server
```

Note: The `--network host` flag is needed for the container to access MPD on localhost.

## Changes

See the [commit history](https://github.com/graysonlee123/radoise-server/commits/main) for recent changes and updates.

## Dependencies

- `github.com/fhs/gompd/v2` - MPD client library for Go
