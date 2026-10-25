<template>
  <div>
    <div class="gva-search-box">
      <el-alert title="仅检索当前账号已加入且扫描为公开、正常的群组/频道；请只保存已获授权、公开领域或自有内容。" type="warning" show-icon :closable="false" class="mb-4" />
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="资源名称"><el-input v-model="form.keyword" maxlength="200" placeholder="例如：自有项目培训资料" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="文件类型"><el-select v-model="form.file_types" multiple filterable class="w-full"><el-option v-for="item in options.file_types" :key="item" :label="item" :value="item" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="最大数量"><el-input-number v-model="form.max_downloads" :min="1" :max="100" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="搜索范围"><el-select v-model="form.source_dialog_ids" multiple filterable collapse-tags collapse-tags-tooltip class="w-full" placeholder="选择已扫描的公开群组或频道"><el-option v-for="item in options.dialogs" :key="item.id" :label="`${item.name}（${item.dialog_type}）`" :value="item.id" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" :disabled="!canCreate" @click="createTask">开始检索</el-button></el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box mb-4">
      <div class="flex justify-between items-center mb-3"><h3>检索任务</h3><el-button @click="loadTasks">刷新</el-button></div>
      <el-table :data="tasks" row-key="id">
        <el-table-column prop="id" label="任务" width="80" /><el-table-column prop="keyword" label="关键词" min-width="180" />
        <el-table-column label="类型" min-width="180"><template #default="scope">{{ scope.row.file_types.join(', ') }}</template></el-table-column>
        <el-table-column label="进度" width="170"><template #default="scope">{{ scope.row.files_completed }} / {{ scope.row.max_downloads }}（发现 {{ scope.row.files_found }}）</template></el-table-column>
        <el-table-column label="状态" width="110"><template #default="scope"><el-tag :type="statusType(scope.row.status)">{{ scope.row.status }}</el-tag></template></el-table-column>
        <el-table-column prop="error_message" label="错误" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="150"><template #default="scope"><el-button v-if="['queued','running'].includes(scope.row.status)" link type="danger" @click="cancelTask(scope.row.id)">取消</el-button><el-button v-if="['failed','cancelled'].includes(scope.row.status)" link type="primary" @click="retryTask(scope.row.id)">重试</el-button></template></el-table-column>
      </el-table>
    </div>

    <div class="gva-search-box">
      <el-form :inline="true" :model="filters">
        <el-form-item label="任务"><el-input-number v-model="filters.task_id" :min="1" controls-position="right" /></el-form-item>
        <el-form-item label="文件名"><el-input v-model="filters.filename" clearable /></el-form-item>
        <el-form-item label="类型"><el-select v-model="filters.file_type" clearable class="w-28"><el-option v-for="item in options.file_types" :key="item" :label="item" :value="item" /></el-select></el-form-item>
        <el-form-item label="状态"><el-select v-model="filters.status" clearable class="w-32"><el-option v-for="item in ['queued','downloading','success','failed','cancelled']" :key="item" :label="item" :value="item" /></el-select></el-form-item>
        <el-form-item label="来源"><el-select v-model="filters.source_id" clearable filterable class="w-52"><el-option v-for="item in options.dialogs" :key="item.id" :label="item.name" :value="item.id" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" @click="searchFiles">筛选</el-button><el-button @click="resetFiles">重置</el-button></el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <el-table :data="files" row-key="id" show-overflow-tooltip>
        <el-table-column prop="original_filename" label="文件名" min-width="240" /><el-table-column prop="file_type" label="类型" width="90" />
        <el-table-column label="大小" width="110"><template #default="scope">{{ formatSize(scope.row.file_size) }}</template></el-table-column>
        <el-table-column prop="source_name" label="来源" min-width="160" /><el-table-column prop="completed_at" label="下载时间" width="190" />
        <el-table-column label="RustFS" width="100"><template #default="scope"><el-tag :type="scope.row.upload_status === 'success' ? 'success' : 'warning'">{{ scope.row.upload_status }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="150"><template #default="scope"><el-button v-if="previewable(scope.row) && scope.row.upload_status === 'success'" link type="primary" @click="openFile(scope.row, 'preview')">预览</el-button><el-button v-if="scope.row.upload_status === 'success'" link type="primary" @click="openFile(scope.row, 'download')">下载</el-button></template></el-table-column>
      </el-table>
      <div class="gva-pagination"><el-pagination v-model:current-page="filters.page" v-model:page-size="filters.page_size" :total="fileTotal" layout="total, sizes, prev, pager, next" @current-change="loadFiles" @size-change="searchFiles" /></div>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { cancelTelegramResourceTask, createTelegramResourceTask, getTelegramResourceFile, getTelegramResourceFiles, getTelegramResourceOptions, getTelegramResourceTasks, retryTelegramResourceTask } from '@/plugin/xianyu/api/telegram-resource'

defineOptions({ name: 'XianyuTelegramResources' })
const options = reactive({ dialogs: [], file_types: [] }); const tasks = ref([]); const files = ref([]); const fileTotal = ref(0)
const form = reactive({ keyword: '', file_types: ['pdf'], source_dialog_ids: [], max_downloads: 10 })
const filters = reactive({ page: 1, page_size: 20, task_id: undefined, filename: '', file_type: '', status: '', source_id: undefined })
const canCreate = computed(() => form.keyword.trim().length >= 2 && form.file_types.length && form.source_dialog_ids.length)
let timer
const loadOptions = async () => { const res = await getTelegramResourceOptions(); if (res.code === 0) Object.assign(options, res.data) }
const loadTasks = async () => { const res = await getTelegramResourceTasks({ page: 1, page_size: 20 }); if (res.code === 0) tasks.value = res.data.list || [] }
const loadFiles = async () => { const res = await getTelegramResourceFiles(filters); if (res.code === 0) { files.value = res.data.list || []; fileTotal.value = res.data.total } }
const createTask = async () => { const res = await createTelegramResourceTask({ ...form, keyword: form.keyword.trim() }); if (res.code === 0) { ElMessage.success('任务已进入队列'); loadTasks() } }
const cancelTask = async (id) => { await cancelTelegramResourceTask(id); loadTasks() }
const retryTask = async (id) => { await retryTelegramResourceTask(id); loadTasks() }
const searchFiles = () => { filters.page = 1; loadFiles() }
const resetFiles = () => { Object.assign(filters, { page: 1, page_size: 20, task_id: undefined, filename: '', file_type: '', status: '', source_id: undefined }); loadFiles() }
const formatSize = (size) => size < 1024 ? `${size} B` : size < 1048576 ? `${(size / 1024).toFixed(1)} KB` : `${(size / 1048576).toFixed(1)} MB`
const previewable = (row) => row.file_type === 'pdf' || (row.mime_type || '').startsWith('image/') || (row.mime_type || '').startsWith('text/')
const statusType = (status) => status === 'completed' ? 'success' : status === 'failed' ? 'danger' : status === 'cancelled' ? 'info' : 'warning'
const openFile = async (row, action) => { const res = await getTelegramResourceFile(row.id, action); const url = URL.createObjectURL(res.data); const link = document.createElement('a'); link.href = url; if (action === 'preview') link.target = '_blank'; else link.download = row.original_filename; link.click(); setTimeout(() => URL.revokeObjectURL(url), 60000) }
onMounted(async () => { await loadOptions(); await Promise.all([loadTasks(), loadFiles()]); timer = setInterval(() => { loadTasks(); loadFiles() }, 5000) })
onBeforeUnmount(() => clearInterval(timer))
</script>
