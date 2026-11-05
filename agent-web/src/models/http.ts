export type ApiEnvelope<T> = {
  code: number;
  data: T;
  msg: string;
};

export type PageResult<T> = {
  page: number;
  pageSize: number;
  total: number;
  list: T[];
};

export type PageQuery = {
  page: number;
  pageSize: number;
};

export type EmptyResponse = Record<string, never>;
