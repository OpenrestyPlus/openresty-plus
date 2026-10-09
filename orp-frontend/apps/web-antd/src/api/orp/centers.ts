/** 中心管理 API */
import type { Center, CenterPayload, PageParams, PageResult } from './types';

import { requestClient } from '#/api/request';

/** 中心列表（支持关键字检索） */
export function listCentersApi(
  params: PageParams & { keyword?: string },
): Promise<PageResult<Center>> {
  return requestClient.get('/orp/centers', { params });
}

export function createCenterApi(data: CenterPayload): Promise<Center> {
  return requestClient.post('/orp/centers', data);
}

export function updateCenterApi(
  id: number,
  data: CenterPayload,
): Promise<Center> {
  return requestClient.put(`/orp/centers/${id}`, data);
}

export function deleteCenterApi(id: number): Promise<null> {
  return requestClient.delete(`/orp/centers/${id}`);
}
