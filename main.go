package main

import (
	"encoding/json"
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
	Data    any    `json:"data"`
}

func okResponse(w http.ResponseWriter, message string, data any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)

	if data != nil {
		encoder.Encode(SuccessResponse{
			OK:      true,
			Message: message,
			Data:    data,
		})
	} else {
		encoder.Encode(SuccessResponse{
			OK:      true,
			Message: message,
		})
	}
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

type SongAtts struct {
	Id           string `json:"id"`
	File         string `json:"file"`
	LastModified string `json:"lastModified"`
	Title        string `json:"title"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		conn, err := getMPDConnection()
		if err != nil {
			errorResponse(w, "Error connecting to MPD: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		switch r.Method {
		case "GET":
			atts, err := conn.CurrentSong()
			if err != nil {
				errorResponse(w, "Error getting current song from MPD: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if len(atts) == 0 {
				errorResponse(w, "Attributes for current song are empty.", http.StatusInternalServerError)
				return
			}

			okResponse(w, "Got attributes for current song.", SongAtts{
				File:         atts["file"],
				LastModified: atts["Last-Modified"],
				Id:           atts["Id"],
				Title:        atts["Title"],
			})
		case "POST":
			file := r.URL.Query().Get("file")

			// If user didn't pass a file, just play
			if file == "" {
				err = conn.Play(-1)
				if err != nil {
					errorResponse(w, "Error playing MPD playlist: "+err.Error(), http.StatusInternalServerError)
					return
				}

				okResponse(w, "Playing.", nil)
				return
			}

			// If user did pass a file, clear the queue, add it, and play
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

			okResponse(w, "Playing "+file, nil)
		default:
			errorResponse(w, "Method not allowed.", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
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

			okResponse(w, "Paused.", nil)
		default:
			errorResponse(w, "Method not allowed.", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/volume", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
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

			okResponse(w, "Volume set to "+volumeStr, nil)
		default:
			errorResponse(w, "Method not allowed.", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/database", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			conn, err := getMPDConnection()
			if err != nil {
				errorResponse(w, "Error connecting to MPD: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer conn.Close()

			files, err := conn.GetFiles()
			if err != nil {
				errorResponse(w, "Error reading MPD database: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if len(files) == 0 {
				errorResponse(w, "No files were found in the MPD database.", http.StatusInternalServerError)
				return
			}

			okResponse(w, "Found database files.", files)
		default:
			errorResponse(w, "Method not allowed.", http.StatusMethodNotAllowed)
		}
	})

	log.Println("API on :3000")
	http.ListenAndServe(":3000", withCORS(mux))
}
