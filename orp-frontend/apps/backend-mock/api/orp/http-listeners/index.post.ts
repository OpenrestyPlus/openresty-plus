import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, readBody, setResponseStatus } from 'h3';

import type { OrpLocationRule } from '~/utils/orp-store';

import {
  centerName,
  certificates,
  httpListeners,
  httpPortConflict,
  nextId,
  nextRouteId,
  now,
  recordAudit,
} from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const DOMAIN_PATTERN = /^(\*\.)?([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/;

function normalizeRoutes(raw: unknown): OrpLocationRule[] {
  if (!Array.isArray(raw)) {
    return [];
  }
  return raw
    .map((route) => ({
      id: nextRouteId(),
      path: String((route as any)?.path ?? '').trim(),
      proxyTimeoutMs:
        Number.parseInt(String((route as any)?.proxyTimeoutMs ?? 30_000), 10) ||
        30_000,
      upstream: String((route as any)?.upstream ?? '').trim(),
    }))
    .filter((route) => route.path && route.upstream);
}

export default defineEventHandler(async (event) => {
  const body = await readBody(event);
  const centerId = Number.parseInt(String(body?.centerId ?? ''), 10);
  const domain = String(body?.domain ?? '').trim();
  const port = Number.parseInt(String(body?.port ?? ''), 10);
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
  const conflict = httpPortConflict(centerId, port);
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

  const listener = {
    centerId,
    createdAt: now(),
    domain,
    id: nextId(httpListeners),
    port,
    routes: normalizeRoutes(body?.routes),
    tlsCertId,
  };
  httpListeners.push(listener);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'create',
    detail: `域名：${domain}；端口：${port}；中心：${centerName(centerId)}；路由 ${listener.routes.length} 条`,
    module: 'http',
    target: `http:${domain}:${port}`,
  });
  return useResponseSuccess(listener);
});
