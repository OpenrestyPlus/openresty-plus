package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"
)

func (s *Server) orpPublishAbort(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.orpIdentity(w, r); !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		orpFailure(w, 400, "批次 ID 无效")
		return
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		orpFailure(w, 500, "中止批次失败")
		return
	}
	defer tx.Rollback()
	var phase string
	if err := tx.QueryRowContext(r.Context(), "SELECT phase FROM orp_publish_batch WHERE id=? FOR UPDATE", id).Scan(&phase); err != nil {
		if err == sql.ErrNoRows {
			orpFailure(w, 404, "批次不存在")
		} else {
			orpFailure(w, 500, "读取批次失败")
		}
		return
	}
	if phase == "done" || phase == "aborted" {
		orpFailure(w, 409, "批次已经结束")
		return
	}
	result, err := tx.ExecContext(r.Context(), "UPDATE orp_publish_item SET status='aborted',error_text='操作员中止批次',finished_at=NOW(6) WHERE batch_id=? AND status='queued'", id)
	if err != nil {
		orpFailure(w, 500, "中止节点任务失败")
		return
	}
	count, _ := result.RowsAffected()
	if _, err := tx.ExecContext(r.Context(), "UPDATE orp_publish_batch SET phase='aborted',finished_at=NOW(6) WHERE id=?", id); err != nil {
		orpFailure(w, 500, "中止批次失败")
		return
	}
	if err := tx.Commit(); err != nil {
		orpFailure(w, 500, "提交中止操作失败")
		return
	}
	orpReply(w, 200, map[string]any{"id": id, "phase": "aborted", "abortedCount": count})
}

func (s *Server) orpPublishAdvance(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.orpIdentity(w, r); !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		orpFailure(w, 400, "批次 ID 无效")
		return
	}
	result, err := s.db.ExecContext(r.Context(), "UPDATE orp_publish_batch SET phase='running',canary_observe_until=NOW(6) WHERE id=? AND phase='canary_observing'", id)
	if err != nil {
		orpFailure(w, 500, "推进批次失败")
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		orpFailure(w, 409, "批次不在观察期")
		return
	}
	orpReply(w, 200, map[string]any{"id": id, "phase": "running"})
}
