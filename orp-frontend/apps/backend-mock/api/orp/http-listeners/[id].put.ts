import { getClientIp, getOperator } from '~/utils/orp-event';

import {
  defineEventHandler,
  getRouterParam,
  readBody,
  setResponseStatus,
} from 'h3';

import type { OrpLocationRule } from '~/utils/orp-store';

import {
  centerName,
  certificates,
  httpListeners,
  httpPortConflict,
  nextRouteId,
  recordAudit,
} from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const DOMAIN_PATTERN = /^(\*\.)?([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/;

function normalizeRoutes(raw: unknown): OrpLocationRule[] {
  if (!Array.isArray(raw)) {
    return [];
  }
  return raw
    .map((route) => {
      const existingId = Number.parseInt(String((route as any)?.id), 10);
      return {
        id: existingId > 0 ? existingId : nextRouteId(),
        path: String((route as any)?.path ?? '').trim(),
        proxyTimeoutMs:
          Number.parseInt(String((route as any)?.proxyTimeoutMs ?? 30_000), 10) ||
          30_000,
        upstream: String((route as any)?.upstream ?? '').trim(),
      };
    })
    .filter((route) => route.path && route.upstream);
}

export default defineEventHandler(async (event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const listener = httpListeners.find((item) => item.id === id);
  if (!listener) {
    setResponseStatus(event, 404);
    return useResponseError('监听配置不存在或已被删除');
  }

  const body = await readBody(event);
  const centerId = Number.parseInt(String(body?.centerId ?? listener.centerId), 10);
  const domain = String(body?.domain ?? listener.domain).trim();
  const port = Number.parseInt(String(body?.port ?? listener.port), 10);
  const tlsCertId = body?.tlsCertId ? Number(body.tlsCertId) : null;

  if (!centerId || !domain || !port) {
    setResponseStatus(event, 400);
    return useResponseError('所属中心、域名与监听端口为必填项');
  }
  if (!DOMAIN_PATTERN.test(domain)) {
    setResponseStatus(event, 400);
    return useResponseError('域名格式不正确，示例：api.example.cn 或 *.example.cn');
  }
  if (!(port >= 1 && port <= 65_535)) {
    setResponseStatus(event, 400);
    return useResponseError('监听端口必须是 1-65535 之间的整数');
  }
  const conflict = httpPortConflict(centerId, port, id);
  if (conflict) {
    setResponseStatus(event, 400);
    return useResponseError(
      `端口冲突：${centerName(centerId)} 下已存在 HTTP 监听端口 ${port}（${conflict.domain}）`,
    );
  }
  if (tlsCertId && !certificates.some((cert) => cert.id === tlsCertId)) {
    setResponseStatus(event, 400);
    return useResponseError('所选 TLS 证书不存在，请刷新后重试');
  }

  const before = `域名：${listener.domain}；端口：${listener.port}；路由 ${listener.routes.length} 条`;
  const routes = normalizeRoutes(body?.routes);
  Object.assign(listener, { centerId, domain, port, routes, tlsCertId });
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'update',
    detail: `变更前【${before}】→ 变更后【域名：${domain}；端口：${port}；路由 ${routes.length} 条】`,
    module: 'http',
    target: `http:${domain}:${port}`,
  });
  return useResponseSuccess(listener);
});
