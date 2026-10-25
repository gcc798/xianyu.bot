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

from analyzer import cleanup_tags, telegram_link
from database import Database
from telegram_service import TelegramService


ROOT = Path(__file__).resolve().parent
load_dotenv(ROOT / ".env")
(ROOT / "data").mkdir(exist_ok=True); (ROOT / "exports").mkdir(exist_ok=True)
logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")
db = Database(ROOT / "data" / "telegram.db")
api_id_text, api_hash = os.getenv("TG_API_ID", ""), os.getenv("TG_API_HASH", "")
api_id = int(api_id_text) if api_id_text.isdigit() and api_id_text != "123456" else None
api_hash = api_hash if api_hash and api_hash != "replace_with_your_api_hash" else None
proxy_host, proxy_port = os.getenv("TG_PROXY_HOST", "").strip(), os.getenv("TG_PROXY_PORT", "").strip()
proxy = ("socks5", proxy_host, int(proxy_port)) if proxy_host and proxy_port.isdigit() else None
service = TelegramService(api_id, api_hash, ROOT / "data" / "telegram", db, proxy)
csrf_token = secrets.token_urlsafe(32)


@asynccontextmanager
async def lifespan(_app):
    # Telegram may be unreachable; local Web UI must still start immediately.
    connect_task = asyncio.create_task(service.connect()) if service.has_session else None
    yield
    if connect_task and not connect_task.done():
        connect_task.cancel()
        with suppress(asyncio.CancelledError):
            await connect_task
    await service.disconnect()


app = FastAPI(title="Telegram Analyzer & Cleaner", lifespan=lifespan)
app.mount("/static", StaticFiles(directory=ROOT / "static"), name="static")
templates = Jinja2Templates(directory=ROOT / "templates")


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
