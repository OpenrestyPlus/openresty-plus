import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, getRouterParam, setResponseStatus } from 'h3';

import { certificates, httpListeners, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const index = certificates.findIndex((item) => item.id === id);
  if (index === -1) {
    setResponseStatus(event, 404);
    return useResponseError('证书不存在或已被删除');
  }

  const referencing = httpListeners.filter((item) => item.tlsCertId === id);
  if (referencing.length > 0) {
    setResponseStatus(event, 400);
    return useResponseError(
      `该证书仍被 ${referencing.length} 个 HTTP 监听域名引用（${referencing.map((item) => item.domain).join('、')}），请先解除绑定`,
    );
  }

  const [removed] = certificates.splice(index, 1);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'delete',
    detail: `名称：${removed.name}；域名：${removed.domains}；有效期：${removed.notBefore} ~ ${removed.notAfter}`,
    module: 'tls',
    target: `tls:${removed.name}`,
  });
  return useResponseSuccess(null);
});
