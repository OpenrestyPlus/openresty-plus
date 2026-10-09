import { defineEventHandler, getRouterParam, setResponseStatus } from 'h3';

import { centerName, nodes } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

/** 基于 ID 的确定性伪指标（Mock 环境演示用） */
export default defineEventHandler((event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const node = nodes.find((item) => item.id === id);
  if (!node) {
    setResponseStatus(event, 404);
    return useResponseError('节点不存在或已被删除');
  }

  const seed = node.id;
  const online = node.status === 'online';
  const round = (value: number, digits = 1) =>
    Number.parseFloat(value.toFixed(digits));

  return useResponseSuccess({
    metrics: {
      bandwidthInMbps: online ? round(80 + (seed * 37) % 260) : 0,
      bandwidthOutMbps: online ? round(120 + (seed * 53) % 340) : 0,
      connections: online ? 1000 + seed * 837 : 0,
      cpuPercent: online ? round(18 + ((seed * 13) % 45)) : 0,
      errorRatePercent: online ? round((seed % 7) / 100, 2) : 0,
      memPercent: online ? round(35 + ((seed * 11) % 50)) : 0,
      qps: online ? 500 + seed * 321 : 0,
      requestTotal24h: online ? 80_0000 + seed * 12_3457 : 0,
      uptimeDays: seed * 3 + 12,
    },
    node: { ...node, centerName: centerName(node.centerId) },
  });
});
