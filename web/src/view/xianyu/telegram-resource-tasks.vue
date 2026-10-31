<template>
  <div>
    <div class="gva-search-box">
      <el-alert title="仅检索已加入且扫描为公开、正常的群组或频道；请只保存已获授权、公开领域或自有内容。" type="warning" show-icon :closable="false" class="mb-4" />
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16">
          <el-col :xs="24" :md="10"><el-form-item label="资源名称"><el-input v-model="form.keyword" maxlength="200" placeholder="描述资源主题，不必给出精确文件名" /></el-form-item></el-col>
          <el-col :xs="24" :md="8"><el-form-item label="文件类型"><el-select v-model="form.file_types" multiple filterable clearable class="w-full" placeholder="不限文件类型"><el-option v-for="item in options.file_types" :key="item" :label="item" :value="item" /></el-select></el-form-item></el-col>
          <el-col :xs="24" :md="6"><el-form-item label="最大数量"><el-input-number v-model="form.max_downloads" :min="1" :max="100" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="搜索方式">
          <el-radio-group v-model="form.source_mode">
            <el-radio-button value="smart">智能筛选</el-radio-button>
            <el-radio-button value="all">全部来源</el-radio-button>
            <el-radio-button value="manual">手动选择</el-radio-button>
          </el-radio-group>
          <span class="mode-tip">{{ sourceModeTip }}</span>
        </el-form-item>
        <el-form-item v-if="form.source_mode === 'manual'" label="搜索范围"><el-select v-model="form.source_dialog_ids" multiple filterable collapse-tags collapse-tags-tooltip class="w-full" placeholder="选择已扫描的公开群组或频道"><el-option v-for="item in options.dialogs" :key="item.id" :label="`${item.name}（${item.dialog_type}）`" :value="item.id" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" :disabled="!canCreate" @click="createTask">创建检索任务</el-button></el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="table-title"><div><h2>检索任务</h2><p>任务由 Telegram 服务串行扫描并把命中文件上传到 RustFS。</p></div><el-button @click="loadTasks">刷新</el-button></div>
      <el-table :data="tasks" row-key="id">
        <el-table-column prop="id" label="任务" width="80" />
        <el-table-column prop="keyword" label="检索主题" min-width="200" />
        <el-table-column label="类型" min-width="140"><template #default="scope">{{ scope.row.file_types?.length ? scope.row.file_types.join(', ') : '不限' }}</template></el-table-column>
        <el-table-column label="范围" min-width="150"><template #default="scope">{{ sourceModeLabel(scope.row) }} · {{ scope.row.source_dialog_ids?.length || 0 }} 个</template></el-table-column>
        <el-table-column label="进度" width="190"><template #default="scope">{{ scope.row.files_completed }} / {{ scope.row.max_downloads }}（发现 {{ scope.row.files_found }}）</template></el-table-column>
        <el-table-column label="状态" width="110"><template #default="scope"><el-tag :type="statusType(scope.row.status)">{{ scope.row.status }}</el-tag></template></el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180"><template #default="scope">{{ formatTime(scope.row.created_at) }}</template></el-table-column>
        <el-table-column prop="error_message" label="错误" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="150"><template #default="scope"><el-button v-if="['queued','running'].includes(scope.row.status)" link type="danger" @click="cancelTask(scope.row.id)">取消</el-button><el-button v-if="['failed','cancelled'].includes(scope.row.status)" link type="primary" @click="retryTask(scope.row.id)">重试</el-button></template></el-table-column>
      </el-table>
      <div class="gva-pagination"><el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.page_size" :total="total" layout="total, sizes, prev, pager, next" @current-change="loadTasks" @size-change="searchTasks" /></div>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { cancelTelegramResourceTask, createTelegramResourceTask, getTelegramResourceOptions, getTelegramResourceTasks, retryTelegramResourceTask } from '@/api/xianyu/telegram-resource'

defineOptions({ name: 'XyTelegramTasks' })

const options = reactive({ dialogs: [], file_types: [] })
const form = reactive({ keyword: '', file_types: [], source_mode: 'smart', source_dialog_ids: [], max_downloads: 10 })
const pagination = reactive({ page: 1, page_size: 20 })
const tasks = ref([])
const total = ref(0)
let timer

const canCreate = computed(() => form.keyword.trim().length >= 2 && (form.source_mode !== 'manual' || form.source_dialog_ids.length > 0))
const sourceModeTip = computed(() => form.source_mode === 'smart'
  ? `优先检索相关来源，结果不足时自动扩展至全部 ${options.dialogs.length} 个合规来源`
  : form.source_mode === 'all' ? `检索全部 ${options.dialogs.length} 个合规来源` : `仅检索已选择的 ${form.source_dialog_ids.length} 个来源`)
const loadOptions = async () => { const res = await getTelegramResourceOptions(); if (res.code === 0) Object.assign(options, res.data) }
const loadTasks = async () => { const res = await getTelegramResourceTasks(pagination); if (res.code === 0) { tasks.value = res.data.list || []; total.value = res.data.total || 0 } }
const searchTasks = () => { pagination.page = 1; loadTasks() }
const createTask = async () => {
  const res = await createTelegramResourceTask({ ...form, keyword: form.keyword.trim() })
  if (res.code === 0) { ElMessage.success('任务已进入队列'); form.keyword = ''; searchTasks() }
}
const cancelTask = async (id) => { const res = await cancelTelegramResourceTask(id); if (res.code === 0) ElMessage.success('任务已取消'); loadTasks() }
const retryTask = async (id) => { const res = await retryTelegramResourceTask(id); if (res.code === 0) ElMessage.success('任务已重新排队'); loadTasks() }
const statusType = (status) => status === 'completed' ? 'success' : status === 'failed' ? 'danger' : status === 'cancelled' ? 'info' : 'warning'
const sourceModeLabel = (task) => task.source_mode === 'smart' ? `智能${task.source_strategy === 'ai' ? '·AI' : '·本地'}` : task.source_mode === 'all' ? '全部' : '手动'
const formatTime = (value) => value ? new Date(value).toLocaleString() : '-'

onMounted(async () => { await Promise.all([loadOptions(), loadTasks()]); timer = setInterval(loadTasks, 5000) })
onBeforeUnmount(() => clearInterval(timer))
</script>

<style scoped>
.table-title { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; }
.table-title h2 { margin: 0; font-size: 18px; }
.table-title p { margin: 5px 0 0; color: var(--el-text-color-secondary); font-size: 13px; }
.mode-tip { margin-left: 12px; color: var(--el-text-color-secondary); font-size: 13px; }
@media (max-width: 768px) { .mode-tip { display: block; width: 100%; margin: 8px 0 0; } }
</style>
