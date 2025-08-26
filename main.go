package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fhs/gompd/v2/mpd"
)

func main() {
    conn, err := mpd.Dial("tcp", "localhost:6600")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
        conn.Clear()
        conn.Add("wn.mp3")
        conn.Play(-1)

        fmt.Fprintln(w, "Playing noise-white.mp3")
    })

    http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
        conn.Pause(true)
        fmt.Fprintln(w, "Paused")
    })

    http.HandleFunc("/volume", func(w http.ResponseWriter, r *http.Request) {
        conn.SetVolume(50)
        fmt.Fprintln(w, "Volume set to 50")
    })

    log.Println("API on :3000")
    http.ListenAndServe(":3000", nil)
}