package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// ImportLegacyCenters exposes existing Center records in the numeric-ID
// console model. The legacy UUID is retained for synchronized updates.
func ImportLegacyCenters(db *sql.DB) error {
	rows, err := db.Query("SELECT id,code,name FROM center ORDER BY code")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var rawID []byte
		var code, name string
		if err := rows.Scan(&rawID, &code, &name); err != nil {
			return err
		}
		legacyID, err := uuid.FromBytes(rawID)
		if err != nil {
			return fmt.Errorf("legacy Center ID: %w", err)
		}
		var exists bool
		if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM orp_resource WHERE kind='centers' AND JSON_UNQUOTE(JSON_EXTRACT(document,'$.legacyId'))=?)", legacyID.String()).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		document, _ := json.Marshal(map[string]any{"code": code, "name": name, "description": "", "legacyId": legacyID.String()})
		if _, err := db.Exec("INSERT INTO orp_resource(kind,document) VALUES('centers',?)", document); err != nil {
			return err
		}
	}
	return rows.Err()
}
