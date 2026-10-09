import { requestClient } from '#/api/request';

export type AlertChannelType = 'feishu' | 'webhook';
export type AlertMessageFormat = 'card' | 'text';

export interface AlertChannel {
  createdAt: string;
  enabled: boolean;
  hasSecret: boolean;
  id: number;
  messageFormat: AlertMessageFormat;
  name: string;
  targetHint: string;
  type: AlertChannelType;
  updatedAt: string;
}

export interface AlertChannelPayload {
  enabled: boolean;
  messageFormat?: AlertMessageFormat;
  name: string;
  secret?: string;
  type: AlertChannelType;
  webhookUrl?: string;
}

export interface TLSAlertRule {
  certificateId: number;
  channelIds: number[];
  channelNames: string[];
  daysBefore: number;
  daysRemaining: number;
  domains: string;
  enabled: boolean;
  name: string;
  notAfter: string;
}

export interface TLSAlertRulePayload {
  channelIds: number[];
  daysBefore: number;
  enabled: boolean;
}

export interface AlertDelivery {
  attemptCount: number;
  certificateId?: number;
  certificateName: string;
  channelId: number;
  channelName: string;
  content: string;
  createdAt: string;
  deliveredAt?: string;
  errorText: string;
  eventType: string;
  id: number;
  responseStatus?: number;
  responseText: string;
  status: 'failed' | 'pending' | 'sending' | 'sent';
  title: string;
}

export interface AlertDeliveryPage {
  items: AlertDelivery[];
  page: number;
  pageSize: number;
  total: number;
}

export interface TLSAlertSweepResult {
  certificatesChecked: number;
  errors?: string[];
  notificationsFailed: number;
  notificationsSent: number;
}

export function listAlertChannelsApi(): Promise<AlertChannel[]> {
  return requestClient.get('/orp/alert-channels');
}

export function createAlertChannelApi(
  payload: AlertChannelPayload,
): Promise<AlertChannel> {
  return requestClient.post('/orp/alert-channels', payload);
}

export function updateAlertChannelApi(
  id: number,
  payload: AlertChannelPayload,
): Promise<AlertChannel> {
  return requestClient.put(`/orp/alert-channels/${id}`, payload);
}

export function deleteAlertChannelApi(id: number): Promise<unknown> {
  return requestClient.delete(`/orp/alert-channels/${id}`);
}

export function testAlertChannelApi(id: number): Promise<AlertDelivery> {
  return requestClient.post(`/orp/alert-channels/${id}/test`);
}

export function listTLSAlertRulesApi(): Promise<TLSAlertRule[]> {
  return requestClient.get('/orp/tls-alert-rules');
}

export function saveTLSAlertRuleApi(
  certificateId: number,
  payload: TLSAlertRulePayload,
): Promise<unknown> {
  return requestClient.put(`/orp/tls-alert-rules/${certificateId}`, payload);
}

export function listAlertDeliveriesApi(
  pageSize = 30,
): Promise<AlertDeliveryPage> {
  return requestClient.get('/orp/alert-deliveries', {
    params: { page: 1, pageSize },
  });
}

export function checkTLSAlertsApi(): Promise<TLSAlertSweepResult> {
  return requestClient.post('/orp/tls-alerts/check');
}
