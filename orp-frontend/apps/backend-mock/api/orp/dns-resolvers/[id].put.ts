import { getClientIp, getOperator } from '~/utils/orp-event';

import {
  defineEventHandler,
  getRouterParam,
  readBody,
  setResponseStatus,
} from 'h3';

import { centerName, dnsResolvers, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const IPV4_PATTERN = /^(\d{1,3}\.){3}\d{1,3}$/;

export default defineEventHandler(async (event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const resolver = dnsResolvers.find((item) => item.id === id);
  if (!resolver) {
    setResponseStatus(event, 404);
    return useResponseError('解析器不存在或已被删除');
  }

  const body = await readBody(event);
  const centerId = Number.parseInt(String(body?.centerId ?? resolver.centerId), 10);
  const address = String(body?.address ?? resolver.address).trim();
  const port = Number.parseInt(String(body?.port ?? resolver.port), 10);
  const timeoutSec = Number.parseInt(String(body?.timeoutSec ?? resolver.timeoutSec), 10);
  const cacheTtlSec = Number.parseInt(String(body?.cacheTtlSec ?? resolver.cacheTtlSec), 10);

  if (!centerId || !address) {
    setResponseStatus(event, 400);
    return useResponseError('所属中心与解析服务器地址为必填项');
  }
  if (!IPV4_PATTERN.test(address)) {
    setResponseStatus(event, 400);
    return useResponseError('解析服务器地址必须是 IPv4 格式，示例：10.60.0.11');
  }
  if (!(port >= 1 && port <= 65_535)) {
    setResponseStatus(event, 400);
    return useResponseError('端口必须是 1-65535 之间的整数');
  }
  if (!(timeoutSec >= 1 && timeoutSec <= 60)) {
    setResponseStatus(event, 400);
    return useResponseError('超时时间必须是 1-60 之间的秒数');
  }
  if (!(cacheTtlSec >= 1 && cacheTtlSec <= 3600)) {
    setResponseStatus(event, 400);
    return useResponseError('有效缓存时间必须是 1-3600 之间的秒数');
  }
  if (
    dnsResolvers.some(
      (item) =>
        item.id !== id &&
        item.centerId === centerId &&
        item.address === address &&
        item.port === port,
    )
  ) {
    setResponseStatus(event, 400);
    return useResponseError(`该中心下已存在解析服务器 ${address}:${port}`);
  }

  const before = `地址：${resolver.address}:${resolver.port}；超时：${resolver.timeoutSec}s；缓存：${resolver.cacheTtlSec}s`;
  Object.assign(resolver, { address, cacheTtlSec, centerId, port, timeoutSec });
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'update',
    detail: `变更前【${before}】→ 变更后【地址：${address}:${port}；超时：${timeoutSec}s；缓存：${cacheTtlSec}s；中心：${centerName(centerId)}】`,
    module: 'dns',
    target: `dns:${address}:${port}`,
  });
  return useResponseSuccess(resolver);
});
