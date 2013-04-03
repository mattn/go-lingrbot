package lingrbot

import (
	"appengine"
	"appengine/urlfetch"
	"code.google.com/p/go.net/html"
	"code.google.com/p/mahonia"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
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
				u := urlfetch.Client(c)
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
					} else if strings.HasPrefix(r.Header.Get("Content-Type"), "text/html") {
						re := regexp.MustCompile(`(?:^|[^a-zA-Z0-9])(https?://[a-zA-Z][a-zA-Z0-9_-]*(\.[a-zA-Z0-9][a-zA-Z0-9_-]*)*(:\d+)?(?:/[a-zA-Z0-9_/.\-+%#?&=;@$,!*~]*)?)`)
						ss := re.FindAllStringSubmatch(event.Message.Text, -1)

						results := ""
						for _, s := range ss {
							url := s[1]
							r, _ := u.Get(url)
							doc, _ := html.Parse(r.Body)

							charset := ""
							var f func(*html.Node)
							f = func(n *html.Node) {
								if charset == "" && n.Type == html.ElementNode && n.Data == "meta" {
									kv := make(map[string]string)
									for _, a := range n.Attr {
										kv[strings.ToLower(a.Key)] = strings.ToLower(a.Val)
									}
									if v, ok := kv["http-equiv"]; ok {
										if v == "content-type" {
											if v, ok = kv["content"]; ok {
												for _, t := range strings.Split(v, ";") {
													tt := strings.Split(strings.TrimSpace(t), "=")
													if len(tt) == 2 && strings.ToLower(tt[0]) == "charset" {
														charset = tt[1]
														break
													}
												}
											}
										}
									}
									if v, ok := kv["charset"]; ok {
											charset = v
									}
								}
								if n.Type == html.ElementNode && n.Data == "title" {
									if charset == "" {
										charset = "utf-8"
									}
									title := mahonia.NewDecoder(charset).ConvertString(n.FirstChild.Data)
									results += "Title: " + title + "\n"
									return
								}
								for c := n.FirstChild; c != nil; c = c.NextSibling {
									f(c)
								}
							}
							f(doc)
						}
						if len(results) > 0 {
							w.Write([]byte(results))
						}
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
