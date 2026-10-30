<template>
  <div class="settings-page">
    <header class="page-heading">
      <div><p class="eyebrow">AI CONTROL PLANE</p><h1>基础配置</h1><p>集中管理 XyAgent 使用的模型、能力服务与单轮运行边界。</p></div>
      <el-tag effect="plain">仅管理员</el-tag>
    </header>

    <el-tabs v-model="activeTab" class="settings-tabs">
      <el-tab-pane label="模型配置" name="models">
        <section class="gva-table-box panel">
          <div class="panel-heading"><div><h2>可用模型</h2><p>默认模型必须已启用并支持工具调用。</p></div><el-button type="primary" @click="openModel()">新增模型</el-button></div>
          <el-table v-loading="modelLoading" :data="models" row-key="ID">
            <el-table-column label="模型" min-width="190"><template #default="{ row }"><strong>{{ row.name }}</strong><div class="muted">{{ row.provider }} · {{ row.model }}</div></template></el-table-column>
            <el-table-column label="地址" prop="baseUrl" min-width="220" show-overflow-tooltip />
            <el-table-column label="能力" width="180"><template #default="{ row }"><el-tag v-if="row.supportsTools" size="small" type="success">Tool</el-tag><el-tag v-if="row.supportsStreaming" size="small" class="tag-gap">Stream</el-tag></template></el-table-column>
            <el-table-column label="密钥" width="100"><template #default="{ row }"><el-tag :type="row.hasApiKey ? 'success' : 'warning'" size="small">{{ row.hasApiKey ? '已配置' : '未配置' }}</el-tag></template></el-table-column>
            <el-table-column label="状态" width="130"><template #default="{ row }"><el-tag v-if="row.isDefault" type="success">默认</el-tag><el-tag v-else :type="row.enabled ? 'info' : 'warning'">{{ row.enabled ? '启用' : '停用' }}</el-tag></template></el-table-column>
            <el-table-column label="操作" width="210" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openModel(row)">编辑</el-button><el-button v-if="!row.isDefault" link type="success" @click="makeDefault(row)">设为默认</el-button><el-popconfirm title="确定删除该模型吗？" @confirm="removeModel(row)"><template #reference><el-button link type="danger">删除</el-button></template></el-popconfirm></template></el-table-column>
          </el-table>
        </section>
      </el-tab-pane>

      <el-tab-pane label="MCP 服务" name="mcp">
        <section class="gva-table-box panel">
          <div class="panel-heading"><div><h2>能力服务</h2><p>服务启用后，XyAgent 将读取其提供的全部工具。</p></div><el-button type="primary" @click="openMCP()">新增服务</el-button></div>
          <el-table v-loading="mcpLoading" :data="mcpServers" row-key="ID">
            <el-table-column label="服务" min-width="180"><template #default="{ row }"><strong>{{ row.name }}</strong><div class="muted">{{ row.code }}</div></template></el-table-column>
            <el-table-column label="地址" prop="baseUrl" min-width="240" show-overflow-tooltip />
            <el-table-column label="认证" width="100"><template #default="{ row }"><el-tag :type="row.hasAuthToken ? 'success' : 'info'" size="small">{{ row.hasAuthToken ? '已配置' : '无' }}</el-tag></template></el-table-column>
            <el-table-column label="状态" width="150"><template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag><span class="status-note">{{ row.lastStatus || 'unknown' }}</span></template></el-table-column>
            <el-table-column label="操作" width="130" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openMCP(row)">编辑</el-button><el-popconfirm title="确定删除该服务吗？" @confirm="removeMCP(row)"><template #reference><el-button link type="danger">删除</el-button></template></el-popconfirm></template></el-table-column>
          </el-table>
        </section>
      </el-tab-pane>

      <el-tab-pane label="Agent 配置" name="agent">
        <section class="gva-table-box panel config-panel" v-loading="configLoading">
          <div class="panel-heading"><div><h2>单轮运行边界</h2><p>达到任一限制时结束本轮，并保留已获得的部分结果。</p></div><el-button type="primary" :loading="configSaving" @click="saveAgentConfig">保存配置</el-button></div>
          <el-form label-position="top" class="limit-grid">
            <el-form-item v-for="field in configFields" :key="field.key" :label="field.label"><el-input-number v-model="agentConfig[field.key]" :min="1" :max="field.max" controls-position="right" /><div class="field-help">{{ field.help }}</div></el-form-item>
          </el-form>
        </section>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="modelDialog" :title="modelForm.ID ? '编辑模型' : '新增模型'" width="700px">
      <el-form label-position="top" class="dialog-grid">
        <el-form-item label="显示名称" required><el-input v-model="modelForm.name" /></el-form-item>
        <el-form-item label="提供商" required><el-input v-model="modelForm.provider" placeholder="openai" /></el-form-item>
        <el-form-item label="服务地址" required class="span-2"><el-input v-model="modelForm.baseUrl" placeholder="https://api.openai.com/v1" /></el-form-item>
        <el-form-item label="模型标识" required><el-input v-model="modelForm.model" /></el-form-item>
        <el-form-item label="API Key"><el-input v-model="modelForm.apiKey" type="password" show-password autocomplete="new-password" :placeholder="modelForm.hasApiKey ? '留空则保留现有密钥' : '输入 API Key'" /></el-form-item>
        <el-form-item label="上下文窗口"><el-input-number v-model="modelForm.contextWindow" :min="1" /></el-form-item>
        <el-form-item label="最大输出 Token"><el-input-number v-model="modelForm.maxOutputTokens" :min="1" /></el-form-item>
        <el-form-item label="请求超时（秒）"><el-input-number v-model="modelForm.requestTimeoutSeconds" :min="1" :max="600" /></el-form-item>
        <el-form-item label="开关"><div class="switches"><el-switch v-model="modelForm.enabled" active-text="启用" /><el-switch v-model="modelForm.supportsTools" active-text="Tool" /><el-switch v-model="modelForm.supportsStreaming" active-text="Stream" /><el-switch v-model="modelForm.isDefault" active-text="默认" /></div></el-form-item>
        <el-form-item label="扩展配置 JSON" class="span-2"><el-input v-model="modelForm.extraConfig" type="textarea" :rows="3" placeholder="{}" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="modelDialog = false">取消</el-button><el-button type="primary" :loading="dialogSaving" @click="saveModel">保存</el-button></template>
    </el-dialog>

    <el-dialog v-model="mcpDialog" :title="mcpForm.ID ? '编辑 MCP 服务' : '新增 MCP 服务'" width="680px">
      <el-form label-position="top" class="dialog-grid">
        <el-form-item label="服务名称" required><el-input v-model="mcpForm.name" /></el-form-item>
        <el-form-item label="服务编码" required><el-input v-model="mcpForm.code" placeholder="telegram" /></el-form-item>
        <el-form-item label="服务地址" required class="span-2"><el-input v-model="mcpForm.baseUrl" placeholder="http://telegram:8080/mcp" /></el-form-item>
        <el-form-item label="认证 Header"><el-input v-model="mcpForm.authHeader" placeholder="Authorization" /></el-form-item>
        <el-form-item label="认证 Token"><el-input v-model="mcpForm.authToken" type="password" show-password autocomplete="new-password" :placeholder="mcpForm.hasAuthToken ? '留空则保留现有 Token' : '可选'" /></el-form-item>
        <el-form-item label="请求超时（秒）"><el-input-number v-model="mcpForm.requestTimeoutSeconds" :min="1" :max="600" /></el-form-item>
        <el-form-item label="任务超时（秒）"><el-input-number v-model="mcpForm.taskTimeoutSeconds" :min="1" :max="86400" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="mcpForm.enabled" active-text="启用" inactive-text="停用" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="mcpDialog = false">取消</el-button><el-button type="primary" :loading="dialogSaving" @click="saveMCP">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createAIModel, createMCPServer, deleteAIModel, deleteMCPServer, getAgentConfig, getAIModels, getMCPServers, setDefaultAIModel, updateAgentConfig, updateAIModel, updateMCPServer } from '@/plugin/xy/api/ai-config'

defineOptions({ name: 'XyAISettings' })

const activeTab = ref('models')
const models = ref([])
const mcpServers = ref([])
const modelLoading = ref(false)
const mcpLoading = ref(false)
const configLoading = ref(false)
const configSaving = ref(false)
const dialogSaving = ref(false)
const modelDialog = ref(false)
const mcpDialog = ref(false)
const emptyModel = () => ({ ID: 0, name: '', provider: 'openai', baseUrl: '', apiKey: '', model: '', enabled: true, isDefault: false, supportsTools: true, supportsStreaming: true, contextWindow: 128000, maxOutputTokens: 2000, requestTimeoutSeconds: 60, extraConfig: '', hasApiKey: false })
const emptyMCP = () => ({ ID: 0, name: '', code: '', baseUrl: '', authHeader: 'Authorization', authToken: '', enabled: true, requestTimeoutSeconds: 15, taskTimeoutSeconds: 300, hasAuthToken: false })
const modelForm = reactive(emptyModel())
const mcpForm = reactive(emptyMCP())
const agentConfig = reactive({})
const configFields = [
  { key: 'maxWallTimeSeconds', label: '最长运行时间', help: '单轮墙钟时间（秒）', max: 600 },
  { key: 'maxIterations', label: '最大循环次数', help: 'Agent Loop 迭代上限', max: 50 },
  { key: 'maxModelCalls', label: '模型调用次数', help: '包含初始与后续模型请求', max: 100 },
  { key: 'maxToolCalls', label: '工具调用次数', help: '单轮全部工具调用总数', max: 200 },
  { key: 'maxConcurrentToolCalls', label: '工具并发数', help: 'Agent Loop 同时执行的工具调用数', max: 20 },
  { key: 'maxSearchCalls', label: '检索调用次数', help: '只读检索工具子预算', max: 200 },
  { key: 'maxWriteToolCalls', label: '写工具调用次数', help: '需确认的写操作子预算', max: 200 },
  { key: 'maxOutputTokens', label: '最大输出 Token', help: '最终回答输出限制', max: 100000 },
  { key: 'conversationHistoryLimit', label: '历史消息数量', help: '每轮加载的最近消息数', max: 1000 },
  { key: 'summaryTriggerTokens', label: '摘要触发 Token', help: '超过后压缩历史上下文', max: 1000000 }
]

const loadModels = async () => { modelLoading.value = true; try { const res = await getAIModels(); if (res.code === 0) models.value = res.data.list || [] } finally { modelLoading.value = false } }
const loadMCP = async () => { mcpLoading.value = true; try { const res = await getMCPServers(); if (res.code === 0) mcpServers.value = res.data.list || [] } finally { mcpLoading.value = false } }
const loadConfig = async () => { configLoading.value = true; try { const res = await getAgentConfig(); if (res.code === 0) Object.assign(agentConfig, res.data.config) } finally { configLoading.value = false } }
const openModel = (row) => { Object.assign(modelForm, emptyModel(), row || {}, { apiKey: '' }); modelDialog.value = true }
const openMCP = (row) => { Object.assign(mcpForm, emptyMCP(), row || {}, { authToken: '' }); mcpDialog.value = true }
const saveModel = async () => {
  if (!modelForm.name.trim() || !modelForm.provider.trim() || !modelForm.baseUrl.trim() || !modelForm.model.trim()) return ElMessage.warning('请完整填写模型必填项')
  dialogSaving.value = true
  try { const data = { ...modelForm }; const res = modelForm.ID ? await updateAIModel(modelForm.ID, data) : await createAIModel(data); if (res.code === 0) { ElMessage.success('模型配置已保存'); modelDialog.value = false; loadModels() } } finally { dialogSaving.value = false }
}
const saveMCP = async () => {
  if (!mcpForm.name.trim() || !mcpForm.code.trim() || !mcpForm.baseUrl.trim()) return ElMessage.warning('请完整填写服务必填项')
  dialogSaving.value = true
  try { const data = { ...mcpForm }; const res = mcpForm.ID ? await updateMCPServer(mcpForm.ID, data) : await createMCPServer(data); if (res.code === 0) { ElMessage.success('MCP 服务已保存'); mcpDialog.value = false; loadMCP() } } finally { dialogSaving.value = false }
}
const removeModel = async (row) => { const res = await deleteAIModel(row.ID); if (res.code === 0) { ElMessage.success('已删除'); loadModels() } }
const removeMCP = async (row) => { const res = await deleteMCPServer(row.ID); if (res.code === 0) { ElMessage.success('已删除'); loadMCP() } }
const makeDefault = async (row) => { const res = await setDefaultAIModel(row.ID); if (res.code === 0) { ElMessage.success('默认模型已更新'); loadModels() } }
const saveAgentConfig = async () => { configSaving.value = true; try { const res = await updateAgentConfig(agentConfig); if (res.code === 0) ElMessage.success('Agent 配置已保存') } finally { configSaving.value = false } }
onMounted(() => { loadModels(); loadMCP(); loadConfig() })
</script>

<style scoped>
.settings-page { max-width: 1180px; margin: 0 auto; }
.page-heading, .panel-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 20px; }
.page-heading { padding: 24px 4px 16px; }
.page-heading h1 { margin: 5px 0 7px; font-size: 28px; color: var(--el-text-color-primary); }
.page-heading p, .panel-heading p { margin: 0; color: var(--el-text-color-secondary); }
.eyebrow { color: #0f8f86 !important; font: 600 11px/1.2 ui-monospace, monospace; letter-spacing: .2em; }
.settings-tabs { --el-color-primary: #0f8f86; }
.panel { border-top: 3px solid #0f8f86; }
.panel-heading { margin-bottom: 20px; }
.panel-heading h2 { margin: 0 0 5px; font-size: 19px; color: var(--el-text-color-primary); }
.muted, .status-note, .field-help { color: var(--el-text-color-secondary); font-size: 12px; }
.muted { margin-top: 3px; }
.status-note { display: block; margin-top: 4px; }
.tag-gap { margin-left: 6px; }
.limit-grid, .dialog-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 24px; }
.limit-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.limit-grid :deep(.el-input-number) { width: 100%; }
.field-help { width: 100%; margin-top: 5px; }
.span-2 { grid-column: 1 / -1; }
.switches { display: flex; flex-wrap: wrap; gap: 14px; }
@media (max-width: 820px) { .limit-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 620px) { .page-heading, .panel-heading { flex-direction: column; } .limit-grid, .dialog-grid { grid-template-columns: 1fr; } .span-2 { grid-column: auto; } }
</style>
