package httpapi

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// auditEventView intentionally excludes detail, which can contain operational metadata.
type auditEventView struct {
	ID           string    `json:"id"`
	Actor        string    `json:"actor"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceId"`
	Result       string    `json:"result"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (server *Server) listAuditEvents(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.auditEvents(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, values)
}
func (server *Server) pageAuditEvents(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	values, err := server.auditEvents(center)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, paginate(r, values))
}
func (server *Server) auditEvents(center uuid.UUID) ([]auditEventView, error) {
	rows, err := server.db.Query("SELECT id,actor,action,resource_type,resource_id,result,created_at FROM audit_event WHERE center_id=? OR actor=? ORDER BY created_at DESC", center[:], center.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []auditEventView{}
	for rows.Next() {
		var raw []byte
		var item auditEventView
		if err := rows.Scan(&raw, &item.Actor, &item.Action, &item.ResourceType, &item.ResourceID, &item.Result, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.ID = uuidText(raw)
		values = append(values, item)
	}
	return values, rows.Err()
}
