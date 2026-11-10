export type TelegramAPIConfig = {
  ID: number;
  name: string;
  apiId: number;
  hasApiHash: boolean;
  floodWaitUntil: string | null;
  lastFloodWaitSeconds: number;
};

export type TelegramAPIConfigInput = {
  name: string;
  apiId: number;
  apiHash: string;
};

export type TelegramAPIConfigPage = {
  page: number;
  pageSize: number;
  total: number;
  list: TelegramAPIConfig[];
};
