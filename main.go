package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/cego/go-lib/v2/logger"
	"github.com/cego/go-lib/v2/renderer"
	"github.com/cego/go-lib/v2/serve"
	_ "github.com/go-sql-driver/mysql"

	"github.com/cego/mysql-admin/internal/config"
	"github.com/cego/mysql-admin/internal/handler"
)

//go:embed templates
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

var funcMap = template.FuncMap{
	"stringToColor": stringToColor,
	"blackOrWhite":  blackOrWhite,
	"formatNumber":  formatNumber,
	"join":          strings.Join,
	"nextDir":       nextDir,
	"sortIndicator": sortIndicator,
	"baseParams":    baseParams,
}

func main() {
	cfg := config.Load()

	l := logger.New()
	slog.SetDefault(l)

	r := renderer.New(l)

	homeTmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/layout.html", "templates/home.html"))
	instanceTmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/layout.html", "templates/instance.html", "templates/partials/process_table.html"))
	tableTmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/partials/process_table.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handler.Home(cfg, homeTmpl))
	mux.HandleFunc("GET /instance/{name}", handler.Instance(cfg, instanceTmpl, tableTmpl))
	mux.HandleFunc("POST /instance/{name}/kill", handler.Kill(cfg, r, tableTmpl))

	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticSub)))

	srv := serve.WithDefaults(&http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	})

	slog.Info("starting server", "port", cfg.Port, "instances", cfg.InstanceNames())

	if err := serve.ListenAndServe(context.Background(), srv, l); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func stringToColor(str string) string {
	if str == "" {
		return "#000000"
	}
	var hash int32
	for _, c := range str {
		hash = c + ((hash << 5) - hash)
	}
	colour := "#"
	for i := 0; i < 3; i++ {
		value := (hash >> (i * 8)) & 0xff
		colour += fmt.Sprintf("%02x", value)
	}
	return colour
}

func blackOrWhite(hex string) string {
	if len(hex) < 7 {
		return "#ffffff"
	}
	r, _ := strconv.ParseInt(hex[1:3], 16, 64)
	g, _ := strconv.ParseInt(hex[3:5], 16, 64)
	b, _ := strconv.ParseInt(hex[5:7], 16, 64)
	brightness := (r*299 + g*587 + b*114) / 1000
	if brightness > 155 {
		return "#000000"
	}
	return "#ffffff"
}

func formatNumber(num int64) string {
	s := strconv.FormatInt(num, 10)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ' ')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

// baseParams returns the persistent query-string tail (starting with &) for all
// params that should survive sort/navigation changes: refresh, hidesleep,
// filteruser, filterdb. Append directly to ?sort=X&dir=Y in hx-get URLs.
func baseParams(ar, hs bool, fu, fd string) string {
	var b strings.Builder
	if ar {
		b.WriteString("&refresh=on")
	}
	if hs {
		b.WriteString("&hidesleep=on")
	}
	if fu != "" {
		b.WriteString("&filteruser=" + url.QueryEscape(fu))
	}
	if fd != "" {
		b.WriteString("&filterdb=" + url.QueryEscape(fd))
	}
	return b.String()
}

func nextDir(currentCol, currentDir, col string) string {
	if currentCol == col {
		if currentDir == "asc" {
			return "desc"
		}
		return "asc"
	}
	return "asc"
}

func sortIndicator(currentCol, currentDir, col string) string {
	if currentCol == col {
		if currentDir == "asc" {
			return " ▲"
		}
		return " ▼"
	}
	return ""
}
