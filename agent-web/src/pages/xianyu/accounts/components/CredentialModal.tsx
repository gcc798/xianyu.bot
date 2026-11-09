import { Alert, Form, Input, Modal } from "antd";
import type { AccountCredentialInput, XianyuAccount } from "../../../../models/account";

type CredentialForm = Omit<AccountCredentialInput, "ID">;

type Props = {
  account: XianyuAccount | null;
  loading: boolean;
  onCancel: () => void;
  onSave: (input: AccountCredentialInput) => Promise<void>;
};

export default function CredentialModal({ account, loading, onCancel, onSave }: Props) {
  const [form] = Form.useForm<CredentialForm>();
  return (
    <Modal
      open={Boolean(account)}
      title="设置账号凭证"
      width={680}
      okText="保存凭证"
      cancelText="取消"
      confirmLoading={loading}
      destroyOnHidden
      onCancel={onCancel}
      onOk={() => form.submit()}
      afterOpenChange={(visible) => {
        if (visible) form.setFieldsValue({ cookie: "", token: "" });
      }}
    >
      <Alert
        type="warning"
        showIcon
        title="凭证保存后不再回显；请仅填写需要更新的账号凭证。"
        style={{ marginBottom: 18 }}
      />
      <Form
        form={form}
        layout="vertical"
        onFinish={(input) => {
          if (account) return onSave({ ...input, ID: account.ID });
        }}
      >
        <Form.Item
          name="cookie"
          label="Cookie"
          rules={[{ required: true, whitespace: true, message: "请输入 Cookie" }]}
        >
          <Input.TextArea rows={5} autoComplete="off" />
        </Form.Item>
        <Form.Item name="token" label="Token">
          <Input.Password autoComplete="new-password" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
