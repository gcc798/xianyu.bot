import { Alert, App, Button, Input, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { automationApi } from "../../../api/automation";
import { dictionaryName, useDictionaryOptions } from "../../../hooks/useDictionaryOptions";
import { productsApi } from "../../../api/products";
import type { ReplyRule, ReplyRuleInput } from "../../../models/automation";
import type { Product } from "../../../models/product";
import styles from "../../../styles/page.module.css";
import { ReplyRuleModal } from "./components/ReplyRuleModal";

const ruleTypeColor: Record<string, string> = {
  exact: "default",
  fuzzy: "processing",
  fallback: "warning",
};

export function Component() {
  const { message, modal } = App.useApp();
  const { options: ruleTypeOptions } = useDictionaryOptions("reply_rule_type");
  const [rows, setRows] = useState<ReplyRule[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [productId, setProductId] = useState<number>();
  const [type, setType] = useState<string>();
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<ReplyRule | null | undefined>();

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await automationApi.replyRules({
        page,
        pageSize,
        productId,
        type,
        keyword: query,
      });
      setRows(data.list);
      setTotal(data.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "回复规则加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, productId, type, query]);

  useEffect(() => {
    void load();
  }, [load]);
  useEffect(() => {
    void productsApi
      .list({ page: 1, pageSize: 100, status: "on_sale" })
      .then((data) => setProducts(data.list))
      .catch(() => undefined);
  }, []);

  const save = async (input: ReplyRuleInput) => {
    setSaving(true);
    try {
      await automationApi.saveReplyRule({
        ...input,
        ID: editing?.ID ?? 0,
        keyword: input.type === "fallback" ? "" : input.keyword.trim(),
        content: input.content.trim(),
      });
      message.success("回复规则已保存");
      setEditing(undefined);
      await load();
    } finally {
      setSaving(false);
    }
  };

  const remove = (record: ReplyRule) => {
    modal.confirm({
      title: "删除回复规则？",
      content: "删除后该规则不再参与自动回复。",
      okText: "删除",
      okType: "danger",
      cancelText: "取消",
      onOk: async () => {
        await automationApi.deleteReplyRule(record.ID);
        message.success("规则已删除");
        await load();
      },
    });
  };

  const columns: TableColumnsType<ReplyRule> = [
    {
      title: "序号",
      width: 64,
      align: "center",
      render: (_value, _record, index) => (page - 1) * pageSize + index + 1,
    },
    { title: "商品", dataIndex: "productTitle", ellipsis: true },
    {
      title: "类型",
      dataIndex: "type",
      width: 100,
      render: (value: string) => (
        <Tag color={ruleTypeColor[value] ?? "default"}>
          {dictionaryName(ruleTypeOptions, value)}
        </Tag>
      ),
    },
    {
      title: "触发关键词",
      dataIndex: "keyword",
      width: 190,
      ellipsis: true,
      render: (value: string, record) => (record.type === "fallback" ? "未匹配其他规则" : value),
    },
    { title: "回复内容", dataIndex: "content", width: 300, ellipsis: true },
    { title: "优先级", dataIndex: "priority", width: 80 },
    {
      title: "状态",
      dataIndex: "enabled",
      width: 80,
      render: (value: boolean) => (
        <Tag color={value ? "success" : "default"}>{value ? "启用" : "停用"}</Tag>
      ),
    },
    {
      title: "操作",
      width: 100,
      fixed: "right",
      render: (_, record) => (
        <Space size={2}>
          <Button type="link" className={styles.linkButton} onClick={() => setEditing(record)}>
            编辑
          </Button>
          <Button type="link" danger className={styles.linkButton} onClick={() => remove(record)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  const search = () => {
    setQuery(keyword.trim());
    setPage(1);
  };
  const reset = () => {
    setProductId(undefined);
    setType(undefined);
    setKeyword("");
    setQuery("");
    setPage(1);
  };

  return (
    <div className={styles.page}>
      <section className={styles.surface}>
        <div className={styles.filters}>
          <Select
            allowClear
            showSearch
            value={productId}
            placeholder="全部商品"
            style={{ width: 220 }}
            options={products.map((item) => ({ value: item.ID, label: item.title }))}
            onChange={(value) => {
              setProductId(value);
              setPage(1);
            }}
          />
          <Select
            allowClear
            value={type}
            placeholder="全部类型"
            style={{ width: 130 }}
            options={ruleTypeOptions.map(({ code, name }) => ({ value: code, label: name }))}
            onChange={(value) => {
              setType(value);
              setPage(1);
            }}
          />
          <Input.Search
            value={keyword}
            allowClear
            placeholder="关键词或回复内容"
            style={{ width: 240 }}
            onChange={(event) => setKeyword(event.target.value)}
            onSearch={search}
          />
          <Space>
            <Button type="primary" onClick={search}>
              查询
            </Button>
            <Button onClick={reset}>重置</Button>
            <Button type="primary" onClick={() => setEditing(null)}>
              新增规则
            </Button>
          </Space>
        </div>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <div className={styles.table}>
          <Table<ReplyRule>
            rowKey="ID"
            size="small"
            loading={loading}
            columns={columns}
            dataSource={rows}
            pagination={false}
            scroll={{ x: 1100 }}
          />
        </div>
        <div className={styles.pagination}>
          <Pagination
            current={page}
            pageSize={pageSize}
            total={total}
            showSizeChanger
            pageSizeOptions={[10, 20, 50]}
            onChange={(nextPage, nextSize) => {
              setPage(nextSize === pageSize ? nextPage : 1);
              setPageSize(nextSize);
            }}
          />
        </div>
      </section>
      <ReplyRuleModal
        record={editing}
        products={products}
        defaultProductId={productId}
        saving={saving}
        onCancel={() => setEditing(undefined)}
        onSave={save}
      />
    </div>
  );
}
