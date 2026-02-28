package handler

import (
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/cego/mysql-admin/internal/config"
	"github.com/cego/mysql-admin/internal/db"
	"github.com/cego/mysql-admin/internal/model"
)

type instanceData struct {
	Name         string
	Processes    []model.ProcessWithTransaction
	InnoDBStatus string
	SortColumn   string
	SortDir      string
	AutoRefresh  bool
	HideSleep    bool
	FilterUser   string
	FilterDB     string
}

func applyFilters(processes []model.ProcessWithTransaction, hideSleep bool, filterUser, filterDB string) []model.ProcessWithTransaction {
	out := make([]model.ProcessWithTransaction, 0, len(processes))
	for _, p := range processes {
		if hideSleep && p.Command == "Sleep" {
			continue
		}
		if filterUser != "" && p.User != filterUser {
			continue
		}
		if filterDB != "" && p.DB != filterDB {
			continue
		}
		out = append(out, p)
	}
	return out
}

// buildInstanceData parses the common sort/filter query params from r, applies
// them to processes, and returns a populated instanceData ready for rendering.
func buildInstanceData(name string, processes []model.ProcessWithTransaction, r *http.Request) instanceData {
	sortCol := r.URL.Query().Get("sort")
	sortDir := r.URL.Query().Get("dir")
	autoRefresh := r.URL.Query().Get("refresh") == "on"
	hideSleep := r.URL.Query().Get("hidesleep") == "on"
	filterUser := r.URL.Query().Get("filteruser")
	filterDB := r.URL.Query().Get("filterdb")

	model.SortProcesses(processes, sortCol, sortDir)
	processes = applyFilters(processes, hideSleep, filterUser, filterDB)

	return instanceData{
		Name:        name,
		Processes:   processes,
		SortColumn:  sortCol,
		SortDir:     sortDir,
		AutoRefresh: autoRefresh,
		HideSleep:   hideSleep,
		FilterUser:  filterUser,
		FilterDB:    filterDB,
	}
}

func Instance(cfg *config.Config, dbs map[string]*sql.DB, fullTmpl, tableTmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if _, ok := cfg.Instances[name]; !ok {
			http.NotFound(w, r)
			return
		}

		processes, innoDBStatus, err := db.GetProcessList(dbs[name])
		if err != nil {
			slog.Error("failed to get process list", "instance", name, "error", err)
			http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
			return
		}

		data := buildInstanceData(name, processes, r)
		data.InnoDBStatus = innoDBStatus

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
