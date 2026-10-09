import { getClientIp, getOperator } from '~/utils/orp-event';

import {
  defineEventHandler,
  getRouterParam,
  readBody,
  setResponseStatus,
} from 'h3';

import { centers, now, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const CODE_PATTERN = /^[a-z0-9-]{2,32}$/;

export default defineEventHandler(async (event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const center = centers.find((item) => item.id === id);
  if (!center) {
    setResponseStatus(event, 404);
    return useResponseError('中心不存在或已被删除');
  }

  const body = await readBody(event);
  const code = String(body?.code ?? center.code).trim();
  const name = String(body?.name ?? '').trim();
  const description = String(body?.description ?? '').trim();

  if (!name) {
    setResponseStatus(event, 400);
    return useResponseError('中心名称为必填项');
  }
  if (!CODE_PATTERN.test(code)) {
    setResponseStatus(event, 400);
    return useResponseError(
      '标识码格式不正确：仅允许小写字母、数字与中划线，长度 2-32',
    );
  }
  if (centers.some((item) => item.code === code && item.id !== id)) {
    setResponseStatus(event, 400);
    return useResponseError(`中心标识码 ${code} 已存在`);
  }

  const before = `名称：${center.name}；标识：${center.code}；描述：${center.description || '无'}`;
  Object.assign(center, { code, description, name, updatedAt: now() });
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'update',
    detail: `变更前【${before}】→ 变更后【名称：${name}；标识：${code}；描述：${description || '无'}】`,
    module: 'center',
    target: `center:${code}`,
  });
  return useResponseSuccess(center);
});
