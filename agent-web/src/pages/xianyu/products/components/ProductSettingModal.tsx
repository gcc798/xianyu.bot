import { Alert, Form, Input, InputNumber, Modal, Space, Switch } from "antd";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { automationApi } from "../../../../api/automation";
import type { Product, ProductSettingInput } from "../../../../models/product";
import styles from "../../../../styles/business.module.css";

type Props = {
  product: Product | null;
  loading: boolean;
  onCancel: () => void;
  onSave: (input: ProductSettingInput) => Promise<void>;
};

type ProductSettingForm = Omit<ProductSettingInput, "productId">;

export default function ProductSettingModal({ product, loading, onCancel, onSave }: Props) {
  const [form] = Form.useForm<ProductSettingForm>();
  const autoRelist = Form.useWatch("autoRelistEnabled", form);
  const [ruleStatus, setRuleStatus] = useState({
    loading: false,
    failed: false,
    reply: false,
    delivery: false,
  });

  useEffect(() => {
    if (!product) return;
    let active = true;
    setRuleStatus({ loading: true, failed: false, reply: false, delivery: false });
    void Promise.all([
      automationApi.replyRules({ page: 1, pageSize: 1, productId: product.ID, enabledOnly: true }),
      automationApi.deliveryRules({
        page: 1,
        pageSize: 1,
        productId: product.ID,
        enabledOnly: true,
      }),
    ])
      .then(([replyRules, deliveryRules]) => {
        if (active)
          setRuleStatus({
            loading: false,
            failed: false,
            reply: replyRules.total > 0,
            delivery: deliveryRules.total > 0,
          });
      })
      .catch(() => {
        if (active) setRuleStatus({ loading: false, failed: true, reply: false, delivery: false });
      });
    return () => {
      active = false;
    };
  }, [product]);

  return (
    <Modal
      open={Boolean(product)}
      title="商品业务设置"
      width={840}
      okText="保存"
      cancelText="取消"
      confirmLoading={loading}
      destroyOnHidden
      onCancel={onCancel}
      onOk={() => form.submit()}
      afterOpenChange={(visible) => {
        if (visible && product)
          form.setFieldsValue({
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
      }}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={(input) => {
          if (product) return onSave({ ...input, productId: product.ID });
        }}
        className={styles.modalGrid}
      >
        <Form.Item name="virtualProduct" label="虚拟商品" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="autoReplyEnabled" label="自动回复" valuePropName="checked">
          <Switch
            loading={ruleStatus.loading}
            disabled={ruleStatus.loading || (!ruleStatus.reply && !product?.autoReplyEnabled)}
          />
        </Form.Item>
        <Form.Item name="autoDeliveryEnabled" label="自动发货" valuePropName="checked">
          <Switch
            loading={ruleStatus.loading}
            disabled={ruleStatus.loading || (!ruleStatus.delivery && !product?.autoDeliveryEnabled)}
          />
        </Form.Item>
        {!ruleStatus.loading &&
          (ruleStatus.failed || !ruleStatus.reply || !ruleStatus.delivery) && (
            <Alert
              className="ant-form-item-wide"
              type={ruleStatus.failed ? "error" : "warning"}
              showIcon
              title={
                ruleStatus.failed
                  ? "规则状态检查失败，无法确认自动化是否已配置。"
                  : "配置并启用对应规则后，才能开启商品自动化。"
              }
              action={
                <Space size={12}>
                  {!ruleStatus.reply && (
                    <Link to={`/xianyu/reply-rules?productId=${product?.ID}`}>配置回复规则</Link>
                  )}
                  {!ruleStatus.delivery && (
                    <Link to={`/xianyu/delivery-rules?productId=${product?.ID}`}>配置发货规则</Link>
                  )}
                </Space>
              }
            />
          )}
        <Form.Item name="autoRelistEnabled" label="自动续售" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item
          name="replyDelaySeconds"
          label="回复延迟（秒）"
          rules={[{ required: true, message: "请输入回复延迟" }]}
        >
          <InputNumber min={0} max={3600} style={{ width: "100%" }} />
        </Form.Item>
        <Form.Item
          name="relistDelaySeconds"
          label="续售延迟（秒）"
          rules={autoRelist ? [{ required: true, message: "请输入续售延迟" }] : []}
        >
          <InputNumber min={0} max={86400} disabled={!autoRelist} style={{ width: "100%" }} />
        </Form.Item>
        {autoRelist && (
          <>
            <Form.Item
              name="relistAddress"
              label="发布所在地"
              className="ant-form-item-wide"
              rules={[
                { required: true, whitespace: true, message: "开启自动续售时必须填写发布所在地" },
              ]}
            >
              <Input placeholder="例如：杭州市西湖区" />
            </Form.Item>
            <Form.Item name="relistDescription" label="续售描述" className="ant-form-item-wide">
              <Input.TextArea
                rows={3}
                maxLength={5000}
                showCount
                placeholder="留空时使用原商品描述"
              />
            </Form.Item>
            <Form.Item name="relistImageUrls" label="续售图片" className="ant-form-item-wide">
              <Input.TextArea rows={2} placeholder="每行一个图片 URL，留空时使用原商品图片" />
            </Form.Item>
            <Form.Item className="ant-form-item-wide">
              <Alert type="info" showIcon title="自动续售发布需要所在地；虚拟商品同样需要填写。" />
            </Form.Item>
          </>
        )}
      </Form>
    </Modal>
  );
}
