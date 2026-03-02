package handler

import (
	"database/sql"
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/cego/mysql-admin/internal/config"
	"github.com/cego/mysql-admin/internal/db"
)

// htmlError writes an HTML error snippet suitable for HTMX to swap into the
// process-table area, so errors are displayed as styled text rather than raw JSON.
func htmlError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, _ = fmt.Fprintf(w, `<div class="px-4 py-10 text-center text-sm text-red-500 dark:text-red-400">%s</div>`, html.EscapeString(msg))
}

func Kill(cfg *config.Config, dbs map[string]*sql.DB, tableTmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Reject requests that did not originate from HTMX. Browsers will not send
		// custom headers (HX-Request) on cross-origin form submissions, so this is
		// a lightweight CSRF mitigation that costs nothing for legitimate clients.
		if r.Header.Get("HX-Request") != "true" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		name := r.PathValue("name")
		if _, ok := cfg.Instances[name]; !ok {
			htmlError(w, http.StatusNotFound, "instance not found")
			return
		}

		idStr := r.FormValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			htmlError(w, http.StatusBadRequest, "invalid process id")
			return
		}

		if err := db.KillProcess(dbs[name], id); err != nil {
			slog.Error("failed to kill process", "instance", name, "id", id, "error", err)
			htmlError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if cfg.UserHeader != "" {
			user := r.Header.Get(cfg.UserHeader)
			slog.Info("killed process", "user", user, "process_id", id, "instance", name)
		}

		processes, _, err := db.GetProcessList(dbs[name])
		if err != nil {
			slog.Error("failed to refresh process list", "instance", name, "error", err)
			htmlError(w, http.StatusInternalServerError, err.Error())
			return
		}

		data := buildInstanceData(name, processes, r)

		if err := tableTmpl.ExecuteTemplate(w, "process_table", data); err != nil {
			slog.Error("failed to render table partial", "error", err)
		}
	}
}
