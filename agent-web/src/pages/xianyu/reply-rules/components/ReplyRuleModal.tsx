import { Form, Input, InputNumber, Modal, Select, Space, Switch } from "antd";
import { useEffect } from "react";
import type { ReplyRule, ReplyRuleInput } from "../../../../models/automation";
import type { Product } from "../../../../models/product";

type Props = {
  record: ReplyRule | null | undefined;
  products: Product[];
  defaultProductId?: number;
  saving: boolean;
  onCancel: () => void;
  onSave: (input: ReplyRuleInput) => Promise<void>;
};

export function ReplyRuleModal({
  record,
  products,
  defaultProductId,
  saving,
  onCancel,
  onSave,
}: Props) {
  const [form] = Form.useForm<ReplyRuleInput>();
  const type = Form.useWatch("type", form);

  useEffect(() => {
    if (record === undefined) return;
    form.setFieldsValue(
      record ?? {
        ID: 0,
        productId: defaultProductId,
        type: "fuzzy",
        keyword: "",
        priority: 100,
        enabled: true,
        content: "",
      },
    );
  }, [defaultProductId, form, record]);

  return (
    <Modal
      open={record !== undefined}
      title={record ? "编辑回复规则" : "新增回复规则"}
      okText="保存"
      cancelText="取消"
      confirmLoading={saving}
      destroyOnHidden
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
        <Form.Item name="type" label="规则类型" rules={[{ required: true }]}>
          <Select
            disabled={record?.type === "exact"}
            options={
              record?.type === "exact"
                ? [{ value: "exact", label: "精确匹配" }]
                : [
                    { value: "fuzzy", label: "模糊匹配" },
                    { value: "fallback", label: "兜底回复" },
                  ]
            }
          />
        </Form.Item>
        {type !== "fallback" && (
          <Form.Item
            name="keyword"
            label="触发关键词"
            rules={[
              { required: true, whitespace: true, message: "请输入触发关键词" },
              { max: 255 },
            ]}
          >
            <Input />
          </Form.Item>
        )}
        <Space size={24} align="start">
          <Form.Item name="priority" label="优先级" rules={[{ required: true }]}>
            <InputNumber min={0} max={10000} />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Space>
        <Form.Item
          name="content"
          label="回复内容"
          rules={[{ required: true, whitespace: true }, { max: 4000 }]}
        >
          <Input.TextArea rows={5} maxLength={4000} showCount />
        </Form.Item>
      </Form>
    </Modal>
  );
}
