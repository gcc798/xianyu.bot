import { Form, Input, Modal } from "antd";
import type { AccountInput, XianyuAccount } from "../../../../models/account";

export type AccountEditorInput = Pick<AccountInput, "label" | "note">;

type Props = {
  account: XianyuAccount | null | undefined;
  loading: boolean;
  onCancel: () => void;
  onSave: (input: AccountEditorInput) => Promise<void>;
};

export default function AccountEditorModal({ account, loading, onCancel, onSave }: Props) {
  const [form] = Form.useForm<AccountEditorInput>();
  const open = account !== undefined;

  return (
    <Modal
      open={open}
      title={account ? "编辑账号" : "新增账号"}
      okText="保存"
      cancelText="取消"
      confirmLoading={loading}
      destroyOnHidden
      onCancel={onCancel}
      onOk={() => form.submit()}
      afterOpenChange={(visible) => {
        if (visible)
          form.setFieldsValue({ label: account?.label || "", note: account?.note || "" });
      }}
    >
      <Form form={form} layout="vertical" onFinish={onSave}>
        <Form.Item
          name="label"
          label="账号名称"
          rules={[
            { required: true, whitespace: true, message: "请输入账号名称" },
            { max: 80, message: "最多 80 个字符" },
          ]}
        >
          <Input autoFocus />
        </Form.Item>
        <Form.Item name="note" label="备注" rules={[{ max: 500, message: "最多 500 个字符" }]}>
          <Input.TextArea rows={3} showCount maxLength={500} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
