import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, getRouterParam, setResponseStatus } from 'h3';

import { nodes, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const index = nodes.findIndex((item) => item.id === id);
  if (index === -1) {
    setResponseStatus(event, 404);
    return useResponseError('节点不存在或已被删除');
  }

  const [removed] = [nodes[index]!];
  if (removed.status === 'online') {
    setResponseStatus(event, 400);
    return useResponseError(
      `节点 ${removed.name} 处于在线状态，请先执行下线操作后再移除`,
    );
  }

  nodes.splice(index, 1);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'delete',
    detail: `名称：${removed.name}；地址：${removed.host}`,
    module: 'node',
    target: `node:${removed.name}`,
  });
  return useResponseSuccess(null);
});
