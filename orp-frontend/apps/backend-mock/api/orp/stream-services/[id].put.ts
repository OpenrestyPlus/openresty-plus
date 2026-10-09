import { getClientIp, getOperator } from '~/utils/orp-event';

import {
  defineEventHandler,
  getRouterParam,
  readBody,
  setResponseStatus,
} from 'h3';

import {
  centerName,
  recordAudit,
  streamPortConflict,
  streamServices,
} from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const BACKEND_PATTERN = /^[a-zA-Z0-9.-]+:\d{1,5}$/;

export default defineEventHandler(async (event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const service = streamServices.find((item) => item.id === id);
  if (!service) {
    setResponseStatus(event, 404);
    return useResponseError('Stream 服务不存在或已被删除');
  }

  const body = await readBody(event);
  const centerId = Number.parseInt(String(body?.centerId ?? service.centerId), 10);
  const protocol = body?.protocol === 'udp' ? 'udp' : body?.protocol === 'tcp' ? 'tcp' : service.protocol;
  const listenPort = Number.parseInt(String(body?.listenPort ?? service.listenPort), 10);
  const listenAddress = String(body?.listenAddress ?? service.listenAddress).trim();
  const description = String(body?.description ?? service.description).trim();
  const backends = Array.isArray(body?.backends)
    ? body.backends.map((item) => String(item).trim()).filter(Boolean)
    : service.backends;

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
  const conflict = streamPortConflict(centerId, listenPort, protocol, id);
  if (conflict) {
    setResponseStatus(event, 400);
    return useResponseError(
      `端口冲突：${centerName(centerId)} 下已存在 ${protocol.toUpperCase()} 监听端口 ${listenPort}（${conflict.description || '未命名服务'}）`,
    );
  }

  const before = `监听：${service.listenAddress}:${service.listenPort}（${service.protocol.toUpperCase()}）；后端：${service.backends.join(', ')}`;
  Object.assign(service, {
    backends,
    centerId,
    description,
    listenAddress: listenAddress || '0.0.0.0',
    listenPort,
    protocol,
  });
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'update',
    detail: `变更前【${before}】→ 变更后【监听：${listenAddress}:${listenPort}（${protocol.toUpperCase()}）；后端：${backends.join(', ')}】`,
    module: 'stream',
    target: `stream:${protocol}/${listenPort}`,
  });
  return useResponseSuccess(service);
});
