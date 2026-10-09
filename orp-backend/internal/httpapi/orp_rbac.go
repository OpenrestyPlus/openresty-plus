package httpapi

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) requireSuper(w http.ResponseWriter, r *http.Request) bool {
	username, ok := s.orpIdentity(w, r)
	if !ok {
		return false
	}
	roles, err := s.accountRoles(r, username)
	if err != nil {
		orpFailure(w, 500, "读取权限失败")
		return false
	}
	for _, role := range roles {
		for _, permission := range role.Permissions {
			if permission == "*" {
				return true
			}
		}
	}
	orpFailure(w, 403, "仅超级管理员可操作")
	return false
}

func orpPage(r *http.Request, items []map[string]any) map[string]any {
	pageNo, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}
	start := (pageNo - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return map[string]any{"items": items[start:end], "total": len(items)}
}

func (s *Server) rbacUsersList(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	rows, err := s.db.QueryContext(r.Context(), "SELECT id,username,real_name,email,phone,status,role_ids,created_at,last_login_at FROM orp_account ORDER BY id DESC")
	if err != nil {
		orpFailure(w, 500, "读取用户失败")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	keyword := strings.ToLower(r.URL.Query().Get("keyword"))
	status := r.URL.Query().Get("status")
	roleFilter := r.URL.Query().Get("roleId")
	for rows.Next() {
		var id int64
		var username, realName, state string
		var email, phone sql.NullString
		var rolesJSON []byte
		var created time.Time
		var lastLogin sql.NullTime
		if err := rows.Scan(&id, &username, &realName, &email, &phone, &state, &rolesJSON, &created, &lastLogin); err != nil {
			orpFailure(w, 500, "读取用户失败")
			return
		}
		if status != "" && status != state {
			continue
		}
		if keyword != "" && !strings.Contains(strings.ToLower(username+" "+realName), keyword) {
			continue
		}
		var roleIDs []int64
		if err := json.Unmarshal(rolesJSON, &roleIDs); err != nil {
			orpFailure(w, 500, "用户角色数据损坏")
			return
		}
		if !matchesRoleFilter(roleIDs, roleFilter) {
			continue
		}
		roleNames := []string{}
		for _, roleID := range roleIDs {
			var name string
			if s.db.QueryRowContext(r.Context(), "SELECT name FROM orp_role WHERE id=?", roleID).Scan(&name) == nil {
				roleNames = append(roleNames, name)
			}
		}
		var last any
		if lastLogin.Valid {
			last = lastLogin.Time.Format("2006-01-02 15:04:05")
		}
		items = append(items, map[string]any{"id": id, "username": username, "realName": realName, "email": orpNullableString(email), "phone": orpNullableString(phone), "status": state, "roleIds": roleIDs, "roleNames": roleNames, "createdAt": created.Format("2006-01-02 15:04:05"), "lastLoginAt": last})
	}
	if err := rows.Err(); err != nil {
		orpFailure(w, 500, "读取用户失败")
		return
	}
	orpReply(w, 200, orpPage(r, items))
}

func matchesRoleFilter(roleIDs []int64, roleFilter string) bool {
	if roleFilter == "" || roleFilter == "all" {
		return true
	}
	for _, roleID := range roleIDs {
		if strconv.FormatInt(roleID, 10) == roleFilter {
			return true
		}
	}
	return false
}

func orpNullableString(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func (s *Server) rbacUserCreate(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	body, ok := orpBody(w, r)
	if !ok {
		return
	}
	username, password, realName := strings.TrimSpace(toString(body["username"])), toString(body["password"]), strings.TrimSpace(toString(body["realName"]))
	if username == "" || len(password) < 12 || realName == "" {
		orpFailure(w, 400, "用户名、姓名不能为空，密码至少 12 位")
		return
	}
	roleIDs, err := parseRoleIDs(body["roleIds"])
	if err != nil {
		orpFailure(w, 400, err.Error())
		return
	}
	if err := s.validateRoleIDs(r, roleIDs); err != nil {
		orpFailure(w, 400, err.Error())
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		orpFailure(w, 500, "加密密码失败")
		return
	}
	rawRoles, _ := json.Marshal(roleIDs)
	result, err := s.db.ExecContext(r.Context(), "INSERT INTO orp_account(username,password_hash,real_name,email,phone,status,role_ids) VALUES(?,?,?,?,?,?,?)", username, string(hash), realName, nullableInput(body["email"]), nullableInput(body["phone"]), statusInput(body["status"]), rawRoles)
	if err != nil {
		orpFailure(w, 409, "用户名已存在或保存失败")
		return
	}
	id, _ := result.LastInsertId()
	orpReply(w, 201, map[string]any{"id": id, "username": username, "realName": realName, "roleIds": roleIDs, "status": statusInput(body["status"])})
}

func parseRoleIDs(value any) ([]int64, error) {
	list, ok := value.([]any)
	if !ok {
		return []int64{}, nil
	}
	ids := []int64{}
	for _, raw := range list {
		number, ok := raw.(float64)
		if !ok || number < 1 || number != float64(int64(number)) {
			return nil, errors.New("角色 ID 无效")
		}
		ids = append(ids, int64(number))
	}
	return ids, nil
}

func (s *Server) validateRoleIDs(r *http.Request, ids []int64) error {
	for _, id := range ids {
		var enabled bool
		if err := s.db.QueryRowContext(r.Context(), "SELECT status='enabled' FROM orp_role WHERE id=?", id).Scan(&enabled); err != nil || !enabled {
			return fmt.Errorf("角色 %d 不存在或已禁用", id)
		}
	}
	return nil
}
func nullableInput(value any) any {
	if text := strings.TrimSpace(toString(value)); text != "" {
		return text
	}
	return nil
}
func statusInput(value any) string {
	if value == "disabled" {
		return "disabled"
	}
	return "enabled"
}

func (s *Server) rbacUserUpdate(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	id, ok := orpID(w, r)
	if !ok {
		return
	}
	body, ok := orpBody(w, r)
	if !ok {
		return
	}
	roleIDs, err := parseRoleIDs(body["roleIds"])
	if err != nil {
		orpFailure(w, 400, err.Error())
		return
	}
	if err := s.validateRoleIDs(r, roleIDs); err != nil {
		orpFailure(w, 400, err.Error())
		return
	}
	var currentUsername string
	if err := s.db.QueryRowContext(r.Context(), "SELECT username FROM orp_account WHERE id=?", id).Scan(&currentUsername); err != nil {
		orpFailure(w, 404, "用户不存在")
		return
	}
	if currentUsername == usernameFromToken(r.Header.Get("Authorization")) {
		if statusInput(body["status"]) != "enabled" {
			orpFailure(w, 409, "不能禁用当前登录账号")
			return
		}
		var superID int64
		if err := s.db.QueryRowContext(r.Context(), "SELECT id FROM orp_role WHERE code='super'").Scan(&superID); err != nil {
			orpFailure(w, 500, "读取超级管理员角色失败")
			return
		}
		found := false
		for _, roleID := range roleIDs {
			if roleID == superID {
				found = true
			}
		}
		if !found {
			orpFailure(w, 409, "不能移除当前登录账号的超级管理员角色")
			return
		}
	}
	rolesJSON, _ := json.Marshal(roleIDs)
	name := strings.TrimSpace(toString(body["realName"]))
	if name == "" {
		orpFailure(w, 400, "姓名不能为空")
		return
	}
	password := toString(body["password"])
	if password != "" && len(password) < 12 {
		orpFailure(w, 400, "密码至少 12 位")
		return
	}
	var hash []byte
	if password != "" {
		hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			orpFailure(w, 500, "加密密码失败")
			return
		}
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		orpFailure(w, 500, "创建事务失败")
		return
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(r.Context(), "UPDATE orp_account SET real_name=?,email=?,phone=?,status=?,role_ids=? WHERE id=?", name, nullableInput(body["email"]), nullableInput(body["phone"]), statusInput(body["status"]), rolesJSON, id)
	if err != nil {
		orpFailure(w, 500, "更新用户失败")
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		orpFailure(w, 404, "用户不存在")
		return
	}
	if password != "" {
		if _, err := tx.ExecContext(r.Context(), "UPDATE orp_account SET password_hash=? WHERE id=?", string(hash), id); err != nil {
			orpFailure(w, 500, "更新密码失败")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		orpFailure(w, 500, "提交更新失败")
		return
	}
	orpReply(w, 200, map[string]any{"id": id, "realName": name, "roleIds": roleIDs, "status": statusInput(body["status"])})
}

func (s *Server) rbacUserDelete(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	id, ok := orpID(w, r)
	if !ok {
		return
	}
	var username string
	if err := s.db.QueryRowContext(r.Context(), "SELECT username FROM orp_account WHERE id=?", id).Scan(&username); errors.Is(err, sql.ErrNoRows) {
		orpFailure(w, 404, "用户不存在")
		return
	} else if err != nil {
		orpFailure(w, 500, "读取用户失败")
		return
	}
	if username == usernameFromToken(r.Header.Get("Authorization")) {
		orpFailure(w, 409, "不能删除当前登录账号")
		return
	}
	if _, err := s.db.ExecContext(r.Context(), "DELETE FROM orp_account WHERE id=?", id); err != nil {
		orpFailure(w, 500, "删除用户失败")
		return
	}
	orpReply(w, 200, nil)
}

func (s *Server) rbacUserResetPassword(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	id, ok := orpID(w, r)
	if !ok {
		return
	}
	var username string
	if err := s.db.QueryRowContext(r.Context(), "SELECT username FROM orp_account WHERE id=?", id).Scan(&username); err != nil {
		orpFailure(w, 404, "用户不存在")
		return
	}
	password, err := randomPassword()
	if err != nil {
		orpFailure(w, 500, "生成密码失败")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		orpFailure(w, 500, "加密密码失败")
		return
	}
	if _, err := s.db.ExecContext(r.Context(), "UPDATE orp_account SET password_hash=? WHERE id=?", string(hash), id); err != nil {
		orpFailure(w, 500, "重置密码失败")
		return
	}
	orpReply(w, 200, map[string]any{"username": username, "resetTo": password})
}

func (s *Server) rbacRolesList(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	rows, err := s.db.QueryContext(r.Context(), "SELECT id,code,name,description,status,builtin,permission_keys,created_at FROM orp_role ORDER BY id")
	if err != nil {
		orpFailure(w, 500, "读取角色失败")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var code, name, status string
		var desc sql.NullString
		var builtin bool
		var raw []byte
		var created time.Time
		if err := rows.Scan(&id, &code, &name, &desc, &status, &builtin, &raw, &created); err != nil {
			orpFailure(w, 500, "读取角色失败")
			return
		}
		var keys []string
		if err := json.Unmarshal(raw, &keys); err != nil {
			orpFailure(w, 500, "角色权限数据损坏")
			return
		}
		items = append(items, map[string]any{"id": id, "code": code, "name": name, "description": orpNullableString(desc), "status": status, "builtin": builtin, "permissionKeys": keys, "createdAt": created.Format("2006-01-02 15:04:05")})
	}
	if err := rows.Err(); err != nil {
		orpFailure(w, 500, "读取角色失败")
		return
	}
	orpReply(w, 200, orpPage(r, items))
}

func (s *Server) rbacRoleCreate(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	body, ok := orpBody(w, r)
	if !ok {
		return
	}
	name := strings.TrimSpace(toString(body["name"]))
	if name == "" {
		orpFailure(w, 400, "角色名称不能为空")
		return
	}
	code := strings.TrimSpace(toString(body["code"]))
	if code == "" {
		code = "role-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	keys, _ := json.Marshal(body["permissionKeys"])
	if string(keys) == "null" {
		keys = []byte("[]")
	}
	result, err := s.db.ExecContext(r.Context(), "INSERT INTO orp_role(code,name,description,status,builtin,permission_keys) VALUES(?,?,?,?,false,?)", code, name, nullableInput(body["description"]), statusInput(body["status"]), keys)
	if err != nil {
		orpFailure(w, 409, "角色代码已存在或保存失败")
		return
	}
	id, _ := result.LastInsertId()
	orpReply(w, 201, map[string]any{"id": id, "code": code, "name": name, "builtin": false, "status": statusInput(body["status"]), "permissionKeys": body["permissionKeys"]})
}

func (s *Server) rbacRoleUpdate(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	id, ok := orpID(w, r)
	if !ok {
		return
	}
	body, ok := orpBody(w, r)
	if !ok {
		return
	}
	name := strings.TrimSpace(toString(body["name"]))
	if name == "" {
		orpFailure(w, 400, "角色名称不能为空")
		return
	}
	var builtin bool
	if err := s.db.QueryRowContext(r.Context(), "SELECT builtin FROM orp_role WHERE id=?", id).Scan(&builtin); err != nil {
		orpFailure(w, 404, "角色不存在")
		return
	}
	if builtin {
		orpFailure(w, 409, "内置角色不可修改")
		return
	}
	keys, _ := json.Marshal(body["permissionKeys"])
	if string(keys) == "null" {
		keys = []byte("[]")
	}
	if _, err := s.db.ExecContext(r.Context(), "UPDATE orp_role SET name=?,description=?,status=?,permission_keys=? WHERE id=?", name, nullableInput(body["description"]), statusInput(body["status"]), keys, id); err != nil {
		orpFailure(w, 500, "更新角色失败")
		return
	}
	orpReply(w, 200, map[string]any{"id": id, "name": name, "status": statusInput(body["status"]), "permissionKeys": body["permissionKeys"]})
}

func (s *Server) rbacRoleDelete(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	id, ok := orpID(w, r)
	if !ok {
		return
	}
	var builtin bool
	if err := s.db.QueryRowContext(r.Context(), "SELECT builtin FROM orp_role WHERE id=?", id).Scan(&builtin); err != nil {
		orpFailure(w, 404, "角色不存在")
		return
	}
	if builtin {
		orpFailure(w, 409, "内置角色不可删除")
		return
	}
	rows, err := s.db.QueryContext(r.Context(), "SELECT role_ids FROM orp_account")
	if err != nil {
		orpFailure(w, 500, "检查角色引用失败")
		return
	}
	for rows.Next() {
		var raw []byte
		if rows.Scan(&raw) != nil {
			continue
		}
		var ids []int64
		_ = json.Unmarshal(raw, &ids)
		for _, value := range ids {
			if value == id {
				rows.Close()
				orpFailure(w, 409, "角色仍被用户引用")
				return
			}
		}
	}
	rows.Close()
	if _, err := s.db.ExecContext(r.Context(), "DELETE FROM orp_role WHERE id=?", id); err != nil {
		orpFailure(w, 500, "删除角色失败")
		return
	}
	orpReply(w, 200, nil)
}

func (s *Server) rbacPermissions(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	orpReply(w, 200, []map[string]any{
		{"key": "orp:read", "title": "查看配置", "type": "menu"},
		{"key": "orp:write", "title": "编辑配置", "type": "action"},
		{"key": "orp:secrets:read", "title": "查看敏感信息", "type": "action"},
	})
}

func (s *Server) rbacRolePermissionsGet(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	id, ok := orpID(w, r)
	if !ok {
		return
	}
	var raw []byte
	if err := s.db.QueryRowContext(r.Context(), "SELECT permission_keys FROM orp_role WHERE id=?", id).Scan(&raw); err != nil {
		orpFailure(w, 404, "角色不存在")
		return
	}
	var keys []string
	if err := json.Unmarshal(raw, &keys); err != nil {
		orpFailure(w, 500, "权限数据损坏")
		return
	}
	orpReply(w, 200, map[string]any{"permissionKeys": keys})
}

func (s *Server) rbacRolePermissionsPut(w http.ResponseWriter, r *http.Request) {
	if !s.requireSuper(w, r) {
		return
	}
	id, ok := orpID(w, r)
	if !ok {
		return
	}
	body, ok := orpBody(w, r)
	if !ok {
		return
	}
	var builtin bool
	if err := s.db.QueryRowContext(r.Context(), "SELECT builtin FROM orp_role WHERE id=?", id).Scan(&builtin); err != nil {
		orpFailure(w, 404, "角色不存在")
		return
	}
	if builtin {
		orpFailure(w, 409, "内置角色权限不可修改")
		return
	}
	keys, ok := body["permissionKeys"].([]any)
	if !ok {
		orpFailure(w, 400, "permissionKeys 必须是数组")
		return
	}
	raw, _ := json.Marshal(keys)
	if _, err := s.db.ExecContext(r.Context(), "UPDATE orp_role SET permission_keys=? WHERE id=?", raw, id); err != nil {
		orpFailure(w, 500, "保存权限失败")
		return
	}
	orpReply(w, 200, map[string]any{"permissionKeys": keys})
}

func randomPassword() (string, error) {
	buffer := make([]byte, 18)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", buffer), nil
}
