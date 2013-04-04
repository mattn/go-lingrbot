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

type PlusPlus struct {
	Count int
}

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

var re = regexp.MustCompile(`(?:^|[^a-zA-Z0-9])(https?://[a-zA-Z][a-zA-Z0-9_-]*(\.[a-zA-Z0-9][a-zA-Z0-9_-]*)*(:\d+)?(?:/[a-zA-Z0-9_/.\-+%#?&=;@$,!*~]*)?)`)

func urlTitle(client *http.Client, url string) string {
	r, _ := client.Get(url)

	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") && !strings.HasPrefix(ct, "application/xhtml+xml") {
		return ""
	}
	doc, _ := html.Parse(r.Body)

	title := ""
	charset := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if charset == "" && n.Type == html.ElementNode && n.Data == "meta" {
			kv := make(map[string]string)
			for _, a := range n.Attr {
				kv[strings.ToLower(a.Key)] = strings.ToLower(a.Val)
			}
			if v, ok := kv["http-equiv"]; ok && v == "content-type" {
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
			if v, ok := kv["charset"]; ok {
				charset = v
			}
		}
		if n.Type == html.ElementNode && n.Data == "title" {
			if charset == "" {
				charset = "utf-8"
			}
			title = mahonia.NewDecoder(charset).ConvertString(n.FirstChild.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return title
}

func goDoc(client *http.Client, url string) string {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "curl/7.16.2")
	req.Header.Set("Accept", "text/plain")
	res, _ := client.Do(req)
	if res.StatusCode != 200 {
		return ""
	}
	b, _ := ioutil.ReadAll(res.Body)
	lines := strings.Split(string(b), "\n")

	var doc []string
	for n := 5; n < len(lines); n++ {
		line := lines[n]
		if len(line) > 0 && line[0] != ' ' {
			break
		}
		doc = append(doc, strings.TrimSpace(line))
	}
	return strings.TrimSpace(strings.Join(doc, "\n"))
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
					} else if len(tokens) == 2 && tokens[0] == "!godoc" {
						url := "http://godoc.org/" + tokens[1]
						results := goDoc(u, url)
						if len(results) > 0 {
							results = url + "\n" + results
							runes := []rune(results)
							if len(runes) > 1000 {
								results = string(runes[0:999])
							}
							w.Write([]byte(results))
						} else {
							w.Write([]byte("No such documents"))
						}
					} else {
						ss := re.FindAllStringSubmatch(event.Message.Text, -1)

						results := ""
						for _, s := range ss {
							if title := urlTitle(u, s[1]); title != "" {
								results += "Title: " + title + "\n"
							}
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
