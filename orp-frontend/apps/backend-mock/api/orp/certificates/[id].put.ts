import { getClientIp, getOperator } from '~/utils/orp-event';

import {
  defineEventHandler,
  getRouterParam,
  readBody,
  setResponseStatus,
} from 'h3';

import { certificates, recordAudit, validateCertificatePem } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

export default defineEventHandler(async (event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const cert = certificates.find((item) => item.id === id);
  if (!cert) {
    setResponseStatus(event, 404);
    return useResponseError('证书不存在或已被删除');
  }

  const body = await readBody(event);
  const name = String(body?.name ?? cert.name).trim();
  const domains = String(body?.domains ?? cert.domains).trim();
  const notBefore = String(body?.notBefore ?? cert.notBefore).trim();
  const notAfter = String(body?.notAfter ?? cert.notAfter).trim();
  const certificate = String(body?.certificate ?? cert.certificate).trim();
  const privateKey = String(body?.privateKey ?? cert.privateKey).trim();
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
      ? cert.centerScope
      : Array.isArray(scopeRaw)
        ? scopeRaw.map(Number).filter((item) => Number.isFinite(item))
        : cert.centerScope;
  if (centerScope !== 'all' && centerScope.length === 0) {
    setResponseStatus(event, 400);
    return useResponseError('证书适用中心范围不能为空，或选择全部中心');
  }

  const before = `名称：${cert.name}；域名：${cert.domains}；有效期：${cert.notBefore} ~ ${cert.notAfter}`;
  Object.assign(cert, {
    certificate,
    centerScope,
    domains,
    name,
    notAfter,
    notBefore,
    privateKey,
  });
  recordAudit(getOperator(event), getClientIp(event), {
    action: 'update',
    detail: `变更前【${before}】→ 变更后【名称：${name}；域名：${domains}；有效期：${notBefore} ~ ${notAfter}】`,
    module: 'tls',
    target: `tls:${name}`,
  });
  return useResponseSuccess({ ...cert, privateKey: '******' });
});
