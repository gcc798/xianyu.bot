import { Alert, App, Button, Input, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { automationApi } from "../../../api/automation";
import { productsApi } from "../../../api/products";
import type { DeliveryRule, DeliveryRuleInput } from "../../../models/automation";
import type { Product } from "../../../models/product";
import styles from "../../../styles/page.module.css";
import { DeliveryRuleModal } from "./components/DeliveryRuleModal";
import { ResourceImportModal } from "./components/ResourceImportModal";

export function Component() {
  const { message, modal } = App.useApp();
  const [rows, setRows] = useState<DeliveryRule[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [productId, setProductId] = useState<number>();
  const [mode, setMode] = useState<string>();
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<DeliveryRule | null | undefined>();
  const [importRule, setImportRule] = useState<DeliveryRule | null>(null);
  const [resourceText, setResourceText] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await automationApi.deliveryRules({ page, pageSize, productId, mode, keyword: query });
      setRows(data.list);
      setTotal(data.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "发货规则加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, productId, mode, query]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => {
    void productsApi.list({ page: 1, pageSize: 100, status: "on_sale" })
      .then((data) => setProducts(data.list))
      .catch(() => undefined);
  }, []);

  const save = async (input: DeliveryRuleInput) => {
    if (input.mode === "card" && !input.template.includes("{card}")) {
      message.warning("卡密模板必须包含 {card}");
      return;
    }
    setSaving(true);
    try {
      await automationApi.saveDeliveryRule({
        ...input,
        ID: editing?.ID ?? 0,
        skuKey: input.skuKey.trim(),
        skuName: input.skuName.trim(),
        template: input.template.trim(),
      });
      message.success("发货规则已保存");
      setEditing(undefined);
      await load();
    } finally {
      setSaving(false);
    }
  };

  const remove = (record: DeliveryRule) => {
    modal.confirm({
      title: "删除发货规则？",
      content: "删除后新订单不会再使用该规则。",
      okText: "删除",
      okType: "danger",
      cancelText: "取消",
      onOk: async () => {
        await automationApi.deleteDeliveryRule(record.ID);
        message.success("规则已删除");
        await load();
      },
    });
  };

  const importResources = async () => {
    if (!importRule) return;
    const values = resourceText.split(/\r?\n/).map((value) => value.trim()).filter(Boolean);
    if (!values.length) {
      message.warning("请输入卡密");
      return;
    }
    setSaving(true);
    try {
      const data = await automationApi.importResources(importRule.ID, values);
      message.success(`已导入 ${data.count} 条卡密`);
      setImportRule(null);
      setResourceText("");
    } finally {
      setSaving(false);
    }
  };

  const showInventory = async (record: DeliveryRule) => {
    const data = await automationApi.inventory(record.ID);
    modal.info({
      title: "卡密库存",
      content: `可用 ${data.available}，已预留 ${data.reserved}，已交付 ${data.consumed}`,
    });
  };

  const columns: TableColumnsType<DeliveryRule> = [
    { title: "序号", width: 64, render: (_value, _record, index) => (page - 1) * pageSize + index + 1 },
    { title: "商品", dataIndex: "productTitle", ellipsis: true },
    {
      title: "SKU", dataIndex: "skuName", width: 170, ellipsis: true,
      render: (value: string, record) => value || record.skuKey || "通用规则",
    },
    {
      title: "模式", dataIndex: "mode", width: 100,
      render: (value: string) => <Tag color={value === "card" ? "processing" : "default"}>{value === "card" ? "卡密发货" : "文本发货"}</Tag>,
    },
    { title: "模板", dataIndex: "template", width: 330, ellipsis: true },
    {
      title: "状态", dataIndex: "enabled", width: 80,
      render: (value: boolean) => <Tag color={value ? "success" : "default"}>{value ? "启用" : "停用"}</Tag>,
    },
    {
      title: "操作", width: 220, fixed: "right",
      render: (_, record) => (
        <Space size={2}>
          <Button type="link" className={styles.linkButton} onClick={() => setEditing(record)}>编辑</Button>
          {record.mode === "card" && (
            <>
              <Button type="link" className={styles.linkButton} onClick={() => setImportRule(record)}>导入卡密</Button>
              <Button type="link" className={styles.linkButton} onClick={() => void showInventory(record)}>库存</Button>
            </>
          )}
          <Button type="link" danger className={styles.linkButton} onClick={() => remove(record)}>删除</Button>
        </Space>
      ),
    },
  ];

  const search = () => { setQuery(keyword.trim()); setPage(1); };
  const reset = () => {
    setProductId(undefined);
    setMode(undefined);
    setKeyword("");
    setQuery("");
    setPage(1);
  };

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <div className={styles.filters}>
          <Select
            allowClear showSearch value={productId} placeholder="全部商品" style={{ width: 220 }}
            options={products.map((item) => ({ value: item.ID, label: item.title }))}
            onChange={(value) => { setProductId(value); setPage(1); }}
          />
          <Select
            allowClear value={mode} placeholder="全部模式" style={{ width: 130 }}
            options={[{ value: "text", label: "文本发货" }, { value: "card", label: "卡密发货" }]}
            onChange={(value) => { setMode(value); setPage(1); }}
          />
          <Input.Search
            value={keyword} allowClear placeholder="SKU 或模板内容" style={{ width: 220 }}
            onChange={(event) => setKeyword(event.target.value)} onSearch={search}
          />
          <Space>
            <Button type="primary" onClick={search}>查询</Button>
            <Button onClick={reset}>重置</Button>
            <Button type="primary" onClick={() => setEditing(null)}>新增规则</Button>
          </Space>
        </div>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<DeliveryRule>
            rowKey="ID" size="small" loading={loading} columns={columns} dataSource={rows}
            pagination={false} scroll={{ x: 1100 }}
          />
        </div>
        <div className={styles.pagination}>
          <Pagination
            current={page} pageSize={pageSize} total={total} showSizeChanger pageSizeOptions={[10, 20, 50]}
            onChange={(nextPage, nextSize) => { setPage(nextSize === pageSize ? nextPage : 1); setPageSize(nextSize); }}
          />
        </div>
      </section>
      <DeliveryRuleModal
        record={editing}
        products={products}
        defaultProductId={productId}
        saving={saving}
        onCancel={() => setEditing(undefined)}
        onSave={save}
      />
      <ResourceImportModal
        rule={importRule}
        value={resourceText}
        saving={saving}
        onChange={setResourceText}
        onCancel={() => setImportRule(null)}
        onImport={importResources}
      />
    </div>
  );
}
