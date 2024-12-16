package main

import (
	"bytes"
	"embed"
	"html/template"
	"log"
	"net/http"
	"strings"
)

//go:embed templates
var templates embed.FS

func main() {
	tmpl, err := template.ParseFS(templates, "templates/*.html")

	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	http.HandleFunc("POST /process", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Println(err)
			return
		}

		resources := strings.Split(strings.TrimSpace(r.FormValue("resources")), "\r\n")

		if len(resources) == 0 {
			return
		}

		var buf bytes.Buffer

		for _, upload := range r.MultipartForm.File["statefiles"] {
			f, err := upload.Open()

			if err != nil {
				log.Println(err)
				continue
			}

			state := readState(f)

			_ = f.Close()

			if state.Version != 4 {
				continue
			}

			for _, resource := range state.Resources {
				if resource.Mode != "managed" {
					continue
				}

				for _, prefix := range resources {
					if strings.HasPrefix(resource.ID(), prefix) {
						importBlock(&buf, resource, r.FormValue("condition"))
					}
				}
			}

			for _, resource := range state.Resources {
				if resource.Mode != "managed" {
					continue
				}

				for _, prefix := range resources {
					if strings.HasPrefix(resource.ID(), prefix) {
						removedBlock(&buf, resource)
					}
				}
			}
		}

		tmpl.ExecuteTemplate(w, "process.html", map[string]interface{}{
			"Output": buf.String(),
		})
	})

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "start.html", map[string]interface{}{
			"Resources": []string{},
		})
	})

	http.ListenAndServe("localhost:8080", http.DefaultServeMux)
}
