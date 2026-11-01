import { request } from "./api";

export type DataRow = Record<string, unknown>;
export type PageData = {
  list: DataRow[];
  total: number;
  page: number;
  pageSize: number;
};
export type CurrentUser = DataRow & {
  ID: number;
  userName: string;
  nickName: string;
  headerImg?: string;
};

function query(path: string, params: Record<string, unknown>) {
  const values = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== "" && value !== undefined && value !== null)
      values.set(key, String(value));
  });
  return `${path}?${values}`;
}

export const adminApi = {
  captcha: () =>
    request<{
      captchaId: string;
      picPath: string;
      openCaptcha: boolean;
      captchaLength: number;
    }>("/base/captcha", { method: "POST" }),
  login: (data: DataRow) =>
    request<{ token: string; user: CurrentUser; expiresAt: number }>(
      "/base/login",
      { method: "POST", body: JSON.stringify(data) }
    ),
  logout: () =>
    request<Record<string, never>>("/jwt/jsonInBlacklist", { method: "POST" }),
  currentUser: () => request<{ userInfo: CurrentUser }>("/user/getUserInfo"),

  users: (data: DataRow) =>
    request<PageData>("/user/getUserList", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  createUser: (data: DataRow) =>
    request<DataRow>("/user/admin_register", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateUser: (data: DataRow) =>
    request<DataRow>("/user/setUserInfo", {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  deleteUser: (id: unknown) =>
    request<DataRow>("/user/deleteUser", {
      method: "DELETE",
      body: JSON.stringify({ id }),
    }),
  resetPassword: (ID: unknown, password: string) =>
    request<DataRow>("/user/resetPassword", {
      method: "POST",
      body: JSON.stringify({ ID, password }),
    }),

  authorities: () =>
    request<DataRow[]>("/authority/getAuthorityList", {
      method: "POST",
      body: "{}",
    }),
  createAuthority: (data: DataRow) =>
    request<DataRow>("/authority/createAuthority", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateAuthority: (data: DataRow) =>
    request<DataRow>("/authority/updateAuthority", {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  deleteAuthority: (authorityId: unknown) =>
    request<DataRow>("/authority/deleteAuthority", {
      method: "POST",
      body: JSON.stringify({ authorityId }),
    }),

  menus: (data: DataRow) =>
    request<PageData>("/menu/getMenuList", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  createMenu: (data: DataRow) =>
    request<DataRow>("/menu/addBaseMenu", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateMenu: (data: DataRow) =>
    request<DataRow>("/menu/updateBaseMenu", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  deleteMenu: (ID: unknown) =>
    request<DataRow>("/menu/deleteBaseMenu", {
      method: "POST",
      body: JSON.stringify({ ID }),
    }),

  apis: (data: DataRow) =>
    request<PageData>("/api/getApiList", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  createApi: (data: DataRow) =>
    request<DataRow>("/api/createApi", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateApi: (data: DataRow) =>
    request<DataRow>("/api/updateApi", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  deleteApi: (data: DataRow) =>
    request<DataRow>("/api/deleteApi", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  syncApis: () => request<DataRow>("/api/syncApi"),

  dictionaries: (params: DataRow) =>
    request<DataRow[]>(query("/sysDictionary/getSysDictionaryList", params)),
  createDictionary: (data: DataRow) =>
    request<DataRow>("/sysDictionary/createSysDictionary", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateDictionary: (data: DataRow) =>
    request<DataRow>("/sysDictionary/updateSysDictionary", {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  deleteDictionary: (ID: unknown) =>
    request<DataRow>("/sysDictionary/deleteSysDictionary", {
      method: "DELETE",
      body: JSON.stringify({ ID }),
    }),
  dictionaryDetails: (sysDictionaryID: unknown) =>
    request<PageData>(
      query("/sysDictionaryDetail/getSysDictionaryDetailList", {
        page: 1,
        pageSize: 1000,
        sysDictionaryID,
      })
    ),
  createDictionaryDetail: (data: DataRow) =>
    request<DataRow>("/sysDictionaryDetail/createSysDictionaryDetail", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateDictionaryDetail: (data: DataRow) =>
    request<DataRow>("/sysDictionaryDetail/updateSysDictionaryDetail", {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  deleteDictionaryDetail: (ID: unknown) =>
    request<DataRow>("/sysDictionaryDetail/deleteSysDictionaryDetail", {
      method: "DELETE",
      body: JSON.stringify({ ID }),
    }),

  loginLogs: (params: DataRow) =>
    request<PageData>(query("/sysLoginLog/getLoginLogList", params)),
  operationLogs: (params: DataRow) =>
    request<PageData>(
      query("/sysOperationRecord/getSysOperationRecordList", params)
    ),
  errorLogs: (params: DataRow) =>
    request<PageData>(query("/sysError/getSysErrorList", params)),
  deleteLoginLog: (ID: unknown) =>
    request<DataRow>("/sysLoginLog/deleteLoginLog", {
      method: "DELETE",
      body: JSON.stringify({ ID }),
    }),
  deleteOperationLog: (ID: unknown) =>
    request<DataRow>("/sysOperationRecord/deleteSysOperationRecord", {
      method: "DELETE",
      body: JSON.stringify({ ID }),
    }),
  deleteErrorLog: (ID: unknown) =>
    request<DataRow>(query("/sysError/deleteSysError", { ID }), {
      method: "DELETE",
    }),
};
