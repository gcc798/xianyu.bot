import { Button, Input, Space, Tabs, Tag } from "antd";

const sessions = [
  { id: "session-1", title: "查询今日待发货订单", summary: "3 笔订单待处理", time: "10:36", active: true },
  { id: "session-2", title: "检查自动回复命中情况", summary: "分析最近 24 小时消息", time: "昨天", active: false },
  { id: "session-3", title: "整理商品同步异常", summary: "定位 2 个失败账号", time: "11 月 2 日", active: false },
  { id: "session-4", title: "统计本周售后问题", summary: "按原因汇总订单", time: "10 月 31 日", active: false },
];

const traceEvents = [
  { role: "用户", title: "查询待发货订单", detail: "今天新增且仍待发货的订单", tone: "input" },
  { role: "模型", title: "识别查询范围", detail: "今天 00:00 至当前时间，状态为待发货", tone: "model" },
  { role: "工具", title: "xianyu_order_list", detail: "返回 3 条订单记录 · 182 ms", tone: "tool" },
  { role: "模型", title: "整理优先级", detail: "标记超时订单与自动交付订单", tone: "model" },
  { role: "助手", title: "生成回复", detail: "输出订单摘要与详情入口", tone: "assistant" },
];

export default function BotWorkspacePreview() {
  return (
    <div className="bot-page">
      <section className="bot-shell">
        <aside className="bot-sessions">
          <div className="bot-section-header">
            <strong>会话</strong>
            <Button type="primary" size="small">新建会话</Button>
          </div>
          <Input allowClear placeholder="搜索会话" aria-label="搜索会话" />
          <div className="session-list">
            {sessions.map((session) => (
              <button key={session.id} className={session.active ? "active" : ""}>
                <span className="session-title-row">
                  <strong>{session.title}</strong>
                  <time>{session.time}</time>
                </span>
                <span>{session.summary}</span>
              </button>
            ))}
          </div>
        </aside>

        <main className="bot-main">
          <div className="conversation-header">
            <div>
              <strong>查询今日待发货订单</strong>
              <span>DeepSeek · 当前用户权限范围</span>
            </div>
            <Button>更多</Button>
          </div>
          <Tabs
            className="bot-view-tabs"
            classNames={{ body: "bot-tabs-body", content: "bot-tabs-content" }}
            defaultActiveKey="conversation"
            items={[
              { key: "conversation", label: "对话", children: <ConversationView /> },
              { key: "trace", label: "轨迹", children: <TraceView /> },
            ]}
          />
        </main>
      </section>
    </div>
  );
}

function ConversationView() {
  return (
    <div className="conversation-view">
      <div className="message-stream">
        <div className="chat-message user-message">列出今天新增且仍待发货的订单。</div>
        <div className="chat-message bot-message">
          <div>
            <p>今天共有 3 笔新增待发货订单，其中 1 笔已超过 2 小时。</p>
            <div className="bot-result-row">
              <div><strong>XY2026110300186</strong><span>Apple 妙控键盘 2 代</span></div>
              <Tag color="warning">待发货 2 小时</Tag>
              <Button type="link">查看订单</Button>
            </div>
            <div className="bot-result-row">
              <div><strong>XY2026110300214</strong><span>Notion 商业运营模板</span></div>
              <Tag color="processing">待自动交付</Tag>
              <Button type="link">查看订单</Button>
            </div>
          </div>
        </div>
      </div>
      <div className="bot-composer">
        <Input.TextArea autoSize={{ minRows: 2, maxRows: 5 }} placeholder="输入运营问题或任务" />
        <Button type="primary">发送</Button>
      </div>
    </div>
  );
}

function TraceView() {
  return (
    <div className="trace-view">
      <div className="trace-toolbar">
        <Space size={20} wrap>
          <span>用时 <strong>1.8 秒</strong></span>
          <span>轮次 <strong>1</strong></span>
          <span>调用 <strong>1</strong></span>
        </Space>
        <Input allowClear placeholder="搜索轨迹" aria-label="搜索轨迹" />
      </div>

      <div className="trace-timeline" aria-label="执行时间概览">
        <div className="trace-track"><span>输入</span><i className="input-segment" /></div>
        <div className="trace-track"><span>模型</span><i className="model-segment first" /><i className="model-segment second" /></div>
        <div className="trace-track"><span>工具</span><i className="tool-segment" /></div>
      </div>

      <div className="trace-body">
        <div className="trace-events" role="list" aria-label="执行事件">
          {traceEvents.map((event, index) => (
            <button key={`${event.role}-${event.title}`} className={index === 2 ? "selected" : ""} role="listitem">
              <Tag className={`trace-role trace-role-${event.tone}`}>{event.role}</Tag>
              <div><strong>{event.title}</strong><span>{event.detail}</span></div>
              <time>{index === 0 ? "0 ms" : `${index * 362} ms`}</time>
            </button>
          ))}
        </div>

        <aside className="trace-inspector">
          <div className="inspector-title"><strong>工具调用 #1</strong><Tag color="success">已完成</Tag></div>
          <Tabs
            defaultActiveKey="overview"
            items={[
              {
                key: "overview",
                label: "概述",
                children: (
                  <dl className="trace-facts">
                    <div><dt>工具</dt><dd>xianyu_order_list</dd></div>
                    <div><dt>状态</dt><dd>成功</dd></div>
                    <div><dt>返回记录</dt><dd>3 条</dd></div>
                    <div><dt>耗时</dt><dd>182 ms</dd></div>
                  </dl>
                ),
              },
              { key: "input", label: "输入", children: <pre className="trace-code">{`{\n  "status": "pending_delivery",\n  "createdToday": true\n}`}</pre> },
              { key: "output", label: "输出", children: <div className="trace-output">返回 3 条符合条件的订单记录。</div> },
            ]}
          />
        </aside>
      </div>
    </div>
  );
}
