import { Form, Input, Modal, Select, Space, Switch } from "antd";
import { useEffect } from "react";
import type { DeliveryRule, DeliveryRuleInput } from "../../../../models/automation";
import type { Product } from "../../../../models/product";

type Props = {
  record: DeliveryRule | null | undefined;
  products: Product[];
  defaultProductId?: number;
  saving: boolean;
  onCancel: () => void;
  onSave: (input: DeliveryRuleInput) => Promise<void>;
};

export function DeliveryRuleModal({
  record,
  products,
  defaultProductId,
  saving,
  onCancel,
  onSave,
}: Props) {
  const [form] = Form.useForm<DeliveryRuleInput>();
  const mode = Form.useWatch("mode", form);

  useEffect(() => {
    if (record === undefined) return;
    form.setFieldsValue(
      record ?? {
        ID: 0,
        productId: defaultProductId,
        skuKey: "",
        skuName: "",
        mode: "text",
        template: "",
        enabled: true,
      },
    );
  }, [defaultProductId, form, record]);

  return (
    <Modal
      open={record !== undefined}
      title={record ? "编辑发货规则" : "新增发货规则"}
      okText="保存"
      cancelText="取消"
      confirmLoading={saving}
      destroyOnHidden
      width={720}
      onCancel={onCancel}
      onOk={() => form.submit()}
    >
      <Form form={form} layout="vertical" onFinish={onSave}>
        <Form.Item
          name="productId"
          label="商品"
          rules={[{ required: true, message: "请选择商品" }]}
        >
          <Select
            showSearch
            options={products.map((item) => ({ value: item.ID, label: item.title }))}
          />
        </Form.Item>
        <Space size={16} align="start" style={{ width: "100%" }}>
          <Form.Item name="skuKey" label="SKU 标识">
            <Input maxLength={128} />
          </Form.Item>
          <Form.Item name="skuName" label="SKU 名称">
            <Input maxLength={255} />
          </Form.Item>
          <Form.Item name="mode" label="发货模式" rules={[{ required: true }]}>
            <Select
              style={{ width: 150 }}
              options={[
                { value: "text", label: "文本发货" },
                { value: "card", label: "卡密发货" },
              ]}
            />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Space>
        <Form.Item
          name="template"
          label="发货模板"
          extra={mode === "card" ? "使用 {card} 表示卡密内容" : undefined}
          rules={[{ required: true, whitespace: true }, { max: 4000 }]}
        >
          <Input.TextArea rows={6} maxLength={4000} showCount />
        </Form.Item>
      </Form>
    </Modal>
  );
}
