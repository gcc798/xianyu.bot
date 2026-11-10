import type { EmptyResponse, PageResult } from "../models/http";
import type {
  CreateUserRequest,
  Dictionary,
  DictionaryDetail,
  DictionaryDetailInput,
  DictionaryInput,
  DictionaryOptions,
  LoginLog,
  UpdateUserRequest,
  UserRecord,
  UserSearch,
} from "../models/system";
import { queryString, request } from "./client";

export const systemApi = {
  dictionaryOptions: () => request<DictionaryOptions>("/sysDictionary/options"),
  users: (input: UserSearch) =>
    request<PageResult<UserRecord>>("/user/getUserList", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  createUser: (input: CreateUserRequest) =>
    request<{ user: UserRecord }>("/user/admin_register", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  updateUser: (input: UpdateUserRequest) =>
    request<EmptyResponse>("/user/setUserInfo", {
      method: "PUT",
      body: JSON.stringify(input),
    }),
  deleteUser: (id: number) =>
    request<EmptyResponse>("/user/deleteUser", {
      method: "DELETE",
      body: JSON.stringify({ id }),
    }),
  resetPassword: (ID: number, password: string) =>
    request<EmptyResponse>("/user/resetPassword", {
      method: "POST",
      body: JSON.stringify({ ID, password }),
    }),
  dictionaries: (name = "") =>
    request<Dictionary[]>(`/sysDictionary/getSysDictionaryList${queryString({ name })}`),
  createDictionary: (input: DictionaryInput) =>
    request<EmptyResponse>("/sysDictionary/createSysDictionary", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  updateDictionary: (input: DictionaryInput) =>
    request<EmptyResponse>("/sysDictionary/updateSysDictionary", {
      method: "PUT",
      body: JSON.stringify(input),
    }),
  deleteDictionary: (ID: number) =>
    request<EmptyResponse>("/sysDictionary/deleteSysDictionary", {
      method: "DELETE",
      body: JSON.stringify({ ID }),
    }),
  dictionaryDetails: (sysDictionaryID: number) =>
    request<PageResult<DictionaryDetail>>(
      `/sysDictionaryDetail/getSysDictionaryDetailList${queryString({ page: 1, pageSize: 100, sysDictionaryID })}`,
    ),
  createDictionaryDetail: (input: DictionaryDetailInput) =>
    request<EmptyResponse>("/sysDictionaryDetail/createSysDictionaryDetail", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  updateDictionaryDetail: (input: DictionaryDetailInput) =>
    request<EmptyResponse>("/sysDictionaryDetail/updateSysDictionaryDetail", {
      method: "PUT",
      body: JSON.stringify(input),
    }),
  deleteDictionaryDetail: (ID: number) =>
    request<EmptyResponse>("/sysDictionaryDetail/deleteSysDictionaryDetail", {
      method: "DELETE",
      body: JSON.stringify({ ID }),
    }),
  loginLogs: (page: number, pageSize: number, username: string) =>
    request<PageResult<LoginLog>>(
      `/sysLoginLog/getLoginLogList${queryString({ page, pageSize, username })}`,
    ),
  deleteLoginLog: (ID: number) =>
    request<EmptyResponse>("/sysLoginLog/deleteLoginLog", {
      method: "DELETE",
      body: JSON.stringify({ ID }),
    }),
};
