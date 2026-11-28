#!/usr/bin/env python3
"""按固定版本准备 Office 样本，可选下载仅限研究用途的 OmniDocBench。"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import shutil
import tempfile
import urllib.request


DEFAULT_POI_REVISION = "f63b77bb7497d950073969f9d418f2270dd87edd"
SUPPORTED_SUFFIXES = (".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx")
COMPLEXITY_HINTS = (
    "chart",
    "drawing",
    "formula",
    "image",
    "picture",
    "table",
    "textbox",
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=pathlib.Path, required=True)
    parser.add_argument("--poi-revision", default=DEFAULT_POI_REVISION)
    parser.add_argument("--max-per-format", type=int, default=3)
    parser.add_argument("--include-omnidocbench", action="store_true")
    parser.add_argument("--omni-revision", help="OmniDocBench 数据集完整提交 SHA")
    parser.add_argument(
        "--accept-research-only",
        action="store_true",
        help="确认 OmniDocBench 仅用于研究评估且不会进入产品数据",
    )
    return parser.parse_args()


def sha256(path: pathlib.Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def sample_score(path: str, size: int) -> tuple[int, int, str]:
    name = pathlib.PurePosixPath(path).name.lower()
    hints = sum(hint in name for hint in COMPLEXITY_HINTS)
    return hints, min(size, 20 * 1024 * 1024), name


def fetch_json(url: str) -> dict[str, object]:
    request = urllib.request.Request(url, headers={"User-Agent": "xianyu.bot-rag-eval/1"})
    with urllib.request.urlopen(request, timeout=120) as response:
        return json.load(response)


def download(url: str, destination: pathlib.Path) -> None:
    request = urllib.request.Request(url, headers={"User-Agent": "xianyu.bot-rag-eval/1"})
    destination.parent.mkdir(parents=True, exist_ok=True)
    with urllib.request.urlopen(request, timeout=120) as response:
        with tempfile.NamedTemporaryFile(dir=destination.parent, delete=False) as temporary:
            shutil.copyfileobj(response, temporary)
            temporary_path = pathlib.Path(temporary.name)
    temporary_path.replace(destination)


def prepare_poi(output_dir: pathlib.Path, revision: str, max_per_format: int) -> list[dict[str, object]]:
    if max_per_format <= 0:
        raise SystemExit("--max-per-format 必须大于零")
    tree = fetch_json(f"https://api.github.com/repos/apache/poi/git/trees/{revision}?recursive=1")
    if tree.get("truncated"):
        raise SystemExit("Apache POI 文件树响应被截断，无法保证样本选择可复现")
    entries = tree.get("tree")
    if not isinstance(entries, list):
        raise SystemExit("Apache POI 文件树响应格式无效")

    selected: list[dict[str, object]] = []
    for suffix in SUPPORTED_SUFFIXES:
        candidates = [
            entry
            for entry in entries
            if isinstance(entry, dict)
            and entry.get("type") == "blob"
            and str(entry.get("path", "")).startswith("test-data/")
            and str(entry.get("path", "")).lower().endswith(suffix)
            and int(entry.get("size", 0)) <= 20 * 1024 * 1024
        ]
        candidates.sort(
            key=lambda entry: sample_score(str(entry["path"]), int(entry.get("size", 0))),
            reverse=True,
        )
        selected.extend(candidates[:max_per_format])

    sample_dir = output_dir / "samples" / "apache-poi"
    sample_dir.mkdir(parents=True, exist_ok=True)
    manifest: list[dict[str, object]] = []
    for index, entry in enumerate(selected, start=1):
        upstream_path = str(entry["path"])
        destination = sample_dir / f"{index:02d}-{pathlib.PurePosixPath(upstream_path).name}"
        download(
            f"https://raw.githubusercontent.com/apache/poi/{revision}/{upstream_path}",
            destination,
        )
        manifest.append(
            {
                "path": str(destination.relative_to(output_dir)),
                "upstreamPath": upstream_path,
                "size": destination.stat().st_size,
                "sha256": sha256(destination),
            }
        )
    return manifest


def prepare_omnidocbench(output_dir: pathlib.Path, revision: str) -> pathlib.Path:
    try:
        from huggingface_hub import snapshot_download
    except ImportError as error:
        raise SystemExit(
            "下载 OmniDocBench 需要执行：uv run --with huggingface-hub==0.36.0 "
            "scripts/rag-eval/prepare_document_samples.py ..."
        ) from error
    destination = output_dir / "sources" / "omnidocbench"
    snapshot_download(
        repo_id="opendatalab/OmniDocBench",
        repo_type="dataset",
        revision=revision,
        local_dir=destination,
    )
    return destination


def main() -> None:
    args = parse_args()
    args.output_dir.mkdir(parents=True, exist_ok=True)
    poi_samples = prepare_poi(args.output_dir, args.poi_revision, args.max_per_format)

    omni_path: str | None = None
    if args.include_omnidocbench:
        if not args.accept_research_only:
            raise SystemExit("OmniDocBench 声明仅限研究用途；请审查后显式使用 --accept-research-only")
        if not args.omni_revision or len(args.omni_revision) != 40:
            raise SystemExit("下载 OmniDocBench 必须通过 --omni-revision 固定完整提交 SHA")
        omni_path = str(prepare_omnidocbench(args.output_dir, args.omni_revision))

    manifest = {
        "apachePoi": {
            "source": "https://github.com/apache/poi/tree/trunk/test-data",
            "revision": args.poi_revision,
            "licenseNote": "Apache 项目测试数据；重新分发前仍需核验单个样本来源",
            "samples": poi_samples,
        },
        "omniDocBench": {
            "source": "https://huggingface.co/datasets/opendatalab/OmniDocBench",
            "revision": args.omni_revision,
            "usage": "research_only" if omni_path else "not_downloaded",
            "path": omni_path,
        },
    }
    manifest_path = args.output_dir / "document-samples-manifest.json"
    manifest_path.write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(manifest_path)


if __name__ == "__main__":
    main()
