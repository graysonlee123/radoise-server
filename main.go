package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

type SuccessResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func okResponse(w http.ResponseWriter, message string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	encoder.Encode(SuccessResponse{
		OK:      true,
		Message: message,
	})
}

type ErrorResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func errorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	encoder := json.NewEncoder(w)
	encoder.Encode(ErrorResponse{
		OK:      false,
		Message: message,
	})
}

func main() {
	file := "wn.mp3"
	mux := http.NewServeMux()

	mux.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		conn, err := getMPDConnection()
		if err != nil {
			errorResponse(w, "Error connecting to MPD: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		err = conn.Clear()
		if err != nil {
			errorResponse(w, "Error clearing MPD playlist: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = conn.Add(file)
		if err != nil {
			errorResponse(w, "Error adding file to MPD playlist: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = conn.Play(-1)
		if err != nil {
			errorResponse(w, "Error playing MPD playlist: "+err.Error(), http.StatusInternalServerError)
			return
		}

		okResponse(w, "Playing "+file)
	})

	mux.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		conn, err := getMPDConnection()
		if err != nil {
			errorResponse(w, "Error connecting to MPD: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		err = conn.Pause(true)
		if err != nil {
			errorResponse(w, "Error pausing MPD playback: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Paused")
	})

	mux.HandleFunc("/volume", func(w http.ResponseWriter, r *http.Request) {
		conn, err := getMPDConnection()
		if err != nil {
			errorResponse(w, "Error connecting to MPD: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		volumeStr := r.URL.Query().Get("level")
		if volumeStr == "" {
			errorResponse(w, "Missing 'level' query parameter.", http.StatusBadRequest)
			return
		}

		volume, err := strconv.Atoi(volumeStr)
		if err != nil {
			errorResponse(w, "Invalid volume level: must be an integer.", http.StatusBadRequest)
			return
		}

		minVolume := 0
		maxVolume := 100
		if volume < minVolume || volume > maxVolume {
			errorResponse(w, "Invalid volume level: out of bounds", http.StatusBadRequest)
			return
		}

		err = conn.SetVolume(volume)
		if err != nil {
			errorResponse(w, "Error setting MPD volume: "+err.Error(), http.StatusInternalServerError)
			return
		}

		okResponse(w, "Volume set to "+volumeStr)
	})

	log.Println("API on :3000")
	http.ListenAndServe(":3000", withCORS(mux))
}
