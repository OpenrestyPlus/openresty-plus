/** DNS 解析器管理 API */
import type {
  DnsResolver,
  DnsResolverPayload,
  PageParams,
  PageResult,
} from './types';

import { requestClient } from '#/api/request';

export function listDnsResolversApi(
  params: PageParams & { centerId?: number },
): Promise<PageResult<DnsResolver>> {
  return requestClient.get('/orp/dns-resolvers', { params });
}

export function createDnsResolverApi(
  data: DnsResolverPayload,
): Promise<DnsResolver> {
  return requestClient.post('/orp/dns-resolvers', data);
}

export function updateDnsResolverApi(
  id: number,
  data: DnsResolverPayload,
): Promise<DnsResolver> {
  return requestClient.put(`/orp/dns-resolvers/${id}`, data);
}

export function deleteDnsResolverApi(id: number): Promise<null> {
  return requestClient.delete(`/orp/dns-resolvers/${id}`);
}
