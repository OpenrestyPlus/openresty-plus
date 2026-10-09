import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, getRouterParam, setResponseStatus } from 'h3';

import { centerUsage, centers, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const index = centers.findIndex((item) => item.id === id);
  if (index === -1) {
    setResponseStatus(event, 404);
    return useResponseError('中心不存在或已被删除');
  }

  const usage = centerUsage(id);
  if (usage.total > 0) {
    setResponseStatus(event, 400);
    return useResponseError(
      `中心下仍存在关联资源（${usage.detail}），请先清空后再删除`,
    );
  }

  const [removed] = centers.splice(index, 1);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'delete',
    detail: `名称：${removed.name}；标识：${removed.code}`,
    module: 'center',
    target: `center:${removed.code}`,
  });
  return useResponseSuccess(null);
});
