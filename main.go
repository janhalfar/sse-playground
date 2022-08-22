package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "embed"
)

//go:embed index.html
var index []byte

func writeEvent(w http.ResponseWriter, flusher http.Flusher, name, id string, data interface{}) error {
	jsonTimeBytes, err := json.Marshal(time.Now())
	if err != nil {
		http.Error(w, "how could that happen", http.StatusInternalServerError)
	}
	msg := "event: time\n" +
		"id: " + id + "\n" +
		"data: " + string(jsonTimeBytes) + "\n\n"
	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func main() {
	http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	}))

	getFlusher := func(w http.ResponseWriter) (flusher http.Flusher, ok bool) {
		w.Header().Set("Content-Type", "text/event-stream")
		newFlusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "could not create flusher, I hate you", 999)
			return nil, false
		}
		return newFlusher, true
	}

	http.HandleFunc("/sse/time", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := getFlusher(w)
		if !ok {
			return
		}
		i := 0
		for i < 100 {
			i++
			err := writeEvent(w, flusher, "time", fmt.Sprint(i), time.Now())
			if err != nil {
				fmt.Println("connection pbly closed", err)
				return
			}
			time.Sleep(time.Millisecond * 15)
		}

	}))

	fmt.Println("oops", http.ListenAndServe(":8080", nil))
}
