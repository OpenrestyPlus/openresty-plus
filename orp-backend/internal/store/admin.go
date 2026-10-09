package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// SeedAdmin creates the first local administrator after schema migration.
// Existing credentials are never reset by process restarts.
func SeedAdmin(db *sql.DB) error {
	username := strings.TrimSpace(os.Getenv("OPENRESTY_ADMIN_USERNAME"))
	if username == "" {
		username = "vben"
	}
	password := os.Getenv("OPENRESTY_ADMIN_PASSWORD")
	if password == "" {
		return fmt.Errorf("OPENRESTY_ADMIN_PASSWORD is required")
	}
	if _, err := db.Exec("INSERT IGNORE INTO orp_role(code,name,status,builtin,permission_keys) VALUES('super','超级管理员','enabled',true,JSON_ARRAY('*'))"); err != nil {
		return err
	}
	var roleID int64
	if err := db.QueryRow("SELECT id FROM orp_role WHERE code='super'").Scan(&roleID); err != nil {
		return err
	}
	var exists bool
	if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM orp_account WHERE username=?)", username).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	roles, _ := json.Marshal([]int64{roleID})
	_, err = db.Exec("INSERT INTO orp_account(username,password_hash,real_name,status,role_ids) VALUES(?,?,?,'enabled',?)", username, string(hash), username, roles)
	return err
}
