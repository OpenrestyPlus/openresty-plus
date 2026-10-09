import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, getRouterParam, setResponseStatus } from 'h3';

import { centerName, nodes, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const node = nodes.find((item) => item.id === id);
  if (!node) {
    setResponseStatus(event, 404);
    return useResponseError('节点不存在或已被删除');
  }

  const before = node.status;
  node.status = 'offline';
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'offline',
    detail: `节点 ${node.name}（${centerName(node.centerId)}）状态由 ${before} 变更为 offline`,
    module: 'node',
    target: `node:${node.name}`,
  });
  return useResponseSuccess(node);
});
