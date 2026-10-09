/** 系统配置管理 API */
import type {
  OrpSetting,
  OrpSettingGroup,
  PageParams,
  PageResult,
  SettingGroupPayload,
  SettingPayload,
} from './types';

import { requestClient } from '#/api/request';

/** 系统配置列表（支持分组筛选与名称/Key 检索） */
export function listSettingsApi(
  params: PageParams & { group?: string; keyword?: string },
): Promise<PageResult<OrpSetting>> {
  return requestClient.get('/orp/settings', { params });
}

/** 全部配置分组（去重） */
export function listSettingGroupsApi(): Promise<string[]> {
  return requestClient.get('/orp/settings/groups');
}

/** 配置项明文详情（敏感信息编辑时查看） */
export function getSettingPlainApi(id: number): Promise<OrpSetting> {
  return requestClient.get(`/orp/settings/${id}/plain`);
}

export function createSettingApi(data: SettingPayload): Promise<OrpSetting> {
  return requestClient.post('/orp/settings', data);
}

export function updateSettingApi(
  id: number,
  data: SettingPayload,
): Promise<OrpSetting> {
  return requestClient.put(`/orp/settings/${id}`, data);
}

/** 启用/禁用状态切换 */
export function toggleSettingStatusApi(
  id: number,
  status: 'disabled' | 'enabled',
): Promise<OrpSetting> {
  return requestClient.put(`/orp/settings/${id}/status`, { status });
}

export function deleteSettingApi(id: number): Promise<null> {
  return requestClient.delete(`/orp/settings/${id}`);
}

/** 分组实体列表（含类型/序号/组内配置计数，按 sortOrder 升序） */
export function listSettingGroupEntitiesApi(): Promise<OrpSettingGroup[]> {
  return requestClient.get('/orp/setting-groups');
}

export function createSettingGroupApi(
  data: SettingGroupPayload,
): Promise<OrpSettingGroup> {
  return requestClient.post('/orp/setting-groups', data);
}

/** 编辑分组：系统分组仅允许调整序号与描述；自定义分组重命名会级联更新组内配置项归属 */
export function updateSettingGroupApi(
  id: number,
  data: SettingGroupPayload,
): Promise<OrpSettingGroup> {
  return requestClient.put(`/orp/setting-groups/${id}`, data);
}

/** 拖拽排序保存：按传入 id 顺序重编序号 1..n */
export function reorderSettingGroupsApi(
  ids: number[],
): Promise<OrpSettingGroup[]> {
  return requestClient.put('/orp/setting-groups/reorder', { ids });
}

/** 删除自定义分组：组内含配置项时必须指定处理方式（move 转移 / cascade 级联） */
export function deleteSettingGroupApi(
  id: number,
  options?: {
    mode?: 'cascade' | 'move';
    riskConfirmed?: boolean;
    targetGroup?: string;
  },
): Promise<null> {
  return requestClient.delete(`/orp/setting-groups/${id}`, {
    params: options,
  });
}

/** 批量移动配置项至指定分组 */
export function batchMoveSettingsApi(
  ids: number[],
  targetGroup: string,
): Promise<{ moved: number; targetGroup: string }> {
  return requestClient.post('/orp/settings/batch-move', {
    ids,
    targetGroup,
  });
}
