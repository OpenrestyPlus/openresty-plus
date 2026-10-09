package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type streamUpstreamView struct {
	ID                             string   `json:"id"`
	Name                           string   `json:"name"`
	TargetHost                     string   `json:"targetHost"`
	TargetPort                     int      `json:"targetPort"`
	ResolveEnabled                 bool     `json:"resolveEnabled"`
	ZoneSizeKilobytes              int      `json:"zoneSizeKilobytes"`
	HealthCheckEnabled             bool     `json:"healthCheckEnabled"`
	HealthCheckType                string   `json:"healthCheckType"`
	HealthCheckPath                string   `json:"healthCheckPath"`
	HealthCheckHost                *string  `json:"healthCheckHost"`
	HealthCheckIntervalSeconds     int      `json:"healthCheckIntervalSeconds"`
	HealthCheckTimeoutMilliseconds int      `json:"healthCheckTimeoutMilliseconds"`
	HealthCheckExpectedStatus      int      `json:"healthCheckExpectedStatus"`
	HealthCheckRequestHeaders      []string `json:"healthCheckRequestHeaders"`
	HealthCheckRise                int      `json:"healthCheckRise"`
	HealthCheckFall                int      `json:"healthCheckFall"`
}
type streamUpstreamRequest struct {
	Name                           string   `json:"name"`
	TargetHost                     string   `json:"targetHost"`
	TargetPort                     int      `json:"targetPort"`
	ResolveEnabled                 bool     `json:"resolveEnabled"`
	ZoneSizeKilobytes              int      `json:"zoneSizeKilobytes"`
	HealthCheckEnabled             bool     `json:"healthCheckEnabled"`
	HealthCheckType                string   `json:"healthCheckType"`
	HealthCheckPath                string   `json:"healthCheckPath"`
	HealthCheckHost                *string  `json:"healthCheckHost"`
	HealthCheckIntervalSeconds     int      `json:"healthCheckIntervalSeconds"`
	HealthCheckTimeoutMilliseconds int      `json:"healthCheckTimeoutMilliseconds"`
	HealthCheckExpectedStatus      int      `json:"healthCheckExpectedStatus"`
	HealthCheckRequestHeaders      []string `json:"healthCheckRequestHeaders"`
	HealthCheckRise                int      `json:"healthCheckRise"`
	HealthCheckFall                int      `json:"healthCheckFall"`
}
type streamServerView struct {
	ID                string  `json:"id"`
	ServiceName       string  `json:"serviceName"`
	ListenPort        int     `json:"listenPort"`
	Protocol          string  `json:"protocol"`
	UpstreamID        string  `json:"upstreamId"`
	AccessLog         string  `json:"accessLog"`
	ErrorLog          string  `json:"errorLog"`
	DynamicDNSEnabled bool    `json:"dynamicDnsEnabled"`
	DynamicDNSHost    *string `json:"dynamicDnsHost"`
	DynamicDNSPort    *int    `json:"dynamicDnsPort"`
	IPPolicyEnabled   bool    `json:"ipPolicyEnabled"`
	IPPolicyModeOrder string  `json:"ipPolicyModeOrder"`
}
type streamServerRequest struct {
	ServiceName       string  `json:"serviceName"`
	ListenPort        int     `json:"listenPort"`
	Protocol          string  `json:"protocol"`
	UpstreamID        string  `json:"upstreamId"`
	AccessLog         string  `json:"accessLog"`
	ErrorLog          string  `json:"errorLog"`
	DynamicDNSEnabled bool    `json:"dynamicDnsEnabled"`
	DynamicDNSHost    *string `json:"dynamicDnsHost"`
	DynamicDNSPort    *int    `json:"dynamicDnsPort"`
}
type streamPolicyRequest struct {
	IPPolicyEnabled   bool    `json:"ipPolicyEnabled"`
	IPPolicyModeOrder *string `json:"ipPolicyModeOrder"`
}

func (server *Server) listStreamUpstreams(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.streamUpstreams(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, values)
}
func (server *Server) pageStreamUpstreams(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.streamUpstreams(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, paginate(r, values))
}
func (server *Server) streamUpstreams(center uuid.UUID) ([]streamUpstreamView, error) {
	rows, err := server.db.Query(`SELECT id,name,target_host,target_port,resolve_enabled,zone_size_kilobytes,health_check_enabled,health_check_type,health_check_path,health_check_host,health_check_interval_seconds,health_check_timeout_milliseconds,health_check_expected_status,health_check_request_headers,health_check_rise,health_check_fall FROM stream_upstream WHERE center_id=? ORDER BY name`, center[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []streamUpstreamView{}
	for rows.Next() {
		var raw, headers []byte
		var item streamUpstreamView
		if err := rows.Scan(&raw, &item.Name, &item.TargetHost, &item.TargetPort, &item.ResolveEnabled, &item.ZoneSizeKilobytes, &item.HealthCheckEnabled, &item.HealthCheckType, &item.HealthCheckPath, &item.HealthCheckHost, &item.HealthCheckIntervalSeconds, &item.HealthCheckTimeoutMilliseconds, &item.HealthCheckExpectedStatus, &headers, &item.HealthCheckRise, &item.HealthCheckFall); err != nil {
			return nil, err
		}
		item.ID = uuidText(raw)
		_ = json.Unmarshal(headers, &item.HealthCheckRequestHeaders)
		item.HealthCheckRequestHeaders = nonNilStrings(item.HealthCheckRequestHeaders)
		values = append(values, item)
	}
	return values, rows.Err()
}
func (server *Server) createStreamUpstream(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	var input streamUpstreamRequest
	if !validStreamUpstream(w, r, &input) {
		return
	}
	headers, _ := json.Marshal(nonNilStrings(input.HealthCheckRequestHeaders))
	id := uuid.New()
	_, err := server.db.Exec(`INSERT INTO stream_upstream (id,center_id,name,target_host,target_port,resolve_enabled,zone_size_kilobytes,health_check_enabled,health_check_type,health_check_path,health_check_host,health_check_interval_seconds,health_check_timeout_milliseconds,health_check_expected_status,health_check_request_headers,health_check_rise,health_check_fall,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, id[:], center[:], strings.TrimSpace(input.Name), strings.TrimSpace(input.TargetHost), input.TargetPort, input.ResolveEnabled, input.ZoneSizeKilobytes, input.HealthCheckEnabled, input.HealthCheckType, strings.TrimSpace(input.HealthCheckPath), nullableString(input.HealthCheckHost), input.HealthCheckIntervalSeconds, input.HealthCheckTimeoutMilliseconds, input.HealthCheckExpectedStatus, headers, input.HealthCheckRise, input.HealthCheckFall, time.Now())
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	server.audit(center, "STREAM_UPSTREAM_CREATED", "STREAM_UPSTREAM", id)
	item, _ := server.findStreamUpstream(center, id)
	writeJSON(w, 201, item)
}
func (server *Server) updateStreamUpstream(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "upstreamID")
	if !ok {
		return
	}
	var input streamUpstreamRequest
	if !validStreamUpstream(w, r, &input) {
		return
	}
	headers, _ := json.Marshal(nonNilStrings(input.HealthCheckRequestHeaders))
	result, err := server.db.Exec(`UPDATE stream_upstream SET name=?,target_host=?,target_port=?,resolve_enabled=?,zone_size_kilobytes=?,health_check_enabled=?,health_check_type=?,health_check_path=?,health_check_host=?,health_check_interval_seconds=?,health_check_timeout_milliseconds=?,health_check_expected_status=?,health_check_request_headers=?,health_check_rise=?,health_check_fall=? WHERE id=? AND center_id=?`, strings.TrimSpace(input.Name), strings.TrimSpace(input.TargetHost), input.TargetPort, input.ResolveEnabled, input.ZoneSizeKilobytes, input.HealthCheckEnabled, input.HealthCheckType, strings.TrimSpace(input.HealthCheckPath), nullableString(input.HealthCheckHost), input.HealthCheckIntervalSeconds, input.HealthCheckTimeoutMilliseconds, input.HealthCheckExpectedStatus, headers, input.HealthCheckRise, input.HealthCheckFall, id[:], center[:])
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "Stream upstream not found")
		return
	}
	server.audit(center, "STREAM_UPSTREAM_UPDATED", "STREAM_UPSTREAM", id)
	item, _ := server.findStreamUpstream(center, id)
	writeJSON(w, 200, item)
}
func (server *Server) deleteStreamUpstream(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "upstreamID")
	if !ok {
		return
	}
	result, err := server.db.Exec("DELETE FROM stream_upstream WHERE id=? AND center_id=?", id[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "Stream upstream not found")
		return
	}
	server.audit(center, "STREAM_UPSTREAM_DELETED", "STREAM_UPSTREAM", id)
	w.WriteHeader(204)
}
func (server *Server) findStreamUpstream(center, id uuid.UUID) (streamUpstreamView, error) {
	values, err := server.streamUpstreams(center)
	if err != nil {
		return streamUpstreamView{}, err
	}
	for _, value := range values {
		if value.ID == id.String() {
			return value, nil
		}
	}
	return streamUpstreamView{}, nil
}
func validStreamUpstream(w http.ResponseWriter, r *http.Request, input *streamUpstreamRequest) bool {
	if !decode(r, input) || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.TargetHost) == "" || input.TargetPort < 1 || input.TargetPort > 65535 || input.ZoneSizeKilobytes < 8 || input.ZoneSizeKilobytes > 65536 || !validHealthCheck(*input) {
		writeError(w, 400, "invalid stream upstream")
		return false
	}
	return true
}
func validHealthCheck(input streamUpstreamRequest) bool {
	return (input.HealthCheckType == "TCP" || input.HealthCheckType == "HTTP") && strings.TrimSpace(input.HealthCheckPath) != "" && input.HealthCheckIntervalSeconds >= 1 && input.HealthCheckIntervalSeconds <= 3600 && input.HealthCheckTimeoutMilliseconds >= 50 && input.HealthCheckTimeoutMilliseconds <= 60000 && input.HealthCheckExpectedStatus >= 100 && input.HealthCheckExpectedStatus <= 599 && input.HealthCheckRise >= 1 && input.HealthCheckRise <= 100 && input.HealthCheckFall >= 1 && input.HealthCheckFall <= 100 && validHeaders(input.HealthCheckRequestHeaders)
}

func (server *Server) listStreamServers(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.streamServers(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, values)
}
func (server *Server) pageStreamServers(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.streamServers(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, paginate(r, values))
}
func (server *Server) streamServers(center uuid.UUID) ([]streamServerView, error) {
	rows, err := server.db.Query(`SELECT id,service_name,listen_port,protocol,upstream_id,access_log,error_log,dynamic_dns_enabled,dynamic_dns_host,dynamic_dns_port,ip_policy_enabled,ip_policy_mode_order FROM stream_server WHERE center_id=? ORDER BY listen_port`, center[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []streamServerView{}
	for rows.Next() {
		var raw, upstream []byte
		var item streamServerView
		if err := rows.Scan(&raw, &item.ServiceName, &item.ListenPort, &item.Protocol, &upstream, &item.AccessLog, &item.ErrorLog, &item.DynamicDNSEnabled, &item.DynamicDNSHost, &item.DynamicDNSPort, &item.IPPolicyEnabled, &item.IPPolicyModeOrder); err != nil {
			return nil, err
		}
		item.ID = uuidText(raw)
		item.UpstreamID = uuidText(upstream)
		values = append(values, item)
	}
	return values, rows.Err()
}
func (server *Server) createStreamServer(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	var input streamServerRequest
	if !validStreamServer(w, r, &input) {
		return
	}
	upstream, ok := optionalUUID(w, &input.UpstreamID)
	if !ok {
		return
	}
	if upstream == nil || !validStreamUpstreamID(server.db, center, *upstream) {
		writeError(w, 400, "Stream upstream not found")
		return
	}
	id := uuid.New()
	access, errorLog := streamLogPaths(input)
	_, err := server.db.Exec(`INSERT INTO stream_server (id,center_id,service_name,listen_port,protocol,upstream_id,access_log,error_log,dynamic_dns_enabled,dynamic_dns_host,dynamic_dns_port,ip_policy_enabled,ip_policy_mode_order,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,false,'BLACKLIST_FIRST',?)`, id[:], center[:], strings.TrimSpace(input.ServiceName), input.ListenPort, input.Protocol, (*upstream)[:], access, errorLog, input.DynamicDNSEnabled, nullableString(input.DynamicDNSHost), input.DynamicDNSPort, time.Now())
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	server.audit(center, "STREAM_SERVER_CREATED", "STREAM_SERVER", id)
	item, _ := server.findStreamServer(center, id)
	writeJSON(w, 201, item)
}
func (server *Server) updateStreamServer(w http.ResponseWriter, r *http.Request) {
	center, id, ok := streamServerIDs(w, r)
	if !ok {
		return
	}
	var input streamServerRequest
	if !validStreamServer(w, r, &input) {
		return
	}
	upstream, ok := optionalUUID(w, &input.UpstreamID)
	if !ok {
		return
	}
	if upstream == nil || !validStreamUpstreamID(server.db, center, *upstream) {
		writeError(w, 400, "Stream upstream not found")
		return
	}
	access, errorLog := streamLogPaths(input)
	result, err := server.db.Exec(`UPDATE stream_server SET service_name=?,listen_port=?,protocol=?,upstream_id=?,access_log=?,error_log=?,dynamic_dns_enabled=?,dynamic_dns_host=?,dynamic_dns_port=? WHERE id=? AND center_id=?`, strings.TrimSpace(input.ServiceName), input.ListenPort, input.Protocol, (*upstream)[:], access, errorLog, input.DynamicDNSEnabled, nullableString(input.DynamicDNSHost), input.DynamicDNSPort, id[:], center[:])
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "Stream server not found")
		return
	}
	server.audit(center, "STREAM_SERVER_UPDATED", "STREAM_SERVER", id)
	item, _ := server.findStreamServer(center, id)
	writeJSON(w, 200, item)
}
func (server *Server) deleteStreamServer(w http.ResponseWriter, r *http.Request) {
	center, id, ok := streamServerIDs(w, r)
	if !ok {
		return
	}
	result, err := server.db.Exec("DELETE FROM stream_server WHERE id=? AND center_id=?", id[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "Stream server not found")
		return
	}
	server.audit(center, "STREAM_SERVER_DELETED", "STREAM_SERVER", id)
	w.WriteHeader(204)
}
func (server *Server) putStreamServerDynamicDNS(w http.ResponseWriter, r *http.Request) {
	center, id, ok := streamServerIDs(w, r)
	if !ok {
		return
	}
	var input dynamicDNSRequest
	if !decode(r, &input) || !validDNS(input) {
		writeError(w, 400, "启用动态 DNS 时必须提供域名和端口")
		return
	}
	result, err := server.db.Exec("UPDATE stream_server SET dynamic_dns_enabled=?,dynamic_dns_host=?,dynamic_dns_port=? WHERE id=? AND center_id=?", input.Enabled, nullableString(input.Host), input.Port, id[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "Stream server not found")
		return
	}
	item, _ := server.findStreamServer(center, id)
	writeJSON(w, 200, item)
}
func (server *Server) putStreamServerPolicySettings(w http.ResponseWriter, r *http.Request) {
	center, id, ok := streamServerIDs(w, r)
	if !ok {
		return
	}
	var input streamPolicyRequest
	if !decode(r, &input) {
		writeError(w, 400, "invalid policy settings")
		return
	}
	current, err := server.findStreamServer(center, id)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if current.ID == "" {
		writeError(w, 404, "Stream server not found")
		return
	}
	order := current.IPPolicyModeOrder
	if input.IPPolicyModeOrder != nil {
		order = strings.TrimSpace(*input.IPPolicyModeOrder)
	}
	if !validPolicyOrder(order) {
		writeError(w, 400, "invalid policy mode order")
		return
	}
	_, err = server.db.Exec("UPDATE stream_server SET ip_policy_enabled=?,ip_policy_mode_order=? WHERE id=? AND center_id=?", input.IPPolicyEnabled, order, id[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	server.audit(center, "STREAM_SERVER_POLICY_SETTINGS_UPDATED", "STREAM_SERVER", id)
	item, _ := server.findStreamServer(center, id)
	writeJSON(w, 200, item)
}
func (server *Server) findStreamServer(center, id uuid.UUID) (streamServerView, error) {
	values, err := server.streamServers(center)
	if err != nil {
		return streamServerView{}, err
	}
	for _, value := range values {
		if value.ID == id.String() {
			return value, nil
		}
	}
	return streamServerView{}, nil
}
func streamServerIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}
	id, ok := pathUUID(w, r, "streamServerID")
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}
	return center, id, true
}
func validStreamUpstreamID(db *sql.DB, center, id uuid.UUID) bool {
	var found int
	return db.QueryRow("SELECT 1 FROM stream_upstream WHERE id=? AND center_id=?", id[:], center[:]).Scan(&found) == nil
}
func validStreamServer(w http.ResponseWriter, r *http.Request, input *streamServerRequest) bool {
	if !decode(r, input) || strings.TrimSpace(input.ServiceName) == "" || input.ListenPort < 1 || input.ListenPort > 65535 || (input.Protocol != "TCP" && input.Protocol != "UDP") || strings.TrimSpace(input.UpstreamID) == "" || !validDNS(dynamicDNSRequest{input.DynamicDNSEnabled, input.DynamicDNSHost, input.DynamicDNSPort}) {
		writeError(w, 400, "invalid stream server")
		return false
	}
	return true
}
func streamLogPaths(input streamServerRequest) (string, string) {
	access := strings.TrimSpace(input.AccessLog)
	if access == "" {
		access = "/var/log/nginx/" + strings.TrimSpace(input.ServiceName) + "." + stringPort(input.ListenPort) + ".access.log"
	}
	errorLog := strings.TrimSpace(input.ErrorLog)
	if errorLog == "" {
		errorLog = "/var/log/nginx/" + strings.TrimSpace(input.ServiceName) + "." + stringPort(input.ListenPort) + ".error.log"
	}
	return access, errorLog
}
