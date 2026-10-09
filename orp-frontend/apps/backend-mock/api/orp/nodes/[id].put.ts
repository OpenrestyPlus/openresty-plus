import { getClientIp, getOperator } from '~/utils/orp-event';

import {
  defineEventHandler,
  getRouterParam,
  readBody,
  setResponseStatus,
} from 'h3';

import type { OrpNodeStatus } from '~/utils/orp-store';

import { centerName, centers, nodes, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const HOST_PATTERN = /^[a-zA-Z0-9][a-zA-Z0-9.-]*$/;
const STATUS_LIST: OrpNodeStatus[] = ['maintenance', 'offline', 'online'];

export default defineEventHandler(async (event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const node = nodes.find((item) => item.id === id);
  if (!node) {
    setResponseStatus(event, 404);
    return useResponseError('节点不存在或已被删除');
  }

  const body = await readBody(event);
  const centerId = Number.parseInt(String(body?.centerId ?? node.centerId), 10);
  const name = String(body?.name ?? node.name).trim();
  const host = String(body?.host ?? node.host).trim();

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
  if (nodes.some((item) => item.name === name && item.id !== id)) {
    setResponseStatus(event, 400);
    return useResponseError(`节点名称 ${name} 已存在`);
  }

  const before = `名称：${node.name}；地址：${node.host}；中心：${centerName(node.centerId)}`;
  node.centerId = centerId;
  node.name = name;
  node.host = host;
  node.controlEndpoint = String(body?.controlEndpoint ?? node.controlEndpoint).trim() || '未登记';
  node.osInfo = String(body?.osInfo ?? node.osInfo).trim() || '未登记';
  node.activeVersion = String(body?.activeVersion ?? node.activeVersion).trim() || '未发布';
  if (STATUS_LIST.includes(body?.status)) {
    node.status = body.status;
  }
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'update',
    detail: `变更前【${before}】→ 变更后【名称：${name}；地址：${host}；中心：${centerName(centerId)}】`,
    module: 'node',
    target: `node:${name}`,
  });
  return useResponseSuccess(node);
});
