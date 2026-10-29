<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="filters">
        <el-form-item label="任务"><el-input-number v-model="filters.task_id" :min="1" controls-position="right" /></el-form-item>
        <el-form-item label="文件名"><el-input v-model="filters.filename" clearable @keyup.enter="search" /></el-form-item>
        <el-form-item label="类型"><el-select v-model="filters.file_type" clearable class="w-28"><el-option v-for="item in options.file_types" :key="item" :label="item" :value="item" /></el-select></el-form-item>
        <el-form-item label="状态"><el-select v-model="filters.status" clearable class="w-32"><el-option v-for="item in statuses" :key="item" :label="item" :value="item" /></el-select></el-form-item>
        <el-form-item label="来源"><el-select v-model="filters.source_id" clearable filterable class="w-52"><el-option v-for="item in options.dialogs" :key="item.id" :label="item.name" :value="item.id" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" @click="search">查询</el-button><el-button @click="reset">重置</el-button></el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <div class="table-title"><div><h2>资源文件</h2><p>只展示已由检索任务发现的文件，预览和下载经 GVA 权限代理。</p></div><el-button @click="load">刷新</el-button></div>
      <el-table :data="files" row-key="id" show-overflow-tooltip>
        <el-table-column prop="original_filename" label="文件名" min-width="260" />
        <el-table-column prop="file_type" label="类型" width="90" />
        <el-table-column label="大小" width="120"><template #default="scope">{{ formatSize(scope.row.file_size) }}</template></el-table-column>
        <el-table-column prop="source_name" label="来源" min-width="180" />
        <el-table-column prop="task_id" label="任务" width="80" />
        <el-table-column prop="completed_at" label="完成时间" width="180"><template #default="scope">{{ formatTime(scope.row.completed_at) }}</template></el-table-column>
        <el-table-column label="下载" width="105"><template #default="scope"><el-tag :type="scope.row.download_status === 'success' ? 'success' : scope.row.download_status === 'failed' ? 'danger' : 'warning'">{{ scope.row.download_status }}</el-tag></template></el-table-column>
        <el-table-column label="存储" width="105"><template #default="scope"><el-tag :type="scope.row.upload_status === 'success' ? 'success' : scope.row.upload_status === 'failed' ? 'danger' : 'warning'">{{ scope.row.upload_status }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="150"><template #default="scope"><el-button v-if="previewable(scope.row) && scope.row.upload_status === 'success'" link type="primary" @click="openFile(scope.row, 'preview')">预览</el-button><el-button v-if="scope.row.upload_status === 'success'" link type="primary" @click="openFile(scope.row, 'download')">下载</el-button></template></el-table-column>
      </el-table>
      <div class="gva-pagination"><el-pagination v-model:current-page="filters.page" v-model:page-size="filters.page_size" :total="total" layout="total, sizes, prev, pager, next" @current-change="load" @size-change="search" /></div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { getTelegramResourceFile, getTelegramResourceFiles, getTelegramResourceOptions } from '@/plugin/xy/api/telegram-resource'

defineOptions({ name: 'XyTelegramFiles' })

const statuses = ['queued', 'downloading', 'success', 'failed', 'cancelled', 'duplicate']
const options = reactive({ dialogs: [], file_types: [] })
const filters = reactive({ page: 1, page_size: 20, task_id: undefined, filename: '', file_type: '', status: '', source_id: undefined })
const files = ref([])
const total = ref(0)

const load = async () => { const res = await getTelegramResourceFiles(filters); if (res.code === 0) { files.value = res.data.list || []; total.value = res.data.total || 0 } }
const search = () => { filters.page = 1; load() }
const reset = () => { Object.assign(filters, { page: 1, page_size: 20, task_id: undefined, filename: '', file_type: '', status: '', source_id: undefined }); load() }
const formatSize = (size = 0) => size < 1024 ? `${size} B` : size < 1048576 ? `${(size / 1024).toFixed(1)} KB` : `${(size / 1048576).toFixed(1)} MB`
const formatTime = (value) => value ? new Date(value).toLocaleString() : '-'
const previewable = (row) => row.file_type === 'pdf' || (row.mime_type || '').startsWith('image/') || (row.mime_type || '').startsWith('text/')
const openFile = async (row, action) => {
  const res = await getTelegramResourceFile(row.id, action)
  const url = URL.createObjectURL(res.data)
  const link = document.createElement('a')
  link.href = url
  if (action === 'preview') link.target = '_blank'; else link.download = row.original_filename
  link.click()
  setTimeout(() => URL.revokeObjectURL(url), 60000)
}

onMounted(async () => { const res = await getTelegramResourceOptions(); if (res.code === 0) Object.assign(options, res.data); await load() })
</script>

<style scoped>
.table-title { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; }
.table-title h2 { margin: 0; font-size: 18px; }
.table-title p { margin: 5px 0 0; color: var(--el-text-color-secondary); font-size: 13px; }
</style>
