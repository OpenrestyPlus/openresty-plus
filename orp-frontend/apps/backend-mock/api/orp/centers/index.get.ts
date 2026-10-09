import { defineEventHandler, getQuery } from 'h3';

import { centers } from '~/utils/orp-store';
import { usePageResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const query = getQuery(event);
  const keyword = String(query.keyword ?? '').trim();
  const page = Number.parseInt(String(query.page ?? 1), 10) || 1;
  const pageSize = Number.parseInt(String(query.pageSize ?? 20), 10) || 20;

  let list = [...centers].sort((a, b) => b.id - a.id);
  if (keyword) {
    list = list.filter(
      (item) => item.name.includes(keyword) || item.code.includes(keyword),
    );
  }
  return usePageResponseSuccess(page, pageSize, list);
});
