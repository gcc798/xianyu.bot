export const ADMIN_AUTHORITY_ID = 888;
export const USER_AUTHORITY_ID = 9528;

export type CurrentUser = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  uuid: string;
  userName: string;
  nickName: string;
  headerImg: string;
  authorityId: number;
  phone: string;
  email: string;
  enable: number;
};

export type LoginRequest = {
  username: string;
  password: string;
};

export type LoginResponse = {
  user: CurrentUser;
  token: string;
  expiresAt: number;
};

export type AuthSession = LoginResponse;
