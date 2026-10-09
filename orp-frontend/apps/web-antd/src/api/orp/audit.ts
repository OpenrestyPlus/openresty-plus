/** 审计日志 API */
import type { AuditAction, AuditLog, AuditModule, PageParams, PageResult } from './types';

import { requestClient } from '#/api/request';

export function listAuditLogsApi(
  params: PageParams & {
    action?: string;
    endTime?: string;
    module?: string;
    operator?: string;
    startTime?: string;
  },
): Promise<PageResult<AuditLog>> {
  return requestClient.get('/orp/audit-logs', { params });
}

export type { AuditAction, AuditLog, AuditModule };
