import { defineEventHandler, getRouterParam, setResponseStatus } from 'h3';

import { certificates } from '~/utils/orp-store';
import { useResponseError, useResponseSuccess } from '~/utils/response';

/** 证书明文详情（编辑回显用，含私钥） */
export default defineEventHandler((event) => {
  const id = Number.parseInt(String(getRouterParam(event, 'id')), 10);
  const cert = certificates.find((item) => item.id === id);
  if (!cert) {
    setResponseStatus(event, 404);
    return useResponseError('证书不存在或已被删除');
  }
  return useResponseSuccess({
    certificate: cert.certificate,
    privateKey: cert.privateKey,
  });
});
