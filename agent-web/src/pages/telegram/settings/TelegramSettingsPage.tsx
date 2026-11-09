import {
  Alert,
  App,
  Button,
  Descriptions,
  Form,
  Input,
  Space,
  Spin,
  Switch,
  Tabs,
  Tag,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { telegramApi } from "../../../api/telegram";
import type { TelegramAISettingsInput, TelegramSettings } from "../../../models/telegram";
import styles from "../../../styles/page.module.css";
import { formatFileSize } from "../../../utils/format";

export function Component() {
  const { message } = App.useApp();
  const [form] = Form.useForm<TelegramAISettingsInput>();
  const [settings, setSettings] = useState<TelegramSettings>();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [probing, setProbing] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const data = await telegramApi.settings();
      setSettings(data);
      form.setFieldsValue({ base_url: data.ai_base_url, model: data.ai_model, api_key: "" });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Telegram 设置加载失败");
    } finally {
      setLoading(false);
    }
  }, [form]);
  useEffect(() => {
    void load();
  }, [load]);

  const updateAdultMode = async (adultMode: boolean) => {
    try {
      const data = await telegramApi.updateAdultMode(adultMode);
      setSettings((current) =>
        current
          ? { ...current, adult_mode: data.adult_mode, adult_dialog_count: data.adult_dialog_count }
          : current,
      );
      message.success("内容范围已更新");
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "内容范围更新失败");
    }
  };
  const saveAI = async (input: TelegramAISettingsInput) => {
    setSaving(true);
    try {
      const data = await telegramApi.updateAISettings(input);
      setSettings((current) => (current ? { ...current, ...data } : current));
      form.setFieldsValue({
        base_url: data.ai_base_url,
        model: data.ai_model,
        api_key: input.api_key,
      });
      message.success("Telegram AI 设置已保存");
    } catch (cause) {
      message.error(cause instanceof Error ? cause.message : "Telegram AI 设置保存失败");
    } finally {
      setSaving(false);
    }
  };
  const probe = async () => {
    const input = await form.validateFields();
    setProbing(true);
    try {
      const data = await telegramApi.probeAISettings(input);
      if (data.available) message.success("AI 服务连接正常");
      else message.error("AI 服务不可用");
    } finally {
      setProbing(false);
    }
  };

  return (
    <div className={styles.page}>
      <section className={styles.surface} style={{ padding: "0 20px 20px" }}>
        {error && <Alert className={styles.alert} type="error" showIcon title={error} />}
        <Spin spinning={loading}>
          <Tabs
            items={[
              {
                key: "service",
                label: "服务设置",
                children: (
                  <>
                    <Descriptions
                      bordered
                      column={2}
                      size="small"
                      items={[
                        {
                          key: "api",
                          label: "Telegram API",
                          children: (
                            <Tag color={settings?.telegram_api_configured ? "success" : "warning"}>
                              {settings?.telegram_api_configured ? "已配置" : "未配置"}
                            </Tag>
                          ),
                        },
                        {
                          key: "login",
                          label: "登录状态",
                          children: (
                            <Tag color={settings?.authorized ? "success" : "default"}>
                              {settings?.authorized ? "已登录" : "未登录"}
                            </Tag>
                          ),
                        },
                        { key: "account", label: "账号", children: settings?.account || "-" },
                        { key: "proxy", label: "代理", children: settings?.proxy || "未配置" },
                        { key: "listen", label: "监听地址", children: settings?.listen || "-" },
                        {
                          key: "file",
                          label: "单文件上限",
                          children: settings
                            ? formatFileSize(settings.resource_max_file_size)
                            : "-",
                        },
                        {
                          key: "concurrency",
                          label: "资源并发数",
                          children: settings?.resource_concurrency ?? "-",
                        },
                        {
                          key: "adult",
                          label: "内容范围",
                          children: (
                            <Space>
                              <Switch
                                checked={settings?.adult_mode}
                                onChange={(value) => void updateAdultMode(value)}
                              />
                              <span>
                                {settings?.adult_mode
                                  ? `包含成人内容会话（${settings.adult_dialog_count}）`
                                  : "仅常规会话"}
                              </span>
                            </Space>
                          ),
                        },
                      ]}
                    />
                  </>
                ),
              },
              {
                key: "ai",
                label: "资源识别 AI",
                forceRender: true,
                children: (
                  <Form form={form} layout="vertical" onFinish={saveAI} style={{ maxWidth: 760 }}>
                    <Form.Item
                      name="base_url"
                      label="服务地址"
                      rules={[{ required: true, type: "url" }]}
                    >
                      <Input />
                    </Form.Item>
                    <Form.Item
                      name="model"
                      label="模型标识"
                      rules={[{ required: true, whitespace: true }]}
                    >
                      <Input />
                    </Form.Item>
                    <Form.Item name="api_key" label="API Key">
                      <Input.Password
                        autoComplete="new-password"
                        placeholder={
                          settings?.ai_api_key_configured
                            ? settings.ai_api_key_masked
                            : "输入 API Key"
                        }
                      />
                    </Form.Item>
                    <Space>
                      <Button type="primary" htmlType="submit" loading={saving}>
                        保存
                      </Button>
                      <Button loading={probing} onClick={() => void probe()}>
                        测试连接
                      </Button>
                    </Space>
                  </Form>
                ),
              },
            ]}
          />
        </Spin>
      </section>
    </div>
  );
}
