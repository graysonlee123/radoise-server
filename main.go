package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/fhs/gompd/v2/mpd"
)

func withCORS(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        // Handle preflight
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)

            return
        }

        h.ServeHTTP(w, r)
    })
}

// Helper function to get a fresh MPD connection
func getMPDConnection() (*mpd.Client, error) {
    return mpd.Dial("tcp", "localhost:6600")
}

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
        conn, err := getMPDConnection()
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error connecting to MPD:", err.Error())

            return
        }
        defer conn.Close()

        file := r.URL.Query().Get("file")
        if file == "" {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintln(w, "Missing 'file' query parameter")
            return
        }

        files, err := conn.GetFiles()
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error reading MPD database:", err.Error())

            return
        }

        if !slices.Contains(files, file) {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "That file was not found in the MPD database.")

            return
        }

        err = conn.Clear()
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error clearing MPD playlist:", err.Error())

            return
        }

        err = conn.Add(file)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error adding file to MPD playlist:", err.Error())

            return
        }

        err = conn.Play(-1)
        if err != nil {
            fmt.Fprintln(w, "Error playing MPD playlist:", err.Error())
            w.WriteHeader(http.StatusInternalServerError)

            return
        }

        fmt.Fprintln(w, "Playing " + file)
    })

    mux.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
        conn, err := getMPDConnection()
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error connecting to MPD:", err.Error())

            return
        }
        defer conn.Close()

        err = conn.Pause(true)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error pausing MPD playback:", err.Error())

            return
        }

        fmt.Fprintln(w, "Paused")
    })

    mux.HandleFunc("/database", func(w http.ResponseWriter, r *http.Request) {
        conn, err := getMPDConnection()
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error connecting to MPD:", err.Error())

            return
        }
        defer conn.Close()

        files, err := conn.GetFiles()
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error reading MPD database:", err.Error())

            return
        }

        if (len(files) == 0) {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "No files were found in the MPD database.")

            return
        }

        encoder := json.NewEncoder(w)
        err = encoder.Encode(files)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Could not encode a JSON response.")

            return
        }

        fmt.Fprintln(w)
    })

    mux.HandleFunc("/volume", func(w http.ResponseWriter, r *http.Request) {
        conn, err := getMPDConnection()
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error connecting to MPD:", err.Error())

            return
        }
        defer conn.Close()

        volumeStr := r.URL.Query().Get("level")
        if volumeStr == "" {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintln(w, "Missing 'level' query parameter")
            return
        }

        volume, err := strconv.Atoi(volumeStr)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintln(w, "Invalid volume level: must be an integer")
            return
        }

        minVolume := 0
        maxVolume := 100
        if volume < minVolume || volume > maxVolume {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintln(w, "Invalid volume level: out of bounds")
            return
        }

        err = conn.SetVolume(volume)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintln(w, "Error setting MPD volume:", err.Error())
            return
        }

        fmt.Fprintf(w, "Volume set to %d", volume)
    })

    log.Println("API on :3000")
    http.ListenAndServe(":3000", withCORS(mux))
}