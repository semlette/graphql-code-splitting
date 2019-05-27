package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/semlette/graphql-code-splitting/interpreter/lexer"
	"github.com/semlette/graphql-code-splitting/parser"
	"github.com/semlette/graphql-code-splitting/parser/ast"
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
	TypeName  string    `json:"__typename" graphql:"__typename"`
	Timestamp time.Time `json:"timestamp" graphql:"timestamp"`
	Text      string    `json:"text" graphql:"text"`
}

func (textPost) post() {}

type photoPost struct {
	TypeName  string    `json:"__typename" graphql:"__typename"`
	Timestamp time.Time `json:"timestamp" graphql:"__timestamp"`
	PhotoURL  string    `json:"photoURL" graphql:"photoURL"`
}

func (photoPost) post() {}

type importMap struct {
	targetObject string
	importPath   string
}

func findImports(ss *ast.SelectionSet) []*importMap {
	imports := []*importMap{}
	if ss == nil {
		return imports
	}
	for _, field := range ss.Fields {
		imports = append(imports, findImports(field.SelectionSet)...)
	}
	for _, fs := range ss.FragmentSpreads {
		if fs.Directive != nil {
			if fs.Directive.Name.Value == "push" && fs.Directive.Arguments != nil {
				for _, arg := range fs.Directive.Arguments {
					if arg.Name.Value == "module" {
						imports = append(imports, &importMap{
							targetObject: fs.FragmentName.Value,
							importPath:   arg.Value,
						})
						break
					}
				}
			}
		}
	}
	return imports
}

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
		}

		var posts []post
		posts = append(posts, textPost{
			TypeName:  "TextPost",
			Timestamp: time.Now(),
			Text:      "Hello world!",
		})
		posts = append(posts, photoPost{
			TypeName:  "PhotoPost",
			Timestamp: time.Now(),
			PhotoURL:  "/hackerman.png",
		})

		pusher, ok := w.(http.Pusher)
		if ok {
			importMaps := findImports(doc.Operation.SelectionSet)
			if len(importMaps) > 0 {
				presentTypes := make(map[string]bool)
				for _, post := range posts {
					tp, isTextPost := post.(textPost)
					pp, isPhotoPost := post.(photoPost)
					switch {
					case isTextPost:
						presentTypes[tp.TypeName] = true
					case isPhotoPost:
						presentTypes[pp.TypeName] = true
					}
				}
				pushed := make(map[string]bool)
				for _, importMap := range importMaps {
					if !pushed[importMap.importPath] {
						if err := pusher.Push(filepath.Join("/", importMap.importPath), nil); err != nil {
							log.Printf("failed push: %s", err)
						}
						pushed[importMap.importPath] = true
					}
				}
			} else {
				log.Printf("no imports")
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
