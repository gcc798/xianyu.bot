export type XianyuAccount = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  externalUserId: string;
  label: string;
  note: string;
  status: string;
  enabled: boolean;
  hasCredential: boolean;
};

export type AccountSearch = {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: string;
  enabled?: boolean;
};

export type AccountInput = {
  ID?: number;
  label: string;
  note: string;
  enabled: boolean;
};

export type AccountCredentialInput = {
  ID: number;
  cookie: string;
  token: string;
};

export type LoginStatus = {
  status: string;
  cookieUpdated: boolean;
  credentialVersion: number;
};

export type QRLogin = {
  sessionId: string;
  accountId: number;
  status: string;
  qrCodeDataUrl?: string;
  message: string;
  verificationUrl?: string;
};

export type WebSocketStatus = {
  accountId: number;
  state: string;
  queueDepth: number;
  reconnects: number;
  running: boolean;
  lastError: string;
  accountStatus: string;
};
