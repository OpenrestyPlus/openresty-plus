import { defineEventHandler, getQuery } from 'h3';

import { certificates } from '~/utils/orp-store';
import { usePageResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const query = getQuery(event);
  const keyword = String(query.keyword ?? '').trim();
  const page = Number.parseInt(String(query.page ?? 1), 10) || 1;
  const pageSize = Number.parseInt(String(query.pageSize ?? 20), 10) || 20;

  let list = [...certificates].sort((a, b) => a.id - b.id);
  if (keyword) {
    list = list.filter(
      (item) =>
        item.name.includes(keyword) || item.domains.includes(keyword),
    );
  }
  // 私钥不随列表下发
  const items = list.map(({ privateKey: _key, ...rest }) => rest);
  return usePageResponseSuccess(page, pageSize, items);
});
