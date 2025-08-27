package main

import (
	"fmt"
	"log"
	"net/http"

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

func main() {
    conn, err := mpd.Dial("tcp", "localhost:6600")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    file := "wn.mp3"
    mux := http.NewServeMux()

    mux.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
        err := conn.Clear()
        if err != nil {
            fmt.Fprintln(w, "Error clearing MPD playlist:", err.Error())
            w.WriteHeader(500)

            return
        }

        err = conn.Add(file)
        if err != nil {
            fmt.Fprintln(w, "Error adding file to MPD playlist:", err.Error())
            w.WriteHeader(500)

            return
        }

        err = conn.Play(-1)
        if err != nil {
            fmt.Fprintln(w, "Error playing MPD playlist:", err.Error())
            w.WriteHeader(500)

            return
        }

        fmt.Fprintln(w, "Playing " + file)
    })

    mux.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
        err := conn.Pause(true)
        if err != nil {
            fmt.Fprintln(w, "Error pausing MPD playback:", err.Error())
            w.WriteHeader(500)

            return
        }

        fmt.Fprintln(w, "Paused")
    })

    mux.HandleFunc("/volume", func(w http.ResponseWriter, r *http.Request) {
        conn.SetVolume(50)
        fmt.Fprintln(w, "Volume set to 50")
    })

    log.Println("API on :3000")
    http.ListenAndServe(":3000", withCORS(mux))
}