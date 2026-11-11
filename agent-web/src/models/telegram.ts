export type TelegramQR = { status: string; image: string | null; error: string | null };
export type TelegramScan = {
  status: string;
  current: number;
  total: number;
  name: string;
  message: string;
};
export type TelegramSummary = {
  total: number;
  groups_count: number;
  channels_count: number;
  abnormal: number;
  muted: number;
  archived: number;
  created_by_me: number;
  unread: number;
};
export type TelegramStatus = {
  authorized: boolean;
  configured: boolean;
  account: string | null;
  summary: TelegramSummary;
  latest_scan: {
    id: number;
    started_at: string;
    finished_at: string | null;
    status: string;
    dialogs_total: number;
    dialogs_success: number;
    dialogs_failed: number;
    error_message: string | null;
  } | null;
  qr: TelegramQR;
};
export type TelegramDialog = {
  id: number;
  telegram_id: number;
  name: string;
  username: string | null;
  dialog_type: string;
  member_count: number | null;
  is_group: boolean;
  is_broadcast: boolean;
  is_archived: boolean;
  is_muted: boolean;
  unread_count: number;
  last_message_at: string | null;
  status: string;
  is_creator: boolean;
  tags: string[];
  telegram_url: string | null;
};
export type TelegramDialogPage = {
  list: TelegramDialog[];
  total: number;
  page: number;
  page_size: number;
};

export type TelegramSettings = {
  telegram_api_configured: boolean;
  authorized: boolean;
  account: string | null;
  proxy: string;
  listen: string;
  resource_max_file_size: number;
  resource_concurrency: number;
  adult_mode: boolean;
  adult_dialog_count: number;
  ai_base_url: string;
  ai_model: string;
  ai_api_key_configured: boolean;
  ai_api_key_masked: string;
};

export type TelegramAISettingsInput = { base_url: string; api_key: string; model: string };

export type TelegramResourceDialog = {
  id: number;
  telegramId: number;
  name: string;
  username: string | null;
  dialogType: string;
  isGroup: boolean;
  isBroadcast: boolean;
};
export type TelegramFilePreviewKind = "image" | "video" | "audio" | "pdf" | "text";
export type TelegramFileTypeOption = {
  value: string;
  label: string;
  category: "document" | "image" | "video" | "audio" | "archive";
  preview: TelegramFilePreviewKind | null;
};
export type TelegramResourceOptions = {
  dialogs: TelegramResourceDialog[];
  file_types: TelegramFileTypeOption[];
  max_file_size: number;
};
export type TelegramResourceTaskOptions = {
  dialogs: TelegramResourceDialog[];
  fileTypes: string[];
  maxFileSize: number;
};
export type TelegramResourceTaskInput = {
  keyword: string;
  fileTypes: string[];
  sourceMode: "smart" | "all" | "manual";
  sourceDialogIds: number[];
  maxDownloads: number;
};
export type TelegramTaskFailure = {
  stage: string;
  code: string;
  message: string;
  sourceId?: number;
  retryable: boolean;
  retryAfterSeconds?: number;
  durationMs?: number;
};
export type TelegramResourceTask = TelegramResourceTaskInput & {
  id: number;
  status: "queued" | "running" | "completed" | "failed" | "cancelled";
  filesFound: number;
  filesCompleted: number;
  errorMessage: string | null;
  errorCode: string;
  errorStage: string;
  failureCount: number;
  failureDetails: TelegramTaskFailure[];
  createdAt: string;
  startedAt: string | null;
  completedAt: string | null;
  cancelRequested: boolean;
  sourceStrategy: string;
  sourceDialogIds: number[];
};
export type TelegramResourceTaskPage = {
  list: TelegramResourceTask[];
  total: number;
  page: number;
  pageSize: number;
};

export type TelegramResourceFile = {
  id: number;
  task_id: number | null;
  original_filename: string;
  file_type: string;
  file_size: number;
  mime_type: string | null;
  download_status: "queued" | "downloading" | "success" | "failed" | "cancelled" | "duplicate";
  upload_status: "pending" | "uploading" | "success" | "failed" | "cancelled";
  bytes_downloaded: number;
  attempts: number;
  error_message: string | null;
  created_at: string;
  completed_at: string | null;
  source_name: string;
  source_username: string | null;
};
export type TelegramResourceFilePage = {
  list: TelegramResourceFile[];
  total: number;
  page: number;
  page_size: number;
};
