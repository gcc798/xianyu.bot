import { Alert, Form, Input, InputNumber, Modal, Switch } from "antd";
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
        if (visible && product) form.setFieldsValue({
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
      <Form form={form} layout="vertical" onFinish={(input) => { if (product) return onSave({ ...input, productId: product.ID }); }} className={styles.modalGrid}>
        <Form.Item name="virtualProduct" label="虚拟商品" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="autoReplyEnabled" label="自动回复" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="autoDeliveryEnabled" label="自动发货" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="autoRelistEnabled" label="自动续售" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="replyDelaySeconds" label="回复延迟（秒）" rules={[{ required: true, message: "请输入回复延迟" }]}><InputNumber min={0} max={3600} style={{ width: "100%" }} /></Form.Item>
        <Form.Item name="relistDelaySeconds" label="续售延迟（秒）" rules={autoRelist ? [{ required: true, message: "请输入续售延迟" }] : []}><InputNumber min={0} max={86400} disabled={!autoRelist} style={{ width: "100%" }} /></Form.Item>
        {autoRelist && <>
          <Form.Item name="relistAddress" label="发布所在地" className="ant-form-item-wide" rules={[{ required: true, whitespace: true, message: "开启自动续售时必须填写发布所在地" }]}><Input placeholder="例如：杭州市西湖区" /></Form.Item>
          <Form.Item name="relistDescription" label="续售描述" className="ant-form-item-wide"><Input.TextArea rows={3} maxLength={5000} showCount placeholder="留空时使用原商品描述" /></Form.Item>
          <Form.Item name="relistImageUrls" label="续售图片" className="ant-form-item-wide"><Input.TextArea rows={2} placeholder="每行一个图片 URL，留空时使用原商品图片" /></Form.Item>
          <Form.Item className="ant-form-item-wide"><Alert type="info" showIcon title="自动续售发布需要所在地；虚拟商品同样需要填写。" /></Form.Item>
        </>}
      </Form>
    </Modal>
  );
}
