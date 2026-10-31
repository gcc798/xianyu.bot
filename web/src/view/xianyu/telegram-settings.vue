<template>
  <div class="gva-table-box settings-page">
    <div class="settings-heading"><div><div class="eyebrow">SERVICE BOUNDARY</div><h1>Telegram 服务设置</h1><p>运行状态与智能检索配置集中管理；Telegram 登录密钥仍由容器环境变量维护。</p></div><el-button @click="load">刷新</el-button></div>
    <el-alert title="修改 TG_API_ID、TG_API_HASH、代理或内部访问令牌后，需要重建 telegram 与 server 服务。登录会话保存在数据库中。" type="info" show-icon :closable="false" class="mb-5" />
    <el-descriptions v-loading="loading" :column="1" border>
      <el-descriptions-item label="Telegram API"><el-tag :type="settings.telegram_api_configured ? 'success' : 'danger'">{{ settings.telegram_api_configured ? '已配置' : '未配置' }}</el-tag></el-descriptions-item>
      <el-descriptions-item label="登录状态"><el-tag :type="settings.authorized ? 'success' : 'warning'">{{ settings.authorized ? `已登录 · ${settings.account || ''}` : '未登录' }}</el-tag></el-descriptions-item>
      <el-descriptions-item label="SOCKS5 代理">{{ settings.proxy || '未配置' }}</el-descriptions-item>
      <el-descriptions-item label="容器监听">{{ settings.listen || '-' }}</el-descriptions-item>
      <el-descriptions-item label="资源并发">{{ settings.resource_concurrency || 1 }} 个任务</el-descriptions-item>
      <el-descriptions-item label="单文件上限">{{ formatSize(settings.resource_max_file_size) }}</el-descriptions-item>
      <el-descriptions-item label="成人模式">
        <el-switch v-model="settings.adult_mode" :loading="savingAdultMode" active-text="显示疑似成人频道" inactive-text="过滤疑似成人频道" @change="saveAdultMode" />
        <span v-if="!settings.adult_mode" class="setting-note">已缓存 {{ settings.adult_dialog_count || 0 }} 个 AI 判定结果；AI 不可用时继续沿用。</span>
      </el-descriptions-item>
      <el-descriptions-item label="调用策略">串行扫描；FloodWait 等待后仅重试一次普通 Telegram API 调用</el-descriptions-item>
    </el-descriptions>

    <section class="ai-panel">
      <div class="ai-panel__heading">
        <div><div class="eyebrow">INTELLIGENCE</div><h2>AI 智能筛选</h2><p>用于检索来源排序与成人频道识别，兼容 OpenAI Chat Completions API。</p></div>
        <el-tag :type="settings.ai_api_key_configured ? 'success' : 'warning'">{{ settings.ai_api_key_configured ? '密钥已配置' : '待配置' }}</el-tag>
      </div>
      <el-form label-position="top" class="ai-form" @submit.prevent="saveAI">
        <el-form-item label="服务地址" required>
          <el-input v-model="aiForm.base_url" placeholder="例如：https://api.openai.com/v1" clearable />
        </el-form-item>
        <el-form-item label="模型" required>
          <el-input v-model="aiForm.model" placeholder="例如：gpt-4.1-mini" clearable />
        </el-form-item>
        <el-form-item label="API Key" class="ai-form__key">
          <el-input v-model="aiForm.api_key" :type="apiKeyVisible ? 'text' : 'password'" autocomplete="new-password" :placeholder="settings.ai_api_key_masked || '输入 API Key'">
            <template #suffix>
              <el-button text :loading="revealingAPIKey" :aria-label="apiKeyVisible ? '隐藏 API Key' : '查看 API Key'" @click="toggleAPIKey">
                <el-icon v-if="!revealingAPIKey"><Hide v-if="apiKeyVisible" /><View v-else /></el-icon>
              </el-button>
            </template>
          </el-input>
        </el-form-item>
        <div class="ai-form__action">
          <span>保存后立即生效；探测不会保存表单，也不会刷新成人识别结果。</span>
          <div class="ai-form__buttons"><el-button :loading="probingAI" @click="probeAI">探测 AI</el-button><el-button type="primary" native-type="submit" :loading="savingAI">保存 AI 配置</el-button></div>
        </div>
      </el-form>
    </section>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Hide, View } from '@element-plus/icons-vue'
import { getTelegramSettings, probeTelegramAI, revealTelegramAIKey, updateTelegramAdultMode, updateTelegramAISettings } from '@/api/xianyu/telegram-resource'

defineOptions({ name: 'XyTelegramSettings' })

const settings = reactive({})
const loading = ref(false)
const savingAdultMode = ref(false)
const savingAI = ref(false)
const probingAI = ref(false)
const apiKeyVisible = ref(false)
const revealingAPIKey = ref(false)
const aiForm = reactive({ base_url: '', api_key: '', model: '' })
const load = async () => {
  loading.value = true
  try {
    const res = await getTelegramSettings()
    if (res.code === 0) {
      Object.assign(settings, res.data)
      Object.assign(aiForm, { base_url: res.data.ai_base_url || '', api_key: '', model: res.data.ai_model || '' })
      apiKeyVisible.value = false
    }
  } finally { loading.value = false }
}
const saveAdultMode = async (enabled) => {
  savingAdultMode.value = true
  try {
    const res = await updateTelegramAdultMode(enabled)
    if (res.code === 0) {
      Object.assign(settings, res.data)
      ElMessage.success(enabled ? '成人模式已开启' : '成人内容过滤已开启')
    }
  } catch {
    settings.adult_mode = !enabled
  } finally { savingAdultMode.value = false }
}
const saveAI = async () => {
  if (!validateAIForm()) return
  savingAI.value = true
  try {
    const res = await updateTelegramAISettings({ base_url: aiForm.base_url.trim(), api_key: aiForm.api_key.trim(), model: aiForm.model.trim() })
    if (res.code === 0) {
      Object.assign(settings, res.data)
      aiForm.api_key = ''
      apiKeyVisible.value = false
      ElMessage.success('AI 配置已保存')
    }
  } finally { savingAI.value = false }
}
const validateAIForm = () => {
  if (aiForm.base_url.trim() && aiForm.model.trim() && (settings.ai_api_key_configured || aiForm.api_key.trim())) return true
  ElMessage.warning('请完整填写服务地址、模型和 API Key')
  return false
}
const probeAI = async () => {
  if (!validateAIForm()) return
  probingAI.value = true
  try {
    const res = await probeTelegramAI({ base_url: aiForm.base_url.trim(), api_key: aiForm.api_key.trim(), model: aiForm.model.trim() })
    if (res.code === 0) ElMessage[res.data.available ? 'success' : 'error'](res.data.available ? 'AI 服务可用' : 'AI 服务不可用，请检查地址、密钥和模型')
  } finally { probingAI.value = false }
}
const toggleAPIKey = async () => {
  if (apiKeyVisible.value) {
    apiKeyVisible.value = false
    return
  }
  if (!aiForm.api_key && settings.ai_api_key_configured) {
    revealingAPIKey.value = true
    try {
      const res = await revealTelegramAIKey()
      if (res.code === 0) aiForm.api_key = res.data.api_key
    } finally { revealingAPIKey.value = false }
  }
  apiKeyVisible.value = true
}
const formatSize = (size = 0) => size ? `${Math.round(size / 1024 / 1024)} MB` : '-'
onMounted(load)
</script>

<style scoped>
.settings-page { max-width: 980px; }
.settings-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 24px; }
.settings-heading h1 { margin: 4px 0 8px; font-size: 24px; }
.settings-heading p { margin: 0; color: var(--el-text-color-secondary); }
.eyebrow { color: #0f8f86; font-size: 11px; letter-spacing: .2em; }
.setting-note { margin-left: 12px; color: var(--el-text-color-secondary); font-size: 13px; }
.ai-panel { margin-top: 24px; padding: 22px; border: 1px solid var(--el-border-color); border-top: 3px solid #0f8f86; border-radius: 8px; background: var(--el-bg-color); }
.ai-panel__heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 20px; }
.ai-panel__heading h2 { margin: 4px 0 6px; font-size: 19px; }
.ai-panel__heading p { margin: 0; color: var(--el-text-color-secondary); }
.ai-form { display: grid; grid-template-columns: 1.5fr 1fr; gap: 0 18px; }
.ai-form__key, .ai-form__action { grid-column: 1 / -1; }
.ai-form__action { display: flex; align-items: center; justify-content: space-between; gap: 16px; color: var(--el-text-color-secondary); font-size: 13px; }
.ai-form__buttons { display: flex; flex: none; }
@media (max-width: 720px) { .ai-form { grid-template-columns: 1fr; } .ai-form__key, .ai-form__action { grid-column: auto; } .ai-form__action { align-items: stretch; flex-direction: column; } }
</style>
