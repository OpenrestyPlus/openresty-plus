import { defineEventHandler, getQuery } from 'h3';

import { centerName, dnsResolvers } from '~/utils/orp-store';
import { usePageResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const query = getQuery(event);
  const centerId = Number.parseInt(String(query.centerId ?? ''), 10);
  const page = Number.parseInt(String(query.page ?? 1), 10) || 1;
  const pageSize = Number.parseInt(String(query.pageSize ?? 20), 10) || 20;

  let list = [...dnsResolvers].sort((a, b) => a.id - b.id);
  if (centerId) {
    list = list.filter((item) => item.centerId === centerId);
  }
  const items = list.map((item) => ({
    ...item,
    centerName: centerName(item.centerId),
  }));
  return usePageResponseSuccess(page, pageSize, items);
});
