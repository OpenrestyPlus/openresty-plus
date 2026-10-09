package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/redis/go-redis/v9"
	"net.daoke/orp-backend/internal/webui"
)

// New builds the control-plane HTTP surface. Store-backed API modules are
// registered here as they are migrated from the Java control plane.
type Server struct {
	db    *sql.DB
	redis *redis.Client
}

func New(database *sql.DB) http.Handler {
	return localCORS(NewMux(database))
}

func NewWithRedis(database *sql.DB, redisClient *redis.Client) http.Handler {
	return localCORS(NewMuxWithRedis(database, redisClient))
}

// NewMux exposes route registration for contract checks without invoking handlers.
func NewMux(database *sql.DB) *http.ServeMux {
	return NewMuxWithRedis(database, nil)
}

// NewMuxWithRedis installs Redis-backed dashboard snapshot caching and refresh queueing.
func NewMuxWithRedis(database *sql.DB, redisClient *redis.Client) *http.ServeMux {
	mux := http.NewServeMux()
	server := &Server{db: database, redis: redisClient}
	mux.HandleFunc("POST /api/orp/nodes/{id}/agent/reload", server.agentReloadTaskCreate)
	mux.HandleFunc("GET /api/orp/nodes/{id}/agent/tasks", server.agentTasksList)
	mux.HandleFunc("PUT /api/orp/nodes/{id}/agent/certificate", server.agentCertificateRegister)
	mux.HandleFunc("POST /api/agent/v1/heartbeat", server.agentHeartbeat)
	mux.HandleFunc("GET /api/agent/v1/tasks/next", server.agentNextTask)
	mux.HandleFunc("POST /api/agent/v1/tasks/{taskId}/result", server.agentTaskResult)
	if redisClient != nil {
		StartDashboardSnapshotWorker(context.Background(), database, redisClient)
	}
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("POST /api/auth/login", server.login)
	mux.HandleFunc("POST /api/auth/refresh", server.refreshToken)
	mux.HandleFunc("POST /api/auth/logout", server.logout)
	mux.HandleFunc("GET /api/auth/codes", server.accessCodes)
	mux.HandleFunc("GET /api/user/info", server.userInfo)
	mux.HandleFunc("GET /api/user/profile", server.profileGet)
	mux.HandleFunc("PUT /api/user/profile", server.profilePut)
	mux.HandleFunc("PUT /api/user/password", server.profilePasswordPut)
	mux.HandleFunc("GET /api/user/otp/setup", server.profileOTPSetup)
	mux.HandleFunc("POST /api/user/otp/enable", server.profileOTPEnable)
	mux.HandleFunc("POST /api/user/otp/disable", server.profileOTPDisable)
	mux.HandleFunc("GET /api/user/profile/audit", server.profileAudit)
	mux.HandleFunc("GET /api/menu/all", server.menuAll)
	mux.HandleFunc("GET /api/orp/certificates/stats", server.orpCertificateStats)
	mux.HandleFunc("GET /api/orp/certificates/{id}/detail", server.orpCertificateDetail)
	mux.HandleFunc("GET /api/orp/geoip-database", server.orpGeoIPDatabaseGet)
	mux.HandleFunc("POST /api/orp/geoip-database/import", server.orpGeoIPDatabaseImport)
	mux.HandleFunc("GET /api/orp/alert-channels", server.orpAlertChannelsList)
	mux.HandleFunc("POST /api/orp/alert-channels", server.orpAlertChannelCreate)
	mux.HandleFunc("PUT /api/orp/alert-channels/{id}", server.orpAlertChannelUpdate)
	mux.HandleFunc("DELETE /api/orp/alert-channels/{id}", server.orpAlertChannelDelete)
	mux.HandleFunc("POST /api/orp/alert-channels/{id}/test", server.orpAlertChannelTest)
	mux.HandleFunc("GET /api/orp/tls-alert-rules", server.orpTLSAlertRulesList)
	mux.HandleFunc("PUT /api/orp/tls-alert-rules/{certificateID}", server.orpTLSAlertRuleSave)
	mux.HandleFunc("GET /api/orp/alert-deliveries", server.orpAlertDeliveriesList)
	mux.HandleFunc("POST /api/orp/tls-alerts/check", server.orpTLSAlertsCheck)
	mux.HandleFunc("GET /api/orp/settings/{id}/plain", server.orpSettingPlain)
	mux.HandleFunc("GET /api/orp/audit-logs", server.orpAuditList)
	mux.HandleFunc("GET /api/orp/settings/groups", server.orpSettingGroupNames)
	mux.HandleFunc("POST /api/orp/settings/batch-move", server.orpSettingBatchMove)
	mux.HandleFunc("PUT /api/orp/settings/{id}/status", server.orpSettingStatus)
	mux.HandleFunc("PUT /api/orp/setting-groups/reorder", server.orpSettingGroupReorder)
	mux.HandleFunc("DELETE /api/orp/setting-groups/{id}", server.orpSettingGroupDelete)
	for _, singleton := range []string{"http-directives", "stream-directives", "http-log-format", "stream-log-format", "http-error-pages"} {
		mux.HandleFunc("GET /api/orp/"+singleton, func(w http.ResponseWriter, r *http.Request) {
			r.SetPathValue("name", singleton)
			server.orpSingletonGet(w, r)
		})
		mux.HandleFunc("PUT /api/orp/"+singleton, func(w http.ResponseWriter, r *http.Request) {
			r.SetPathValue("name", singleton)
			server.orpSingletonPut(w, r)
		})
	}
	mux.HandleFunc("GET /api/orp/{kind}", server.orpList)
	mux.HandleFunc("POST /api/orp/{kind}", server.orpCreate)
	mux.HandleFunc("PUT /api/orp/{kind}/{id}", server.orpUpdate)
	mux.HandleFunc("DELETE /api/orp/{kind}/{id}", server.orpDelete)
	mux.HandleFunc("GET /api/orp/rbac/users", server.rbacUsersList)
	mux.HandleFunc("POST /api/orp/rbac/users", server.rbacUserCreate)
	mux.HandleFunc("PUT /api/orp/rbac/users/{id}", server.rbacUserUpdate)
	mux.HandleFunc("DELETE /api/orp/rbac/users/{id}", server.rbacUserDelete)
	mux.HandleFunc("POST /api/orp/rbac/users/{id}/reset-password", server.rbacUserResetPassword)
	mux.HandleFunc("GET /api/orp/rbac/roles", server.rbacRolesList)
	mux.HandleFunc("POST /api/orp/rbac/roles", server.rbacRoleCreate)
	mux.HandleFunc("PUT /api/orp/rbac/roles/{id}", server.rbacRoleUpdate)
	mux.HandleFunc("DELETE /api/orp/rbac/roles/{id}", server.rbacRoleDelete)
	mux.HandleFunc("GET /api/orp/rbac/permissions", server.rbacPermissions)
	mux.HandleFunc("GET /api/orp/rbac/roles/{id}/permissions", server.rbacRolePermissionsGet)
	mux.HandleFunc("PUT /api/orp/rbac/roles/{id}/permissions", server.rbacRolePermissionsPut)
	mux.HandleFunc("POST /api/orp/publish/precheck", server.orpPublishPrecheck)
	mux.HandleFunc("POST /api/orp/publish", server.orpPublishCreate)
	mux.HandleFunc("GET /api/orp/publish/batches/{id}", server.orpPublishBatch)
	mux.HandleFunc("GET /api/orp/publish/summary", server.orpPublishSummary)
	mux.HandleFunc("GET /api/orp/publish/diff", server.orpPublishDiff)
	mux.HandleFunc("GET /api/orp/publish/history", server.orpPublishHistory)
	mux.HandleFunc("GET /api/orp/publish/history/stats", server.orpPublishHistoryStats)
	mux.HandleFunc("GET /api/orp/logs", server.orpLogs)
	mux.HandleFunc("GET /api/orp/logs/events", server.orpLogEvents)
	mux.HandleFunc("GET /api/orp/nodes/{id}/metrics", server.orpNodeMetrics)
	mux.HandleFunc("GET /api/orp/nodes/{id}/latency-history", server.orpNodeLatencyHistory)
	mux.HandleFunc("PUT /api/orp/nodes/{id}/status", server.orpNodeStatus)
	mux.HandleFunc("GET /api/orp/dashboard/metrics", server.orpDashboardMetrics)
	mux.HandleFunc("GET /api/orp/dashboard/events", server.orpDashboardEvents)
	mux.HandleFunc("GET /api/orp/dashboard/trends", server.orpDashboardTrends)
	mux.HandleFunc("GET /api/orp/dashboard/top-rankings", server.orpDashboardTopRankings)
	mux.HandleFunc("POST /api/orp/publish/rollback", server.orpPublishRollback)
	mux.HandleFunc("POST /api/orp/publish/batches/{id}/abort", server.orpPublishAbort)
	mux.HandleFunc("POST /api/orp/publish/batches/{id}/advance", server.orpPublishAdvance)
	mux.HandleFunc("GET /api/centers", server.listCenters)
	mux.HandleFunc("GET /api/centers/paged", server.pageCenters)
	mux.HandleFunc("POST /api/centers", server.createCenter)
	mux.HandleFunc("PUT /api/centers/{centerID}", server.updateCenter)
	mux.HandleFunc("DELETE /api/centers/{centerID}", server.deleteCenter)
	mux.HandleFunc("GET /api/centers/{centerID}/nodes", server.listNodes)
	mux.HandleFunc("GET /api/centers/{centerID}/nodes/paged", server.pageNodes)
	mux.HandleFunc("POST /api/centers/{centerID}/nodes", server.createNode)
	mux.HandleFunc("PUT /api/centers/{centerID}/nodes/{nodeID}", server.updateNode)
	mux.HandleFunc("DELETE /api/centers/{centerID}/nodes/{nodeID}", server.deleteNode)
	mux.HandleFunc("GET /api/centers/{centerID}/node-metrics", server.listNodeMetrics)
	mux.HandleFunc("GET /api/centers/{centerID}/http/upstreams", server.listHTTPUpstreams)
	mux.HandleFunc("GET /api/centers/{centerID}/http/upstreams/paged", server.pageHTTPUpstreams)
	mux.HandleFunc("POST /api/centers/{centerID}/http/upstreams", server.createHTTPUpstream)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/upstreams/{upstreamID}", server.updateHTTPUpstream)
	mux.HandleFunc("DELETE /api/centers/{centerID}/http/upstreams/{upstreamID}", server.deleteHTTPUpstream)
	mux.HandleFunc("GET /api/centers/{centerID}/http/upstreams/{upstreamID}/targets/paged", server.pageHTTPUpstreamTargets)
	mux.HandleFunc("POST /api/centers/{centerID}/http/upstreams/{upstreamID}/targets", server.createHTTPUpstreamTarget)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/upstreams/{upstreamID}/targets/{targetID}", server.updateHTTPUpstreamTarget)
	mux.HandleFunc("DELETE /api/centers/{centerID}/http/upstreams/{upstreamID}/targets/{targetID}", server.deleteHTTPUpstreamTarget)
	mux.HandleFunc("GET /api/centers/{centerID}/http/settings", server.getHTTPSettings)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/settings", server.putHTTPSettings)
	mux.HandleFunc("GET /api/centers/{centerID}/http/servers", server.listHTTPServers)
	mux.HandleFunc("GET /api/centers/{centerID}/http/servers/paged", server.pageHTTPServers)
	mux.HandleFunc("POST /api/centers/{centerID}/http/servers", server.createHTTPServer)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/servers/{serverID}", server.updateHTTPServer)
	mux.HandleFunc("DELETE /api/centers/{centerID}/http/servers/{serverID}", server.deleteHTTPServer)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/servers/{serverID}/policy-settings", server.putHTTPServerPolicySettings)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/servers/{serverID}/directives", server.putHTTPServerDirectives)
	mux.HandleFunc("GET /api/centers/{centerID}/http/servers/locations/paged", server.pageCenterHTTPLocations)
	mux.HandleFunc("GET /api/centers/{centerID}/http/servers/{serverID}/locations", server.listHTTPLocations)
	mux.HandleFunc("GET /api/centers/{centerID}/http/servers/{serverID}/locations/paged", server.pageHTTPLocations)
	mux.HandleFunc("POST /api/centers/{centerID}/http/servers/{serverID}/locations", server.createHTTPLocation)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/servers/{serverID}/locations/{locationID}", server.updateHTTPLocation)
	mux.HandleFunc("DELETE /api/centers/{centerID}/http/servers/{serverID}/locations/{locationID}", server.deleteHTTPLocation)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/servers/{serverID}/locations/{locationID}/policy-settings", server.putHTTPLocationPolicySettings)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/servers/{serverID}/locations/{locationID}/directives", server.putHTTPLocationDirectives)
	mux.HandleFunc("PUT /api/centers/{centerID}/http/servers/{serverID}/locations/{locationID}/dynamic-dns", server.putHTTPLocationDynamicDNS)
	mux.HandleFunc("GET /api/centers/{centerID}/tls-certificates", server.listTLSCertificates)
	mux.HandleFunc("POST /api/centers/{centerID}/tls-certificates", server.createTLSCertificate)
	mux.HandleFunc("PUT /api/centers/{centerID}/tls-certificates/{certificateID}", server.updateTLSCertificate)
	mux.HandleFunc("DELETE /api/centers/{centerID}/tls-certificates/{certificateID}", server.deleteTLSCertificate)
	mux.HandleFunc("GET /api/centers/{centerID}/stream/upstreams", server.listStreamUpstreams)
	mux.HandleFunc("GET /api/centers/{centerID}/stream/upstreams/paged", server.pageStreamUpstreams)
	mux.HandleFunc("POST /api/centers/{centerID}/stream/upstreams", server.createStreamUpstream)
	mux.HandleFunc("PUT /api/centers/{centerID}/stream/upstreams/{upstreamID}", server.updateStreamUpstream)
	mux.HandleFunc("DELETE /api/centers/{centerID}/stream/upstreams/{upstreamID}", server.deleteStreamUpstream)
	mux.HandleFunc("GET /api/centers/{centerID}/stream/servers", server.listStreamServers)
	mux.HandleFunc("GET /api/centers/{centerID}/stream/servers/paged", server.pageStreamServers)
	mux.HandleFunc("POST /api/centers/{centerID}/stream/servers", server.createStreamServer)
	mux.HandleFunc("PUT /api/centers/{centerID}/stream/servers/{streamServerID}", server.updateStreamServer)
	mux.HandleFunc("DELETE /api/centers/{centerID}/stream/servers/{streamServerID}", server.deleteStreamServer)
	mux.HandleFunc("PUT /api/centers/{centerID}/stream/servers/{streamServerID}/dynamic-dns", server.putStreamServerDynamicDNS)
	mux.HandleFunc("PUT /api/centers/{centerID}/stream/servers/{streamServerID}/policy-settings", server.putStreamServerPolicySettings)
	mux.HandleFunc("GET /api/centers/{centerID}/dns-resolvers", server.listDNSResolvers)
	mux.HandleFunc("GET /api/centers/{centerID}/dns-resolvers/paged", server.pageDNSResolvers)
	mux.HandleFunc("POST /api/centers/{centerID}/dns-resolvers", server.createDNSResolver)
	mux.HandleFunc("PUT /api/centers/{centerID}/dns-resolvers/{resolverID}", server.updateDNSResolver)
	mux.HandleFunc("DELETE /api/centers/{centerID}/dns-resolvers/{resolverID}", server.deleteDNSResolver)
	mux.HandleFunc("GET /api/centers/{centerID}/audit-events", server.listAuditEvents)
	mux.HandleFunc("GET /api/centers/{centerID}/audit-events/paged", server.pageAuditEvents)
	mux.HandleFunc("GET /api/centers/{centerID}/runtime-configurations", server.listRuntimeConfigurations)
	mux.HandleFunc("GET /api/centers/{centerID}/runtime-configurations/paged", server.pageRuntimeConfigurations)
	mux.HandleFunc("GET /api/centers/{centerID}/runtime-configurations/current", server.currentRuntimeConfiguration)
	mux.HandleFunc("GET /api/centers/{centerID}/runtime-configurations/draft", server.draftRuntimeConfiguration)
	mux.HandleFunc("GET /api/centers/{centerID}/runtime-configurations/draft/compare", server.compareRuntimeConfiguration)
	mux.HandleFunc("POST /api/centers/{centerID}/runtime-configurations", server.publishRuntimeConfiguration)
	mux.HandleFunc("GET /api/centers/{centerID}/control-api-reloads", server.listReloadTasks)
	mux.HandleFunc("GET /api/centers/{centerID}/control-api-reloads/paged", server.pageReloadTasks)
	mux.Handle("/", webui.Handler())
	return mux
}

// NewAgentMux exposes only the mTLS-protected node protocol on its dedicated
// listener, keeping browser and user APIs on the regular control-plane port.
func NewAgentMux(database *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	server := &Server{db: database}
	mux.HandleFunc("POST /api/agent/v1/heartbeat", server.agentHeartbeat)
	mux.HandleFunc("GET /api/agent/v1/tasks/next", server.agentNextTask)
	mux.HandleFunc("POST /api/agent/v1/tasks/{taskId}/result", server.agentTaskResult)
	return mux
}

func localCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isLocalOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept-Language")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions && strings.HasPrefix(r.URL.Path, "/api/") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLocalOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	return parsed.Scheme == "http" && (host == "127.0.0.1" || host == "localhost")
}

func healthz(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
func writeError(writer http.ResponseWriter, status int, detail string) {
	writeJSON(writer, status, map[string]string{"detail": detail})
}
