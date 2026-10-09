import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, readBody, setResponseStatus } from 'h3';

import {
  centerName,
  centers,
  dnsResolvers,
  nextId,
  recordAudit,
} from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const IPV4_PATTERN = /^(\d{1,3}\.){3}\d{1,3}$/;

export default defineEventHandler(async (event) => {
  const body = await readBody(event);
  const centerId = Number.parseInt(String(body?.centerId ?? ''), 10);
  const address = String(body?.address ?? '').trim();
  const port = Number.parseInt(String(body?.port ?? 53), 10);
  const timeoutSec = Number.parseInt(String(body?.timeoutSec ?? 2), 10);
  const cacheTtlSec = Number.parseInt(String(body?.cacheTtlSec ?? 30), 10);

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
  if (!centers.some((item) => item.id === centerId)) {
    setResponseStatus(event, 400);
    return useResponseError('所属中心不存在，请刷新后重试');
  }
  if (
    dnsResolvers.some(
      (item) =>
        item.centerId === centerId &&
        item.address === address &&
        item.port === port,
    )
  ) {
    setResponseStatus(event, 400);
    return useResponseError(`该中心下已存在解析服务器 ${address}:${port}`);
  }

  const resolver = {
    address,
    cacheTtlSec,
    centerId,
    id: nextId(dnsResolvers),
    port,
    timeoutSec,
  };
  dnsResolvers.push(resolver);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'create',
    detail: `地址：${address}:${port}；超时：${timeoutSec}s；缓存：${cacheTtlSec}s；中心：${centerName(centerId)}`,
    module: 'dns',
    target: `dns:${address}:${port}`,
  });
  return useResponseSuccess(resolver);
});
