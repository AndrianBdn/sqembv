package sqembv

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed LICENSE
var license string

//go:embed static
var embStatic embed.FS

// API ...
type API struct {
	dbClient *sqlClient
	dbFile   string
}

// NewAPI initializes the API controller with a DB file.
func NewAPI(dbFile string) (*API, error) {
	if !fileExists(dbFile) {
		return nil, errors.New("db file does not exist")
	}

	client, err := newClient(dbFile)
	if err != nil {
		return nil, err
	}
	return &API{client, dbFile}, nil
}

func fileExists(dbFile string) bool {
	info, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// NewAPIFromDB initializes the API controller with a DB.
func NewAPIFromDB(db *sql.DB) (*API, error) {
	client, err := newClientFromDB(db)
	if err != nil {
		return nil, err
	}
	return &API{client, ""}, nil
}

func License() string {
	return license
}

// Handler ...
func (a *API) Handler(browserRoot string) http.Handler {
	if !strings.HasSuffix(browserRoot, "/") {
		browserRoot += "/"
	}

	indexPage, err := embStatic.ReadFile("static/index.html")
	if err != nil {
		panic("can not read index.html: " + err.Error())
	}
	indexTmpl, _ := template.New("name").Parse(string(indexPage))

	staticHandler := http.StripPrefix(browserRoot, http.FileServer(http.FS(embStatic)))

	staticRoot := browserRoot + "static/"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case browserRoot + "api/info":
			a.Info(w, r)
		case browserRoot + "api/tables":
			a.Tables(w, r)
		case browserRoot + "api/table":
			a.Table(w, r)
		case browserRoot + "api/table/info":
			a.TableInfo(w, r)
		case browserRoot + "api/table/sql":
			a.TableSQL(w, r)
		case browserRoot + "api/table/indexes":
			a.TableIndexes(w, r)
		case browserRoot + "api/query":
			a.Query(w, r)
		case browserRoot:
			indexTmpl.Execute(w, map[string]string{"root": browserRoot, "static": staticRoot})
		default:
			staticHandler.ServeHTTP(w, r)
		}
	})
}

// Info ...
func (a *API) Info(w http.ResponseWriter, req *http.Request) {
	info, err := a.dbClient.Info()
	if err != nil {
		renderError(w, http.StatusInternalServerError, err)
	}

	filePath, err := filepath.Abs(a.dbFile)
	if err != nil {
		filePath = ""
	}

	dbName := filepath.Base(a.dbFile)
	size, _ := fileSize(filePath)

	result := map[string]interface{}{
		"number_of_tables":  info.Rows[0][0],
		"number_of_indexes": info.Rows[0][1],
		"filename":          dbName,
		"fullname":          filePath,
		"size":              size,
	}
	renderJSON(w, http.StatusOK, result)
}

// Tables ...
func (a *API) Tables(w http.ResponseWriter, req *http.Request) {
	tables, err := a.dbClient.Tables()
	if err != nil {
		renderError(w, http.StatusInternalServerError, err)
	}

	result := map[string]interface{}{
		"tables": tables,
	}
	renderJSON(w, http.StatusOK, result)
}

// Table ...
func (a *API) Table(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("table")
	result, err := a.dbClient.Table(name)
	if err != nil {
		renderError(w, http.StatusInternalServerError, err)
	}

	renderJSON(w, http.StatusOK, result.Format())
}

// TableInfo ...
func (a *API) TableInfo(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("table")
	result, err := a.dbClient.TableInfo(name)
	if err != nil {
		renderError(w, http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{
		"row_count":     result.Rows[0][0],
		"indexes_count": 0,
	}

	renderJSON(w, http.StatusOK, data)
}

// TableSQL ...
func (a *API) TableSQL(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("table")
	result, err := a.dbClient.TableSQL(name)
	if err != nil {
		renderError(w, http.StatusInternalServerError, err)
	}

	data := map[string]interface{}{
		"sql": result[0],
	}

	renderJSON(w, http.StatusOK, data)
}

// TableIndexes ...
func (a *API) TableIndexes(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("table")
	result, err := a.dbClient.TableIndexes(name)
	if err != nil {
		renderError(w, http.StatusInternalServerError, err)
	}

	renderJSON(w, http.StatusOK, result.Format())
}

// Query ...
func (a *API) Query(w http.ResponseWriter, req *http.Request) {
	query := strings.TrimSpace(req.FormValue("query"))

	if query == "" {
		renderError(w, http.StatusBadRequest, errors.New("Query missing"))
		return
	}

	result, err := a.dbClient.QuerySQL(req.FormValue("query"))
	if err != nil {
		renderError(w, http.StatusInternalServerError, err)
		return
	}

	q := req.URL.Query()
	if len(q["format"]) > 0 {
		if q["format"][0] == "csv" {
			renderCSV(w, http.StatusOK, result.CSV())
			return
		} else if q["format"][0] == "json" {
			// Format the returned JSON instead of returning in the Result format
			renderJSON(w, http.StatusOK, result.Format())
			return
		}
	}

	renderJSON(w, http.StatusOK, result)
}

// renderError renders a JSON response with the given error message.
func renderError(w http.ResponseWriter, status int, err error) {
	result := map[string]interface{}{
		"code":    "error",
		"message": err.Error(),
	}
	renderJSON(w, status, result)
}

func renderCSV(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "text/csv")
	w.WriteHeader(status)
	w.Write(data)
}

func renderJSON(w http.ResponseWriter, status int, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	w.Write(data)
}

func fileSize(fileName string) (int64, error) {
	fi, err := os.Stat(fileName)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
