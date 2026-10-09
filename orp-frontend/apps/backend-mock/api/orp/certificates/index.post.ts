import { getClientIp, getOperator } from '~/utils/orp-event';

import { defineEventHandler, readBody, setResponseStatus } from 'h3';

import {
  certificates,
  nextId,
  now,
  recordAudit,
  validateCertificatePem,
} from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

export default defineEventHandler(async (event) => {
  const body = await readBody(event);
  const name = String(body?.name ?? '').trim();
  const domains = String(body?.domains ?? '').trim();
  const notBefore = String(body?.notBefore ?? '').trim();
  const notAfter = String(body?.notAfter ?? '').trim();
  const certificate = String(body?.certificate ?? '').trim();
  const privateKey = String(body?.privateKey ?? '').trim();
  const scopeRaw = body?.centerScope;

  if (!name || !domains || !notBefore || !notAfter) {
    setResponseStatus(event, 400);
    return useResponseError('证书名称、关联域名与有效期范围为必填项');
  }
  const pemError = validateCertificatePem(certificate, privateKey);
  if (pemError) {
    setResponseStatus(event, 400);
    return useResponseError(pemError);
  }
  if (Date.parse(notAfter) <= Date.parse(notBefore)) {
    setResponseStatus(event, 400);
    return useResponseError('有效期范围不正确：结束时间必须晚于开始时间');
  }
  const centerScope: 'all' | number[] =
    scopeRaw === 'all' || scopeRaw === undefined || scopeRaw === null
      ? 'all'
      : Array.isArray(scopeRaw)
        ? scopeRaw.map(Number).filter((item) => Number.isFinite(item))
        : 'all';
  if (centerScope !== 'all' && centerScope.length === 0) {
    setResponseStatus(event, 400);
    return useResponseError('证书适用中心范围不能为空，或选择全部中心');
  }

  const cert = {
    certificate,
    centerScope,
    createdAt: now(),
    domains,
    id: nextId(certificates),
    name,
    notAfter,
    notBefore,
    privateKey,
  };
  certificates.push(cert);
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'create',
    detail: `名称：${name}；域名：${domains}；有效期：${notBefore} ~ ${notAfter}`,
    module: 'tls',
    target: `tls:${name}`,
  });
  return useResponseSuccess({ ...cert, privateKey: '******' });
});
