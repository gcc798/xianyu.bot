#!/usr/bin/env python3
"""调用 Tika HTTP 服务，对本地文档样本记录可复现的解析结果。"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import time
import urllib.error
import urllib.request


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--endpoint", default="http://127.0.0.1:19998/tika")
    parser.add_argument("--input-dir", type=pathlib.Path, action="append", required=True)
    parser.add_argument("--output", type=pathlib.Path, required=True)
    parser.add_argument("--timeout", type=float, default=120)
    return parser.parse_args()


def parse_document(endpoint: str, path: pathlib.Path, timeout: float) -> dict[str, object]:
    content = path.read_bytes()
    request = urllib.request.Request(
        endpoint,
        data=content,
        method="PUT",
        headers={
            "Accept": "text/plain; charset=utf-8",
            "Content-Type": "application/octet-stream",
            "Content-Disposition": f'attachment; filename="{path.name}"',
            "User-Agent": "xianyu.bot-rag-eval/1",
        },
    )
    started = time.perf_counter()
    status = 0
    body = b""
    error = ""
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            status = response.status
            body = response.read()
    except urllib.error.HTTPError as caught:
        status = caught.code
        body = caught.read()
        error = str(caught)
    except (OSError, TimeoutError) as caught:
        error = str(caught)
    elapsed_millis = (time.perf_counter() - started) * 1000
    text = body.decode("utf-8", errors="replace")
    return {
        "path": str(path),
        "sha256": hashlib.sha256(content).hexdigest(),
        "sourceBytes": len(content),
        "httpStatus": status,
        "elapsedMillis": elapsed_millis,
        "extractedCharacters": len(text.strip()),
        "extractedSha256": hashlib.sha256(body).hexdigest(),
        "error": error,
        "preview": text.strip()[:200],
    }


def main() -> None:
    args = parse_args()
    paths = sorted(
        path
        for directory in args.input_dir
        for path in directory.iterdir()
        if path.is_file() and path.name not in {"manifest.json"}
    )
    if not paths:
        raise SystemExit("没有找到待解析文档")
    report = {
        "endpoint": args.endpoint,
        "generatedAt": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
        "limitations": [
            "没有逐文档金标准，不能据此声明文本准确率、表格完整率或阅读顺序正确率",
            "资源占用需要结合容器级监控单独记录",
        ],
        "documents": [
            parse_document(args.endpoint, path, args.timeout) for path in paths
        ],
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(
        json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(args.output)


if __name__ == "__main__":
    main()
