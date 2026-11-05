import { Alert, App as AntdApp, Button, Descriptions, Image, Space, Tabs, Tag, Timeline, Typography } from "antd";
import type { DescriptionsProps } from "antd";
import { previewProducts } from "./previewData";

const orderProduct = previewProducts[0];

export default function OrderDetailPreview({ onBack }: { onBack: () => void }) {
  const { message, modal } = AntdApp.useApp();
  const orderFields: DescriptionsProps["items"] = [
    { key: "orderId", label: "闲鱼订单号", children: <Typography.Text copyable>XY2026110300186</Typography.Text> },
    { key: "status", label: "订单状态", children: <Tag color="warning">待发货</Tag> },
    { key: "account", label: "所属账号", children: "数码严选店" },
    { key: "buyer", label: "买家", children: "林先生" },
    { key: "amount", label: "实付金额", children: "¥468.00" },
    { key: "created", label: "下单时间", children: "2026-11-03 10:28:16" },
  ];

  return (
    <div className="detail-page">
      <div className="detail-heading">
        <div>
          <Button type="link" className="back-button" onClick={onBack}>返回商品详情</Button>
          <Space size={10} wrap>
            <Typography.Title level={2}>订单 XY2026110300186</Typography.Title>
            <Tag color="warning">待发货</Tag>
          </Space>
        </div>
        <Space>
          <Button onClick={() => message.info("预览模式不会向真实买家发送消息")}>联系买家</Button>
          <Button
            type="primary"
            onClick={() =>
              modal.confirm({
                title: "确认发货？",
                content: "请确认物流或交付内容已准备完成。预览操作不会写入真实订单。",
                okText: "确认发货",
                cancelText: "取消",
                onOk: () => message.success("预览发货已确认，不会写入真实订单"),
              })
            }
          >
            确认发货
          </Button>
        </Space>
      </div>

      <section className="detail-workspace">
        <Tabs
          className="detail-tabs"
          defaultActiveKey="detail"
          items={[
            {
              key: "detail",
              label: "订单详情",
              children: (
                <div className="detail-tab-body">
                  <Alert
                    type="warning"
                    showIcon
                    title="订单尚未发货"
                    description="请在确认交付内容无误后执行发货；预览中的操作不会写入真实订单。"
                    className="order-alert"
                  />
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>订单信息</Typography.Title>
                    </div>
                    <Descriptions bordered size="small" column={{ xs: 1, md: 2, xl: 3 }} items={orderFields} />
                  </section>
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>商品信息</Typography.Title>
                    </div>
                    <div className="order-product-row">
                      <Image
                        className="order-product-image"
                        src={orderProduct.imageUrls[0]}
                        alt="Apple 妙控键盘 2 代商品图片"
                        width={44}
                        height={44}
                      />
                      <div>
                        <strong>Apple 妙控键盘 2 代｜深空灰｜成色 95 新</strong>
                        <span>商品 ID：873291604521</span>
                      </div>
                      <strong>¥468.00</strong>
                    </div>
                  </section>
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>买家与收货信息</Typography.Title>
                    </div>
                    <Descriptions
                      bordered
                      size="small"
                      column={{ xs: 1, md: 2 }}
                      items={[
                        { key: "name", label: "收货人", children: "林先生" },
                        { key: "phone", label: "联系电话", children: "138****6201" },
                        { key: "address", label: "收货地址", span: 2, children: "浙江省杭州市西湖区 ···· 18 号" },
                        { key: "note", label: "买家留言", span: 2, children: "工作日发货即可，谢谢。" },
                      ]}
                    />
                  </section>
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>交付策略</Typography.Title>
                      <Button type="link" onClick={() => message.info("当前商品未启用自动发货规则")}>查看发货规则</Button>
                    </div>
                    <Descriptions
                      bordered
                      size="small"
                      column={{ xs: 1, md: 2 }}
                      items={[
                        { key: "method", label: "交付方式", children: <Tag>人工发货</Tag> },
                        { key: "rule", label: "规则状态", children: "未启用自动发货规则" },
                        { key: "note", label: "处理说明", span: 2, children: "成交后由运营人员确认物流信息。" },
                      ]}
                    />
                  </section>
                </div>
              ),
            },
            {
              key: "fulfillment",
              label: "履约记录",
              children: (
                <div className="detail-tab-body">
                  <section className="detail-section">
                    <div className="detail-section-heading">
                      <Typography.Title level={4}>履约进度</Typography.Title>
                    </div>
                    <Timeline
                      className="detail-timeline"
                      items={[
                        { color: "green", children: <><strong>买家付款成功</strong><span>11-03 10:28:16</span></> },
                        { color: "blue", children: <><strong>等待卖家发货</strong><span>自动发货规则未启用</span></> },
                        { color: "gray", children: <><strong>买家确认收货</strong><span>尚未发生</span></> },
                      ]}
                    />
                  </section>
                </div>
              ),
            },
          ]}
        />
      </section>
    </div>
  );
}
