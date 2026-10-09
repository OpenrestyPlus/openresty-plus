package httpapi

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type runtimeVersion struct {
	ID        string    `json:"id"`
	VersionNo int64     `json:"versionNo"`
	Checksum  string    `json:"checksum"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
}
type runtimeDraft struct {
	PublishedVersionNo *int64   `json:"publishedVersionNo,omitempty"`
	ChangeCount        int      `json:"changeCount"`
	ChangedSections    []string `json:"changedSections"`
	Checksum           string   `json:"checksum"`
}

func (s *Server) runtimeModel(center uuid.UUID) (map[string]any, error) {
	model := map[string]any{"schemaVersion": 1, "centerId": center.String()}
	for table, section := range map[string]string{
		"http_configuration": "httpConfiguration",
		"http_upstream":      "httpUpstreams",
		"http_server":        "httpServers",
		"stream_upstream":    "streamUpstreams",
		"stream_server":      "streamServers",
	} {
		rows, err := s.tableRows(table, center)
		if err != nil {
			return nil, err
		}
		model[section] = rows
	}
	locations, err := s.db.Query(`SELECT l.* FROM http_location l JOIN http_server s ON s.id=l.server_id WHERE s.center_id=?`, center[:])
	if err == nil {
		defer locations.Close()
		model["httpLocations"] = scanRows(locations)
	} else {
		model["httpLocations"] = []map[string]any{}
	}
	return model, nil
}
func (s *Server) tableRows(table string, center uuid.UUID) ([]map[string]any, error) {
	// These are fixed internal table names, not user input.
	rows, err := s.db.Query("SELECT * FROM "+table+" WHERE center_id=?", center[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	out := []map[string]any{}
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		out = append(out, rowMap(cols, values))
	}
	return out, rows.Err()
}
func scanRows(rows *sql.Rows) []map[string]any {
	cols, _ := rows.Columns()
	out := []map[string]any{}
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if rows.Scan(ptrs...) == nil {
			out = append(out, rowMap(cols, values))
		}
	}
	return out
}
func rowMap(cols []string, values []any) map[string]any {
	item := map[string]any{}
	for i, col := range cols {
		if b, ok := values[i].([]byte); ok {
			// MySQL 的 BINARY(16) UUID 不能直接转成 UTF-8 字符串，否则前端会显示乱码。
			if len(b) == 16 && (col == "id" || col == "center_id" || col == "server_id" || col == "upstream_id" || strings.HasSuffix(col, "_id")) {
				item[col] = uuidText(b)
			} else {
				item[col] = string(b)
			}
		} else {
			item[col] = values[i]
		}
	}
	return item
}
func runtimeJSON(model map[string]any) ([]byte, string, error) {
	b, err := json.Marshal(model)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(b)
	return b, hex.EncodeToString(sum[:]), nil
}
func (s *Server) listRuntimeConfigurations(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	rows, err := s.db.Query("SELECT id,version_no,checksum,state,created_at FROM runtime_configuration_version WHERE center_id=? ORDER BY version_no DESC", center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer rows.Close()
	out := []runtimeVersion{}
	for rows.Next() {
		var raw []byte
		var v runtimeVersion
		if err := rows.Scan(&raw, &v.VersionNo, &v.Checksum, &v.State, &v.CreatedAt); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		v.ID = uuidText(raw)
		out = append(out, v)
	}
	writeJSON(w, 200, out)
}
func (s *Server) pageRuntimeConfigurations(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size <= 0 {
		size = 10
	}
	all := []runtimeVersion{}
	rr := httptestRequest(r)
	_ = rr
	_ = page
	_ = size // list logic is kept in one response for compatibility with the existing UI.
	rows, err := s.db.Query("SELECT id,version_no,checksum,state,created_at FROM runtime_configuration_version WHERE center_id=? ORDER BY version_no DESC", center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var v runtimeVersion
		if err := rows.Scan(&raw, &v.VersionNo, &v.Checksum, &v.State, &v.CreatedAt); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		v.ID = uuidText(raw)
		all = append(all, v)
	}
	start := page * size
	if start > len(all) {
		start = len(all)
	}
	end := start + size
	if end > len(all) {
		end = len(all)
	}
	writeJSON(w, 200, map[string]any{"records": all[start:end], "total": len(all), "page": page, "size": size})
}
func httptestRequest(r *http.Request) *http.Request { return r }
func (s *Server) currentRuntimeConfiguration(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	var raw, content []byte
	var v runtimeVersion
	if err := s.db.QueryRow("SELECT id,version_no,checksum,state,created_at,content FROM runtime_configuration_version WHERE center_id=? ORDER BY version_no DESC LIMIT 1", center[:]).Scan(&raw, &v.VersionNo, &v.Checksum, &v.State, &v.CreatedAt, &content); err != nil {
		writeError(w, 404, "No runtime configuration has been published")
		return
	}
	v.ID = uuidText(raw)
	var value any
	_ = json.Unmarshal(content, &value)
	writeJSON(w, 200, map[string]any{"id": v.ID, "versionNo": v.VersionNo, "checksum": v.Checksum, "content": value, "changed": true})
}
func (s *Server) draftRuntimeConfiguration(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	model, err := s.runtimeModel(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	_, sum, err := runtimeJSON(model)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	var published sql.NullInt64
	var old string
	_ = s.db.QueryRow("SELECT version_no,checksum FROM runtime_configuration_version WHERE center_id=? ORDER BY version_no DESC LIMIT 1", center[:]).Scan(&published, &old)
	changed := 1
	if old == sum {
		changed = 0
	}
	var version *int64
	if published.Valid {
		version = &published.Int64
	}
	changedSections := []string{}
	if changed > 0 {
		changedSections = []string{"httpConfiguration", "httpUpstreams", "httpServers", "httpLocations", "streamUpstreams", "streamServers"}
	}
	writeJSON(w, 200, runtimeDraft{version, changed, changedSections, sum})
}
func (s *Server) compareRuntimeConfiguration(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	model, err := s.runtimeModel(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	_, sum, err := runtimeJSON(model)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	var oldContent []byte
	var versionValue sql.NullInt64
	var old string
	_ = s.db.QueryRow("SELECT version_no,checksum,content FROM runtime_configuration_version WHERE center_id=? ORDER BY version_no DESC LIMIT 1", center[:]).Scan(&versionValue, &old, &oldContent)
	var version *int64
	if versionValue.Valid {
		version = &versionValue.Int64
	}
	var published any
	_ = json.Unmarshal(oldContent, &published)
	sections := []string{}
	if old != sum {
		sections = []string{"httpConfiguration", "httpUpstreams", "httpServers", "httpLocations", "streamUpstreams", "streamServers"}
	}
	writeJSON(w, 200, map[string]any{"publishedVersionNo": version, "published": published, "draft": model, "changedSections": sections})
}
func (s *Server) publishRuntimeConfiguration(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "旧版运行配置发布入口已停用；请使用 /api/orp/publish/precheck 和 /api/orp/publish 完成节点校验与发布")
}
func (s *Server) listReloadTasks(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	s.reloadTasks(w, center, false)
}
func (s *Server) pageReloadTasks(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	s.reloadTasks(w, center, true)
}
func (s *Server) reloadTasks(w http.ResponseWriter, center uuid.UUID, paged bool) {
	rows, err := s.db.Query("SELECT id,status,created_at,completed_at FROM control_api_reload_task WHERE center_id=? ORDER BY created_at DESC", center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var raw []byte
		var status string
		var created, completed sql.NullTime
		if err := rows.Scan(&raw, &status, &created, &completed); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		out = append(out, map[string]any{"id": uuidText(raw), "status": status, "createdAt": created.Time, "completedAt": completed.Time})
	}
	if paged {
		writeJSON(w, 200, map[string]any{"records": out, "total": len(out), "page": 0, "size": len(out)})
	} else {
		writeJSON(w, 200, out)
	}
}
