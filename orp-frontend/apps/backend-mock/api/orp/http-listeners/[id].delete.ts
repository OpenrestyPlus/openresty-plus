import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, getRouterParam, setResponseStatus } from 'h3';

import { centerName, httpListeners, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const index = httpListeners.findIndex((item) => item.id === id);
  if (index === -1) {
    setResponseStatus(event, 404);
    return useResponseError('监听配置不存在或已被删除');
  }

  const [removed] = httpListeners.splice(index, 1);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'delete',
    detail: `域名：${removed.domain}；端口：${removed.port}；中心：${centerName(removed.centerId)}；路由 ${removed.routes.length} 条`,
    module: 'http',
    target: `http:${removed.domain}:${removed.port}`,
  });
  return useResponseSuccess(null);
});
