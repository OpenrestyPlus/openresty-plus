/** IP 组管理 API */
import type {
  OrpIpGroup,
  OrpIpGroupPayload,
  PageParams,
  PageResult,
} from './types';

import { requestClient } from '#/api/request';

export function listIpGroupsApi(
  params: PageParams & { keyword?: string },
): Promise<PageResult<OrpIpGroup>> {
  return requestClient.get('/orp/ip-groups', { params });
}

export function createIpGroupApi(
  data: OrpIpGroupPayload,
): Promise<OrpIpGroup> {
  return requestClient.post('/orp/ip-groups', data);
}

export function updateIpGroupApi(
  id: number,
  data: OrpIpGroupPayload,
): Promise<OrpIpGroup> {
  return requestClient.put(`/orp/ip-groups/${id}`, data);
}

export function deleteIpGroupApi(id: number): Promise<null> {
  return requestClient.delete(`/orp/ip-groups/${id}`);
}
