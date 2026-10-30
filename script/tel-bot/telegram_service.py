from __future__ import annotations

import asyncio
import base64
import io
import logging
import re
from datetime import datetime, timezone

import qrcode
from telethon import TelegramClient, errors, functions, types
from telethon.sessions import StringSession

log = logging.getLogger(__name__)


class TelegramService:
    def __init__(self, api_id: int | None, api_hash: str | None, db, proxy: tuple | None = None):
        self.db = db
        self.proxy = proxy
        session_string = db.setting("telegram_session") or ""
        self.has_session = bool(session_string)
        self.client = TelegramClient(
            StringSession(session_string), api_id, api_hash, flood_sleep_threshold=0,
            connection_retries=0, retry_delay=1, auto_reconnect=False,
            receive_updates=False, proxy=proxy,
        ) if api_id and api_hash else None
        self.authorized = False
        self.qr = {"status": "idle", "image": None, "error": None}
        self.qr_task = None
        self.connect_lock = asyncio.Lock()
        self.scan_task = None
        self.refresh_adult_dialogs = None
        self.progress = {"status": "idle", "current": 0, "total": 0, "name": "", "message": ""}

    async def connect(self):
        if not self.client:
            return
        for attempt in range(3):
            try:
                await self._ensure_connected()
                self.authorized = await self.client.is_user_authorized()
                if self.authorized:
                    me = await self.call(self.client.get_me)
                    self.db.setting("account_name", " ".join(x for x in (me.first_name, me.last_name) if x) or str(me.id))
                return
            except Exception as exc:
                log.warning("Telegram connection failed: %s", exc)
                if attempt < 2:
                    await asyncio.sleep(attempt + 1)

    async def _ensure_connected(self):
        async with self.connect_lock:
            if not self.client.is_connected():
                await asyncio.wait_for(self.client.connect(), timeout=12)

    async def disconnect(self):
        if self.qr_task:
            self.qr_task.cancel()
        if self.client:
            if self.authorized:
                self.db.setting("telegram_session", self.client.session.save())
            await self.client.disconnect()

    async def call(self, func, *args, **kwargs):
        for attempt in range(2):
            try:
                return await func(*args, **kwargs)
            except errors.FloodWaitError as exc:
                if attempt:
                    raise
                log.warning("Telegram FloodWait: waiting %s seconds", exc.seconds)
                await asyncio.sleep(exc.seconds)

    def start_qr(self):
        if not self.client:
            raise RuntimeError("请先在 .env 配置 TG_API_ID 和 TG_API_HASH")
        if self.qr_task:
            self.qr_task.cancel()
        self.qr = {"status": "connecting", "image": None, "error": None}
        self.qr_task = asyncio.create_task(self._start_qr())

    async def _start_qr(self):
        try:
            await self._ensure_connected()
            login = await self.call(self.client.qr_login)
            if not await self._prepare_qr(login):
                return
            image = qrcode.make(login.url)
            buffer = io.BytesIO(); image.save(buffer, format="PNG")
            self.qr = {"status": "waiting", "image": "data:image/png;base64," + base64.b64encode(buffer.getvalue()).decode(), "error": None}
            await self._wait_qr(login)
        except asyncio.CancelledError:
            pass
        except Exception as exc:
            log.warning("QR login start failed: %s", exc)
            self.qr.update(status="error", error=f"无法连接 Telegram：{exc}")

    async def _prepare_qr(self, login):
        # shortcut: Telethon 1.x has no public initial-migration hook; remove when QRLogin handles it.
        response = login._resp
        if isinstance(response, types.auth.LoginTokenMigrateTo):
            await self.client._switch_dc(response.dc_id)
            response = await self.call(self.client, functions.auth.ImportLoginTokenRequest(response.token))
            login._resp = response
        if isinstance(response, types.auth.LoginTokenSuccess):
            await self.client._on_login(response.authorization.user)
            await self._login_done()
            return False
        if not isinstance(response, types.auth.LoginToken):
            raise TypeError(f"Unexpected QR login response: {type(response).__name__}")
        return True

    async def _wait_qr(self, login):
        try:
            seconds = max(1, (login.expires - datetime.now(timezone.utc)).total_seconds())
            await login.wait(timeout=seconds)
            await self._login_done()
        except errors.SessionPasswordNeededError:
            self.qr.update(status="2fa", error=None)
        except (asyncio.TimeoutError, errors.AuthTokenExpiredError):
            self.qr.update(status="expired", error="二维码已过期，请刷新")
        except asyncio.CancelledError:
            pass
        except Exception as exc:
            log.exception("QR login failed")
            self.qr.update(status="error", error=str(exc))

    async def submit_2fa(self, password: str):
        try:
            await self.call(self.client.sign_in, password=password)
            await self._login_done()
        except Exception as exc:
            self.qr.update(status="2fa", error=f"两步验证失败：{exc}")

    async def _login_done(self):
        self.authorized = True
        me = await self.call(self.client.get_me)
        name = " ".join(x for x in (me.first_name, me.last_name) if x) or str(me.id)
        self.db.setting("account_name", name)
        self.db.setting("telegram_session", self.client.session.save())
        self.qr.update(status="success", error=None)

    def start_scan(self):
        if self.scan_task and not self.scan_task.done():
            return None
        scan_id = self.db.start_scan()
        self.scan_task = asyncio.create_task(self.scan(scan_id))
        return scan_id

    def cancel_scan(self, scan_id: int) -> bool:
        task = self.db.scan_history(scan_id)
        if not task or task["status"] != "running" or not self.scan_task or self.scan_task.done():
            return False
        self.scan_task.cancel()
        self.db.update_scan(scan_id, finished_at=datetime.now(timezone.utc).isoformat(), status="cancelled")
        return True

    async def scan(self, scan_id: int):
        self.progress = {"status": "running", "current": 0, "total": 0, "name": "", "message": "正在读取会话列表"}
        success = failed = 0
        try:
            await self._ensure_connected()
            dialogs = [d async for d in self.client.iter_dialogs() if isinstance(d.entity, (types.Chat, types.ChatForbidden, types.Channel, types.ChannelForbidden))]
            self.progress["total"] = len(dialogs)
            self.db.update_scan(scan_id, dialogs_total=len(dialogs))
            for number, dialog in enumerate(dialogs, 1):
                self.progress.update(current=number, name=dialog.name or "(未命名)", message="保存会话信息")
                try:
                    item = self._dialog_data(dialog)
                    self.db.upsert_dialog(item)
                    success += 1
                except Exception:
                    failed += 1
                    log.exception("Failed to scan %s", dialog.name)
                self.db.update_scan(scan_id, dialogs_success=success, dialogs_failed=failed)
            if self.refresh_adult_dialogs:
                try:
                    self.progress.update(name="", message="正在刷新成人频道识别结果")
                    await self.refresh_adult_dialogs()
                except Exception:
                    log.exception("Adult dialog classification failed after scan")
            self.db.update_scan(scan_id, finished_at=datetime.now(timezone.utc).isoformat(), status="completed")
            self.progress.update(status="completed", name="", message=f"扫描完成：成功 {success}，失败 {failed}")
        except Exception as exc:
            log.exception("Scan failed")
            self.db.update_scan(scan_id, finished_at=datetime.now(timezone.utc).isoformat(), status="failed", error_message=str(exc))
            self.progress.update(status="failed", message=str(exc))

    def _dialog_data(self, dialog) -> dict:
        entity = dialog.entity
        forbidden = isinstance(entity, (types.ChatForbidden, types.ChannelForbidden))
        deleted = bool(getattr(entity, "deactivated", False))
        is_channel = isinstance(entity, (types.Channel, types.ChannelForbidden))
        is_megagroup = bool(getattr(entity, "megagroup", False))
        notify = getattr(dialog.dialog, "notify_settings", None)
        mute_until = getattr(notify, "mute_until", None)
        muted = bool(mute_until and (not isinstance(mute_until, datetime) or mute_until > datetime.now(timezone.utc)))
        return {
            "telegram_id": entity.id, "access_hash": getattr(entity, "access_hash", None),
            "name": dialog.name or "(未命名)", "username": getattr(entity, "username", None),
            "dialog_type": "超级群" if is_megagroup else ("频道" if is_channel else "普通群"),
            "member_count": getattr(entity, "participants_count", None), "is_channel": is_channel,
            "is_group": not is_channel or is_megagroup, "is_megagroup": is_megagroup,
            "is_broadcast": bool(getattr(entity, "broadcast", False)), "is_archived": bool(dialog.archived),
            "is_muted": muted, "unread_count": dialog.unread_count,
            "last_message_id": getattr(dialog.message, "id", None),
            "last_message_at": getattr(dialog.message, "date", None).isoformat() if dialog.message else None,
            "last_read_message_id": getattr(dialog.dialog, "read_inbox_max_id", None),
            "is_creator": bool(getattr(entity, "creator", False)),
            "status": "forbidden" if forbidden else ("deleted" if deleted else "normal"),
        }

    async def leave(self, dialog: dict):
        try:
            await self.call(self.client.delete_dialog, dialog["telegram_id"])
            self.db.record_leave(dialog, "success")
            return True, None
        except Exception as exc:
            self.db.record_leave(dialog, "failed", str(exc))
            return False, str(exc)

    async def join(self, reference: str) -> dict:
        value = reference.strip()
        invite = re.fullmatch(r"(?:https?://)?t\.me/(?:joinchat/|\+)([A-Za-z0-9_-]+)", value, re.IGNORECASE)
        if invite:
            updates = await self.call(self.client, functions.messages.ImportChatInviteRequest(invite.group(1)))
            joined_ids = {chat.id for chat in getattr(updates, "chats", [])}
        else:
            username = re.sub(r"^(?:https?://)?t\.me/", "", value, flags=re.IGNORECASE).strip("/@")
            entity = await self.call(self.client.get_entity, username)
            await self.call(self.client, functions.channels.JoinChannelRequest(entity))
            joined_ids = {entity.id}
        dialogs = await self.call(self.client.get_dialogs, limit=None)
        joined = next((dialog for dialog in dialogs if dialog.entity.id in joined_ids), None)
        if joined is None:
            raise RuntimeError("加入成功但未读取到会话")
        dialog_id = self.db.upsert_dialog(self._dialog_data(joined))
        return self.db.dialog(dialog_id)
