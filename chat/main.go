package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	_ "embed"
)

//go:embed index.html
var index []byte

type Post struct {
	User string `json:"user"`
	Msg  string `json:"msg"`
}

func sendEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(
		"event: " + event + "\n" +
			"data: " + string(dataBytes) + "\n\n",
	))
	if err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

type chansEvent struct {
	chanPost    chan Post
	chanConnect chan string
}

func broker(
	chanConnect chan string,
	chanPost chan Post,
	chanChans chan chansEvent,

) {
	users := map[string]chansEvent{}
	for {
		select {
		case newUser := <-chanConnect:
			users[newUser] = chansEvent{
				chanPost:    make(chan Post),
				chanConnect: make(chan string),
			}
			chanChans <- users[newUser]
			for _, chansUser := range users {
				chansUser.chanConnect <- newUser
			}
		case post := <-chanPost:
			for _, chansUser := range users {
				chansUser.chanPost <- post
			}
		}
	}

}

func main() {

	http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	}))

	chanConnect := make(chan string)
	chanPost := make(chan Post)
	chanChans := make(chan chansEvent)

	go broker(chanConnect, chanPost, chanChans)

	http.HandleFunc("/post", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		chanPost <- Post{
			User: r.URL.Query().Get("user"),
			Msg:  string(bodyBytes),
		}
	}))

	http.HandleFunc("/sse/connect", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		flusher, _ := w.(http.Flusher)

		chanConnect <- r.URL.Query().Get("user")
		myChans := <-chanChans

		for {
			select {
			case post := <-myChans.chanPost:
				sendEvent(w, flusher, "post", post)
			case user := <-myChans.chanConnect:
				sendEvent(w, flusher, "user", user)
			}
		}
	}))
	fmt.Println("shit went sideways", http.ListenAndServe(":8080", nil))
}
