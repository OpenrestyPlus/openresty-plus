/** Upstream 上游服务器组管理 API */
import type {
  PageParams,
  PageResult,
  UpstreamGroup,
  UpstreamGroupPayload,
} from './types';

import { requestClient } from '#/api/request';

export function listUpstreamGroupsApi(
  params: PageParams & { centerId?: number; keyword?: string },
): Promise<PageResult<UpstreamGroup>> {
  return requestClient.get('/orp/upstream-groups', { params });
}

export function createUpstreamGroupApi(
  data: UpstreamGroupPayload,
): Promise<UpstreamGroup> {
  return requestClient.post('/orp/upstream-groups', data);
}

export function updateUpstreamGroupApi(
  id: number,
  data: UpstreamGroupPayload,
): Promise<UpstreamGroup> {
  return requestClient.put(`/orp/upstream-groups/${id}`, data);
}

export function deleteUpstreamGroupApi(id: number): Promise<null> {
  return requestClient.delete(`/orp/upstream-groups/${id}`);
}
