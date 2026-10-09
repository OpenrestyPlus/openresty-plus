package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type httpLocationView struct {
	ID                    string   `json:"id"`
	Path                  string   `json:"path"`
	Methods               []string `json:"methods"`
	ContentTypes          []string `json:"contentTypes"`
	HeaderLengthMin       int      `json:"headerLengthMin"`
	HeaderLengthMax       int      `json:"headerLengthMax"`
	BodyLengthMin         int64    `json:"bodyLengthMin"`
	BodyLengthMax         int64    `json:"bodyLengthMax"`
	UpstreamID            *string  `json:"upstreamId"`
	ProxyConnectTimeoutMs int      `json:"proxyConnectTimeoutMs"`
	ProxyReadTimeoutMs    int      `json:"proxyReadTimeoutMs"`
	ProxySendTimeoutMs    int      `json:"proxySendTimeoutMs"`
	RateLimitEnabled      bool     `json:"rateLimitEnabled"`
	RatePerSecond         int      `json:"ratePerSecond"`
	RateLimitBurst        int      `json:"rateLimitBurst"`
	RateLimitNodelay      bool     `json:"rateLimitNodelay"`
	DynamicDNSEnabled     bool     `json:"dynamicDnsEnabled"`
	DynamicDNSHost        *string  `json:"dynamicDnsHost"`
	DynamicDNSPort        *int     `json:"dynamicDnsPort"`
	IPPolicyEnabled       bool     `json:"ipPolicyEnabled"`
	APIPolicyEnabled      bool     `json:"apiPolicyEnabled"`
	Action                string   `json:"action"`
	RootPath              *string  `json:"rootPath"`
	AliasPath             *string  `json:"aliasPath"`
	ReturnStatus          *int     `json:"returnStatus"`
	ReturnBody            *string  `json:"returnBody"`
	ReturnContentTypeMode string   `json:"returnContentTypeMode"`
	ReturnContentType     *string  `json:"returnContentType"`
	ResponseHeaders       []string `json:"responseHeaders"`
}
type httpLocationRequest struct {
	Path                  string   `json:"path"`
	Methods               []string `json:"methods"`
	ContentTypes          []string `json:"contentTypes"`
	HeaderLengthMin       int      `json:"headerLengthMin"`
	HeaderLengthMax       int      `json:"headerLengthMax"`
	BodyLengthMin         int64    `json:"bodyLengthMin"`
	BodyLengthMax         int64    `json:"bodyLengthMax"`
	UpstreamID            *string  `json:"upstreamId"`
	ProxyConnectTimeoutMs int      `json:"proxyConnectTimeoutMs"`
	ProxyReadTimeoutMs    int      `json:"proxyReadTimeoutMs"`
	ProxySendTimeoutMs    int      `json:"proxySendTimeoutMs"`
	RateLimitEnabled      bool     `json:"rateLimitEnabled"`
	RatePerSecond         int      `json:"ratePerSecond"`
	RateLimitBurst        int      `json:"rateLimitBurst"`
	RateLimitNodelay      bool     `json:"rateLimitNodelay"`
	DynamicDNSEnabled     bool     `json:"dynamicDnsEnabled"`
	DynamicDNSHost        *string  `json:"dynamicDnsHost"`
	DynamicDNSPort        *int     `json:"dynamicDnsPort"`
}
type locationPolicyRequest struct {
	IPPolicyEnabled  bool `json:"ipPolicyEnabled"`
	APIPolicyEnabled bool `json:"apiPolicyEnabled"`
}
type locationDirectivesRequest struct {
	Action                string   `json:"action"`
	RootPath              *string  `json:"rootPath"`
	AliasPath             *string  `json:"aliasPath"`
	ReturnStatus          *int     `json:"returnStatus"`
	ReturnBody            *string  `json:"returnBody"`
	ReturnContentTypeMode string   `json:"returnContentTypeMode"`
	ReturnContentType     *string  `json:"returnContentType"`
	ResponseHeaders       []string `json:"responseHeaders"`
}
type dynamicDNSRequest struct {
	Enabled bool    `json:"enabled"`
	Host    *string `json:"host"`
	Port    *int    `json:"port"`
}
type locationTargetView struct {
	ID               string `json:"id"`
	ServerID         string `json:"serverId"`
	Label            string `json:"label"`
	IPPolicyEnabled  bool   `json:"ipPolicyEnabled"`
	APIPolicyEnabled bool   `json:"apiPolicyEnabled"`
}

func (server *Server) listHTTPLocations(w http.ResponseWriter, r *http.Request) {
	center, serverID, ok := serverIDs(w, r)
	if !ok {
		return
	}
	if !serverExists(center, serverID, server.db) {
		writeError(w, 404, "HTTP server not found")
		return
	}
	values, err := server.httpLocations(serverID)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, values)
}
func (server *Server) pageHTTPLocations(w http.ResponseWriter, r *http.Request) {
	center, serverID, ok := serverIDs(w, r)
	if !ok {
		return
	}
	if !serverExists(center, serverID, server.db) {
		writeError(w, 404, "HTTP server not found")
		return
	}
	values, err := server.httpLocations(serverID)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, paginate(r, values))
}
func (server *Server) pageCenterHTTPLocations(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	rows, err := server.db.Query(`SELECT l.id,l.server_id,CONCAT(s.domain,':',s.listen_port,l.path),l.ip_policy_enabled,l.api_policy_enabled FROM http_location l JOIN http_server s ON s.id=l.server_id WHERE s.center_id=? ORDER BY s.domain,s.listen_port,l.path`, center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer rows.Close()
	values := []locationTargetView{}
	for rows.Next() {
		var id, serverID []byte
		var item locationTargetView
		if err := rows.Scan(&id, &serverID, &item.Label, &item.IPPolicyEnabled, &item.APIPolicyEnabled); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		item.ID = uuidText(id)
		item.ServerID = uuidText(serverID)
		values = append(values, item)
	}
	writeJSON(w, 200, paginate(r, values))
}
func (server *Server) httpLocations(serverID uuid.UUID) ([]httpLocationView, error) {
	rows, err := server.db.Query(`SELECT id,path,methods,content_types,header_length_min,header_length_max,body_length_min,body_length_max,upstream_id,proxy_connect_timeout_ms,proxy_read_timeout_ms,proxy_send_timeout_ms,rate_limit_enabled,rate_per_second,rate_limit_burst,rate_limit_nodelay,dynamic_dns_enabled,dynamic_dns_host,dynamic_dns_port,ip_policy_enabled,api_policy_enabled,action,root_path,alias_path,return_status,return_body,return_content_type_mode,return_content_type,response_headers FROM http_location WHERE server_id=? ORDER BY path`, serverID[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []httpLocationView{}
	for rows.Next() {
		var id, upstream, methods, types, headers []byte
		var item httpLocationView
		if err := rows.Scan(&id, &item.Path, &methods, &types, &item.HeaderLengthMin, &item.HeaderLengthMax, &item.BodyLengthMin, &item.BodyLengthMax, &upstream, &item.ProxyConnectTimeoutMs, &item.ProxyReadTimeoutMs, &item.ProxySendTimeoutMs, &item.RateLimitEnabled, &item.RatePerSecond, &item.RateLimitBurst, &item.RateLimitNodelay, &item.DynamicDNSEnabled, &item.DynamicDNSHost, &item.DynamicDNSPort, &item.IPPolicyEnabled, &item.APIPolicyEnabled, &item.Action, &item.RootPath, &item.AliasPath, &item.ReturnStatus, &item.ReturnBody, &item.ReturnContentTypeMode, &item.ReturnContentType, &headers); err != nil {
			return nil, err
		}
		item.ID = uuidText(id)
		if len(upstream) > 0 {
			value := uuidText(upstream)
			item.UpstreamID = &value
		}
		_ = json.Unmarshal(methods, &item.Methods)
		_ = json.Unmarshal(types, &item.ContentTypes)
		_ = json.Unmarshal(headers, &item.ResponseHeaders)
		item.Methods = nonNilStrings(item.Methods)
		item.ContentTypes = nonNilStrings(item.ContentTypes)
		item.ResponseHeaders = nonNilStrings(item.ResponseHeaders)
		values = append(values, item)
	}
	return values, rows.Err()
}
func (server *Server) createHTTPLocation(w http.ResponseWriter, r *http.Request) {
	center, serverID, ok := serverIDs(w, r)
	if !ok {
		return
	}
	if !serverExists(center, serverID, server.db) {
		writeError(w, 404, "HTTP server not found")
		return
	}
	var input httpLocationRequest
	if !validLocation(w, r, &input) {
		return
	}
	upstream, ok := optionalUUID(w, input.UpstreamID)
	if !ok {
		return
	}
	if !validCenterUpstream(server.db, center, upstream) {
		writeError(w, 400, "Upstream does not belong to this center")
		return
	}
	methods, _ := json.Marshal(input.Methods)
	types, _ := json.Marshal(input.ContentTypes)
	id := uuid.New()
	_, err := server.db.Exec(`INSERT INTO http_location (id,server_id,path,methods,content_types,header_length_min,header_length_max,body_length_min,body_length_max,upstream_id,proxy_connect_timeout_ms,proxy_read_timeout_ms,proxy_send_timeout_ms,rate_limit_enabled,rate_per_second,rate_limit_burst,rate_limit_nodelay,dynamic_dns_enabled,dynamic_dns_host,dynamic_dns_port,ip_policy_enabled,api_policy_enabled,action,return_content_type_mode,response_headers,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,false,false,'PROXY','CUSTOM',JSON_ARRAY(),?)`, id[:], serverID[:], strings.TrimSpace(input.Path), methods, types, input.HeaderLengthMin, input.HeaderLengthMax, input.BodyLengthMin, input.BodyLengthMax, uuidBytes(upstream), input.ProxyConnectTimeoutMs, input.ProxyReadTimeoutMs, input.ProxySendTimeoutMs, input.RateLimitEnabled, input.RatePerSecond, input.RateLimitBurst, input.RateLimitNodelay, input.DynamicDNSEnabled, nullableString(input.DynamicDNSHost), input.DynamicDNSPort, time.Now())
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	server.audit(center, "HTTP_LOCATION_CREATED", "HTTP_LOCATION", id)
	item, _ := server.findHTTPLocation(serverID, id)
	writeJSON(w, 201, item)
}
func (server *Server) updateHTTPLocation(w http.ResponseWriter, r *http.Request) {
	center, serverID, locationID, ok := serverLocationIDs(w, r)
	if !ok {
		return
	}
	if !serverExists(center, serverID, server.db) {
		writeError(w, 404, "HTTP server not found")
		return
	}
	var input httpLocationRequest
	if !validLocation(w, r, &input) {
		return
	}
	upstream, ok := optionalUUID(w, input.UpstreamID)
	if !ok {
		return
	}
	if !validCenterUpstream(server.db, center, upstream) {
		writeError(w, 400, "Upstream does not belong to this center")
		return
	}
	methods, _ := json.Marshal(input.Methods)
	types, _ := json.Marshal(input.ContentTypes)
	result, err := server.db.Exec(`UPDATE http_location SET path=?,methods=?,content_types=?,header_length_min=?,header_length_max=?,body_length_min=?,body_length_max=?,upstream_id=?,proxy_connect_timeout_ms=?,proxy_read_timeout_ms=?,proxy_send_timeout_ms=?,rate_limit_enabled=?,rate_per_second=?,rate_limit_burst=?,rate_limit_nodelay=?,dynamic_dns_enabled=?,dynamic_dns_host=?,dynamic_dns_port=? WHERE id=? AND server_id=?`, strings.TrimSpace(input.Path), methods, types, input.HeaderLengthMin, input.HeaderLengthMax, input.BodyLengthMin, input.BodyLengthMax, uuidBytes(upstream), input.ProxyConnectTimeoutMs, input.ProxyReadTimeoutMs, input.ProxySendTimeoutMs, input.RateLimitEnabled, input.RatePerSecond, input.RateLimitBurst, input.RateLimitNodelay, input.DynamicDNSEnabled, nullableString(input.DynamicDNSHost), input.DynamicDNSPort, locationID[:], serverID[:])
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "HTTP location not found")
		return
	}
	server.audit(center, "HTTP_LOCATION_UPDATED", "HTTP_LOCATION", locationID)
	item, _ := server.findHTTPLocation(serverID, locationID)
	writeJSON(w, 200, item)
}
func (server *Server) deleteHTTPLocation(w http.ResponseWriter, r *http.Request) {
	center, serverID, locationID, ok := serverLocationIDs(w, r)
	if !ok {
		return
	}
	if !serverExists(center, serverID, server.db) {
		writeError(w, 404, "HTTP server not found")
		return
	}
	result, err := server.db.Exec("DELETE FROM http_location WHERE id=? AND server_id=?", locationID[:], serverID[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "HTTP location not found")
		return
	}
	server.audit(center, "HTTP_LOCATION_DELETED", "HTTP_LOCATION", locationID)
	w.WriteHeader(204)
}
func (server *Server) putHTTPLocationPolicySettings(w http.ResponseWriter, r *http.Request) {
	center, serverID, locationID, ok := serverLocationIDs(w, r)
	if !ok {
		return
	}
	var input locationPolicyRequest
	if !decode(r, &input) {
		writeError(w, 400, "invalid policy settings")
		return
	}
	result, err := server.db.Exec("UPDATE http_location SET ip_policy_enabled=?,api_policy_enabled=? WHERE id=? AND server_id=?", input.IPPolicyEnabled, input.APIPolicyEnabled, locationID[:], serverID[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "HTTP location not found")
		return
	}
	server.audit(center, "HTTP_LOCATION_POLICY_SETTINGS_UPDATED", "HTTP_LOCATION", locationID)
	item, _ := server.findHTTPLocation(serverID, locationID)
	writeJSON(w, 200, item)
}
func (server *Server) putHTTPLocationDirectives(w http.ResponseWriter, r *http.Request) {
	center, serverID, locationID, ok := serverLocationIDs(w, r)
	if !ok {
		return
	}
	var input locationDirectivesRequest
	if !decode(r, &input) || !validLocationDirectives(input) {
		writeError(w, 400, "invalid location directives")
		return
	}
	current, err := server.findHTTPLocation(serverID, locationID)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if current.ID == "" {
		writeError(w, 404, "HTTP location not found")
		return
	}
	if input.Action == "PROXY" && current.UpstreamID == nil {
		writeError(w, 400, "代理模式需要配置 Upstream")
		return
	}
	headers, _ := json.Marshal(nonNilStrings(input.ResponseHeaders))
	_, err = server.db.Exec("UPDATE http_location SET action=?,root_path=?,alias_path=?,return_status=?,return_body=?,return_content_type_mode=?,return_content_type=?,response_headers=? WHERE id=? AND server_id=?", input.Action, nullableString(input.RootPath), nullableString(input.AliasPath), input.ReturnStatus, nullableString(input.ReturnBody), input.ReturnContentTypeMode, nullableString(input.ReturnContentType), headers, locationID[:], serverID[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	server.audit(center, "HTTP_LOCATION_DIRECTIVES_UPDATED", "HTTP_LOCATION", locationID)
	item, _ := server.findHTTPLocation(serverID, locationID)
	writeJSON(w, 200, item)
}
func (server *Server) putHTTPLocationDynamicDNS(w http.ResponseWriter, r *http.Request) {
	center, serverID, locationID, ok := serverLocationIDs(w, r)
	if !ok {
		return
	}
	var input dynamicDNSRequest
	if !decode(r, &input) || !validDNS(input) {
		writeError(w, 400, "启用动态 DNS 时必须提供域名和端口")
		return
	}
	result, err := server.db.Exec("UPDATE http_location SET dynamic_dns_enabled=?,dynamic_dns_host=?,dynamic_dns_port=? WHERE id=? AND server_id=?", input.Enabled, nullableString(input.Host), input.Port, locationID[:], serverID[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "HTTP location not found")
		return
	}
	server.audit(center, "HTTP_LOCATION_DYNAMIC_DNS_UPDATED", "HTTP_LOCATION", locationID)
	item, _ := server.findHTTPLocation(serverID, locationID)
	writeJSON(w, 200, item)
}
func (server *Server) findHTTPLocation(serverID, id uuid.UUID) (httpLocationView, error) {
	values, err := server.httpLocations(serverID)
	if err != nil {
		return httpLocationView{}, err
	}
	for _, value := range values {
		if value.ID == id.String() {
			return value, nil
		}
	}
	return httpLocationView{}, nil
}
func serverLocationIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	center, serverID, ok := serverIDs(w, r)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	locationID, ok := pathUUID(w, r, "locationID")
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return center, serverID, locationID, true
}
func validCenterUpstream(db interface{ QueryRow(string, ...any) *sql.Row }, center uuid.UUID, upstream *uuid.UUID) bool {
	return upstream == nil || db.QueryRow("SELECT 1 FROM http_upstream WHERE id=? AND center_id=?", (*upstream)[:], center[:]).Scan(new(int)) == nil
}
func validLocation(w http.ResponseWriter, r *http.Request, input *httpLocationRequest) bool {
	if !decode(r, input) || strings.TrimSpace(input.Path) == "" || len(input.Methods) == 0 || input.HeaderLengthMin < 0 || input.HeaderLengthMax < input.HeaderLengthMin || input.BodyLengthMin < 0 || input.BodyLengthMax < input.BodyLengthMin || input.ProxyConnectTimeoutMs < 1 || input.ProxyReadTimeoutMs < 1 || input.ProxySendTimeoutMs < 1 || input.RatePerSecond < 1 || input.RatePerSecond > 100000 || input.RateLimitBurst < 0 || input.RateLimitBurst > 100000 || !validDNS(dynamicDNSRequest{input.DynamicDNSEnabled, input.DynamicDNSHost, input.DynamicDNSPort}) {
		writeError(w, 400, "invalid HTTP location")
		return false
	}
	return true
}
func validLocationDirectives(input locationDirectivesRequest) bool {
	if input.Action != "PROXY" && input.Action != "STATIC" && input.Action != "RETURN" || !validPath(input.RootPath) || !validPath(input.AliasPath) || !validHeaders(input.ResponseHeaders) || input.ReturnContentTypeMode != "CUSTOM" && input.ReturnContentTypeMode != "TEXT" && input.ReturnContentTypeMode != "JSON" {
		return false
	}
	if input.RootPath != nil && strings.TrimSpace(*input.RootPath) != "" && input.AliasPath != nil && strings.TrimSpace(*input.AliasPath) != "" {
		return false
	}
	if input.Action == "STATIC" && (input.RootPath == nil || strings.TrimSpace(*input.RootPath) == "") && (input.AliasPath == nil || strings.TrimSpace(*input.AliasPath) == "") {
		return false
	}
	return input.Action != "RETURN" || (input.ReturnStatus != nil && *input.ReturnStatus >= 100 && *input.ReturnStatus <= 599)
}
func validDNS(input dynamicDNSRequest) bool {
	if !input.Enabled {
		return true
	}
	return input.Host != nil && strings.TrimSpace(*input.Host) != "" && input.Port != nil && *input.Port >= 1 && *input.Port <= 65535 && isDNSHost(*input.Host)
}
func isDNSHost(value string) bool {
	if len(value) > 253 {
		return false
	}
	for _, char := range value {
		if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '.' || char == '-') {
			return false
		}
	}
	return true
}
