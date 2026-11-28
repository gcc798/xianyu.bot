#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "openpyxl==3.1.5",
#   "pillow==12.0.0",
#   "python-docx==1.2.0",
#   "python-pptx==1.0.2",
#   "reportlab==4.4.4",
# ]
# ///
"""生成带文本、表格、阅读顺序和 OCR 期望的复杂文档金标准。"""

from __future__ import annotations

import argparse
import json
import pathlib

from docx import Document
from openpyxl import Workbook
from PIL import Image, ImageDraw, ImageFont
from pptx import Presentation
from pptx.util import Inches
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.cidfonts import UnicodeCIDFont
from reportlab.pdfgen import canvas


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=pathlib.Path, required=True)
    parser.add_argument("--cjk-font", type=pathlib.Path)
    return parser.parse_args()


def create_docx(output_dir: pathlib.Path) -> dict[str, object]:
    path = output_dir / "complex-customer-service.docx"
    document = Document()
    document.add_heading("闲鱼商品售后说明", level=1)
    document.add_paragraph("第一部分：验货。买家应先检查商品外观和配件。")
    document.add_paragraph("第二部分：退款。已经发货的订单应根据物流状态处理。")
    table = document.add_table(rows=1, cols=3)
    table.rows[0].cells[0].text = "场景"
    table.rows[0].cells[1].text = "所需材料"
    table.rows[0].cells[2].text = "首次响应"
    for row in (("质量问题", "照片或视频", "二十四小时"), ("物流异常", "订单号", "十二小时")):
        cells = table.add_row().cells
        for index, value in enumerate(row):
            cells[index].text = value
    document.save(path)
    return golden(
        path,
        ["闲鱼商品售后说明", "质量问题", "照片或视频", "物流异常", "订单号"],
        ["第一部分：验货", "第二部分：退款"],
        min_tables=1,
    )


def create_xlsx(output_dir: pathlib.Path) -> dict[str, object]:
    path = output_dir / "customer-service-rules.xlsx"
    workbook = Workbook()
    sheet = workbook.active
    sheet.title = "售后规则"
    sheet.append(("规则编号", "场景", "处理要求", "时效小时"))
    sheet.append(("R-001", "未发货退款", "平台内申请", 24))
    sheet.append(("R-002", "质量问题", "提交照片或视频", 48))
    sheet.append(("R-003", "物流异常", "使用订单工具查询", 12))
    sheet["E1"] = "总时效"
    sheet["E2"] = "=SUM(D2:D4)"
    workbook.save(path)
    return golden(
        path,
        ["售后规则", "R-001", "未发货退款", "R-002", "质量问题", "R-003", "物流异常"],
        ["R-001", "R-002", "R-003"],
        min_tables=1,
    )


def create_pptx(output_dir: pathlib.Path) -> dict[str, object]:
    path = output_dir / "customer-service-workflow.pptx"
    presentation = Presentation()
    title_slide = presentation.slides.add_slide(presentation.slide_layouts[0])
    title_slide.shapes.title.text = "闲鱼客服处理流程"
    title_slide.placeholders[1].text = "知识回答与实时工具边界"

    detail_slide = presentation.slides.add_slide(presentation.slide_layouts[5])
    detail_slide.shapes.title.text = "处理顺序"
    table = detail_slide.shapes.add_table(4, 2, Inches(1), Inches(1.8), Inches(8), Inches(3)).table
    for row, values in enumerate(
        (("步骤", "动作"), ("第一步", "识别问题"), ("第二步", "检索知识"), ("第三步", "查询实时状态"))
    ):
        for column, value in enumerate(values):
            table.cell(row, column).text = value
    presentation.save(path)
    return golden(
        path,
        ["闲鱼客服处理流程", "知识回答与实时工具边界", "识别问题", "检索知识", "查询实时状态"],
        ["第一步", "第二步", "第三步"],
        min_tables=1,
    )


def create_multicolumn_pdf(output_dir: pathlib.Path) -> dict[str, object]:
    path = output_dir / "multicolumn-customer-service.pdf"
    pdfmetrics.registerFont(UnicodeCIDFont("STSong-Light"))
    document = canvas.Canvas(str(path), pagesize=(595, 842))
    document.setFont("STSong-Light", 16)
    document.drawString(50, 790, "闲鱼售后双栏说明")
    document.setFont("STSong-Light", 11)
    left = ("左栏第一条：签收后先检查外观。", "左栏第二条：发现问题保留照片。")
    right = ("右栏第一条：物流状态通过工具查询。", "右栏第二条：敏感操作需要确认。")
    for index, line in enumerate(left):
        document.drawString(50, 740 - index * 28, line)
    for index, line in enumerate(right):
        document.drawString(315, 740 - index * 28, line)
    document.save()
    return golden(
        path,
        ["闲鱼售后双栏说明", *left, *right],
        [*left, *right],
        min_tables=0,
    )


def find_cjk_font(explicit: pathlib.Path | None) -> pathlib.Path:
    candidates = [
        explicit,
        pathlib.Path("/System/Library/Fonts/PingFang.ttc"),
        pathlib.Path("/System/Library/Fonts/STHeiti Medium.ttc"),
        pathlib.Path("/System/Library/Fonts/Supplemental/Songti.ttc"),
        pathlib.Path("/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc"),
        pathlib.Path("/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc"),
    ]
    for candidate in candidates:
        if candidate and candidate.is_file():
            return candidate
    raise SystemExit("生成中文扫描件需要通过 --cjk-font 指定可用的中文字体")


def create_scanned_pdf(output_dir: pathlib.Path, font_path: pathlib.Path) -> dict[str, object]:
    path = output_dir / "scanned-customer-service.pdf"
    image = Image.new("RGB", (1654, 2339), "white")
    drawing = ImageDraw.Draw(image)
    title_font = ImageFont.truetype(str(font_path), 58)
    body_font = ImageFont.truetype(str(font_path), 38)
    drawing.text((120, 150), "闲鱼扫描件测试", font=title_font, fill="black")
    lines = ("订单编号：XY-20261128", "处理要求：先验货再确认收货", "安全提示：不要脱离平台付款")
    for index, line in enumerate(lines):
        drawing.text((120, 300 + index * 90), line, font=body_font, fill="black")
    image.save(path, "PDF", resolution=150)
    result = golden(path, ["闲鱼扫描件测试", *lines], list(lines), min_tables=0)
    result["requiresOCR"] = True
    return result


def golden(
    path: pathlib.Path,
    expected_snippets: list[str],
    ordered_snippets: list[str],
    min_tables: int,
) -> dict[str, object]:
    return {
        "file": path.name,
        "expectedHTTPStatus": 200,
        "expectedSnippets": expected_snippets,
        "orderedSnippets": ordered_snippets,
        "minTables": min_tables,
    }


def main() -> None:
    args = parse_args()
    args.output_dir.mkdir(parents=True, exist_ok=True)
    font_path = find_cjk_font(args.cjk_font)
    documents = [
        create_docx(args.output_dir),
        create_xlsx(args.output_dir),
        create_pptx(args.output_dir),
        create_multicolumn_pdf(args.output_dir),
        create_scanned_pdf(args.output_dir, font_path),
    ]
    manifest = {
        "schemaVersion": "1",
        "source": "synthetic document parsing golden set",
        "reviewStatus": "generated_expectations_pending_human_review",
        "cjkFont": str(font_path),
        "documents": documents,
    }
    manifest_path = args.output_dir / "golden.json"
    manifest_path.write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(manifest_path)


if __name__ == "__main__":
    main()
