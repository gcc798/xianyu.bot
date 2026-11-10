import { Alert, App, Button, Image, Input, Pagination, Select, Space, Table, Tag } from "antd";
import type { TableColumnsType } from "antd";
import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { accountsApi } from "../../../api/accounts";
import { productsApi } from "../../../api/products";
import { dictionaryName, useDictionaryOptions } from "../../../hooks/useDictionaryOptions";
import type { XianyuAccount } from "../../../models/account";
import type { Product, ProductSettingInput } from "../../../models/product";
import { formatMoney } from "../../../utils/format";
import businessStyles from "../../../styles/business.module.css";
import pageStyles from "../../../styles/page.module.css";
import ProductSettingModal from "./components/ProductSettingModal";

const productStatusColor: Record<string, string> = {
  on_sale: "success",
  offline: "default",
  sold: "warning",
};

export function Component() {
  const { message } = App.useApp();
  const { options: productStatuses } = useDictionaryOptions("product_status");
  const [rows, setRows] = useState<Product[]>([]);
  const [accounts, setAccounts] = useState<XianyuAccount[]>([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [accountId, setAccountId] = useState<number>();
  const [status, setStatus] = useState<string>();
  const [keyword, setKeyword] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [settingProduct, setSettingProduct] = useState<Product | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const response = await productsApi.list({
        page,
        pageSize,
        accountId,
        titleOrId: query,
        status,
      });
      setRows(response.list);
      setTotal(response.total);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "商品加载失败");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, accountId, query, status]);

  useEffect(() => {
    void load();
  }, [load]);
  useEffect(() => {
    void accountsApi
      .list({ page: 1, pageSize: 100 })
      .then((response) => setAccounts(response.list))
      .catch((cause) => setError(cause instanceof Error ? cause.message : "账号加载失败"));
  }, []);

  const syncProducts = async () => {
    const targetAccountID = accountId || (accounts.length === 1 ? accounts[0].ID : undefined);
    if (!targetAccountID) {
      message.warning("请先选择要同步的闲鱼账号");
      return;
    }
    setSyncing(true);
    try {
      const response = await productsApi.sync(targetAccountID);
      message.success(`已同步 ${response.count} 个商品`);
      await load();
    } finally {
      setSyncing(false);
    }
  };

  const saveSetting = async (input: ProductSettingInput) => {
    setSaving(true);
    try {
      await productsApi.updateSetting({
        ...input,
        relistAddress: input.relistAddress.trim(),
        relistDescription: input.relistDescription.trim(),
        relistImageUrls: input.relistImageUrls.trim(),
      });
      message.success("商品设置已保存");
      setSettingProduct(null);
      await load();
    } finally {
      setSaving(false);
    }
  };

  const renderImage = (record: Product) => {
    const images = record.imageUrls.length
      ? record.imageUrls
      : record.imageUrl
        ? [record.imageUrl]
        : [];
    if (!images.length) return "-";
    return (
      <Image.PreviewGroup items={images}>
        <Image
          src={images[0]}
          width={36}
          height={36}
          className={businessStyles.thumbnail}
          preview={{ mask: false }}
        />
      </Image.PreviewGroup>
    );
  };

  const renderAutomation = (record: Product) => (
    <div className={businessStyles.statusGroup}>
      <Tag color={record.autoReplyEnabled ? "processing" : "default"}>
        回复{record.autoReplyEnabled ? "开启" : "关闭"}
      </Tag>
      <Tag color={record.autoDeliveryEnabled ? "processing" : "default"}>
        发货{record.autoDeliveryEnabled ? "开启" : "关闭"}
      </Tag>
      <Tag color={record.autoRelistEnabled ? "processing" : "default"}>
        续售{record.autoRelistEnabled ? "开启" : "关闭"}
      </Tag>
    </div>
  );

  const columns: TableColumnsType<Product> = [
    {
      title: "序号",
      width: 64,
      align: "center",
      render: (_value, _record, index) => (page - 1) * pageSize + index + 1,
    },
    {
      title: "商品名称",
      dataIndex: "title",
      width: 215,
      ellipsis: true,
      render: (value: string, record) => <Link to={`/xianyu/products/${record.ID}`}>{value}</Link>,
    },
    { title: "商品编码", dataIndex: "externalProductId", width: 125, ellipsis: true },
    {
      title: "图片",
      dataIndex: "imageUrl",
      width: 55,
      render: (_: string, record) => renderImage(record),
    },
    { title: "账号", dataIndex: "accountLabel", width: 110, ellipsis: true },
    { title: "价格", dataIndex: "priceCent", width: 85, render: formatMoney },
    {
      title: "状态",
      dataIndex: "status",
      width: 74,
      render: (value: string) => {
        return (
          <Tag color={productStatusColor[value] ?? "default"}>
            {dictionaryName(productStatuses, value) || "未知"}
          </Tag>
        );
      },
    },
    {
      title: "自动化",
      key: "automation",
      width: 240,
      render: (_, record) => renderAutomation(record),
    },
    {
      title: "最后同步",
      dataIndex: "UpdatedAt",
      width: 115,
      render: (value: string) =>
        new Intl.DateTimeFormat("zh-CN", { dateStyle: "short", timeStyle: "short" }).format(
          new Date(value),
        ),
    },
    {
      title: "操作",
      key: "actions",
      width: 130,
      fixed: "right",
      render: (_, record) => (
        <Space size={4}>
          <Link to={`/xianyu/products/${record.ID}`}>查看</Link>
          {record.status === "on_sale" ? (
            <Button
              type="link"
              className={pageStyles.linkButton}
              onClick={() => setSettingProduct(record)}
            >
              业务设置
            </Button>
          ) : (
            <span style={{ color: "#89919d" }}>不可设置</span>
          )}
        </Space>
      ),
    },
  ];

  const reset = () => {
    setAccountId(undefined);
    setStatus(undefined);
    setKeyword("");
    setQuery("");
    setPage(1);
  };
  return (
    <div className={pageStyles.page}>
      <section className={pageStyles.surface}>
        <div className={pageStyles.filters}>
          <Select
            allowClear
            value={accountId}
            placeholder="全部账号"
            style={{ width: 170 }}
            options={accounts.map((account) => ({ value: account.ID, label: account.label }))}
            onChange={(value) => {
              setAccountId(value);
              setPage(1);
            }}
          />
          <Input.Search
            value={keyword}
            allowClear
            placeholder="标题或闲鱼商品 ID"
            style={{ width: 260 }}
            onChange={(event) => setKeyword(event.target.value)}
            onSearch={() => {
              setQuery(keyword.trim());
              setPage(1);
            }}
          />
          <Select
            allowClear
            value={status}
            placeholder="全部状态"
            style={{ width: 140 }}
            options={productStatuses.map(({ code, name }) => ({ value: code, label: name }))}
            onChange={(value) => {
              setStatus(value);
              setPage(1);
            }}
          />
          <Space>
            <Button
              type="primary"
              onClick={() => {
                setQuery(keyword.trim());
                setPage(1);
              }}
            >
              查询
            </Button>
            <Button onClick={reset}>重置</Button>
            <Button type="primary" loading={syncing} onClick={() => void syncProducts()}>
              同步商品
            </Button>
          </Space>
        </div>
        {error && (
          <Alert
            className={pageStyles.alert}
            type="error"
            showIcon
            title={error}
            closable
            onClose={() => setError("")}
          />
        )}
        <div className={pageStyles.table}>
          <Table<Product>
            rowKey="ID"
            size="small"
            loading={loading}
            columns={columns}
            dataSource={rows}
            pagination={false}
            scroll={{ x: 1100 }}
          />
        </div>
        <div className={pageStyles.pagination}>
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
      <ProductSettingModal
        product={settingProduct}
        loading={saving}
        onCancel={() => setSettingProduct(null)}
        onSave={saveSetting}
      />
    </div>
  );
}
