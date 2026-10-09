package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type centerView struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}
type centerRequest struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}
type page[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Size       int `json:"size"`
	TotalPages int `json:"totalPages"`
}

func (server *Server) listCenters(writer http.ResponseWriter, _ *http.Request) {
	values, err := server.centers()
	if err != nil {
		writeError(writer, 500, err.Error())
		return
	}
	writeJSON(writer, 200, values)
}
func (server *Server) pageCenters(writer http.ResponseWriter, request *http.Request) {
	values, err := server.centers()
	if err != nil {
		writeError(writer, 500, err.Error())
		return
	}
	writeJSON(writer, 200, paginate(request, values))
}
func (server *Server) centers() ([]centerView, error) {
	if server.db == nil {
		return nil, sql.ErrConnDone
	}
	rows, err := server.db.Query("SELECT id, code, name, enabled FROM center ORDER BY code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []centerView{}
	for rows.Next() {
		var id []byte
		var value centerView
		if err := rows.Scan(&id, &value.Code, &value.Name, &value.Enabled); err != nil {
			return nil, err
		}
		value.ID = uuidText(id)
		values = append(values, value)
	}
	return values, rows.Err()
}
func (server *Server) createCenter(writer http.ResponseWriter, request *http.Request) {
	var input centerRequest
	if !decode(request, &input) || strings.TrimSpace(input.Code) == "" || strings.TrimSpace(input.Name) == "" {
		writeError(writer, 400, "code and name are required")
		return
	}
	id := uuid.New()
	_, err := server.db.Exec("INSERT INTO center (id, code, name, enabled, created_at) VALUES (?, ?, ?, true, ?)", id[:], strings.TrimSpace(input.Code), strings.TrimSpace(input.Name), time.Now())
	if err != nil {
		writeError(writer, 409, err.Error())
		return
	}
	server.audit(id, "CENTER_CREATED", "CENTER", id)
	writeJSON(writer, 201, centerView{ID: id.String(), Code: strings.TrimSpace(input.Code), Name: strings.TrimSpace(input.Name), Enabled: true})
}
func (server *Server) updateCenter(writer http.ResponseWriter, request *http.Request) {
	id, ok := pathUUID(writer, request, "centerID")
	if !ok {
		return
	}
	var input centerRequest
	if !decode(request, &input) || strings.TrimSpace(input.Code) == "" || strings.TrimSpace(input.Name) == "" {
		writeError(writer, 400, "code and name are required")
		return
	}
	result, err := server.db.Exec("UPDATE center SET code=?, name=?, enabled=? WHERE id=?", strings.TrimSpace(input.Code), strings.TrimSpace(input.Name), input.Enabled, id[:])
	if err != nil {
		writeError(writer, 500, err.Error())
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		writeError(writer, 404, "Center not found")
		return
	}
	server.audit(id, "CENTER_UPDATED", "CENTER", id)
	writeJSON(writer, 200, centerView{ID: id.String(), Code: strings.TrimSpace(input.Code), Name: strings.TrimSpace(input.Name), Enabled: input.Enabled})
}
func (server *Server) deleteCenter(writer http.ResponseWriter, request *http.Request) {
	id, ok := pathUUID(writer, request, "centerID")
	if !ok {
		return
	}
	result, err := server.db.Exec("DELETE FROM center WHERE id=?", id[:])
	if err != nil {
		writeError(writer, 500, err.Error())
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		writeError(writer, 404, "Center not found")
		return
	}
	server.audit(id, "CENTER_DELETED", "CENTER", id)
	writer.WriteHeader(http.StatusNoContent)
}
func decode(request *http.Request, target any) bool {
	return json.NewDecoder(request.Body).Decode(target) == nil
}
func pathUUID(writer http.ResponseWriter, request *http.Request, key string) (uuid.UUID, bool) {
	value, err := uuid.Parse(request.PathValue(key))
	if err != nil {
		writeError(writer, 400, "invalid identifier")
		return uuid.Nil, false
	}
	return value, true
}
func uuidText(value []byte) string {
	id, err := uuid.FromBytes(value)
	if err != nil {
		return ""
	}
	return id.String()
}
func paginate[T any](request *http.Request, values []T) page[T] {
	number, _ := strconv.Atoi(request.URL.Query().Get("page"))
	size, _ := strconv.Atoi(request.URL.Query().Get("size"))
	if number < 0 {
		number = 0
	}
	if size < 1 {
		size = 10
	}
	if size > 200 {
		size = 200
	}
	start := number * size
	if start > len(values) {
		start = len(values)
	}
	end := start + size
	if end > len(values) {
		end = len(values)
	}
	pages := 0
	if len(values) > 0 {
		pages = (len(values) + size - 1) / size
	}
	return page[T]{Items: values[start:end], Total: len(values), Page: number, Size: size, TotalPages: pages}
}
