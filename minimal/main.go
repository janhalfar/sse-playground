package main

import (
	"fmt"
	"net/http"
	"time"

	_ "embed"
)

//go:embed index.html
var index []byte

func main() {

	http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	}))

	http.HandleFunc("/sse/time", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		flusher, _ := w.(http.Flusher)

		i := 0
		for i < 100 {
			i++
			eventID := fmt.Sprint(i)
			_, err := w.Write([]byte(
				"event: time\n" +
					"id: " + eventID + "\n" +
					"data: " + time.Now().Format(time.RFC3339Nano) + "\n\n",
			))

			if err != nil {
				fmt.Println("connection pbly closed", i, err)
				return
			}
			flusher.Flush()
			time.Sleep(time.Millisecond * 15)
			fmt.Println("sent", i, "to", r.RemoteAddr)
		}
	}))

	fmt.Println("shit went sideways", http.ListenAndServe(":8080", nil))
}
