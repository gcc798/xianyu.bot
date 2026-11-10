-- +goose Up
WITH dictionaries(type, name, description) AS (
    VALUES
        ('product_status', '商品状态', '闲鱼商品状态'),
        ('order_status', '订单状态', '闲鱼订单状态'),
        ('account_status', '账号状态', '闲鱼账号登录状态'),
        ('account_connection_status', '账号连接状态', '闲鱼账号 WebSocket 连接状态'),
        ('reply_rule_type', '回复规则类型', '自动回复匹配方式'),
        ('message_direction', '消息方向', '闲鱼消息的收发方向'),
        ('business_event_type', '业务事件类型', '自动化审计事件类别'),
        ('business_event_status', '业务事件状态', '自动化审计事件结果'),
        ('agent_event_type', 'Agent 事件类型', 'Agent 运行轨迹事件类别'),
        ('delivery_task_status', '发货任务状态', '自动发货任务状态'),
        ('delivery_mode', '发货模式', '发货规则内容类型'),
        ('agent_resource_task_state', '资源任务状态', 'Agent 资源获取任务状态'),
        ('resource_state', '资源状态', '本地资源可用状态'),
        ('telegram_resource_task_status', 'Telegram 采集任务状态', 'Telegram 批量采集状态'),
        ('telegram_resource_source_mode', 'Telegram 来源模式', 'Telegram 采集来源选择模式'),
        ('telegram_resource_source_strategy', 'Telegram 选源策略', 'Telegram 智能模式实际选源方式'),
        ('telegram_file_status', 'Telegram 文件状态', 'Telegram 资源文件处理状态'),
        ('telegram_dialog_type', 'Telegram 会话类型', 'Telegram 群组与频道类型'),
        ('telegram_dialog_status', 'Telegram 会话状态', 'Telegram 会话可用状态'),
        ('telegram_dialog_filter', 'Telegram 会话筛选', 'Telegram 会话筛选条件'),
        ('mcp_health_status', 'MCP 健康状态', 'MCP 服务最近检查状态')
)
INSERT INTO s_dictionaries (created_at, updated_at, name, type, status, "desc")
SELECT now(), now(), dictionaries.name, dictionaries.type, true, dictionaries.description
FROM dictionaries
WHERE NOT EXISTS (
    SELECT 1 FROM s_dictionaries existing
    WHERE existing.type = dictionaries.type AND existing.deleted_at IS NULL
);

WITH entries(type, code, name, sort) AS (
    VALUES
        ('product_status', 'on_sale', '在售', 1),
        ('product_status', 'offline', '已下架', 2),
        ('product_status', 'sold', '已售出', 3),
        ('order_status', 'pending_delivery', '待发货', 1),
        ('order_status', 'shipped', '已发货', 2),
        ('order_status', 'unknown', '待识别', 3),
        ('account_status', 'ready', '正常', 1),
        ('account_status', 'expired', '已失效', 2),
        ('account_status', 'verification_required', '需验证', 3),
        ('account_status', 'unconfigured', '未配置', 4),
        ('account_connection_status', 'connected', '已连接', 1),
        ('account_connection_status', 'connecting', '连接中', 2),
        ('account_connection_status', 'reconnecting', '重连中', 3),
        ('account_connection_status', 'failed', '需处理', 4),
        ('reply_rule_type', 'exact', '精确匹配', 1),
        ('reply_rule_type', 'fuzzy', '模糊匹配', 2),
        ('reply_rule_type', 'fallback', '兜底回复', 3),
        ('message_direction', 'inbound', '买家发来', 1),
        ('message_direction', 'outbound', '卖家发出', 2),
        ('business_event_type', 'message', '消息接收', 1),
        ('business_event_type', 'delivery', '自动发货', 2),
        ('business_event_type', 'relist', '自动续售', 3),
        ('business_event_type', 'order_sync', '订单同步', 4),
        ('business_event_type', 'product_sync', '商品同步', 5),
        ('business_event_type', 'login_check', '登录检查', 6),
        ('business_event_type', 'websocket', 'WebSocket 登录', 7),
        ('business_event_type', 'websocket_alert', '连接告警', 8),
        ('agent_event_type', 'run.started', '开始运行', 1),
        ('agent_event_type', 'run.completed', '运行完成', 2),
        ('agent_event_type', 'run.failed', '运行失败', 3),
        ('agent_event_type', 'run.cancelled', '运行取消', 4),
        ('agent_event_type', 'turn.started', 'Turn 开始', 5),
        ('agent_event_type', 'turn.completed', 'Turn 完成', 6),
        ('agent_event_type', 'turn.failed', 'Turn 失败', 7),
        ('agent_event_type', 'step.started', 'Step 开始', 8),
        ('agent_event_type', 'step.completed', 'Step 完成', 9),
        ('agent_event_type', 'step.failed', 'Step 失败', 10),
        ('agent_event_type', 'system.prompt', 'System Prompt', 11),
        ('agent_event_type', 'request.snapshot', '模型请求快照', 12),
        ('agent_event_type', 'model.attempt.started', '模型请求', 13),
        ('agent_event_type', 'model.attempt.completed', '模型返回', 14),
        ('agent_event_type', 'model.attempt.failed', '模型失败', 15),
        ('agent_event_type', 'model.fallback', '模型降级', 16),
        ('agent_event_type', 'model.unavailable', '模型不可用', 17),
        ('agent_event_type', 'assistant.message', 'Assistant 消息', 18),
        ('agent_event_type', 'tool.started', 'Tool 调用', 19),
        ('agent_event_type', 'tool.completed', 'Tool 返回', 20),
        ('agent_event_type', 'tool.failed', 'Tool 失败', 21),
        ('agent_event_type', 'approval.required', '等待确认', 22),
        ('agent_event_type', 'skill.loaded', '加载 Skill', 23),
        ('agent_event_type', 'task.created', '创建任务', 24),
        ('agent_event_type', 'task.updated', '任务更新', 25),
        ('agent_event_type', 'task.completed', '任务完成', 26),
        ('business_event_status', 'success', '成功', 1),
        ('business_event_status', 'failed', '失败', 2),
        ('business_event_status', 'unknown', '结果未知', 3),
        ('business_event_status', 'ready', '账号正常', 4),
        ('business_event_status', 'expired', '账号已失效', 5),
        ('business_event_status', 'verification_required', '需要安全验证', 6),
        ('delivery_task_status', 'pending', '等待执行', 1),
        ('delivery_task_status', 'sending', '发送中', 2),
        ('delivery_task_status', 'sent_unconfirmed', '待确认', 3),
        ('delivery_task_status', 'failed', '失败', 4),
        ('delivery_task_status', 'unknown', '结果未知', 5),
        ('delivery_task_status', 'confirmed', '已确认', 6),
        ('delivery_mode', 'card', '卡密发货', 1),
        ('delivery_mode', 'text', '文本发货', 2),
        ('agent_resource_task_state', 'queued', '排队中', 1),
        ('agent_resource_task_state', 'running', '执行中', 2),
        ('agent_resource_task_state', 'completed', '已完成', 3),
        ('agent_resource_task_state', 'failed', '失败', 4),
        ('agent_resource_task_state', 'cancelled', '已取消', 5),
        ('resource_state', 'available', '可用', 1),
        ('resource_state', 'deleted', '已删除', 2),
        ('telegram_resource_task_status', 'queued', '排队中', 1),
        ('telegram_resource_task_status', 'running', '执行中', 2),
        ('telegram_resource_task_status', 'completed', '已完成', 3),
        ('telegram_resource_task_status', 'failed', '失败', 4),
        ('telegram_resource_task_status', 'cancelled', '已取消', 5),
        ('telegram_resource_source_mode', 'smart', '智能', 1),
        ('telegram_resource_source_mode', 'all', '全部', 2),
        ('telegram_resource_source_mode', 'manual', '手动选择', 3),
        ('telegram_resource_source_strategy', 'ai', 'AI 智能检索', 1),
        ('telegram_resource_source_strategy', 'local', '本地规则筛选', 2),
        ('telegram_resource_source_strategy', 'pending', '等待选源', 3),
        ('telegram_resource_source_strategy', 'all', '全部来源', 4),
        ('telegram_resource_source_strategy', 'manual', '手动选择', 5),
        ('telegram_file_status', 'queued', '排队中', 1),
        ('telegram_file_status', 'downloading', '下载中', 2),
        ('telegram_file_status', 'success', '成功', 3),
        ('telegram_file_status', 'failed', '失败', 4),
        ('telegram_file_status', 'cancelled', '已取消', 5),
        ('telegram_file_status', 'duplicate', '重复', 6),
        ('telegram_file_status', 'pending', '待上传', 7),
        ('telegram_file_status', 'uploading', '上传中', 8),
        ('telegram_file_status', 'no_result', '无结果', 9),
        ('telegram_dialog_type', 'group', '普通群', 1),
        ('telegram_dialog_type', 'megagroup', '超级群', 2),
        ('telegram_dialog_type', 'channel', '频道', 3),
        ('telegram_dialog_status', 'normal', '正常', 1),
        ('telegram_dialog_status', 'forbidden', '无访问权限', 2),
        ('telegram_dialog_status', 'deleted', '已删除', 3),
        ('telegram_dialog_status', 'left', '已退出', 4),
        ('telegram_dialog_filter', 'groups', '群组', 1),
        ('telegram_dialog_filter', 'channels', '频道', 2),
        ('telegram_dialog_filter', 'normal', '正常', 3),
        ('telegram_dialog_filter', 'invalid', '异常', 4),
        ('telegram_dialog_filter', 'muted', '已静音', 5),
        ('telegram_dialog_filter', 'archived', '已归档', 6),
        ('telegram_dialog_filter', 'creator', '我创建的', 7),
        ('mcp_health_status', 'available', '可用', 1),
        ('mcp_health_status', 'unavailable', '不可用', 2),
        ('mcp_health_status', 'unknown', '未检查', 3)
)
INSERT INTO s_dictionary_details (created_at, updated_at, label, value, status, sort, sys_dictionary_id, level, path)
SELECT now(), now(), entries.name, entries.code, true, entries.sort, dictionaries.id, 0, ''
FROM entries
JOIN s_dictionaries dictionaries ON dictionaries.type = entries.type AND dictionaries.deleted_at IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM s_dictionary_details existing
    WHERE existing.sys_dictionary_id = dictionaries.id
      AND existing.value = entries.code
      AND existing.deleted_at IS NULL
);

UPDATE tel_dialogs
SET dialog_type = CASE dialog_type
    WHEN '普通群' THEN 'group'
    WHEN '群组' THEN 'group'
    WHEN '超级群' THEN 'megagroup'
    WHEN '频道' THEN 'channel'
    ELSE dialog_type
END
WHERE dialog_type IN ('普通群', '群组', '超级群', '频道');

-- +goose Down
-- 字典项可能已被管理员编辑或业务数据引用，因此回滚时保留字典数据。
