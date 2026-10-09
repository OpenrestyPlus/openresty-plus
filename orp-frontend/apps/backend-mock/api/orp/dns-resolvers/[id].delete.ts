import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, getRouterParam, setResponseStatus } from 'h3';

import { centerName, dnsResolvers, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const index = dnsResolvers.findIndex((item) => item.id === id);
  if (index === -1) {
    setResponseStatus(event, 404);
    return useResponseError('解析器不存在或已被删除');
  }

  const [removed] = dnsResolvers.splice(index, 1);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'delete',
    detail: `地址：${removed.address}:${removed.port}；中心：${centerName(removed.centerId)}`,
    module: 'dns',
    target: `dns:${removed.address}:${removed.port}`,
  });
  return useResponseSuccess(null);
});
