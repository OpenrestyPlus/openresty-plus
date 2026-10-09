import { defineEventHandler, getQuery } from 'h3';

import { centerName, nodes } from '~/utils/orp-store';
import { usePageResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const query = getQuery(event);
  const centerId = Number.parseInt(String(query.centerId ?? ''), 10);
  const keyword = String(query.keyword ?? '').trim();
  const status = String(query.status ?? '');
  const page = Number.parseInt(String(query.page ?? 1), 10) || 1;
  const pageSize = Number.parseInt(String(query.pageSize ?? 20), 10) || 20;

  let list = [...nodes].sort((a, b) => a.id - b.id);
  if (centerId) {
    list = list.filter((item) => item.centerId === centerId);
  }
  if (status && status !== 'all') {
    list = list.filter((item) => item.status === status);
  }
  if (keyword) {
    list = list.filter(
      (item) => item.name.includes(keyword) || item.host.includes(keyword),
    );
  }
  const items = list.map((item) => ({
    ...item,
    centerName: centerName(item.centerId),
  }));
  return usePageResponseSuccess(page, pageSize, items);
});
