import { useMemo, useState } from "react";
import {
  Alert,
  App as AntdApp,
  Button,
  ConfigProvider,
  Drawer,
  Form,
  Grid,
  Image,
  Input,
  InputNumber,
  Layout,
  Menu,
  Modal,
  Pagination,
  Select,
  Switch,
  Table,
  Typography,
} from "antd";
import type { TableColumnsType } from "antd";
import {
  PreviewProduct,
  previewAccounts,
  previewProducts,
} from "./previewData";
import BotWorkspacePreview from "./BotWorkspacePreview";
import OrderDetailPreview from "./OrderDetailPreview";
import ProductDetailPreview from "./ProductDetailPreview";
import { neutralPreviewTheme } from "./previewThemes";
import "./product-management-preview.css";

const { Header, Content, Sider } = Layout;

type ProductFilters = {
  accountId?: number;
  titleOrId?: string;
  status?: PreviewProduct["status"];
};

type PreviewPage = "products" | "product-detail" | "order-detail" | "bot";

type ProductSettingValues = Pick<
  PreviewProduct,
  | "virtualProduct"
  | "autoReplyEnabled"
  | "autoDeliveryEnabled"
  | "autoRelistEnabled"
  | "replyDelaySeconds"
  | "relistDelaySeconds"
  | "relistDescription"
  | "relistImageUrls"
  | "relistAddress"
>;

const navigation = [
  { key: "overview", label: "总览" },
  { key: "bot", label: "Bot" },
  {
    key: "xianyu-automation",
    label: "闲鱼自动化",
    children: [
      { key: "accounts", label: "账号管理" },
      { key: "products", label: "商品管理" },
      { key: "orders", label: "订单管理" },
      { key: "messages", label: "消息记录" },
      { key: "reply-rules", label: "回复规则" },
      { key: "delivery-rules", label: "发货规则" },
      { key: "events", label: "业务事件" },
    ],
  },
  { key: "system", label: "系统设置" },
];

const statusLabels: Record<PreviewProduct["status"], string> = {
  on_sale: "在售",
  offline: "已下架",
  sold: "已售出",
};

export default function ProductManagementPreview() {
  return (
    <ConfigProvider theme={neutralPreviewTheme}>
      <AntdApp>
        <PreviewConsole />
      </AntdApp>
    </ConfigProvider>
  );
}

function PreviewConsole() {
  const { message } = AntdApp.useApp();
  const screens = Grid.useBreakpoint();
  const [filters] = Form.useForm<ProductFilters>();
  const [settingForm] = Form.useForm<ProductSettingValues>();
  const [products, setProducts] = useState(previewProducts);
  const [previewPage, setPreviewPage] = useState<PreviewPage>("products");
  const [detailProduct, setDetailProduct] = useState(previewProducts[0]);
  const [appliedFilters, setAppliedFilters] = useState<ProductFilters>({});
  const [selectedProduct, setSelectedProduct] = useState<PreviewProduct>();
  const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const filteredProducts = useMemo(() => {
    const keyword = appliedFilters.titleOrId?.trim().toLocaleLowerCase();
    return products.filter((product) => {
      if (appliedFilters.accountId && product.accountId !== appliedFilters.accountId)
        return false;
      if (appliedFilters.status && product.status !== appliedFilters.status) return false;
      return (
        !keyword ||
        product.title.toLocaleLowerCase().includes(keyword) ||
        product.externalProductId.toLocaleLowerCase().includes(keyword)
      );
    });
  }, [appliedFilters, products]);

  const openSetting = (product: PreviewProduct) => {
    setSelectedProduct(product);
    settingForm.setFieldsValue({
      virtualProduct: product.virtualProduct,
      autoReplyEnabled: product.autoReplyEnabled,
      autoDeliveryEnabled: product.autoDeliveryEnabled,
      autoRelistEnabled: product.autoRelistEnabled,
      replyDelaySeconds: product.replyDelaySeconds,
      relistDelaySeconds: product.relistDelaySeconds,
      relistDescription: product.relistDescription,
      relistImageUrls: product.relistImageUrls,
      relistAddress: product.relistAddress,
    });
  };

  const saveSetting = async () => {
    if (!selectedProduct) return;
    const values = await settingForm.validateFields();
    setProducts((current) =>
      current.map((product) =>
        product.ID === selectedProduct.ID ? { ...product, ...values } : product
      )
    );
    setSelectedProduct(undefined);
    message.success("预览设置已更新，不会写入真实业务数据");
  };

  const syncProducts = () => {
    const accountId = filters.getFieldValue("accountId");
    if (!accountId) {
      message.warning("请先选择需要同步的闲鱼账号");
      return;
    }
    setSyncing(true);
    window.setTimeout(() => {
      setSyncing(false);
      message.success("模拟同步完成，共刷新 6 个商品");
    }, 700);
  };

  const openProductDetail = (product: PreviewProduct) => {
    setDetailProduct(product);
    setPreviewPage("product-detail");
  };

  const navigateFromMenu = (key: string) => {
    if (key === "bot") setPreviewPage("bot");
    if (key === "products") setPreviewPage("products");
    if (key === "orders") setPreviewPage("order-detail");
    setMobileNavigationOpen(false);
  };

  const selectedMenuKey = previewPage === "bot" ? "bot" : previewPage === "order-detail" ? "orders" : "products";
  const pageTitle = previewPage === "bot" ? "Bot" : previewPage === "order-detail" ? "订单详情" : previewPage === "product-detail" ? "商品详情" : "商品管理";

  const columns: TableColumnsType<PreviewProduct> = [
    {
      title: "商品名称",
      key: "product",
      width: screens.sm ? 210 : 190,
      render: (_, product) => (
        <Typography.Link className="product-name" ellipsis onClick={() => openProductDetail(product)}>
          {product.title}
        </Typography.Link>
      ),
    },
    {
      title: "商品编码",
      dataIndex: "externalProductId",
      width: 132,
      render: (externalProductId: string) => (
        <Typography.Text className="product-code" copyable>
          {externalProductId}
        </Typography.Text>
      ),
    },
    {
      title: "图片",
      key: "image",
      width: 56,
      align: "center",
      render: (_, product) => (
        <Image.PreviewGroup items={product.imageUrls}>
          <Image
            className="product-list-image"
            src={product.imageUrls[0]}
            alt={`${product.title}商品图片`}
            width={30}
            height={30}
          />
        </Image.PreviewGroup>
      ),
    },
    {
      title: "账号",
      dataIndex: "accountLabel",
      width: 116,
      responsive: ["xxl"],
    },
    {
      title: "价格",
      dataIndex: "priceCent",
      width: 84,
      responsive: ["xl"],
      render: (priceCent: number) => (
        <span className="product-price">¥{(priceCent / 100).toFixed(2)}</span>
      ),
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 78,
      render: (status: PreviewProduct["status"]) => (
        <StatusTag status={status} />
      ),
    },
    {
      title: "自动化",
      key: "automation",
      width: 196,
      responsive: ["xl"],
      render: (_, product) => (
        <div className="capability-group">
          <CapabilityTag label="回复" enabled={product.autoReplyEnabled} />
          <CapabilityTag label="发货" enabled={product.autoDeliveryEnabled} />
          <CapabilityTag label="续售" enabled={product.autoRelistEnabled} />
        </div>
      ),
    },
    {
      title: "最后同步",
      dataIndex: "updatedAt",
      width: 100,
      responsive: ["xxl"],
    },
    {
      title: "操作",
      key: "operation",
      width: screens.sm ? 86 : 78,
      align: "center",
      render: (_, product) =>
        product.status === "on_sale" ? (
          <Button type="link" onClick={() => openSetting(product)}>
            业务设置
          </Button>
        ) : (
          <Typography.Text type="secondary">不可设置</Typography.Text>
        ),
    },
  ];

  return (
    <div className="product-preview">
      <Layout className="preview-layout">
        <Sider className="preview-sider" width={screens.xl ? 228 : 196}>
          <Brand />
          <Menu
            mode="inline"
            items={navigation}
            selectedKeys={[selectedMenuKey]}
            defaultOpenKeys={["xianyu-automation"]}
            onClick={({ key }) => navigateFromMenu(key)}
          />
        </Sider>

        <Layout>
          <Header className="preview-header">
            <Button
              className="mobile-menu-button"
              onClick={() => setMobileNavigationOpen(true)}
              aria-label="打开导航"
            >
              导航
            </Button>
            <div className="header-context">
              <Typography.Text type="secondary">闲鱼自动化</Typography.Text>
              <span>/</span>
              <Typography.Text>{pageTitle}</Typography.Text>
            </div>
          </Header>

          <Content
            className={`preview-content ${previewPage === "products" ? "product-list-content" : ""} ${previewPage === "bot" ? "bot-preview-content" : ""}`}
          >
            {previewPage === "products" && (
              <>
            <section className="filter-panel" aria-label="商品筛选">
              <Form<ProductFilters>
                form={filters}
                layout="horizontal"
                onFinish={(values) => {
                  setAppliedFilters(values);
                  setPage(1);
                }}
              >
                <div className="filter-grid">
                  <Form.Item name="accountId" label="闲鱼账号">
                    <Select allowClear placeholder="全部账号" options={previewAccounts} />
                  </Form.Item>
                  <Form.Item name="titleOrId" label="标题或商品 ID">
                    <Input allowClear placeholder="输入标题或闲鱼商品 ID" />
                  </Form.Item>
                  <Form.Item name="status" label="商品状态">
                    <Select
                      allowClear
                      placeholder="全部状态"
                      options={[
                        { label: "在售", value: "on_sale" },
                        { label: "已下架", value: "offline" },
                        { label: "已售出", value: "sold" },
                      ]}
                    />
                  </Form.Item>
                  <div className="filter-actions">
                    <Button type="primary" htmlType="submit">查询</Button>
                    <Button
                      onClick={() => {
                        filters.resetFields();
                        setAppliedFilters({});
                        setPage(1);
                      }}
                    >
                      重置
                    </Button>
                    <Button type="primary" loading={syncing} onClick={syncProducts}>
                      同步商品
                    </Button>
                  </div>
                </div>
              </Form>
            </section>

            <section className="product-panel">
              <Table<PreviewProduct>
                columns={columns}
                dataSource={filteredProducts}
                rowKey="ID"
                pagination={false}
                tableLayout="fixed"
                locale={{ emptyText: "没有符合当前筛选条件的商品" }}
              />
              <div className="table-pagination">
                <Pagination
                  current={page}
                  pageSize={pageSize}
                  total={36}
                  showSizeChanger
                  showQuickJumper
                  showTotal={(total) => `共 ${total} 条`}
                  onChange={(nextPage, nextPageSize) => {
                    setPage(nextPageSize === pageSize ? nextPage : 1);
                    setPageSize(nextPageSize);
                  }}
                />
              </div>
            </section>
              </>
            )}
            {previewPage === "product-detail" && (
              <ProductDetailPreview
                product={detailProduct}
                onBack={() => setPreviewPage("products")}
                onOpenSetting={() => openSetting(detailProduct)}
                onSync={() => message.success("模拟同步完成，不会修改真实商品")}
              />
            )}
            {previewPage === "order-detail" && (
              <OrderDetailPreview onBack={() => setPreviewPage("product-detail")} />
            )}
            {previewPage === "bot" && <BotWorkspacePreview />}
          </Content>
        </Layout>
      </Layout>

      <Drawer
        title={<Brand />}
        placement="left"
        width={280}
        open={mobileNavigationOpen}
        onClose={() => setMobileNavigationOpen(false)}
        className="mobile-navigation"
      >
        <Menu
          mode="inline"
          items={navigation}
          selectedKeys={[selectedMenuKey]}
          defaultOpenKeys={["xianyu-automation"]}
          onClick={({ key }) => navigateFromMenu(key)}
        />
      </Drawer>

      <ProductSettingModal
        product={selectedProduct}
        form={settingForm}
        onCancel={() => setSelectedProduct(undefined)}
        onSave={() => void saveSetting()}
      />
    </div>
  );
}

function Brand() {
  return (
    <div className="preview-brand">
      <strong>闲鱼运营</strong>
    </div>
  );
}

function StatusTag({ status }: { status: PreviewProduct["status"] }) {
  return <span className={`status-badge status-${status}`}>{statusLabels[status]}</span>;
}

function CapabilityTag({ label, enabled }: { label: string; enabled: boolean }) {
  return (
    <span className={`capability-state ${enabled ? "enabled" : "disabled"}`}>
      {label}{enabled ? "开启" : "关闭"}
    </span>
  );
}

function ProductSettingModal({
  product,
  form,
  onCancel,
  onSave,
}: {
  product?: PreviewProduct;
  form: ReturnType<typeof Form.useForm<ProductSettingValues>>[0];
  onCancel: () => void;
  onSave: () => void;
}) {
  const relistEnabled = Form.useWatch("autoRelistEnabled", form);

  return (
    <Modal
      title="商品业务设置"
      open={Boolean(product)}
      width={920}
      className="product-setting-modal"
      okText="保存预览设置"
      cancelText="取消"
      onCancel={onCancel}
      onOk={onSave}
      destroyOnHidden
    >
      {product && (
        <div className="setting-product">
          <Typography.Text strong>{product.title}</Typography.Text>
          <Typography.Text type="secondary">{product.externalProductId}</Typography.Text>
        </div>
      )}
      <Form<ProductSettingValues> form={form} layout="vertical" className="setting-form">
        <div className="setting-form-grid">
          <Form.Item name="virtualProduct" label="虚拟商品" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="autoReplyEnabled" label="自动回复" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="autoDeliveryEnabled" label="自动发货" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="autoRelistEnabled" label="自动续售" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="replyDelaySeconds" label="回复延迟">
            <InputNumber min={0} max={3600} suffix="秒" />
          </Form.Item>
          {relistEnabled && (
            <>
              <Form.Item name="relistDelaySeconds" label="续售延迟">
                <InputNumber min={0} max={86400} suffix="秒" />
              </Form.Item>
              <Form.Item
                name="relistAddress"
                label="发布所在地"
                className="setting-field-span-2"
                rules={[{ required: true, whitespace: true, message: "请输入发布所在地" }]}
              >
                <Input placeholder="例如：杭州市西湖区" />
              </Form.Item>
              <Form.Item name="relistDescription" label="续售描述" className="setting-field-span-2">
                <Input.TextArea
                  rows={3}
                  maxLength={5000}
                  showCount
                  placeholder="留空则读取原商品完整描述"
                />
              </Form.Item>
              <Form.Item name="relistImageUrls" label="续售图片" className="setting-field-span-2">
                <Input.TextArea rows={3} placeholder="每行一个图片 URL，留空则使用原商品图片" />
              </Form.Item>
              <Alert
                className="setting-note"
                type="info"
                showIcon
                title="自动续售发布时需要所在地；虚拟商品也需要填写。"
              />
            </>
          )}
        </div>
      </Form>
    </Modal>
  );
}
