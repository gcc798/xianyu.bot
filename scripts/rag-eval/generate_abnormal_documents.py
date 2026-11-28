#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["pypdf==6.3.0"]
# ///
"""生成可重复的空、损坏、加密及错误后缀文档解析样本。"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib

from pypdf import PdfWriter


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=pathlib.Path, required=True)
    return parser.parse_args()


def sha256(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    args = parse_args()
    args.output_dir.mkdir(parents=True, exist_ok=True)

    (args.output_dir / "empty.pdf").write_bytes(b"")
    (args.output_dir / "corrupted.pdf").write_bytes(b"%PDF-1.7\n1 0 obj\ntruncated")
    (args.output_dir / "wrong-suffix.pdf").write_text(
        "这实际上是纯文本，不是 PDF。", encoding="utf-8"
    )
    (args.output_dir / "corrupted.docx").write_bytes(b"PK\x03\x04truncated-ooxml")

    writer = PdfWriter()
    writer.add_blank_page(width=300, height=300)
    writer.encrypt("xianyu-rag-eval")
    with (args.output_dir / "encrypted.pdf").open("wb") as destination:
        writer.write(destination)

    samples = []
    for path in sorted(args.output_dir.iterdir()):
        if path.name == "manifest.json" or not path.is_file():
            continue
        samples.append(
            {"name": path.name, "size": path.stat().st_size, "sha256": sha256(path)}
        )
    manifest = {"generatorVersion": "1", "samples": samples}
    manifest_path = args.output_dir / "manifest.json"
    manifest_path.write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(manifest_path)


if __name__ == "__main__":
    main()
