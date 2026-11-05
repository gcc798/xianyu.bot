import { Alert, Input, Modal } from "antd";
import type { DeliveryRule } from "../../../../models/automation";

type Props = {
  rule: DeliveryRule | null;
  value: string;
  saving: boolean;
  onChange: (value: string) => void;
  onCancel: () => void;
  onImport: () => Promise<void>;
};

export function ResourceImportModal({ rule, value, saving, onChange, onCancel, onImport }: Props) {
  return (
    <Modal
      open={Boolean(rule)}
      title="导入卡密"
      okText="导入"
      cancelText="取消"
      confirmLoading={saving}
      onCancel={onCancel}
      onOk={() => void onImport()}
    >
      <Alert type="info" showIcon title="每行一条，重复内容自动忽略。" style={{ marginBottom: 16 }} />
      <Input.TextArea
        value={value}
        rows={10}
        onChange={(event) => onChange(event.target.value)}
        placeholder={"CARD-001\nCARD-002"}
      />
    </Modal>
  );
}
