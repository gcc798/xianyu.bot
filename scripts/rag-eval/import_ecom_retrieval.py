#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["pyarrow==22.0.0"]
# ///
"""按固定版本下载 EcomRetrieval，并转换为项目统一 RAG 评估 JSON。"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import shutil
import tempfile
import urllib.request

import pyarrow.parquet as parquet


DEFAULT_REVISION = "94a400f33e9951f2fa474c1527e29f6a99f7cfe4"
FILES = {
    "corpus": "corpus/dev-00000-of-00001.parquet",
    "queries": "queries/dev-00000-of-00001.parquet",
    "qrels": "data/dev-00000-of-00001.parquet",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=pathlib.Path, required=True)
    parser.add_argument("--revision", default=DEFAULT_REVISION)
    parser.add_argument(
        "--accept-license",
        action="store_true",
        help="确认已审查上游数据许可；数据卡当前未声明明确许可证",
    )
    return parser.parse_args()


def download(url: str, destination: pathlib.Path) -> str:
    destination.parent.mkdir(parents=True, exist_ok=True)
    request = urllib.request.Request(url, headers={"User-Agent": "xianyu.bot-rag-eval/1"})
    with urllib.request.urlopen(request, timeout=120) as response:
        with tempfile.NamedTemporaryFile(dir=destination.parent, delete=False) as temporary:
            shutil.copyfileobj(response, temporary)
            temporary_path = pathlib.Path(temporary.name)
    temporary_path.replace(destination)
    return sha256(destination)


def sha256(path: pathlib.Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def value(row: dict[str, object], *names: str) -> object:
    for name in names:
        if name in row:
            return row[name]
    raise KeyError(f"缺少字段，候选为: {', '.join(names)}")


def main() -> None:
    args = parse_args()
    if not args.accept_license:
        raise SystemExit(
            "EcomRetrieval 数据卡未声明明确许可证；审查数据用途后使用 --accept-license 显式确认"
        )

    raw_dir = args.output_dir / "raw"
    checksums: dict[str, str] = {}
    for name, relative_path in FILES.items():
        destination = raw_dir / relative_path
        url = (
            "https://huggingface.co/datasets/mteb/EcomRetrieval/resolve/"
            f"{args.revision}/{relative_path}?download=true"
        )
        checksums[name] = download(url, destination)

    corpus_rows = parquet.read_table(raw_dir / FILES["corpus"]).to_pylist()
    query_rows = parquet.read_table(raw_dir / FILES["queries"]).to_pylist()
    qrel_rows = parquet.read_table(raw_dir / FILES["qrels"]).to_pylist()

    dataset = {
        "schemaVersion": "1",
        "name": "mteb-ecom-retrieval",
        "version": args.revision,
        "kind": "general_retrieval",
        "source": "https://huggingface.co/datasets/mteb/EcomRetrieval",
        "license": "upstream dataset card does not declare a license; local evaluation only after review",
        "reviewStatus": "public_benchmark_unverified_license",
        "corpus": [
            {
                "id": str(value(row, "_id", "id")),
                "title": str(row.get("title") or ""),
                "text": str(value(row, "text")),
            }
            for row in corpus_rows
        ],
        "queries": [
            {"id": str(value(row, "_id", "id")), "text": str(value(row, "text"))}
            for row in query_rows
        ],
        "qrels": [
            {
                "queryId": str(value(row, "query-id", "query_id")),
                "documentId": str(value(row, "corpus-id", "corpus_id")),
                "relevance": int(value(row, "score")),
            }
            for row in qrel_rows
        ],
    }

    args.output_dir.mkdir(parents=True, exist_ok=True)
    dataset_path = args.output_dir / "dataset.json"
    dataset_path.write_text(json.dumps(dataset, ensure_ascii=False), encoding="utf-8")
    manifest = {
        "source": dataset["source"],
        "revision": args.revision,
        "rawSha256": checksums,
        "corpusCount": len(dataset["corpus"]),
        "queryCount": len(dataset["queries"]),
        "qrelCount": len(dataset["qrels"]),
    }
    (args.output_dir / "manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(dataset_path)


if __name__ == "__main__":
    main()
