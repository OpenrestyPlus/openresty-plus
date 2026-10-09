package store

import (
	"database/sql"
	"encoding/json"
)

func SeedSettingGroups(db *sql.DB) error {
	groups := []struct{ Code, Name string }{
		{"system", "系统基础"},
		{"business", "业务参数"},
		{"api-key", "接口密钥"},
		{"security", "安全策略"},
	}
	for index, group := range groups {
		var exists bool
		if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM orp_resource WHERE kind='setting-groups' AND JSON_UNQUOTE(JSON_EXTRACT(document,'$.code'))=?)", group.Code).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		document, _ := json.Marshal(map[string]any{"code": group.Code, "name": group.Name, "description": "", "isSystem": true, "sortOrder": index + 1})
		if _, err := db.Exec("INSERT INTO orp_resource(kind,document) VALUES('setting-groups',?)", document); err != nil {
			return err
		}
	}
	return nil
}
