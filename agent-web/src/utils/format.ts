const dateTimeFormatter = new Intl.DateTimeFormat("zh-CN", { dateStyle: "short", timeStyle: "short" });

export function formatMoney(cent: number) {
  return `¥${(cent / 100).toFixed(2)}`;
}

export function formatDateTime(value: string | number) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "-" : dateTimeFormatter.format(date);
}
