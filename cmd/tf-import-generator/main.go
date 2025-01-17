package main

import (
	"bytes"
	"embed"
	"fmt"
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
			w.WriteHeader(http.StatusInternalServerError)
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

		condition := r.FormValue("condition")

		check := func(address string) bool {
			resources := strings.Split(strings.TrimSpace(r.FormValue("resources")), "\r\n")

			if len(resources) == 0 {
				return true
			}

			for _, prefix := range resources {
				if strings.HasPrefix(address, prefix) {
					return true
				}
			}
			return false
		}

		var buf bytes.Buffer

		// Only consider resources in common with all uploaded files.

		for ri, resourceAddress := range states.CommonResources() {
			// Reduce the list of resources further according to user preferences.

			if !check(resourceAddress) {
				continue
			}

			_, _ = fmt.Fprintf(&buf, "# %s\n\n", resourceAddress)

			resources := states.MapResources(resourceAddress)

			for _, refResource := range resources {

				switch {
				case refResource.Enumeration() == "native":
					_, _ = fmt.Fprintf(&buf, "locals {\n")
					_, _ = fmt.Fprintf(&buf, "  import_%d = {\n", ri+1)

					for key, resource := range resources {
						_, _ = fmt.Fprintf(&buf, `    "%s" = "%s"`+"\n", key, resource.Instances[0].Import(refResource.Type))
					}

					_, _ = fmt.Fprintf(&buf, "  }\n")
					_, _ = fmt.Fprintf(&buf, "}\n\n")

					_, _ = fmt.Fprintf(&buf, "import {\n")
					_, _ = fmt.Fprintf(&buf, "  for_each = contains(keys(local.import_%d), %s) ? [1] : []\n", ri+1, condition)
					_, _ = fmt.Fprintf(&buf, "  to       = %s\n", refResource.ID())
					_, _ = fmt.Fprintf(&buf, "  id       = local.import_%d[%s]\n", ri+1, condition)
					_, _ = fmt.Fprintf(&buf, "}\n\n")

				case refResource.Enumeration() == "count", refResource.Enumeration() == "for_each":
					index := make(Index)

					for key, resource := range resources {
						for _, instance := range resource.Instances {
							index.Add(instance.Index(), key)
						}
					}

					_, _ = fmt.Fprintf(&buf, "locals {\n")

					index.Walk(func(ii int, index string, keys []string) {
						_, _ = fmt.Fprintf(&buf, "  # %s%s\n\n", resourceAddress, index)
						_, _ = fmt.Fprintf(&buf, "  import_%d_%d = {\n", ri+1, ii)

						for _, key := range keys {
							for _, instance := range resources[key].Instances {
								if instance.Index() != index {
									continue
								}
								_, _ = fmt.Fprintf(&buf, `    "%s" = "%s"`+"\n", key, instance.Import(refResource.Type))
							}
						}

						_, _ = fmt.Fprintf(&buf, "  }\n\n")
					})

					_, _ = fmt.Fprintf(&buf, "}\n\n")

					index.Walk(func(ii int, index string, keys []string) {
						_, _ = fmt.Fprintf(&buf, "import {\n")
						_, _ = fmt.Fprintf(&buf, "  for_each = contains(keys(local.import_%d_%d), %s) ? [1] : []\n", ri+1, ii, condition)
						_, _ = fmt.Fprintf(&buf, "  to       = %s%s\n", refResource.ID(), index)
						_, _ = fmt.Fprintf(&buf, "  id       = local.import_%d_%d[%s]\n", ri+1, ii, condition)
						_, _ = fmt.Fprintf(&buf, "}\n\n")
					})
				}

				break
			}

		}

		output, err := format(&buf)

		if err != nil {
			log.Println(fmt.Errorf("terraform fmt error: %v", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_ = tmpl.ExecuteTemplate(w, "process.html", map[string]interface{}{
			"Output": output,
		})
	})

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		_ = tmpl.ExecuteTemplate(w, "start.html", map[string]interface{}{
			"Resources": []string{},
		})
	})

	_ = http.ListenAndServe("localhost:8080", http.DefaultServeMux)
}
