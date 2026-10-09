/** Stream 流量管理 API */
import type {
  LogFormatConfig,
  OrpDirective,
  PageParams,
  PageResult,
  StreamService,
  StreamServicePayload,
} from './types';

import { requestClient } from '#/api/request';

export function listStreamServicesApi(
  params: PageParams & { centerId?: number; protocol?: string },
): Promise<PageResult<StreamService>> {
  return requestClient.get('/orp/stream-services', { params });
}

export function createStreamServiceApi(
  data: StreamServicePayload,
): Promise<StreamService> {
  return requestClient.post('/orp/stream-services', data);
}

export function updateStreamServiceApi(
  id: number,
  data: StreamServicePayload,
): Promise<StreamService> {
  return requestClient.put(`/orp/stream-services/${id}`, data);
}

export function deleteStreamServiceApi(id: number): Promise<null> {
  return requestClient.delete(`/orp/stream-services/${id}`);
}

/** 获取 stream {} 上下文全局指令列表 */
export function getStreamDirectivesApi(): Promise<OrpDirective[]> {
  return requestClient.get('/orp/stream-directives');
}

/** 保存 stream {} 上下文全局指令列表 */
export function updateStreamDirectivesApi(data: {
  directives: OrpDirective[];
}): Promise<OrpDirective[]> {
  return requestClient.put('/orp/stream-directives', data);
}

/** 获取 stream 块 log_format 日志格式配置 */
export function getStreamLogFormatApi(): Promise<LogFormatConfig> {
  return requestClient.get('/orp/stream-log-format');
}

/** 保存 stream 块 log_format 日志格式配置 */
export function updateStreamLogFormatApi(
  data: LogFormatConfig,
): Promise<LogFormatConfig> {
  return requestClient.put('/orp/stream-log-format', data);
}
