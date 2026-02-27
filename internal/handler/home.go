package handler

import (
	"html/template"
	"net/http"

	"github.com/cego/mysql-admin/internal/config"
)

func Home(cfg *config.Config, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Instances []string
		}{
			Instances: cfg.InstanceNames(),
		}

		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
