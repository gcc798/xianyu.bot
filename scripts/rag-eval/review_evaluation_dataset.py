#!/usr/bin/env python3
"""导出并应用人工 Qrel 核验表，不允许脚本自行冒充人工结论。"""

from __future__ import annotations

import argparse
import csv
import json
import pathlib


REVIEW_FIELDS = (
    "query_id",
    "query_text",
    "document_id",
    "document_title",
    "document_text",
    "proposed_relevance",
    "decision",
    "reviewed_relevance",
    "reviewer",
    "comment",
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    subcommands = parser.add_subparsers(dest="command", required=True)

    export = subcommands.add_parser("export")
    export.add_argument("--dataset", type=pathlib.Path, required=True)
    export.add_argument("--output", type=pathlib.Path, required=True)

    apply = subcommands.add_parser("apply")
    apply.add_argument("--dataset", type=pathlib.Path, required=True)
    apply.add_argument("--review", type=pathlib.Path, required=True)
    apply.add_argument("--version", required=True)
    apply.add_argument("--output", type=pathlib.Path, required=True)
    return parser.parse_args()


def load_dataset(path: pathlib.Path) -> dict[str, object]:
    dataset = json.loads(path.read_text(encoding="utf-8"))
    for field in ("corpus", "queries", "qrels"):
        if not isinstance(dataset.get(field), list):
            raise SystemExit(f"评估数据缺少 {field} 数组")
    return dataset


def export_review(dataset_path: pathlib.Path, output: pathlib.Path) -> None:
    dataset = load_dataset(dataset_path)
    queries = {str(query["id"]): query for query in dataset["queries"]}
    documents = {str(document["id"]): document for document in dataset["corpus"]}
    output.parent.mkdir(parents=True, exist_ok=True)
    with output.open("w", encoding="utf-8-sig", newline="") as destination:
        writer = csv.DictWriter(destination, fieldnames=REVIEW_FIELDS)
        writer.writeheader()
        for qrel in dataset["qrels"]:
            query_id = str(qrel["queryId"])
            document_id = str(qrel["documentId"])
            query = queries.get(query_id)
            document = documents.get(document_id)
            if query is None or document is None:
                raise SystemExit(f"Qrel 引用了不存在的查询或文档: {query_id}/{document_id}")
            writer.writerow(
                {
                    "query_id": query_id,
                    "query_text": query["text"],
                    "document_id": document_id,
                    "document_title": document.get("title", ""),
                    "document_text": document["text"],
                    "proposed_relevance": qrel["relevance"],
                    "decision": "",
                    "reviewed_relevance": "",
                    "reviewer": "",
                    "comment": "",
                }
            )
    print(output)


def apply_review(
    dataset_path: pathlib.Path,
    review_path: pathlib.Path,
    version: str,
    output: pathlib.Path,
) -> None:
    dataset = load_dataset(dataset_path)
    valid_queries = {str(query["id"]) for query in dataset["queries"]}
    valid_documents = {str(document["id"]) for document in dataset["corpus"]}
    expected_pairs = {
        (str(qrel["queryId"]), str(qrel["documentId"])) for qrel in dataset["qrels"]
    }
    accepted: list[dict[str, object]] = []
    reviewed_pairs: set[tuple[str, str]] = set()
    reviewers: set[str] = set()
    with review_path.open(encoding="utf-8-sig", newline="") as source:
        for line, row in enumerate(csv.DictReader(source), start=2):
            pair = (row["query_id"].strip(), row["document_id"].strip())
            if pair[0] not in valid_queries or pair[1] not in valid_documents:
                raise SystemExit(f"第 {line} 行引用了未知查询或文档: {pair[0]}/{pair[1]}")
            if pair in reviewed_pairs:
                raise SystemExit(f"第 {line} 行包含重复 Qrel: {pair[0]}/{pair[1]}")
            decision = row["decision"].strip().lower()
            reviewer = row["reviewer"].strip()
            if decision not in {"accept", "reject"} or not reviewer:
                raise SystemExit(f"第 {line} 行必须填写 decision=accept/reject 和 reviewer")
            reviewed_pairs.add(pair)
            reviewers.add(reviewer)
            if decision == "reject":
                continue
            relevance_text = row["reviewed_relevance"].strip() or row["proposed_relevance"].strip()
            relevance = int(relevance_text)
            if relevance <= 0:
                raise SystemExit(f"第 {line} 行相关性必须大于零")
            accepted.append(
                {"queryId": pair[0], "documentId": pair[1], "relevance": relevance}
            )

    missing_pairs = expected_pairs - reviewed_pairs
    if missing_pairs:
        raise SystemExit(f"仍有 {len(missing_pairs)} 条 Qrel 未核验")
    judged_queries = {str(qrel["queryId"]) for qrel in accepted}
    missing_queries = [
        str(query["id"]) for query in dataset["queries"] if str(query["id"]) not in judged_queries
    ]
    if missing_queries:
        raise SystemExit(f"以下查询没有保留任何相关文档: {', '.join(missing_queries)}")

    dataset["version"] = version
    dataset["reviewStatus"] = "human_reviewed"
    dataset["reviewers"] = sorted(reviewers)
    dataset["qrels"] = accepted
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(dataset, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(output)


def main() -> None:
    args = parse_args()
    if args.command == "export":
        export_review(args.dataset, args.output)
        return
    apply_review(args.dataset, args.review, args.version, args.output)


if __name__ == "__main__":
    main()
