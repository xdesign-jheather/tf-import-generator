package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		states, err := uploadedStateFiles(r, "statefiles")

		if err != nil {
			log.Println(err)
		}

		if err != nil || len(states) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		f1, err := os.CreateTemp("", "tf-import-generator-*.tf")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, state := range states {
			if state.Version != 4 {
				continue
			}

			for _, resource := range state.Resources {
				if resource.Mode != "managed" {
					continue
				}

				for _, prefix := range resources {
					if strings.HasPrefix(resource.ID(), prefix) {
						importBlock(f1, resource, r.FormValue("condition"))
					}
				}
			}

			for _, resource := range state.Resources {
				if resource.Mode != "managed" {
					continue
				}

				for _, prefix := range resources {
					if strings.HasPrefix(resource.ID(), prefix) {
						removedBlock(f1, resource)
					}
				}
			}
		}

		if err = f1.Close(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cmd := exec.Command("terraform", "fmt", f1.Name())

		if err = cmd.Run(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		f2, err := os.ReadFile(f1.Name())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_ = os.Remove(f1.Name())

		tmpl.ExecuteTemplate(w, "process.html", map[string]interface{}{
			"Output": string(f2),
		})
	})

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "start.html", map[string]interface{}{
			"Resources": []string{},
		})
	})

	http.ListenAndServe("localhost:8080", http.DefaultServeMux)
}
