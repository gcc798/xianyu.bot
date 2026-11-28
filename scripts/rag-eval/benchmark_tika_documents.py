#!/usr/bin/env python3
"""调用 Tika HTTP 服务，对本地文档样本记录可复现的解析结果。"""

from __future__ import annotations

import argparse
import hashlib
import html.parser
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
    parser.add_argument("--golden", type=pathlib.Path)
    parser.add_argument("--ocr-language", default="")
    parser.add_argument("--timeout", type=float, default=120)
    return parser.parse_args()


class XHTMLText(html.parser.HTMLParser):
    """XHTMLText 提取可见文本并统计解析器保留的表格结构。"""

    def __init__(self) -> None:
        super().__init__()
        self.text: list[str] = []
        self.table_count = 0

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        if tag.lower() == "table":
            self.table_count += 1

    def handle_data(self, data: str) -> None:
        if data.strip():
            self.text.append(data)


def normalized(value: str) -> str:
    punctuation = str.maketrans({"：": ":", "，": ",", "。": ".", "（": "(", "）": ")"})
    return "".join(value.translate(punctuation).split())


def evaluate_golden(
    text: str,
    table_count: int,
    http_status: int,
    expectation: dict[str, object] | None,
) -> dict[str, object] | None:
    if expectation is None:
        return None
    compact_text = normalized(text)
    snippets = [str(value) for value in expectation.get("expectedSnippets", [])]
    ordered = [str(value) for value in expectation.get("orderedSnippets", [])]
    matched = [snippet for snippet in snippets if normalized(snippet) in compact_text]

    cursor = 0
    reading_order_passed = True
    for snippet in ordered:
        position = compact_text.find(normalized(snippet), cursor)
        if position < 0:
            reading_order_passed = False
            break
        cursor = position + len(normalized(snippet))

    expected_status = int(expectation.get("expectedHTTPStatus", 200))
    min_tables = int(expectation.get("minTables", 0))
    snippet_recall = len(matched) / len(snippets) if snippets else 1.0
    passed = (
        http_status == expected_status
        and snippet_recall == 1.0
        and reading_order_passed
        and table_count >= min_tables
    )
    return {
        "passed": passed,
        "snippetRecall": snippet_recall,
        "missingSnippets": [snippet for snippet in snippets if snippet not in matched],
        "readingOrderPassed": reading_order_passed,
        "tableStructurePassed": table_count >= min_tables,
        "expectedMinTables": min_tables,
        "requiresOCR": bool(expectation.get("requiresOCR", False)),
    }


def parse_document(
    endpoint: str,
    path: pathlib.Path,
    timeout: float,
    expectation: dict[str, object] | None,
    ocr_language: str,
) -> dict[str, object]:
    content = path.read_bytes()
    headers = {
        "Accept": "text/html; charset=utf-8",
        "Content-Type": "application/octet-stream",
        "Content-Disposition": f'attachment; filename="{path.name}"',
        "User-Agent": "xianyu.bot-rag-eval/1",
    }
    if ocr_language:
        headers["X-Tika-OCRLanguage"] = ocr_language
    request = urllib.request.Request(
        endpoint,
        data=content,
        method="PUT",
        headers=headers,
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
    xhtml = body.decode("utf-8", errors="replace")
    parsed = XHTMLText()
    parsed.feed(xhtml)
    text = "\n".join(parsed.text)
    return {
        "path": str(path),
        "sha256": hashlib.sha256(content).hexdigest(),
        "sourceBytes": len(content),
        "httpStatus": status,
        "elapsedMillis": elapsed_millis,
        "extractedCharacters": len(text.strip()),
        "tableCount": parsed.table_count,
        "extractedSha256": hashlib.sha256(body).hexdigest(),
        "error": error,
        "preview": text.strip()[:200],
        "golden": evaluate_golden(text, parsed.table_count, status, expectation),
    }


def load_goldens(path: pathlib.Path | None) -> dict[str, dict[str, object]]:
    if path is None:
        return {}
    manifest = json.loads(path.read_text(encoding="utf-8"))
    documents = manifest.get("documents")
    if not isinstance(documents, list):
        raise SystemExit("金标准缺少 documents 数组")
    goldens: dict[str, dict[str, object]] = {}
    for document in documents:
        if not isinstance(document, dict) or not document.get("file"):
            raise SystemExit("金标准包含无效文档定义")
        goldens[str(document["file"])] = document
    return goldens


def main() -> None:
    args = parse_args()
    goldens = load_goldens(args.golden)
    paths = sorted(
        path
        for directory in args.input_dir
        for path in directory.iterdir()
        if path.is_file() and path.name not in {"manifest.json", "golden.json"}
    )
    if not paths:
        raise SystemExit("没有找到待解析文档")
    documents = [
        parse_document(
            args.endpoint,
            path,
            args.timeout,
            goldens.get(path.name),
            args.ocr_language,
        )
        for path in paths
    ]
    evaluated = [document["golden"] for document in documents if document["golden"] is not None]
    report = {
        "endpoint": args.endpoint,
        "generatedAt": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
        "golden": str(args.golden) if args.golden else None,
        "ocrLanguage": args.ocr_language or None,
        "summary": {
            "documentCount": len(documents),
            "evaluatedCount": len(evaluated),
            "passedCount": sum(bool(result["passed"]) for result in evaluated),
        },
        "limitations": ["资源占用需要结合容器级监控单独记录"],
        "documents": documents,
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(
        json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(args.output)


if __name__ == "__main__":
    main()
