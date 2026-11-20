# Spec Delta

## Purpose

维护可恢复、可刷新且可按任务筛选的 MCP 工具目录，使 Agent 在大量 MCP 和工具存在时只加载获准且相关的能力。

## ADDED Requirements

### Requirement: MCP 工具目录可持久恢复
系统 SHALL 保存最近一次成功发现的 MCP 工具定义，并在服务重启或 MCP 暂时不可用时从该快照恢复目录。

#### Scenario: 服务从目录快照启动
- **WHEN** 服务启动且数据库中存在有效 MCP 工具快照
- **THEN** 系统在后台刷新完成前即可从快照提供工具目录

#### Scenario: 刷新失败保留旧目录
- **WHEN** 某个 MCP 的目录刷新失败且存在上次成功快照
- **THEN** 系统保留旧目录并记录刷新失败，不删除已缓存工具

### Requirement: MCP 工具目录独立刷新
系统 SHALL 在配置变更、目录过期、手动请求或工具列表变化时刷新相应 MCP 的完整分页目录，而不要求普通 Agent Run 重新发现所有 MCP。

#### Scenario: 后台刷新多个 MCP
- **WHEN** 多个启用 MCP 的目录需要刷新
- **THEN** 系统使用受控并发刷新并独立保存每个 MCP 的成功或失败结果

#### Scenario: MCP 返回分页目录
- **WHEN** `tools/list` 响应包含下一页游标
- **THEN** 系统读取全部页面后再原子替换该 MCP 的目录

#### Scenario: 管理员手动刷新
- **WHEN** 管理员请求刷新指定 MCP 的工具目录
- **THEN** 系统只刷新该 MCP 并返回最新目录状态

### Requirement: MCP 路由元数据来源可区分
系统 SHALL 分别保存管理员配置的服务描述和标签、MCP 披露的服务指令以及从工具目录生成的检索文本，并以管理员配置作为主要路由依据。

#### Scenario: MCP 返回服务指令
- **WHEN** MCP 初始化返回 instructions
- **THEN** 系统将其标记为外部不可信元数据，且不允许其覆盖系统权限、审批或安全规则

#### Scenario: 管理员未填写描述
- **WHEN** MCP 没有管理员配置的能力描述
- **THEN** 系统仍可使用服务名称、标签和工具名称描述生成检索文本

### Requirement: 路由前过滤不可用能力
系统 SHALL 在任何模型选择之前过滤停用、禁止或当前主体无权使用的 MCP 和工具。

#### Scenario: 工具位于禁止列表
- **WHEN** 已发现工具位于 MCP 的 disabled tools 配置中
- **THEN** 工具不进入路由候选、模型上下文或可调用目录

#### Scenario: 配置允许列表
- **WHEN** MCP 配置了 enabled tools
- **THEN** 只有允许列表中的已发现工具能够进入候选目录

### Requirement: Agent 使用两级工具路由
系统 SHALL 先按任务选择相关 MCP，再从候选 MCP 中选择具体工具；单步工具数量是可扩大的工作集目标，不是整个任务的累计工具上限。

#### Scenario: 普通任务选择小型工作集
- **WHEN** 用户任务只涉及少量明确的 MCP 能力
- **THEN** 系统仅向主模型提供相关 MCP 中的候选工具定义

#### Scenario: 路由置信度不足
- **WHEN** 路由结果不能可靠覆盖用户任务
- **THEN** 系统扩大 MCP 或工具候选集，而不是直接声称没有可用能力

#### Scenario: 用户显式指定工具
- **WHEN** 用户指令指定一个获准且存在的工具
- **THEN** 系统强制将该工具加入当前工作集

#### Scenario: 长任务跨越多个阶段
- **WHEN** 一个 Agent Run 在后续阶段需要新的工具领域
- **THEN** 系统能够扩展或重建后续模型步骤的工具工作集，且累计使用工具数不受单步目标限制

### Requirement: 工具路由可观测且不泄密
系统 SHALL 记录选中的 MCP、工具编码、路由策略和扩展原因，但不得记录 MCP 凭据或未经压缩的外部指令。

#### Scenario: 完成工具路由
- **WHEN** 系统为 Agent Run 选出工具工作集
- **THEN** 轨迹包含安全的路由摘要，可用于评估召回率和工具数量

### Requirement: MCP 工具目录输入受校验
系统 SHALL 校验外部工具名称、描述、Schema、数量和大小；单个无效工具不得使其他有效工具不可用。

#### Scenario: 一个工具 Schema 无效
- **WHEN** MCP 返回一个无法解析或超出限制的工具 Schema
- **THEN** 系统跳过该工具、记录安全错误并保留同一 MCP 的其他有效工具
