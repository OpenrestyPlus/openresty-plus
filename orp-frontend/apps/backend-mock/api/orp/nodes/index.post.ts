import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, readBody, setResponseStatus } from 'h3';

import type { OrpNodeStatus } from '~/utils/orp-store';

import { centerName, centers, nextId, nodes, now, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const HOST_PATTERN = /^[a-zA-Z0-9][a-zA-Z0-9.-]*$/;
const STATUS_LIST: OrpNodeStatus[] = ['maintenance', 'offline', 'online'];

export default defineEventHandler(async (event) => {
  const body = await readBody(event);
  const centerId = Number.parseInt(String(body?.centerId ?? ''), 10);
  const name = String(body?.name ?? '').trim();
  const host = String(body?.host ?? '').trim();
  const controlEndpoint = String(body?.controlEndpoint ?? '').trim();
  const osInfo = String(body?.osInfo ?? '').trim();
  const activeVersion = String(body?.activeVersion ?? '').trim();

  if (!name || !host || !centerId) {
    setResponseStatus(event, 400);
    return useResponseError('节点名称、主机地址与所属中心为必填项');
  }
  if (!centers.some((item) => item.id === centerId)) {
    setResponseStatus(event, 400);
    return useResponseError('所属中心不存在，请刷新后重试');
  }
  if (!HOST_PATTERN.test(host)) {
    setResponseStatus(event, 400);
    return useResponseError('主机地址格式不正确（支持 IPv4 或主机名）');
  }
  if (nodes.some((item) => item.name === name)) {
    setResponseStatus(event, 400);
    return useResponseError(`节点名称 ${name} 已存在`);
  }
  const status: OrpNodeStatus = STATUS_LIST.includes(body?.status)
    ? body.status
    : 'offline';

  const node = {
    activeVersion: activeVersion || '未发布',
    centerId,
    controlEndpoint: controlEndpoint || '未登记',
    createdAt: now(),
    host,
    id: nextId(nodes),
    name,
    osInfo: osInfo || '未登记',
    status,
  };
  nodes.push(node);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'create',
    detail: `名称：${name}；地址：${host}；所属中心：${centerName(centerId)}`,
    module: 'node',
    target: `node:${name}`,
  });
  return useResponseSuccess(node);
});
