package httpapi

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func (server *Server) audit(centerID uuid.UUID, action, resourceType string, resourceID uuid.UUID) {
	if server.db == nil {
		return
	}
	detail, _ := json.Marshal(map[string]string{"centerId": centerID.String(), "resourceId": resourceID.String()})
	id := uuid.New()
	_, _ = server.db.Exec("INSERT INTO audit_event (id,center_id,actor,action,resource_type,resource_id,result,detail,created_at) VALUES (?,?,?,?,?,?,?,?,?)", id[:], centerID[:], "admin", action, resourceType, resourceID.String(), "SUCCESS", detail, time.Now())
}
