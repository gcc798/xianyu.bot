export type TelegramQR = { status: string; image: string | null; error: string | null };
export type TelegramScan = { status: string; current: number; total: number; name: string; message: string };
export type TelegramSummary = {
  total: number; groups_count: number; channels_count: number; abnormal: number;
  muted: number; archived: number; created_by_me: number; unread: number;
};
export type TelegramStatus = {
  authorized: boolean;
  configured: boolean;
  account: string | null;
  summary: TelegramSummary;
  latest_scan: {
    id: number; started_at: string; finished_at: string | null; status: string;
    dialogs_total: number; dialogs_success: number; dialogs_failed: number; error_message: string | null;
  } | null;
  qr: TelegramQR;
};
export type TelegramDialog = {
  id: number; telegram_id: number; name: string; username: string | null; dialog_type: string;
  member_count: number | null; is_group: boolean; is_broadcast: boolean; is_archived: boolean;
  is_muted: boolean; unread_count: number; last_message_at: string | null; status: string;
  is_creator: boolean; tags: string[]; telegram_url: string | null;
};
export type TelegramDialogPage = { list: TelegramDialog[]; total: number; page: number; page_size: number };
