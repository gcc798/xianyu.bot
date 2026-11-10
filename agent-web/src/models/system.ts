import type { PageQuery } from "./http";

export type UserRecord = {
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

export type UserSearch = PageQuery & {
  username?: string;
  nickName?: string;
  phone?: string;
  email?: string;
  orderKey: "id" | "created_at" | "updated_at" | "username" | "nickname";
  desc: boolean;
};

export type CreateUserRequest = {
  userName: string;
  passWord: string;
  nickName: string;
  headerImg: string;
  authorityId: number;
  enable: number;
  phone: string;
  email: string;
};

export type UpdateUserRequest = Omit<CreateUserRequest, "userName" | "passWord"> & {
  ID: number;
};

export type Dictionary = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  name: string;
  type: string;
  status: boolean | null;
  desc: string;
  parentID: number | null;
  children: Dictionary[];
  sysDictionaryDetails: DictionaryDetail[];
};

export type DictionaryInput = Pick<Dictionary, "name" | "type" | "status" | "desc" | "parentID"> & {
  ID?: number;
};

export type DictionaryDetail = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  label: string;
  value: string;
  extend: string;
  status: boolean | null;
  sort: number;
  sysDictionaryID: number;
  parentID: number | null;
  children: DictionaryDetail[];
  level: number;
  path: string;
};

export type DictionaryDetailInput = Pick<
  DictionaryDetail,
  "label" | "value" | "extend" | "status" | "sort" | "sysDictionaryID" | "parentID"
> & { ID?: number };

export type DictionaryOption = { code: string; name: string };
export type DictionaryOptions = Record<string, DictionaryOption[]>;

export type LoginLog = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  username: string;
  ip: string;
  status: boolean;
  errorMessage: string;
  agent: string;
  userId: number;
};
