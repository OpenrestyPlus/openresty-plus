/** TLS 证书管理 API */
import type {
  Certificate,
  CertificatePayload,
  PageParams,
  PageResult,
} from './types';

import { requestClient } from '#/api/request';

export function listCertificatesApi(
  params: PageParams & { keyword?: string },
): Promise<PageResult<Certificate>> {
  return requestClient.get('/orp/certificates', { params });
}

export function createCertificateApi(
  data: CertificatePayload,
): Promise<Certificate> {
  return requestClient.post('/orp/certificates', data);
}

export function updateCertificateApi(
  id: number,
  data: CertificatePayload,
): Promise<Certificate> {
  return requestClient.put(`/orp/certificates/${id}`, data);
}

export function deleteCertificateApi(id: number): Promise<null> {
  return requestClient.delete(`/orp/certificates/${id}`);
}

/** 临期/过期证书观察条目 */
export interface CertificateWatchItem {
  days: number;
  domains: string;
  id: number;
  name: string;
  notAfter: string;
}

/** 证书状态预警统计 */
export interface CertificateStats {
  criticalCount: number;
  expiredCount: number;
  expiringCount: number;
  total: number;
  watchlist: CertificateWatchItem[];
}

/** 证书状态预警统计（总数 / 临期 / 紧急 / 已过期） */
export function getCertificateStatsApi(): Promise<CertificateStats> {
  return requestClient.get('/orp/certificates/stats');
}
