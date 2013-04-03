package lingrbot

import (
	"appengine"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type Status struct {
	Events []Event `json:"events"`
}

type Event struct {
	Id      int      `json:"event_id"`
	Message *Message `json:"message"`
}

type Message struct {
	Id              string `json:"id"`
	Room            string `json:"room"`
	PublicSessionId string `json:"public_session_id"`
	IconUrl         string `json:"icon_url"`
	Type            string `json:"type"`
	SpeakerId       string `json:"speaker_id"`
	Nickname        string `json:"nickname"`
	Text            string `json:"text"`
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			if r.Method == "POST" {
				var status Status

				c := appengine.NewContext(r)
				e := json.NewDecoder(r.Body).Decode(&status)
				if e != nil {
					c.Errorf("%s", e.Error())
					return
				}
				w.Header().Set("Content-Type", "text/plain; charset=utf8")
				for _, event := range status.Events {
					tokens := strings.SplitN(event.Message.Text, " ", 2)
					if tokens[0] == "!go" {
						w.Write([]byte("日本goでok"))
					}
				}
			} else {
				w.Header().Set("Content-Type", "text/html; charset=utf8")
				b, _ := ioutil.ReadFile("index.html")
				w.Write(b)
			}
		} else {
			http.NotFound(w, r)
		}
	})
}
