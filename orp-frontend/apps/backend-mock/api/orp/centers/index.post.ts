import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, readBody, setResponseStatus } from 'h3';

import { centers, nextId, now, recordAudit } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

const CODE_PATTERN = /^[a-z0-9-]{2,32}$/;

export default defineEventHandler(async (event) => {
  const body = await readBody(event);
  const code = String(body?.code ?? '').trim();
  const name = String(body?.name ?? '').trim();
  const description = String(body?.description ?? '').trim();

  if (!code || !name) {
    setResponseStatus(event, 400);
    return useResponseError('中心名称与标识码为必填项');
  }
  if (!CODE_PATTERN.test(code)) {
    setResponseStatus(event, 400);
    return useResponseError(
      '标识码格式不正确：仅允许小写字母、数字与中划线，长度 2-32',
    );
  }
  if (centers.some((item) => item.code === code)) {
    setResponseStatus(event, 400);
    return useResponseError(`中心标识码 ${code} 已存在`);
  }

  const center = {
    code,
    createdAt: now(),
    description,
    id: nextId(centers),
    name,
    updatedAt: now(),
  };
  centers.push(center);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'create',
    detail: `名称：${name}；标识：${code}；描述：${description || '无'}`,
    module: 'center',
    target: `center:${code}`,
  });
  return useResponseSuccess(center);
});
