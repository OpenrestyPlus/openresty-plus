import { defineEventHandler, getQuery } from 'h3';

import { auditLogs } from '~/utils/orp-store';
import { usePageResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const query = getQuery(event);
  const module = String(query.module ?? '');
  const action = String(query.action ?? '');
  const operator = String(query.operator ?? '').trim();
  const startTime = String(query.startTime ?? '');
  const endTime = String(query.endTime ?? '');
  const page = Number.parseInt(String(query.page ?? 1), 10) || 1;
  const pageSize = Number.parseInt(String(query.pageSize ?? 20), 10) || 20;

  let list = [...auditLogs];
  if (module && module !== 'all') {
    list = list.filter((item) => item.module === module);
  }
  if (action && action !== 'all') {
    list = list.filter((item) => item.action === action);
  }
  if (operator) {
    list = list.filter((item) => item.operator.includes(operator));
  }
  if (startTime) {
    list = list.filter((item) => item.createdAt >= startTime);
  }
  if (endTime) {
    list = list.filter((item) => item.createdAt <= `${endTime} 23:59:59`);
  }
  return usePageResponseSuccess(page, pageSize, list);
});
