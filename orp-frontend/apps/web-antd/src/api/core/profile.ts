import { requestClient } from '#/api/request';

/**
 * 个人中心相关 API（个人信息 / 密码 / OTP 两步验证 / 账户操作审计）
 */

export namespace ProfileApi {
  /** 个人信息（服务端返回掩码手机号/邮箱） */
  export interface ProfileInfo {
    avatarText?: string;
    createdAt?: string;
    email?: string;
    gender?: string;
    homePath?: string;
    lastLoginAt?: string;
    nickname?: string;
    otpEnabled?: boolean;
    phone?: string;
    realName?: string;
    roles?: string[];
    username?: string;
  }

  export interface ProfileUpdateParams {
    email?: string;
    gender?: string;
    nickname?: string;
    phone?: string;
  }

  export interface PasswordUpdateParams {
    newPassword: string;
    oldPassword: string;
  }

  export interface OtpSetupResult {
    otpSecret: string;
    otpauthUrl: string;
  }

  /** 个人中心操作审计条目（复用审计日志结构） */
  export interface ProfileAuditItem {
    action: string;
    createdAt: string;
    detail: string;
    id: number;
    ip: string;
    module: string;
    operator: string;
    target: string;
  }

  /** 分页返回（sw Mock E() 结构） */
  export interface PageResult<T> {
    items: T[];
    total: number;
  }
}

/** 获取当前登录用户个人信息 */
export function getProfileApi() {
  return requestClient.get<ProfileApi.ProfileInfo>('/user/profile');
}

/** 更新个人信息（昵称/性别/手机/邮箱） */
export function updateProfileApi(data: ProfileApi.ProfileUpdateParams) {
  return requestClient.put<ProfileApi.ProfileInfo>('/user/profile', data);
}

/** 修改登录密码 */
export function updatePasswordApi(data: ProfileApi.PasswordUpdateParams) {
  return requestClient.put<null>('/user/password', data);
}

/** 生成 OTP 绑定密钥（二维码内容 + Base32 密钥） */
export function setupOtpApi() {
  return requestClient.get<ProfileApi.OtpSetupResult>('/user/otp/setup');
}

/** 启用 OTP 两步验证 */
export function enableOtpApi(otpCode: string) {
  return requestClient.post<null>('/user/otp/enable', { otpCode });
}

/** 关闭 OTP 两步验证 */
export function disableOtpApi(otpCode: string) {
  return requestClient.post<null>('/user/otp/disable', { otpCode });
}

/** 当前用户操作审计（分页） */
export function getProfileAuditApi(params: { page?: number; pageSize?: number } = {}) {
  return requestClient.get<ProfileApi.PageResult<ProfileApi.ProfileAuditItem>>(
    '/user/profile/audit',
    { params },
  );
}
