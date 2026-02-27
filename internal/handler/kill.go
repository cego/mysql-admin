package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/cego/go-lib/v2/renderer"
	"github.com/cego/mysql-admin/internal/config"
	"github.com/cego/mysql-admin/internal/db"
	"github.com/cego/mysql-admin/internal/model"
)

func Kill(cfg *config.Config, rend *renderer.Renderer, tableTmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		inst, ok := cfg.Instances[name]
		if !ok {
			rend.JSON(w, http.StatusNotFound, map[string]string{"error": "instance not found"})
			return
		}

		idStr := r.FormValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			rend.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid process id"})
			return
		}

		if err := db.KillProcess(inst, id); err != nil {
			slog.Error("failed to kill process", "instance", name, "id", id, "error", err)
			rend.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		if cfg.UserHeader != "" {
			user := r.Header.Get(cfg.UserHeader)
			slog.Info("killed process", "user", user, "process_id", id, "instance", name)
		}

		processes, _, err := db.GetProcessList(inst)
		if err != nil {
			slog.Error("failed to refresh process list", "instance", name, "error", err)
			rend.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		sortCol := r.URL.Query().Get("sort")
		sortDir := r.URL.Query().Get("dir")
		model.SortProcesses(processes, sortCol, sortDir)

		data := instanceData{
			Name:       name,
			Processes:  processes,
			SortColumn: sortCol,
			SortDir:    sortDir,
		}

		if err := tableTmpl.ExecuteTemplate(w, "process_table", data); err != nil {
			slog.Error("failed to render table partial", "error", err)
		}
	}
}
