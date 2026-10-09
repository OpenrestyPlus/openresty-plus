import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, getRouterParam, setResponseStatus } from 'h3';

import { centerName, recordAudit, streamServices } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const index = streamServices.findIndex((item) => item.id === id);
  if (index === -1) {
    setResponseStatus(event, 404);
    return useResponseError('Stream 服务不存在或已被删除');
  }

  const [removed] = streamServices.splice(index, 1);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'delete',
    detail: `协议：${removed.protocol.toUpperCase()}；监听：${removed.listenAddress}:${removed.listenPort}；中心：${centerName(removed.centerId)}`,
    module: 'stream',
    target: `stream:${removed.protocol}/${removed.listenPort}`,
  });
  return useResponseSuccess(null);
});
