package httpapi

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type tlsCertificateView struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CommonName string    `json:"commonName"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
type createTLSCertificateRequest struct {
	Name           string  `json:"name"`
	CertificatePEM string  `json:"certificatePem"`
	PrivateKeyPEM  string  `json:"privateKeyPem"`
	ChainPEM       *string `json:"chainPem"`
	Enabled        *bool   `json:"enabled"`
}
type updateTLSCertificateRequest struct {
	Name           string  `json:"name"`
	CertificatePEM *string `json:"certificatePem"`
	PrivateKeyPEM  *string `json:"privateKeyPem"`
	ChainPEM       *string `json:"chainPem"`
	Enabled        *bool   `json:"enabled"`
}

func (server *Server) listTLSCertificates(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	rows, err := server.db.Query("SELECT id,name,common_name,enabled,created_at,updated_at FROM tls_certificate WHERE center_id=? ORDER BY name", center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer rows.Close()
	values := []tlsCertificateView{}
	for rows.Next() {
		var raw []byte
		var item tlsCertificateView
		if err := rows.Scan(&raw, &item.Name, &item.CommonName, &item.Enabled, &item.CreatedAt, &item.UpdatedAt); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		item.ID = uuidText(raw)
		values = append(values, item)
	}
	writeJSON(w, 200, values)
}
func (server *Server) createTLSCertificate(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	var input createTLSCertificateRequest
	if !decode(r, &input) || strings.TrimSpace(input.Name) == "" || len(strings.TrimSpace(input.Name)) > 128 {
		writeError(w, 400, "certificate name is required")
		return
	}
	commonName, ok := certificateCommonName(input.CertificatePEM, input.PrivateKeyPEM, input.ChainPEM)
	if !ok {
		writeError(w, 422, "证书或私钥 PEM 内容无效")
		return
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	id := uuid.New()
	now := time.Now()
	_, err := server.db.Exec("INSERT INTO tls_certificate (id,center_id,name,common_name,certificate_pem,private_key_pem,chain_pem,enabled,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)", id[:], center[:], strings.TrimSpace(input.Name), commonName, strings.TrimSpace(input.CertificatePEM), strings.TrimSpace(input.PrivateKeyPEM), nullableString(input.ChainPEM), enabled, now, now)
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	server.audit(center, "TLS_CERTIFICATE_CREATED", "TLS_CERTIFICATE", id)
	writeJSON(w, 201, tlsCertificateView{ID: id.String(), Name: strings.TrimSpace(input.Name), CommonName: commonName, Enabled: enabled, CreatedAt: now, UpdatedAt: now})
}
func (server *Server) updateTLSCertificate(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "certificateID")
	if !ok {
		return
	}
	var input updateTLSCertificateRequest
	if !decode(r, &input) || strings.TrimSpace(input.Name) == "" || len(strings.TrimSpace(input.Name)) > 128 {
		writeError(w, 400, "certificate name is required")
		return
	}
	if (input.CertificatePEM == nil) != (input.PrivateKeyPEM == nil) {
		writeError(w, 400, "更新证书内容时必须同时提供证书和私钥")
		return
	}
	var current tlsCertificateView
	var certificate, privateKey, chain string
	err := server.db.QueryRow("SELECT name,common_name,certificate_pem,private_key_pem,COALESCE(chain_pem,''),enabled,created_at,updated_at FROM tls_certificate WHERE id=? AND center_id=?", id[:], center[:]).Scan(&current.Name, &current.CommonName, &certificate, &privateKey, &chain, &current.Enabled, &current.CreatedAt, &current.UpdatedAt)
	if err != nil {
		writeError(w, 404, "TLS certificate not found")
		return
	}
	if input.CertificatePEM != nil {
		commonName, valid := certificateCommonName(*input.CertificatePEM, *input.PrivateKeyPEM, input.ChainPEM)
		if !valid {
			writeError(w, 422, "证书或私钥 PEM 内容无效")
			return
		}
		current.CommonName = commonName
		certificate = strings.TrimSpace(*input.CertificatePEM)
		privateKey = strings.TrimSpace(*input.PrivateKeyPEM)
	}
	if input.ChainPEM != nil {
		chain = strings.TrimSpace(*input.ChainPEM)
		if chain != "" && !strings.Contains(chain, "BEGIN CERTIFICATE") {
			writeError(w, 422, "证书链 PEM 内容无效")
			return
		}
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	now := time.Now()
	_, err = server.db.Exec("UPDATE tls_certificate SET name=?,common_name=?,certificate_pem=?,private_key_pem=?,chain_pem=?,enabled=?,updated_at=? WHERE id=? AND center_id=?", strings.TrimSpace(input.Name), current.CommonName, certificate, privateKey, nullIfBlank(chain), current.Enabled, now, id[:], center[:])
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	server.audit(center, "TLS_CERTIFICATE_UPDATED", "TLS_CERTIFICATE", id)
	current.ID = id.String()
	current.Name = strings.TrimSpace(input.Name)
	current.UpdatedAt = now
	writeJSON(w, 200, current)
}
func (server *Server) deleteTLSCertificate(w http.ResponseWriter, r *http.Request) {
	center, ok := pathUUID(w, r, "centerID")
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "certificateID")
	if !ok {
		return
	}
	result, err := server.db.Exec("DELETE FROM tls_certificate WHERE id=? AND center_id=?", id[:], center[:])
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		writeError(w, 404, "TLS certificate not found")
		return
	}
	server.audit(center, "TLS_CERTIFICATE_DELETED", "TLS_CERTIFICATE", id)
	w.WriteHeader(204)
}
func certificateCommonName(certificate, privateKey string, chain *string) (string, bool) {
	if !strings.Contains(certificate, "BEGIN CERTIFICATE") || !strings.Contains(privateKey, "PRIVATE KEY") {
		return "", false
	}
	if chain != nil && strings.TrimSpace(*chain) != "" && !strings.Contains(*chain, "BEGIN CERTIFICATE") {
		return "", false
	}
	block, _ := pem.Decode([]byte(certificate))
	if block == nil || block.Type != "CERTIFICATE" {
		return "", false
	}
	parsed, err := x509.ParseCertificate(block.Bytes)
	if err != nil || strings.TrimSpace(parsed.Subject.CommonName) == "" {
		return "", false
	}
	return parsed.Subject.CommonName, true
}
func nullIfBlank(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}
