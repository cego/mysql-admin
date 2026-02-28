package handler

import (
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/cego/mysql-admin/internal/config"
	"github.com/cego/mysql-admin/internal/db"
	"github.com/cego/mysql-admin/internal/model"
)

// htmlError writes an HTML error snippet suitable for HTMX to swap into the
// process-table area, so errors are displayed as styled text rather than raw JSON.
func htmlError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, `<div class="px-4 py-10 text-center text-sm text-red-500">%s</div>`, html.EscapeString(msg))
}

func Kill(cfg *config.Config, tableTmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		inst, ok := cfg.Instances[name]
		if !ok {
			htmlError(w, http.StatusNotFound, "instance not found")
			return
		}

		idStr := r.FormValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			htmlError(w, http.StatusBadRequest, "invalid process id")
			return
		}

		if err := db.KillProcess(inst, id); err != nil {
			slog.Error("failed to kill process", "instance", name, "id", id, "error", err)
			htmlError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if cfg.UserHeader != "" {
			user := r.Header.Get(cfg.UserHeader)
			slog.Info("killed process", "user", user, "process_id", id, "instance", name)
		}

		processes, _, err := db.GetProcessList(inst)
		if err != nil {
			slog.Error("failed to refresh process list", "instance", name, "error", err)
			htmlError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sortCol := r.URL.Query().Get("sort")
		sortDir := r.URL.Query().Get("dir")
		autoRefresh := r.URL.Query().Get("refresh") == "on"
		hideSleep := r.URL.Query().Get("hidesleep") == "on"
		filterUser := r.URL.Query().Get("filteruser")
		filterDB := r.URL.Query().Get("filterdb")

		model.SortProcesses(processes, sortCol, sortDir)
		processes = applyFilters(processes, hideSleep, filterUser, filterDB)

		data := instanceData{
			Name:        name,
			Processes:   processes,
			SortColumn:  sortCol,
			SortDir:     sortDir,
			AutoRefresh: autoRefresh,
			HideSleep:   hideSleep,
			FilterUser:  filterUser,
			FilterDB:    filterDB,
		}

		if err := tableTmpl.ExecuteTemplate(w, "process_table", data); err != nil {
			slog.Error("failed to render table partial", "error", err)
		}
	}
}
