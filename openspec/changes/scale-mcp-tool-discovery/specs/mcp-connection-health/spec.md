# Spec Delta

## Purpose

为 MCP 服务提供独立健康状态、按需并发调用和空闲连接回收，使大量用户可以安全共享或隔离 MCP 执行会话。

## ADDED Requirements

### Requirement: MCP 健康检查独立运行
系统 SHALL 使用短期协议级探测检查 MCP 是否可连接和响应，并将健康检查与目录刷新、Agent Tool Call 和执行连接 TTL 分离。

#### Scenario: 定时健康检查成功
- **WHEN** 启用的 MCP 在检查超时内完成协议级探测
- **THEN** 系统记录可用状态、检查时间和响应耗时

#### Scenario: 定时健康检查失败
- **WHEN** MCP 连接、认证、协议协商或探测失败
- **THEN** 系统记录不可用状态和不含凭据的错误摘要

#### Scenario: 健康检查不影响执行连接
- **WHEN** 系统对 MCP 执行健康检查
- **THEN** 检查不进入执行连接缓存、不刷新执行连接空闲时间且不产生 Agent Tool 调用记录

#### Scenario: 管理员手动检查
- **WHEN** 管理员请求检查指定 MCP
- **THEN** 系统立即执行或合并该 MCP 的健康探测并返回最新状态

### Requirement: 健康状态与目录状态分别展示
系统 SHALL 向管理员分别展示 MCP 当前健康状态和最近工具目录状态，避免把缓存目录误认为实时在线能力。

#### Scenario: MCP 离线但存在缓存目录
- **WHEN** MCP 最近健康检查失败且存在工具目录快照
- **THEN** 管理界面同时显示“不可用”和缓存工具数量及最后刷新时间

### Requirement: MCP 执行连接按需建立和复用
系统 SHALL 在实际工具调用时建立 MCP 客户端，并在后续调用中复用同一作用域内的活动客户端；浏览目录和路由选择不得建立执行连接。

#### Scenario: 首次调用 MCP 工具
- **WHEN** 当前连接作用域没有可用客户端
- **THEN** 系统建立并初始化客户端后执行工具调用

#### Scenario: 并发首次调用同一 MCP
- **WHEN** 多个请求同时需要同一连接作用域且尚未建立客户端
- **THEN** 系统合并建连过程并让请求共享成功结果或同一失败结果

### Requirement: 共享 MCP 客户端支持并发 Tool Call
系统 SHALL 允许同一共享 Streamable HTTP MCP 客户端同时执行多个独立 Tool Call，不得用覆盖整个调用过程的互斥锁将热门 MCP 串行化。

#### Scenario: 多用户并发调用热门 MCP
- **WHEN** 多个用户同时调用一个配置为共享作用域的 MCP
- **THEN** 系统通过同一逻辑客户端发出并发请求，并分别应用调用超时、预算和审计

#### Scenario: MCP 配置调用并发限制
- **WHEN** 管理员为某个 MCP 配置了正数并发上限
- **THEN** 超出上限的调用在自身超时范围内等待执行名额

### Requirement: 活跃 MCP 连接不设总数上限
系统 SHALL 允许所有被真实业务使用的 MCP 保持活动连接，不得因全局连接数量上限拒绝新的 MCP；系统只通过空闲 TTL 回收无在途调用的连接。

#### Scenario: 用户同时使用不同 MCP
- **WHEN** 多个用户同时使用任意数量的不同 MCP
- **THEN** 系统允许分别建立活动客户端且不基于活动 MCP 总数拒绝调用

#### Scenario: 空闲连接到期
- **WHEN** MCP 客户端超过空闲 TTL 且没有在途调用
- **THEN** 系统从活动连接表原子移除并安全关闭该客户端

#### Scenario: TTL 与调用发生竞争
- **WHEN** 清理器准备关闭连接时新调用已取得该连接的执行权
- **THEN** 清理器不得关闭该连接，直到在途调用结束并再次满足 TTL

### Requirement: MCP 连接按认证主体隔离
系统 SHALL 根据 MCP 的连接作用域按服务或用户认证主体复用客户端，防止带用户状态的 MCP Session 在不同用户之间共享。

#### Scenario: 系统级共享 MCP
- **WHEN** MCP 配置为 shared 且使用系统级凭据
- **THEN** 所有获准用户可共享以 server code 标识的客户端

#### Scenario: 用户隔离 MCP
- **WHEN** MCP 配置为 owner 作用域
- **THEN** 不同 owner 使用不同客户端和 Session，且连接键不包含可暴露的凭据

### Requirement: MCP 调用结果正确影响健康状态
系统 SHALL 用连接和协议结果更新 MCP 健康状态，但不得把工具业务失败、用户参数错误或审批等待误判为服务离线。

#### Scenario: Tool Call 连接失败
- **WHEN** MCP Tool Call 因连接或 Session 失效失败
- **THEN** 系统使当前客户端失效并记录 MCP 不可用

#### Scenario: Tool 返回业务错误
- **WHEN** MCP 正常响应且 Tool Result 表示业务失败
- **THEN** 系统保留 MCP 可用状态并按工具错误处理结果

### Requirement: 写操作断线后不盲目重试
系统 SHALL 在 MCP 写工具连接中断或结果不明确时避免自动重试，并向 Agent 返回需要核验的安全结果。

#### Scenario: 写工具调用结果未知
- **WHEN** 写工具请求已发出但连接在确认结果前中断
- **THEN** 系统不自动重放调用，并要求通过业务状态或事件核验结果

### Requirement: 审批执行校验工具定义
系统 SHALL 保存写工具提议时的目录定义哈希，并在批准执行前确认工具仍存在、获准且定义未变化。

#### Scenario: 审批期间工具定义变化
- **WHEN** 用户批准操作时工具定义或安全策略哈希不同于提议时
- **THEN** 系统拒绝执行旧参数并要求重新生成操作提议
