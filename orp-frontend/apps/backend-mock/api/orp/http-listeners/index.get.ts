import { defineEventHandler, getQuery } from 'h3';

import { centerName, certificates, httpListeners } from '~/utils/orp-store';
import { usePageResponseSuccess } from '~/utils/response';

export default defineEventHandler((event) => {
  const query = getQuery(event);
  const centerId = Number.parseInt(String(query.centerId ?? ''), 10);
  const keyword = String(query.keyword ?? '').trim();
  const page = Number.parseInt(String(query.page ?? 1), 10) || 1;
  const pageSize = Number.parseInt(String(query.pageSize ?? 20), 10) || 20;

  let list = [...httpListeners].sort((a, b) => a.id - b.id);
  if (centerId) {
    list = list.filter((item) => item.centerId === centerId);
  }
  if (keyword) {
    list = list.filter((item) => item.domain.includes(keyword));
  }
  const items = list.map((item) => ({
    ...item,
    centerName: centerName(item.centerId),
    certName: item.tlsCertId
      ? certificates.find((cert) => cert.id === item.tlsCertId)?.name ?? '-'
      : '-',
    routeCount: item.routes.length,
  }));
  return usePageResponseSuccess(page, pageSize, items);
});
