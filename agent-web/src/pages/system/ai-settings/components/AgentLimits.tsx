import { App, Button, Form, InputNumber, Spin } from "antd";
import { useEffect, useState } from "react";
import { aiConfigApi } from "../../../../api/ai-config";
import type { AgentConfig } from "../../../../models/ai-config";
import styles from "../AISettingsPage.module.css";

const fields: { name: keyof Omit<AgentConfig, "ID">; label: string; max: number }[] = [
  { name: "maxWallTimeSeconds", label: "最长运行时间（秒）", max: 600 },
  { name: "maxIterations", label: "最大循环次数", max: 50 },
  { name: "maxModelCalls", label: "模型调用次数", max: 100 },
  { name: "maxToolCalls", label: "工具调用次数", max: 200 },
  { name: "maxConcurrentToolCalls", label: "工具并发数", max: 20 },
  { name: "maxSearchCalls", label: "检索调用次数", max: 200 },
  { name: "maxWriteToolCalls", label: "写工具调用次数", max: 200 },
  { name: "maxOutputTokens", label: "最大输出 Token", max: 100000 },
  { name: "conversationHistoryLimit", label: "历史消息数量", max: 1000 },
  { name: "summaryTriggerTokens", label: "摘要触发 Token", max: 1000000 },
];

export function AgentLimits() {
  const { message } = App.useApp();
  const [form] = Form.useForm<AgentConfig>();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    void aiConfigApi
      .agentConfig()
      .then(({ config }) => form.setFieldsValue(config))
      .catch((cause) =>
        message.error(cause instanceof Error ? cause.message : "Agent 配置加载失败"),
      )
      .finally(() => setLoading(false));
  }, [form, message]);
  const save = async (input: AgentConfig) => {
    setSaving(true);
    try {
      await aiConfigApi.updateAgentConfig(input);
      message.success("Agent 配置已保存");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Spin spinning={loading}>
      <Form form={form} layout="vertical" onFinish={save}>
        <div className={styles.toolbar}>
          <Button type="primary" loading={saving} onClick={() => form.submit()}>
            保存配置
          </Button>
        </div>
        <div className={styles.grid}>
          {fields.map((field) => (
            <Form.Item
              key={field.name}
              name={field.name}
              label={field.label}
              rules={[{ required: true }]}
            >
              <InputNumber min={1} max={field.max} />
            </Form.Item>
          ))}
        </div>
        <Form.Item name="ID" hidden>
          <InputNumber />
        </Form.Item>
      </Form>
    </Spin>
  );
}
