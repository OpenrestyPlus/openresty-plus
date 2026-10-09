package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type dnsResolverView struct {
	ID                  string   `json:"id"`
	Scope               string   `json:"scope"`
	TargetResourceID    *string  `json:"targetResourceId"`
	ResolverAddresses   []string `json:"resolverAddresses"`
	ValidSeconds        int      `json:"validSeconds"`
	TimeoutMilliseconds int      `json:"timeoutMilliseconds"`
	IPv6Enabled         bool     `json:"ipv6Enabled"`
	Enabled             bool     `json:"enabled"`
}
type dnsResolverRequest struct {
	Scope               string   `json:"scope"`
	TargetResourceID    *string  `json:"targetResourceId"`
	ResolverAddresses   []string `json:"resolverAddresses"`
	ValidSeconds        int      `json:"validSeconds"`
	TimeoutMilliseconds int      `json:"timeoutMilliseconds"`
	IPv6Enabled         bool     `json:"ipv6Enabled"`
	Enabled             bool     `json:"enabled"`
}

func (server *Server) listDNSResolvers(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.dnsResolvers(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, values)
}
func (server *Server) pageDNSResolvers(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.dnsResolvers(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, paginate(r, values))
}
func (server *Server) dnsResolvers(center uuid.UUID) ([]dnsResolverView, error) {
	rows, err := server.db.Query("SELECT id,scope,target_resource_id,resolver_addresses,valid_seconds,timeout_milliseconds,ipv6_enabled,enabled FROM dns_resolver_configuration WHERE center_id=? ORDER BY scope", center[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []dnsResolverView{}
	for rows.Next() {
		var raw, target, addresses []byte
		var item dnsResolverView
		if err := rows.Scan(&raw, &item.Scope, &target, &addresses, &item.ValidSeconds, &item.TimeoutMilliseconds, &item.IPv6Enabled, &item.Enabled); err != nil {
			return nil, err
		}
		item.ID = uuidText(raw)
		if len(target) > 0 {
			value := uuidText(target)
			item.TargetResourceID = &value
		}
		_ = json.Unmarshal(addresses, &item.ResolverAddresses)
		item.ResolverAddresses = nonNilStrings(item.ResolverAddresses)
		values = append(values, item)
	}
	return values, rows.Err()
}
func (server *Server) createDNSResolver(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	var input dnsResolverRequest
	if !validDNSResolver(w, r, &input) {
		return
	}
	target, ok := optionalUUID(w, input.TargetResourceID)
	if !ok {
		return
	}
	if !server.validResolverTarget(center, input.Scope, target) {
		writeError(w, 400, "Resolver 目标不属于当前中心")
		return
	}
	addresses, _ := json.Marshal(input.ResolverAddresses)
	id := uuid.New()
	_, err := server.db.Exec("INSERT INTO dns_resolver_configuration (id,center_id,scope,target_resource_id,resolver_addresses,valid_seconds,timeout_milliseconds,ipv6_enabled,enabled) VALUES (?,?,?,?,?,?,?,?,?)", id[:], center[:], input.Scope, uuidBytes(target), addresses, input.ValidSeconds, input.TimeoutMilliseconds, input.IPv6Enabled, input.Enabled)
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	server.audit(center, "DNS_RESOLVER_CREATED", "DNS_RESOLVER", id)
	item, _ := server.findDNSResolver(center, id)
	writeJSON(w, 201, item)
}
func (server *Server) updateDNSResolver(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "resolverID")
	if !ok {
		return
	}
	var input dnsResolverRequest
	if !validDNSResolver(w, r, &input) {
		return
	}
	target, ok := optionalUUID(w, input.TargetResourceID)
	if !ok {
		return
	}
	if !server.validResolverTarget(center, input.Scope, target) {
		writeError(w, 400, "Resolver 目标不属于当前中心")
		return
	}
	addresses, _ := json.Marshal(input.ResolverAddresses)
	result, err := server.db.Exec("UPDATE dns_resolver_configuration SET scope=?,target_resource_id=?,resolver_addresses=?,valid_seconds=?,timeout_milliseconds=?,ipv6_enabled=?,enabled=? WHERE id=? AND center_id=?", input.Scope, uuidBytes(target), addresses, input.ValidSeconds, input.TimeoutMilliseconds, input.IPv6Enabled, input.Enabled, id[:], center[:])
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "DNS resolver not found")
		return
	}
	server.audit(center, "DNS_RESOLVER_UPDATED", "DNS_RESOLVER", id)
	item, _ := server.findDNSResolver(center, id)
	writeJSON(w, 200, item)
}
func (server *Server) deleteDNSResolver(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "resolverID")
	if !ok {
		return
	}
	result, err := server.db.Exec("DELETE FROM dns_resolver_configuration WHERE id=? AND center_id=?", id[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "DNS resolver not found")
		return
	}
	server.audit(center, "DNS_RESOLVER_DELETED", "DNS_RESOLVER", id)
	w.WriteHeader(204)
}
func (server *Server) findDNSResolver(center, id uuid.UUID) (dnsResolverView, error) {
	values, err := server.dnsResolvers(center)
	if err != nil {
		return dnsResolverView{}, err
	}
	for _, value := range values {
		if value.ID == id.String() {
			return value, nil
		}
	}
	return dnsResolverView{}, nil
}
func validDNSResolver(w http.ResponseWriter, r *http.Request, input *dnsResolverRequest) bool {
	if !decode(r, input) || (input.Scope != "HTTP" && input.Scope != "STREAM" && input.Scope != "HTTP_SERVER" && input.Scope != "HTTP_LOCATION" && input.Scope != "STREAM_SERVER") || len(input.ResolverAddresses) == 0 || input.ValidSeconds < 1 || input.ValidSeconds > 3600 || input.TimeoutMilliseconds < 100 || input.TimeoutMilliseconds > 60000 {
		writeError(w, 400, "invalid DNS resolver")
		return false
	}
	global := input.Scope == "HTTP" || input.Scope == "STREAM"
	if global == (input.TargetResourceID != nil && strings.TrimSpace(*input.TargetResourceID) != "") {
		writeError(w, 400, "全局作用域不应绑定资源，Server 和 Location 作用域必须绑定资源")
		return false
	}
	for _, address := range input.ResolverAddresses {
		if !validResolverAddress(address) {
			writeError(w, 400, "invalid resolver address")
			return false
		}
	}
	return true
}
func validResolverAddress(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if !(char >= '0' && char <= '9' || char >= 'a' && char <= 'f' || char >= 'A' && char <= 'F' || char == '.' || char == ':' || char == '[' || char == ']') {
			return false
		}
	}
	return true
}
func (server *Server) validResolverTarget(center uuid.UUID, scope string, target *uuid.UUID) bool {
	if scope == "HTTP" || scope == "STREAM" {
		return target == nil
	}
	if target == nil {
		return false
	}
	var found int
	switch scope {
	case "HTTP_SERVER":
		return server.db.QueryRow("SELECT 1 FROM http_server WHERE id=? AND center_id=?", (*target)[:], center[:]).Scan(&found) == nil
	case "HTTP_LOCATION":
		return server.db.QueryRow("SELECT 1 FROM http_location l JOIN http_server s ON s.id=l.server_id WHERE l.id=? AND s.center_id=?", (*target)[:], center[:]).Scan(&found) == nil
	case "STREAM_SERVER":
		return server.db.QueryRow("SELECT 1 FROM stream_server WHERE id=? AND center_id=?", (*target)[:], center[:]).Scan(&found) == nil
	}
	return false
}
