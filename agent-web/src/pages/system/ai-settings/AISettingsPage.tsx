import { Tabs } from "antd";
import styles from "./AISettingsPage.module.css";
import { AgentLimits } from "./components/AgentLimits";
import { MCPSettings } from "./components/MCPSettings";
import { ModelSettings } from "./components/ModelSettings";

export function Component() {
  return <div className={styles.page}>
    <section className={styles.surface}>
      <Tabs
        items={[
          { key: "models", label: "模型配置", children: <ModelSettings /> },
          { key: "mcp", label: "MCP 服务", children: <MCPSettings /> },
          { key: "agent", label: "Agent 配置", children: <AgentLimits /> },
        ]}
      />
    </section>
  </div>;
}
