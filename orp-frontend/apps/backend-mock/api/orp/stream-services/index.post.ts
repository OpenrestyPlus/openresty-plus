import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, readBody, setResponseStatus } from 'h3';

import {
  centerName,
  nextId,
  now,
  recordAudit,
  streamPortConflict,
  streamServices,
} from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const BACKEND_PATTERN = /^[a-zA-Z0-9.-]+:\d{1,5}$/;

export default defineEventHandler(async (event) => {
  const body = await readBody(event);
  const centerId = Number.parseInt(String(body?.centerId ?? ''), 10);
  const protocol = body?.protocol === 'udp' ? 'udp' : 'tcp';
  const listenPort = Number.parseInt(String(body?.listenPort ?? ''), 10);
  const listenAddress = String(body?.listenAddress ?? '0.0.0.0').trim();
  const description = String(body?.description ?? '').trim();
  const backends = Array.isArray(body?.backends)
    ? body.backends.map((item) => String(item).trim()).filter(Boolean)
    : [];

  if (!centerId || !listenPort) {
    setResponseStatus(event, 400);
    return useResponseError('所属中心与监听端口为必填项');
  }
  if (!(listenPort >= 1 && listenPort <= 65_535)) {
    setResponseStatus(event, 400);
    return useResponseError('监听端口必须是 1-65535 之间的整数');
  }
  if (backends.length === 0) {
    setResponseStatus(event, 400);
    return useResponseError('至少配置一个后端转发地址（host:port）');
  }
  if (backends.some((item) => !BACKEND_PATTERN.test(item))) {
    setResponseStatus(event, 400);
    return useResponseError('后端地址格式不正确，示例：mysql-cluster-01:3306');
  }
  const conflict = streamPortConflict(centerId, listenPort, protocol);
  if (conflict) {
    setResponseStatus(event, 400);
    return useResponseError(
      `端口冲突：${centerName(centerId)} 下已存在 ${protocol.toUpperCase()} 监听端口 ${listenPort}（${conflict.description || '未命名服务'}）`,
    );
  }

  const service = {
    backends,
    centerId,
    createdAt: now(),
    description,
    id: nextId(streamServices),
    listenAddress: listenAddress || '0.0.0.0',
    listenPort,
    protocol,
  };
  streamServices.push(service);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'create',
    detail: `协议：${protocol.toUpperCase()}；监听：${listenAddress}:${listenPort}；后端：${backends.join(', ')}；中心：${centerName(centerId)}`,
    module: 'stream',
    target: `stream:${protocol}/${listenPort}`,
  });
  return useResponseSuccess(service);
});
