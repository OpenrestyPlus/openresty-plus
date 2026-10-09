package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type nodeView struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Host          string  `json:"host"`
	ServicePort   int     `json:"servicePort"`
	ControlAPIURL *string `json:"controlApiUrl"`
	Enabled       bool    `json:"enabled"`
}
type nodeRequest struct {
	Name          string  `json:"name"`
	Host          string  `json:"host"`
	ServicePort   int     `json:"servicePort"`
	ControlAPIURL *string `json:"controlApiUrl"`
	Enabled       bool    `json:"enabled"`
}

func (server *Server) listNodes(writer http.ResponseWriter, request *http.Request) {
	center, ok := pathUUID(writer, request, "centerID")
	if !ok {
		return
	}
	values, err := server.nodes(center)
	if err != nil {
		writeError(writer, 500, err.Error())
		return
	}
	writeJSON(writer, 200, values)
}
func (server *Server) pageNodes(writer http.ResponseWriter, request *http.Request) {
	center, ok := pathUUID(writer, request, "centerID")
	if !ok {
		return
	}
	values, err := server.nodes(center)
	if err != nil {
		writeError(writer, 500, err.Error())
		return
	}
	writeJSON(writer, 200, paginate(request, values))
}
func (server *Server) nodes(center uuid.UUID) ([]nodeView, error) {
	rows, err := server.db.Query("SELECT id,name,host,service_port,control_api_url,enabled FROM nginx_node WHERE center_id=? ORDER BY name", center[:])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []nodeView{}
	for rows.Next() {
		var id []byte
		var value nodeView
		if err := rows.Scan(&id, &value.Name, &value.Host, &value.ServicePort, &value.ControlAPIURL, &value.Enabled); err != nil {
			return nil, err
		}
		value.ID = uuidText(id)
		values = append(values, value)
	}
	return values, rows.Err()
}
func (server *Server) createNode(writer http.ResponseWriter, request *http.Request) {
	center, ok := pathUUID(writer, request, "centerID")
	if !ok {
		return
	}
	var input nodeRequest
	if !validNode(writer, request, &input) {
		return
	}
	id := uuid.New()
	_, err := server.db.Exec("INSERT INTO nginx_node (id,center_id,name,protocol,host,service_port,control_api_url,enabled,created_at) VALUES (?,?,?,NULL,?,?,?,true,?)", id[:], center[:], strings.TrimSpace(input.Name), strings.TrimSpace(input.Host), input.ServicePort, input.ControlAPIURL, time.Now())
	if err != nil {
		writeError(writer, 409, err.Error())
		return
	}
	server.audit(center, "NGINX_NODE_CREATED", "NGINX_NODE", id)
	writeJSON(writer, 201, nodeView{ID: id.String(), Name: strings.TrimSpace(input.Name), Host: strings.TrimSpace(input.Host), ServicePort: input.ServicePort, ControlAPIURL: input.ControlAPIURL, Enabled: true})
}
func (server *Server) updateNode(writer http.ResponseWriter, request *http.Request) {
	center, ok := pathUUID(writer, request, "centerID")
	if !ok {
		return
	}
	node, ok := pathUUID(writer, request, "nodeID")
	if !ok {
		return
	}
	var input nodeRequest
	if !validNode(writer, request, &input) {
		return
	}
	result, err := server.db.Exec("UPDATE nginx_node SET name=?,host=?,service_port=?,control_api_url=?,enabled=? WHERE id=? AND center_id=?", strings.TrimSpace(input.Name), strings.TrimSpace(input.Host), input.ServicePort, input.ControlAPIURL, input.Enabled, node[:], center[:])
	if err != nil {
		writeError(writer, 500, err.Error())
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		writeError(writer, 404, "Nginx node not found")
		return
	}
	server.audit(center, "NGINX_NODE_UPDATED", "NGINX_NODE", node)
	writeJSON(writer, 200, nodeView{ID: node.String(), Name: strings.TrimSpace(input.Name), Host: strings.TrimSpace(input.Host), ServicePort: input.ServicePort, ControlAPIURL: input.ControlAPIURL, Enabled: input.Enabled})
}
func (server *Server) deleteNode(writer http.ResponseWriter, request *http.Request) {
	center, ok := pathUUID(writer, request, "centerID")
	if !ok {
		return
	}
	node, ok := pathUUID(writer, request, "nodeID")
	if !ok {
		return
	}
	result, err := server.db.Exec("DELETE FROM nginx_node WHERE id=? AND center_id=?", node[:], center[:])
	if err != nil {
		writeError(writer, 500, err.Error())
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		writeError(writer, 404, "Nginx node not found")
		return
	}
	server.audit(center, "NGINX_NODE_DELETED", "NGINX_NODE", node)
	writer.WriteHeader(http.StatusNoContent)
}
func validNode(writer http.ResponseWriter, request *http.Request, input *nodeRequest) bool {
	if !decode(request, input) || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Host) == "" || input.ServicePort < 1 || input.ServicePort > 65535 {
		writeError(writer, 400, "name, host and servicePort are required")
		return false
	}
	return true
}
