import { Descriptions, Drawer, Empty, Skeleton, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import type { ResourceDetail, ResourceSource } from "../../../models/resource";
import { formatDateTime, formatFileSize } from "../../../utils/format";

type Props = {
  resource?: ResourceDetail;
  loading: boolean;
  onClose: () => void;
};

export function ResourceDetailDrawer({ resource, loading, onClose }: Props) {
  const columns: TableColumnsType<ResourceSource> = [
    { title: "来源服务", dataIndex: "mcpServerCode", width: 120 },
    { title: "来源名称", dataIndex: "sourceName", width: 180, ellipsis: true },
    { title: "外部标识", dataIndex: "externalId", width: 200, ellipsis: true },
    { title: "来源引用", dataIndex: "sourceReference", ellipsis: true, render: (value: string) => value || "-" },
  ];

  return <Drawer open={loading || Boolean(resource)} title="资源详情" size={760} onClose={onClose}>
    {loading ? <Skeleton active /> : resource ? <>
      <Descriptions column={2} bordered size="small" items={[
        { key: "name", label: "资源名称", children: resource.name, span: 2 },
        { key: "type", label: "文件类型", children: resource.mimeType || "-" },
        { key: "size", label: "文件大小", children: formatFileSize(resource.size) },
        { key: "state", label: "状态", children: <Tag color={resource.state === "available" ? "success" : "default"}>{resource.state}</Tag> },
        { key: "time", label: "创建时间", children: formatDateTime(resource.CreatedAt) },
        { key: "sha", label: "SHA-256", children: resource.sha256 || "-", span: 2 },
      ]} />
      <Table<ResourceSource>
        rowKey="ID" size="small" columns={columns} dataSource={resource.sources}
        pagination={false} locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无来源记录" /> }}
        style={{ marginTop: 24 }} scroll={{ x: 700 }}
      />
    </> : null}
  </Drawer>;
}
