from __future__ import annotations

import asyncio
import logging
import os
import secrets
import socket
from contextlib import asynccontextmanager, suppress
from datetime import datetime
from pathlib import Path
from urllib.parse import urlencode

import uvicorn
from dotenv import load_dotenv
from fastapi import FastAPI, Form, HTTPException, Request
from fastapi.responses import HTMLResponse, JSONResponse, RedirectResponse, Response
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel, Field, field_validator

from analyzer import cleanup_tags, telegram_link
from database import Database
from resource_service import DEFAULT_TYPES, ResourceService
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
csrf_token = secrets.token_urlsafe(32)


@asynccontextmanager
async def lifespan(_app):
    # Telegram may be unreachable; local Web UI must still start immediately.
    connect_task = asyncio.create_task(service.connect()) if service.has_session else None
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
    file_types: list[str] = Field(min_length=1, max_length=30)
    source_dialog_ids: list[int] = Field(min_length=1, max_length=100)
    max_downloads: int = Field(default=10, ge=1, le=100)

    @field_validator("file_types")
    @classmethod
    def validate_types(cls, values: list[str]) -> list[str]:
        cleaned = list(dict.fromkeys(value.lower().lstrip(".").strip() for value in values))
        if any(value not in DEFAULT_TYPES for value in cleaned):
            raise ValueError("仅支持常见文档、图片和 zip 类型")
        return cleaned


def resource_access(request: Request):
    internal_token = os.getenv("TELEGRAM_INTERNAL_TOKEN", "")
    if internal_token and secrets.compare_digest(request.headers.get("x-internal-token", ""), internal_token):
        return
    require_login()


def check_csrf(token: str):
    if not secrets.compare_digest(token, csrf_token):
        raise HTTPException(403, "请求令牌无效，请刷新页面")


def filters_from(request: Request):
    return {key: request.query_params.get(key, "") for key in ("q", "filter", "members_min", "members_max")}


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


@app.get("/api/resources/options")
async def resource_options(request: Request):
    resource_access(request)
    return {"dialogs": db.resource_dialogs(), "file_types": sorted(DEFAULT_TYPES), "max_file_size": resources.max_file_size}


@app.post("/api/resources/tasks", status_code=201)
async def create_resource_task(payload: ResourceTaskInput, request: Request):
    resource_access(request)
    if not service.authorized:
        raise HTTPException(409, "Telegram 未登录，请先在 Telegram 服务页面完成登录")
    try:
        task = db.create_resource_task(payload.keyword.strip(), payload.file_types, payload.source_dialog_ids, payload.max_downloads)
    except ValueError as exc:
        raise HTTPException(400, str(exc)) from exc
    resources.enqueue(task["id"])
    return task


@app.get("/api/resources/tasks")
async def resource_tasks(request: Request, page: int = 1, page_size: int = 20):
    resource_access(request)
    if page < 1 or not 1 <= page_size <= 100:
        raise HTTPException(400, "分页参数无效")
    rows, total = db.list_resource_tasks(page, page_size)
    return {"list": rows, "total": total, "page": page, "page_size": page_size}


@app.get("/api/resources/tasks/{task_id}")
async def resource_task(task_id: int, request: Request):
    resource_access(request)
    task = db.resource_task(task_id)
    if not task:
        raise HTTPException(404, "任务不存在")
    return task


@app.post("/api/resources/tasks/{task_id}/cancel")
async def cancel_resource_task(task_id: int, request: Request):
    resource_access(request)
    try:
        resources.cancel(task_id)
    except ValueError as exc:
        raise HTTPException(409, str(exc)) from exc
    return {"ok": True}


@app.post("/api/resources/tasks/{task_id}/retry")
async def retry_resource_task(task_id: int, request: Request):
    resource_access(request)
    try:
        resources.retry(task_id)
    except ValueError as exc:
        raise HTTPException(409, str(exc)) from exc
    return {"ok": True}


@app.get("/api/resources/files")
async def resource_files(request: Request, page: int = 1, page_size: int = 20, task_id: int | None = None, filename: str = "", file_type: str = "", status: str = "", source_id: int | None = None):
    resource_access(request)
    if page < 1 or not 1 <= page_size <= 100:
        raise HTTPException(400, "分页参数无效")
    filters = {"task_id": task_id, "filename": filename[:200], "file_type": file_type[:12], "status": status[:20], "source_id": source_id}
    rows, total = db.list_resource_files(filters, page, page_size)
    return {"list": rows, "total": total, "page": page, "page_size": page_size}


def resource_redirect(file_id: int, request: Request, inline: bool):
    resource_access(request)
    item = db.resource_file(file_id)
    if not item or item["upload_status"] != "success" or not item["object_key"]:
        raise HTTPException(404, "文件不存在或尚未上传完成")
    previewable = (item["mime_type"] or "").startswith(("image/", "text/")) or item["file_type"] == "pdf"
    if inline and not previewable:
        raise HTTPException(415, "该文件类型不支持在线预览")
    return RedirectResponse(resources.storage.presigned_url(item["object_key"], item["original_filename"], inline), 307)


@app.get("/api/resources/files/{file_id}/preview")
async def preview_resource_file(file_id: int, request: Request):
    return resource_redirect(file_id, request, True)


@app.get("/api/resources/files/{file_id}/download")
async def download_resource_file(file_id: int, request: Request):
    return resource_redirect(file_id, request, False)


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
