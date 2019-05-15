package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/semlette/graphql-code-splitting/interpreter/lexer"
	"github.com/semlette/graphql-code-splitting/parser"
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

	type query struct {
		Query string `json:"query"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var q query
		err := json.NewDecoder(r.Body).Decode(&q)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Printf("decode error: %s", err)
			return
		}
		l := lexer.New(q.Query)
		p := parser.New(l)
		doc := p.Parse()
		if err := p.Error(); err != nil {
			log.Printf("parser error: %s", err)
		} else {
			for _, ss := range doc.Query.SelectionSets {
				log.Printf("selection set: fields: %d, sub selection sets: %d", len(ss.Fields), len(ss.SelectionSets))
			}
			for _, fragment := range doc.Query.Fragments {
				log.Printf(
					"fragment %s on %s, fields: %d",
					fragment.Name.Value,
					fragment.TargetObject.Value,
					len(fragment.SelectionSet.Fields),
				)
				for _, field := range fragment.SelectionSet.Fields {
					log.Printf("- %s", field.Name.Value)
				}
			}
		}

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
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Printf("encode error: %s", err)
		}
	}
}
