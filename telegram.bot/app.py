from __future__ import annotations

import asyncio
import logging
import os
import secrets
import socket
from contextlib import asynccontextmanager, suppress
from datetime import datetime
from pathlib import Path
from typing import Literal
from urllib.parse import urlencode, urlparse

import uvicorn
from botocore.exceptions import BotoCoreError, ClientError
from dotenv import load_dotenv
from fastapi import FastAPI, Form, HTTPException, Request
from fastapi.responses import HTMLResponse, JSONResponse, RedirectResponse, Response
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel, Field, field_validator, model_validator

from analyzer import cleanup_tags, telegram_link
from database import Database
from mcp_server import build_mcp
from resource_service import DEFAULT_TYPES, FILE_TYPE_DICTIONARY, ResourceService
from storage import RustFS
from telegram_service import TelegramService


ROOT = Path(__file__).resolve().parent
load_dotenv(ROOT / ".env")
(ROOT / "exports").mkdir(exist_ok=True)
logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")
db = Database(os.getenv("DATABASE_URL", "postgresql://xianyu:xianyu_assistant@postgres:5432/xianyu_assistant"))
api_id_text, api_hash = os.getenv("TG_API_ID", ""), os.getenv("TG_API_HASH", "")
api_id = int(api_id_text) if api_id_text.isdigit() and api_id_text != "123456" else None
api_hash = api_hash if api_hash and api_hash != "replace_with_your_api_hash" else None
proxy_host, proxy_port = os.getenv("TG_PROXY_HOST", "").strip(), os.getenv("TG_PROXY_PORT", "").strip()
proxy = ("socks5", proxy_host, int(proxy_port)) if proxy_host and proxy_port.isdigit() else None
service = TelegramService(api_id, api_hash, db, proxy)
resources = ResourceService(service, db, RustFS())
service.refresh_adult_dialogs = resources.refresh_adult_dialogs
telegram_mcp, telegram_mcp_app = build_mcp(service, resources, db)
csrf_token = secrets.token_urlsafe(32)


@asynccontextmanager
async def lifespan(_app):
    # Telegram may be unreachable; local Web UI must still start immediately.
    connect_task = asyncio.create_task(service.connect()) if service.has_session else None
    async with telegram_mcp.session_manager.run():
        await resources.start()
        yield
        if connect_task and not connect_task.done():
            connect_task.cancel()
            with suppress(asyncio.CancelledError):
                await connect_task
        await resources.stop()
        await service.disconnect()


app = FastAPI(title="Telegram Analyzer & Cleaner", lifespan=lifespan)
app.mount("/static", StaticFiles(directory=ROOT / "static"), name="static")
app.mount("/mcp", telegram_mcp_app)
templates = Jinja2Templates(directory=ROOT / "templates")


@app.get("/healthz")
async def healthz():
    try:
        database_ok = await asyncio.to_thread(db.health)
        await asyncio.to_thread(resources.storage.ensure_bucket)
    except Exception as exc:
        raise HTTPException(503, "依赖服务不可用") from exc
    return {"status": "ok", "database": database_ok, "rustfs": True, "telegram_authorized": service.authorized}


@app.middleware("http")
async def disable_script_cache(request: Request, call_next):
    if request.url.path.startswith("/mcp"):
        internal_token = os.getenv("TELEGRAM_INTERNAL_TOKEN", "")
        if not internal_token:
            return JSONResponse({"detail": "服务访问令牌未配置"}, status_code=503)
        if not secrets.compare_digest(request.headers.get("x-internal-token", ""), internal_token):
            return JSONResponse({"detail": "访问令牌无效"}, status_code=403)
    response = await call_next(request)
    if request.url.path == "/static/app.js":
        response.headers["Cache-Control"] = "no-store"
    return response


def context(request: Request, **values):
    return {"request": request, "authorized": service.authorized, "account": db.setting("account_name"), "csrf": csrf_token, **values}


def require_login():
    if not service.authorized:
        raise HTTPException(401, "Telegram 未登录")


class ResourceTaskInput(BaseModel):
    keyword: str = Field(min_length=2, max_length=200)
    file_types: list[str] = Field(default_factory=list, max_length=len(DEFAULT_TYPES))
    source_mode: Literal["smart", "all", "manual"] = "smart"
    source_dialog_ids: list[int] = Field(default_factory=list, max_length=100)
    max_downloads: int = Field(default=10, ge=1, le=100)

    @field_validator("file_types")
    @classmethod
    def validate_types(cls, values: list[str]) -> list[str]:
        cleaned = list(dict.fromkeys(value.lower().lstrip(".").strip() for value in values))
        if any(value not in DEFAULT_TYPES for value in cleaned):
            raise ValueError("文件类型不在资源字典中")
        return cleaned

    @model_validator(mode="after")
    def validate_sources(self):
        if self.source_mode == "manual" and not self.source_dialog_ids:
            raise ValueError("手动选择模式至少需要一个搜索来源")
        return self


class TelegramPasswordInput(BaseModel):
    password: str = Field(min_length=1, max_length=256)


class AdultModeInput(BaseModel):
    adult_mode: bool


class AISettingsInput(BaseModel):
    base_url: str = Field(min_length=1, max_length=1000)
    api_key: str = Field(default="", max_length=4096)
    model: str = Field(min_length=1, max_length=200)

    @model_validator(mode="after")
    def validate_ai_settings(self):
        self.base_url = self.base_url.strip().rstrip("/")
        self.api_key = self.api_key.strip()
        self.model = self.model.strip()
        parsed = urlparse(self.base_url)
        if parsed.scheme not in {"http", "https"} or not parsed.netloc:
            raise ValueError("AI 服务地址必须是有效的 HTTP 或 HTTPS 地址")
        if not self.model:
            raise ValueError("AI 模型不能为空")
        return self


class LeaveDialogsInput(BaseModel):
    dialog_ids: list[int] = Field(min_length=1, max_length=100)
    confirm: str


def internal_access(request: Request):
    internal_token = os.getenv("TELEGRAM_INTERNAL_TOKEN", "")
    # shortcut: an unset token relies on the Docker port staying loopback-only, require a token before external exposure.
    if not internal_token or secrets.compare_digest(request.headers.get("x-internal-token", ""), internal_token):
        return
    raise HTTPException(403, "Telegram 内部访问令牌未配置或不匹配")


def check_csrf(token: str):
    if not secrets.compare_digest(token, csrf_token):
        raise HTTPException(403, "请求令牌无效，请刷新页面")


def filters_from(request: Request):
    return {key: request.query_params.get(key, "") for key in ("q", "filter", "members_min", "members_max")}


def public_ai_settings() -> dict:
    config = resources.ai_settings()
    masked_key = f"sk-{'•' * 12}" if config["api_key"].startswith("sk-") else "•" * 12
    return {
        "ai_base_url": config["base_url"],
        "ai_model": config["model"],
        "ai_api_key_configured": bool(config["api_key"]),
        "ai_api_key_masked": masked_key if config["api_key"] else "",
    }


def requested_ai_settings(payload: AISettingsInput) -> dict[str, str]:
    current = resources.ai_settings()
    api_key = payload.api_key or current["api_key"]
    if not api_key:
        raise HTTPException(400, "请填写 AI API Key")
    return {"base_url": payload.base_url, "api_key": api_key, "model": payload.model}


@app.get("/", response_class=HTMLResponse)
async def index(request: Request):
    if not service.authorized:
        return templates.TemplateResponse("login.html", context(request, qr=service.qr, configured=bool(service.client)))
    return templates.TemplateResponse("index.html", context(request, summary=db.summary(), latest=db.latest_scan()))


@app.post("/login/qr")
async def login_qr(csrf: str = Form(...)):
    check_csrf(csrf)
    try:
        service.start_qr()
    except Exception as exc:
        service.qr.update(status="error", error=str(exc))
    return RedirectResponse("/", 303)


@app.get("/login/qr")
async def login_qr_redirect():
    return RedirectResponse("/", 303)


@app.get("/api/login/status")
async def login_status():
    return {"status": service.qr["status"], "error": service.qr["error"]}


@app.post("/login/2fa")
async def login_2fa(password: str = Form(...), csrf: str = Form(...)):
    check_csrf(csrf)
    await service.submit_2fa(password)
    return RedirectResponse("/", 303)


@app.post("/scan")
async def start_scan(csrf: str = Form(...)):
    require_login(); check_csrf(csrf)
    service.start_scan()
    return RedirectResponse("/", 303)


@app.get("/api/scan/status")
async def scan_status():
    require_login()
    return service.progress


@app.get("/groups", response_class=HTMLResponse)
async def groups(request: Request, sort: str = "last_message", order: str = "desc"):
    require_login()
    filters = filters_from(request)
    rows, total = db.list_dialogs(filters, page_size=100_000, sort=sort, order=order)
    for row in rows: row.update(tags=cleanup_tags(row), telegram_url=telegram_link(row))
    query = dict(request.query_params)
    query.pop("sort", None); query.pop("order", None)
    query.pop("leave_success", None); query.pop("leave_failed", None)
    return templates.TemplateResponse("groups.html", context(request, rows=rows, total=total, sort=sort, order=order, filters=filters, query=urlencode(query)))


@app.get("/rankings", response_class=HTMLResponse)
async def rankings(request: Request):
    require_login()
    return RedirectResponse("/groups?sort=last_message&order=desc", 303)


@app.get("/cleanup")
async def cleanup():
    require_login()
    return RedirectResponse("/groups?sort=last_message&order=asc", 303)


@app.post("/cleanup/leave", response_class=HTMLResponse)
async def cleanup_leave(request: Request, selected: list[int] = Form(default=[]), confirm: str = Form(...), csrf: str = Form(...)):
    require_login(); check_csrf(csrf)
    if confirm.strip() != "退出":
        raise HTTPException(400, "请输入“退出”确认")
    success = failed = 0
    for item_id in selected:
        item = db.dialog(item_id)
        if not item or item["status"] == "left": continue
        ok, _error = await service.leave(item)
        success += int(ok); failed += int(not ok)
        await asyncio.sleep(1)
    return RedirectResponse(f"/groups?leave_success={success}&leave_failed={failed}", 303)


@app.get("/export.csv")
async def export_csv(request: Request):
    require_login()
    content = "\ufeff" + db.export_csv(filters_from(request))
    return Response(content, media_type="text/csv; charset=utf-8", headers={"Content-Disposition": f'attachment; filename="telegram-dialogs-{datetime.now():%Y%m%d-%H%M%S}.csv"'})


@app.get("/settings", response_class=HTMLResponse)
async def settings(request: Request):
    return templates.TemplateResponse("settings.html", context(request, configured=bool(service.client), host=os.getenv("WEB_HOST", "0.0.0.0"), port=os.getenv("WEB_PORT", "8080"), proxy=service.proxy))


@app.get("/api/admin/status")
async def admin_status(request: Request):
    internal_access(request)
    return {
        "authorized": service.authorized,
        "configured": bool(service.client),
        "account": db.setting("account_name"),
        "summary": db.summary(),
        "latest_scan": db.latest_scan(),
        "qr": service.qr,
    }


@app.post("/api/admin/login/qr")
async def admin_login_qr(request: Request):
    internal_access(request)
    try:
        service.start_qr()
    except Exception as exc:
        service.qr.update(status="error", error=str(exc))
        raise HTTPException(409, str(exc)) from exc
    return service.qr


@app.get("/api/admin/login/status")
async def admin_login_status(request: Request):
    internal_access(request)
    return service.qr


@app.post("/api/admin/login/2fa")
async def admin_login_2fa(payload: TelegramPasswordInput, request: Request):
    internal_access(request)
    await service.submit_2fa(payload.password)
    return service.qr


@app.post("/api/admin/scan")
async def admin_start_scan(request: Request):
    internal_access(request)
    require_login()
    scan_id = service.start_scan()
    return {"started": scan_id is not None, "task_id": scan_id, "progress": service.progress}


@app.get("/api/admin/scan/status")
async def admin_scan_status(request: Request):
    internal_access(request)
    require_login()
    return service.progress


@app.get("/api/admin/dialogs")
async def admin_dialogs(
    request: Request,
    page: int = 1,
    page_size: int = 30,
    q: str = "",
    filter: str = "",
    members_min: int | None = None,
    members_max: int | None = None,
    sort: str = "last_message",
    order: str = "desc",
):
    internal_access(request)
    require_login()
    if page < 1 or not 1 <= page_size <= 100:
        raise HTTPException(400, "分页参数无效")
    filters = {"q": q[:200], "filter": filter, "members_min": members_min, "members_max": members_max}
    rows, total = db.list_dialogs(filters, page, page_size, sort, order)
    for row in rows:
        row.update(tags=cleanup_tags(row), telegram_url=telegram_link(row))
    return {"list": rows, "total": total, "page": page, "page_size": page_size}


@app.post("/api/admin/dialogs/leave")
async def admin_leave_dialogs(payload: LeaveDialogsInput, request: Request):
    internal_access(request)
    require_login()
    if payload.confirm.strip() != "退出":
        raise HTTPException(400, "请输入“退出”确认")
    success = failed = 0
    for dialog_id in dict.fromkeys(payload.dialog_ids):
        item = db.dialog(dialog_id)
        if not item or item["status"] == "left":
            continue
        ok, _error = await service.leave(item)
        success += int(ok)
        failed += int(not ok)
        await asyncio.sleep(1)
    return {"success": success, "failed": failed}


@app.get("/api/admin/dialogs/export")
async def admin_export_dialogs(request: Request):
    internal_access(request)
    require_login()
    content = "\ufeff" + db.export_csv(filters_from(request))
    filename = f"telegram-dialogs-{datetime.now():%Y%m%d-%H%M%S}.csv"
    return Response(content, media_type="text/csv; charset=utf-8", headers={"Content-Disposition": f'attachment; filename="{filename}"'})


@app.get("/api/admin/settings")
async def admin_settings(request: Request):
    internal_access(request)
    return {
        "telegram_api_configured": bool(service.client),
        "authorized": service.authorized,
        "account": db.setting("account_name"),
        "proxy": f"{service.proxy[1]}:{service.proxy[2]}" if service.proxy else "",
        "listen": f"{os.getenv('WEB_HOST', '0.0.0.0')}:{os.getenv('WEB_PORT', '8080')}",
        "resource_max_file_size": resources.max_file_size,
        "resource_concurrency": resources.concurrency,
        "adult_mode": db.adult_mode(),
        "adult_dialog_count": len(db.adult_dialog_ids()),
        **public_ai_settings(),
    }


@app.post("/api/admin/settings/adult-mode")
async def update_adult_mode(payload: AdultModeInput, request: Request):
    internal_access(request)
    db.set_adult_mode(payload.adult_mode)
    return {"adult_mode": db.adult_mode(), "adult_dialog_count": len(db.adult_dialog_ids())}


@app.post("/api/admin/settings/ai")
async def update_ai_settings(payload: AISettingsInput, request: Request):
    internal_access(request)
    requested_ai_settings(payload)
    db.setting("ai_base_url", payload.base_url)
    db.setting("ai_model", payload.model)
    if payload.api_key:
        db.setting("ai_api_key", payload.api_key)
    return public_ai_settings()


@app.post("/api/admin/settings/ai/probe")
async def probe_ai_settings(payload: AISettingsInput, request: Request):
    internal_access(request)
    return {"available": await resources.probe_ai(requested_ai_settings(payload))}


@app.get("/api/resources/options")
async def resource_options(request: Request):
    internal_access(request)
    file_types = [{"value": value, **metadata} for value, metadata in FILE_TYPE_DICTIONARY.items()]
    return {"dialogs": db.resource_dialogs(), "file_types": file_types, "max_file_size": resources.max_file_size}


@app.post("/api/resources/tasks", status_code=201)
async def create_resource_task(payload: ResourceTaskInput, request: Request):
    internal_access(request)
    if not service.authorized:
        raise HTTPException(409, "Telegram 未登录，请先在 Telegram 服务页面完成登录")
    try:
        source_ids, strategy = await resources.plan_sources(payload.keyword.strip(), payload.source_mode, payload.source_dialog_ids)
        task = db.create_resource_task(payload.keyword.strip(), payload.file_types, source_ids, payload.max_downloads, payload.source_mode, strategy)
    except ValueError as exc:
        raise HTTPException(400, str(exc)) from exc
    resources.enqueue(task["id"])
    return task


@app.get("/api/resources/tasks")
async def resource_tasks(request: Request, page: int = 1, page_size: int = 20, q: str = "", status: str = ""):
    internal_access(request)
    if page < 1 or not 1 <= page_size <= 100:
        raise HTTPException(400, "分页参数无效")
    if status and status not in {"queued", "running", "completed", "failed", "cancelled"}:
        raise HTTPException(400, "任务状态无效")
    filters = {"q": q.strip()[:200], "status": status}
    rows, total = db.list_resource_tasks(filters, page, page_size)
    return {"list": rows, "total": total, "page": page, "page_size": page_size}


@app.get("/api/resources/tasks/{task_id}")
async def resource_task(task_id: int, request: Request):
    internal_access(request)
    task = db.resource_task(task_id)
    if not task:
        raise HTTPException(404, "任务不存在")
    return task


@app.post("/api/resources/tasks/{task_id}/cancel")
async def cancel_resource_task(task_id: int, request: Request):
    internal_access(request)
    try:
        resources.cancel(task_id)
    except ValueError as exc:
        raise HTTPException(409, str(exc)) from exc
    return {"ok": True}


@app.post("/api/resources/tasks/{task_id}/retry")
async def retry_resource_task(task_id: int, request: Request):
    internal_access(request)
    try:
        resources.retry(task_id)
    except ValueError as exc:
        raise HTTPException(409, str(exc)) from exc
    return {"ok": True}


@app.delete("/api/resources/tasks/{task_id}")
async def delete_resource_task(task_id: int, request: Request, delete_files: bool = False):
    internal_access(request)
    try:
        resources.delete_task(task_id, delete_files)
    except ValueError as exc:
        raise HTTPException(409, str(exc)) from exc
    except (BotoCoreError, ClientError) as exc:
        raise HTTPException(502, "RustFS 文件删除失败") from exc
    return {"ok": True}


@app.get("/api/resources/files")
async def resource_files(request: Request, page: int = 1, page_size: int = 20, task_id: int | None = None, filename: str = "", file_type: str = "", status: str = "", source_id: int | None = None):
    internal_access(request)
    if page < 1 or not 1 <= page_size <= 100:
        raise HTTPException(400, "分页参数无效")
    filters = {"task_id": task_id, "filename": filename[:200], "file_type": file_type[:12], "status": status[:20], "source_id": source_id}
    rows, total = db.list_resource_files(filters, page, page_size)
    return {"list": rows, "total": total, "page": page, "page_size": page_size}


def resource_redirect(file_id: int, request: Request, inline: bool):
    internal_access(request)
    item = db.resource_file(file_id)
    if not item or item["upload_status"] != "success" or not item["object_key"]:
        raise HTTPException(404, "文件不存在或尚未上传完成")
    previewable = FILE_TYPE_DICTIONARY.get(item["file_type"], {}).get("preview") is not None
    if inline and not previewable:
        raise HTTPException(415, "该文件类型不支持在线预览")
    return RedirectResponse(resources.storage.presigned_url(item["object_key"], item["original_filename"], inline), 307)


@app.get("/api/resources/files/{file_id}/preview")
async def preview_resource_file(file_id: int, request: Request):
    return resource_redirect(file_id, request, True)


@app.get("/api/resources/files/{file_id}/download")
async def download_resource_file(file_id: int, request: Request):
    return resource_redirect(file_id, request, False)


@app.delete("/api/resources/files/{file_id}")
async def delete_resource_file(file_id: int, request: Request):
    internal_access(request)
    try:
        resources.delete_file(file_id)
    except ValueError as exc:
        raise HTTPException(409, str(exc)) from exc
    except (BotoCoreError, ClientError) as exc:
        raise HTTPException(502, "RustFS 文件删除失败") from exc
    return {"ok": True}


def local_ip():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    try:
        sock.connect(("8.8.8.8", 80)); return sock.getsockname()[0]
    except OSError:
        return "127.0.0.1"
    finally:
        sock.close()


if __name__ == "__main__":
    host, port = os.getenv("WEB_HOST", "0.0.0.0"), int(os.getenv("WEB_PORT", "8080"))
    print(f"Telegram Analyzer & Cleaner\nLocal:   http://127.0.0.1:{port}\nNetwork: http://{local_ip()}:{port}")
    uvicorn.run(app, host=host, port=port)
