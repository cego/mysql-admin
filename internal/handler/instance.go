package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/cego/mysql-admin/internal/config"
	"github.com/cego/mysql-admin/internal/db"
	"github.com/cego/mysql-admin/internal/model"
)

type instanceData struct {
	Name         string
	Instances    []string
	Processes    []model.ProcessWithTransaction
	InnoDBStatus string
	SortColumn   string
	SortDir      string
}

func Instance(cfg *config.Config, fullTmpl, tableTmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		inst, ok := cfg.Instances[name]
		if !ok {
			http.NotFound(w, r)
			return
		}

		processes, innoDBStatus, err := db.GetProcessList(inst)
		if err != nil {
			slog.Error("failed to get process list", "instance", name, "error", err)
			http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
			return
		}

		sortCol := r.URL.Query().Get("sort")
		sortDir := r.URL.Query().Get("dir")
		model.SortProcesses(processes, sortCol, sortDir)

		data := instanceData{
			Name:         name,
			Instances:    cfg.InstanceNames(),
			Processes:    processes,
			InnoDBStatus: innoDBStatus,
			SortColumn:   sortCol,
			SortDir:      sortDir,
		}

		if r.Header.Get("HX-Request") == "true" {
			if err := tableTmpl.ExecuteTemplate(w, "process_table", data); err != nil {
				slog.Error("failed to render table partial", "error", err)
			}
		} else {
			if err := fullTmpl.ExecuteTemplate(w, "layout", data); err != nil {
				slog.Error("failed to render instance page", "error", err)
			}
		}
	}
}
