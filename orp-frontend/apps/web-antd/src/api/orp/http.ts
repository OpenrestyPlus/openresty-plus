/** HTTP 流量管理 API */
import type {
  HttpListener,
  HttpListenerPayload,
  LogFormatConfig,
  OrpDirective,
  PageParams,
  PageResult,
} from './types';

import { requestClient } from '#/api/request';

export function listHttpListenersApi(
  params: PageParams & { centerId?: number; keyword?: string },
): Promise<PageResult<HttpListener>> {
  return requestClient.get('/orp/http-listeners', { params });
}

export function createHttpListenerApi(
  data: HttpListenerPayload,
): Promise<HttpListener> {
  return requestClient.post('/orp/http-listeners', data);
}

export function updateHttpListenerApi(
  id: number,
  data: HttpListenerPayload,
): Promise<HttpListener> {
  return requestClient.put(`/orp/http-listeners/${id}`, data);
}

export function deleteHttpListenerApi(id: number): Promise<null> {
  return requestClient.delete(`/orp/http-listeners/${id}`);
}

/** 获取 http {} 上下文全局指令列表 */
export function getHttpDirectivesApi(): Promise<OrpDirective[]> {
  return requestClient.get('/orp/http-directives');
}

/** 保存 http {} 上下文全局指令列表 */
export function updateHttpDirectivesApi(data: {
  directives: OrpDirective[];
}): Promise<OrpDirective[]> {
  return requestClient.put('/orp/http-directives', data);
}

/** 获取 http 块 log_format 日志格式配置 */
export function getHttpLogFormatApi(): Promise<LogFormatConfig> {
  return requestClient.get('/orp/http-log-format');
}

/** 保存 http 块 log_format 日志格式配置 */
export function updateHttpLogFormatApi(
  data: LogFormatConfig,
): Promise<LogFormatConfig> {
  return requestClient.put('/orp/http-log-format', data);
}

/** HTTP 块默认错误页面（键为状态码或 状态码|Content-Type） */
export function getHttpErrorPagesApi(): Promise<Record<string, string>> {
  return requestClient.get('/orp/http-error-pages');
}

/** 保存 HTTP 块默认错误页面 */
export function updateHttpErrorPagesApi(data: {
  errorPages: Record<string, string>;
}): Promise<Record<string, string>> {
  return requestClient.put('/orp/http-error-pages', data);
}
