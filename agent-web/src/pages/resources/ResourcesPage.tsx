import { Alert, App, Button, Input, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { resourcesApi } from "../../api/resources";
import type { Resource, ResourceDetail } from "../../models/resource";
import styles from "../../styles/page.module.css";
import { formatDateTime, formatFileSize } from "../../utils/format";
import { ResourceDetailDrawer } from "./components/ResourceDetailDrawer";

export function Component() {
  const { message } = App.useApp();
  const [rows, setRows] = useState<Resource[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [mimeType, setMimeType] = useState<string>();
  const [state, setState] = useState<string>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [detail, setDetail] = useState<ResourceDetail>();
  const [detailLoading, setDetailLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await resourcesApi.list({ page, pageSize, keyword: query, mimeType, state });
      setRows(data.list);
      setTotal(data.total);
    } catch (cause) { setError(cause instanceof Error ? cause.message : "资源加载失败"); }
    finally { setLoading(false); }
  }, [page, pageSize, query, mimeType, state]);
  useEffect(() => { void load(); }, [load]);

  const openDetail = async (ID: number) => {
    setDetailLoading(true);
    try { setDetail((await resourcesApi.get(ID)).resource); }
    catch (cause) { message.error(cause instanceof Error ? cause.message : "资源详情加载失败"); }
    finally { setDetailLoading(false); }
  };
  const download = async (record: Resource) => {
    try {
      const data = await resourcesApi.download(record.ID);
      window.open(data.downloadUrl, "_blank", "noopener,noreferrer");
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "下载地址获取失败");
    }
  };

  const columns: TableColumnsType<Resource> = [
    { title: "序号", width: 64, render: (_value, _record, index) => (page - 1) * pageSize + index + 1 },
    { title: "资源名称", dataIndex: "name", ellipsis: true, render: (value: string, record) => <Button type="link" className={styles.linkButton} onClick={() => void openDetail(record.ID)}>{value}</Button> },
    { title: "文件类型", dataIndex: "mimeType", width: 180, ellipsis: true, render: (value: string) => value || "-" },
    { title: "大小", dataIndex: "size", width: 100, render: formatFileSize },
    { title: "状态", dataIndex: "state", width: 90, render: (value: string) => <Tag color={value === "available" ? "success" : "default"}>{value === "available" ? "可用" : value}</Tag> },
    { title: "创建时间", dataIndex: "CreatedAt", width: 150, render: formatDateTime },
    { title: "操作", width: 80, fixed: "right", render: (_, record) => <Button type="link" className={styles.linkButton} disabled={record.state !== "available"} onClick={() => void download(record)}>下载</Button> },
  ];

  const search = () => { setQuery(keyword.trim()); setPage(1); };
  const reset = () => { setKeyword(""); setQuery(""); setMimeType(undefined); setState(undefined); setPage(1); };
  return <div className={styles.page}>
    <section className={styles.surface}>
      <div className={styles.filters}>
        <Input.Search value={keyword} allowClear placeholder="资源名称" style={{ width: 240 }} onChange={(event) => setKeyword(event.target.value)} onSearch={search} />
        <Input value={mimeType} allowClear placeholder="MIME 类型" style={{ width: 170 }} onChange={(event) => setMimeType(event.target.value || undefined)} />
        <Select allowClear value={state} placeholder="全部状态" style={{ width: 130 }} options={[{ value: "available", label: "可用" }, { value: "deleted", label: "已删除" }]} onChange={(value) => { setState(value); setPage(1); }} />
        <Space><Button type="primary" onClick={search}>查询</Button><Button onClick={reset}>重置</Button></Space>
      </div>
      {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
      <div className={styles.table}><Table<Resource> rowKey="ID" size="small" loading={loading} columns={columns} dataSource={rows} pagination={false} scroll={{ x: 900 }} /></div>
      <div className={styles.pagination}><Pagination current={page} pageSize={pageSize} total={total} showSizeChanger pageSizeOptions={[10, 20, 50]} onChange={(nextPage, nextSize) => { setPage(nextSize === pageSize ? nextPage : 1); setPageSize(nextSize); }} /></div>
    </section>
    <ResourceDetailDrawer resource={detail} loading={detailLoading} onClose={() => setDetail(undefined)} />
  </div>;
}
