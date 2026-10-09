package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type targetView struct {
	ID                 string `json:"id"`
	TargetHost         string `json:"targetHost"`
	TargetPort         int    `json:"targetPort"`
	Weight             int    `json:"weight"`
	MaxFails           int    `json:"maxFails"`
	FailTimeoutSeconds int    `json:"failTimeoutSeconds"`
	ResolveEnabled     bool   `json:"resolveEnabled"`
	Backup             bool   `json:"backup"`
	Enabled            bool   `json:"enabled"`
}
type targetRequest struct {
	TargetHost         string `json:"targetHost"`
	TargetPort         int    `json:"targetPort"`
	Weight             int    `json:"weight"`
	MaxFails           int    `json:"maxFails"`
	FailTimeoutSeconds int    `json:"failTimeoutSeconds"`
	ResolveEnabled     bool   `json:"resolveEnabled"`
	Backup             bool   `json:"backup"`
	Enabled            bool   `json:"enabled"`
}
type upstreamView struct {
	ID                             string       `json:"id"`
	Name                           string       `json:"name"`
	KeepaliveConnections           int          `json:"keepaliveConnections"`
	ZoneSizeKilobytes              int          `json:"zoneSizeKilobytes"`
	HealthCheckEnabled             bool         `json:"healthCheckEnabled"`
	HealthCheckType                string       `json:"healthCheckType"`
	HealthCheckPath                string       `json:"healthCheckPath"`
	HealthCheckIntervalSeconds     int          `json:"healthCheckIntervalSeconds"`
	HealthCheckTimeoutMilliseconds int          `json:"healthCheckTimeoutMilliseconds"`
	HealthCheckExpectedStatus      int          `json:"healthCheckExpectedStatus"`
	HealthCheckHost                *string      `json:"healthCheckHost"`
	HealthCheckRequestHeaders      []string     `json:"healthCheckRequestHeaders"`
	HealthCheckRise                int          `json:"healthCheckRise"`
	HealthCheckFall                int          `json:"healthCheckFall"`
	Targets                        []targetView `json:"targets"`
}
type upstreamRequest struct {
	Name                           string   `json:"name"`
	KeepaliveConnections           int      `json:"keepaliveConnections"`
	ZoneSizeKilobytes              int      `json:"zoneSizeKilobytes"`
	HealthCheckEnabled             bool     `json:"healthCheckEnabled"`
	HealthCheckType                string   `json:"healthCheckType"`
	HealthCheckPath                string   `json:"healthCheckPath"`
	HealthCheckIntervalSeconds     int      `json:"healthCheckIntervalSeconds"`
	HealthCheckTimeoutMilliseconds int      `json:"healthCheckTimeoutMilliseconds"`
	HealthCheckExpectedStatus      int      `json:"healthCheckExpectedStatus"`
	HealthCheckHost                *string  `json:"healthCheckHost"`
	HealthCheckRequestHeaders      []string `json:"healthCheckRequestHeaders"`
	HealthCheckRise                int      `json:"healthCheckRise"`
	HealthCheckFall                int      `json:"healthCheckFall"`
}

func (server *Server) listHTTPUpstreams(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.httpUpstreams(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, values)
}
func (server *Server) pageHTTPUpstreams(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.httpUpstreams(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, paginate(r, values))
}
func (server *Server) httpUpstreams(center uuid.UUID) ([]upstreamView, error) {
	rows, err := server.db.Query(`SELECT id,name,keepalive_connections,zone_size_kilobytes,health_check_enabled,health_check_type,health_check_path,health_check_interval_seconds,health_check_timeout_milliseconds,health_check_expected_status,health_check_host,health_check_request_headers,health_check_rise,health_check_fall FROM http_upstream WHERE center_id=? ORDER BY name`, center[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []upstreamView{}
	for rows.Next() {
		var id []byte
		var headers []byte
		var item upstreamView
		if err := rows.Scan(&id, &item.Name, &item.KeepaliveConnections, &item.ZoneSizeKilobytes, &item.HealthCheckEnabled, &item.HealthCheckType, &item.HealthCheckPath, &item.HealthCheckIntervalSeconds, &item.HealthCheckTimeoutMilliseconds, &item.HealthCheckExpectedStatus, &item.HealthCheckHost, &headers, &item.HealthCheckRise, &item.HealthCheckFall); err != nil {
			return nil, err
		}
		item.ID = uuidText(id)
		_ = json.Unmarshal(headers, &item.HealthCheckRequestHeaders)
		item.Targets, _ = server.httpTargets(uuid.MustParse(item.ID))
		values = append(values, item)
	}
	return values, rows.Err()
}
func (server *Server) createHTTPUpstream(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	var input upstreamRequest
	if !validUpstream(w, r, &input) {
		return
	}
	id := uuid.New()
	headers, _ := json.Marshal(input.HealthCheckRequestHeaders)
	_, err := server.db.Exec(`INSERT INTO http_upstream (id,center_id,name,keepalive_connections,zone_size_kilobytes,health_check_enabled,health_check_type,health_check_path,health_check_interval_seconds,health_check_timeout_milliseconds,health_check_expected_status,health_check_host,health_check_request_headers,health_check_rise,health_check_fall,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, id[:], center[:], strings.TrimSpace(input.Name), input.KeepaliveConnections, input.ZoneSizeKilobytes, input.HealthCheckEnabled, input.HealthCheckType, input.HealthCheckPath, input.HealthCheckIntervalSeconds, input.HealthCheckTimeoutMilliseconds, input.HealthCheckExpectedStatus, input.HealthCheckHost, headers, input.HealthCheckRise, input.HealthCheckFall, time.Now())
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	server.audit(center, "HTTP_UPSTREAM_CREATED", "HTTP_UPSTREAM", id)
	item, _ := server.findHTTPUpstream(center, id)
	writeJSON(w, 201, item)
}
func (server *Server) updateHTTPUpstream(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "upstreamID")
	if !ok {
		return
	}
	var input upstreamRequest
	if !validUpstream(w, r, &input) {
		return
	}
	headers, _ := json.Marshal(input.HealthCheckRequestHeaders)
	result, err := server.db.Exec(`UPDATE http_upstream SET name=?,keepalive_connections=?,zone_size_kilobytes=?,health_check_enabled=?,health_check_type=?,health_check_path=?,health_check_interval_seconds=?,health_check_timeout_milliseconds=?,health_check_expected_status=?,health_check_host=?,health_check_request_headers=?,health_check_rise=?,health_check_fall=? WHERE id=? AND center_id=?`, strings.TrimSpace(input.Name), input.KeepaliveConnections, input.ZoneSizeKilobytes, input.HealthCheckEnabled, input.HealthCheckType, input.HealthCheckPath, input.HealthCheckIntervalSeconds, input.HealthCheckTimeoutMilliseconds, input.HealthCheckExpectedStatus, input.HealthCheckHost, headers, input.HealthCheckRise, input.HealthCheckFall, id[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "HTTP upstream not found")
		return
	}
	server.audit(center, "HTTP_UPSTREAM_UPDATED", "HTTP_UPSTREAM", id)
	item, _ := server.findHTTPUpstream(center, id)
	writeJSON(w, 200, item)
}
func (server *Server) deleteHTTPUpstream(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "upstreamID")
	if !ok {
		return
	}
	result, err := server.db.Exec("DELETE FROM http_upstream WHERE id=? AND center_id=?", id[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "HTTP upstream not found")
		return
	}
	server.audit(center, "HTTP_UPSTREAM_DELETED", "HTTP_UPSTREAM", id)
	w.WriteHeader(204)
}
func validUpstream(w http.ResponseWriter, r *http.Request, input *upstreamRequest) bool {
	if !decode(r, input) || strings.TrimSpace(input.Name) == "" || input.KeepaliveConnections < 1 || input.ZoneSizeKilobytes < 8 || input.HealthCheckPath == "" {
		writeError(w, 400, "invalid HTTP upstream")
		return false
	}
	if input.HealthCheckType == "" {
		input.HealthCheckType = "HTTP"
	}
	if input.HealthCheckIntervalSeconds == 0 {
		input.HealthCheckIntervalSeconds = 10
	}
	if input.HealthCheckTimeoutMilliseconds == 0 {
		input.HealthCheckTimeoutMilliseconds = 1000
	}
	if input.HealthCheckExpectedStatus == 0 {
		input.HealthCheckExpectedStatus = 200
	}
	if input.HealthCheckRise == 0 {
		input.HealthCheckRise = 2
	}
	if input.HealthCheckFall == 0 {
		input.HealthCheckFall = 3
	}
	return true
}
func (server *Server) findHTTPUpstream(center, id uuid.UUID) (upstreamView, error) {
	items, err := server.httpUpstreams(center)
	if err != nil {
		return upstreamView{}, err
	}
	for _, item := range items {
		if item.ID == id.String() {
			return item, nil
		}
	}
	return upstreamView{}, nil
}
func (server *Server) httpTargets(upstream uuid.UUID) ([]targetView, error) {
	rows, err := server.db.Query("SELECT id,target_host,target_port,weight,max_fails,fail_timeout_seconds,resolve_enabled,backup,enabled FROM http_upstream_target WHERE upstream_id=? ORDER BY target_host,target_port", upstream[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []targetView{}
	for rows.Next() {
		var raw []byte
		var item targetView
		if err := rows.Scan(&raw, &item.TargetHost, &item.TargetPort, &item.Weight, &item.MaxFails, &item.FailTimeoutSeconds, &item.ResolveEnabled, &item.Backup, &item.Enabled); err != nil {
			return nil, err
		}
		item.ID = uuidText(raw)
		items = append(items, item)
	}
	return items, rows.Err()
}
func (server *Server) pageHTTPUpstreamTargets(w http.ResponseWriter, r *http.Request) {
	_, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	upstream, ok := pathUUID(w, r, "upstreamID")
	if !ok {
		return
	}
	items, err := server.httpTargets(upstream)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, paginate(r, items))
}
func (server *Server) createHTTPUpstreamTarget(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	upstream, ok := pathUUID(w, r, "upstreamID")
	if !ok {
		return
	}
	var input targetRequest
	if !validTarget(w, r, &input) {
		return
	}
	id := uuid.New()
	_, err := server.db.Exec("INSERT INTO http_upstream_target (id,upstream_id,target_host,target_port,weight,max_fails,fail_timeout_seconds,resolve_enabled,backup,enabled) SELECT ?,id,?,?,?,?,?,?,?,? FROM http_upstream WHERE id=? AND center_id=?", id[:], strings.TrimSpace(input.TargetHost), input.TargetPort, input.Weight, input.MaxFails, input.FailTimeoutSeconds, input.ResolveEnabled, input.Backup, input.Enabled, upstream[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	server.audit(center, "HTTP_UPSTREAM_TARGET_CREATED", "HTTP_UPSTREAM_TARGET", id)
	writeJSON(w, 201, targetView{ID: id.String(), TargetHost: strings.TrimSpace(input.TargetHost), TargetPort: input.TargetPort, Weight: input.Weight, MaxFails: input.MaxFails, FailTimeoutSeconds: input.FailTimeoutSeconds, ResolveEnabled: input.ResolveEnabled, Backup: input.Backup, Enabled: input.Enabled})
}
func (server *Server) updateHTTPUpstreamTarget(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	upstream, ok := pathUUID(w, r, "upstreamID")
	if !ok {
		return
	}
	target, ok := pathUUID(w, r, "targetID")
	if !ok {
		return
	}
	var input targetRequest
	if !validTarget(w, r, &input) {
		return
	}
	result, err := server.db.Exec("UPDATE http_upstream_target t JOIN http_upstream u ON u.id=t.upstream_id SET t.target_host=?,t.target_port=?,t.weight=?,t.max_fails=?,t.fail_timeout_seconds=?,t.resolve_enabled=?,t.backup=?,t.enabled=? WHERE t.id=? AND t.upstream_id=? AND u.center_id=?", strings.TrimSpace(input.TargetHost), input.TargetPort, input.Weight, input.MaxFails, input.FailTimeoutSeconds, input.ResolveEnabled, input.Backup, input.Enabled, target[:], upstream[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "HTTP upstream target not found")
		return
	}
	server.audit(center, "HTTP_UPSTREAM_TARGET_UPDATED", "HTTP_UPSTREAM_TARGET", target)
	writeJSON(w, 200, targetView{ID: target.String(), TargetHost: strings.TrimSpace(input.TargetHost), TargetPort: input.TargetPort, Weight: input.Weight, MaxFails: input.MaxFails, FailTimeoutSeconds: input.FailTimeoutSeconds, ResolveEnabled: input.ResolveEnabled, Backup: input.Backup, Enabled: input.Enabled})
}
func (server *Server) deleteHTTPUpstreamTarget(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	upstream, ok := pathUUID(w, r, "upstreamID")
	if !ok {
		return
	}
	target, ok := pathUUID(w, r, "targetID")
	if !ok {
		return
	}
	result, err := server.db.Exec("DELETE t FROM http_upstream_target t JOIN http_upstream u ON u.id=t.upstream_id WHERE t.id=? AND t.upstream_id=? AND u.center_id=?", target[:], upstream[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "HTTP upstream target not found")
		return
	}
	server.audit(center, "HTTP_UPSTREAM_TARGET_DELETED", "HTTP_UPSTREAM_TARGET", target)
	w.WriteHeader(204)
}
func validTarget(w http.ResponseWriter, r *http.Request, input *targetRequest) bool {
	if !decode(r, input) || strings.TrimSpace(input.TargetHost) == "" || input.TargetPort < 1 || input.TargetPort > 65535 || input.Weight < 1 || input.MaxFails < 0 || input.FailTimeoutSeconds < 1 {
		writeError(w, 400, "invalid HTTP upstream target")
		return false
	}
	return true
}
