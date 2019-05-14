package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.HandleFunc("/graphql", graphqlHandler())
	go log.Printf("listening on :8080")
	err := http.ListenAndServeTLS(":8080", "server.crt", "server.key", mux)
	if err != nil {
		log.Fatalf("serve error: %s", err)
	}
}

type post interface {
	post()
}

type textPost struct {
	TypeName  string    `json:"__typename"`
	Timestamp time.Time `json:"timestamp"`
	Text      string    `json:"text"`
}

func (textPost) post() {}

type photoPost struct {
	TypeName  string    `json:"__typename"`
	Timestamp time.Time `json:"timestamp"`
	PhotoURL  string    `json:"photoURL"`
}

func (photoPost) post() {}

func graphqlHandler() http.HandlerFunc {
	type response struct {
		Data map[string]interface{} `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		imports := []string{}

		var posts []post
		posts = append(posts, textPost{
			TypeName:  "TextPost",
			Timestamp: time.Now(),
			Text:      "Hello world!",
		})
		imports = append(imports, "TextPost.js")
		posts = append(posts, photoPost{
			TypeName:  "PhotoPost",
			Timestamp: time.Now(),
			PhotoURL:  "/hackerman.png",
		})
		imports = append(imports, "PhotoPost.js")

		pusher, ok := w.(http.Pusher)
		if ok {
			for _, pushImport := range imports {
				if err := pusher.Push(filepath.Join("/", pushImport), nil); err != nil {
					log.Printf("failed push: %s", err)
				}
			}
		}

		resp := response{
			Data: map[string]interface{}{
				"posts": posts,
			},
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Printf("encode error: %s", err)
		}
	}
}
