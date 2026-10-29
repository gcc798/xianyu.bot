<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="filters">
        <el-form-item label="名称"><el-input v-model="filters.q" clearable placeholder="群组或频道名称" @keyup.enter="search" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.filter" clearable class="w-36">
            <el-option v-for="option in filterOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="成员数"><el-input-number v-model="filters.members_min" :min="0" controls-position="right" placeholder="最少" /><span class="range-separator">—</span><el-input-number v-model="filters.members_max" :min="0" controls-position="right" placeholder="最多" /></el-form-item>
        <el-form-item><el-button type="primary" @click="search">查询</el-button><el-button @click="reset">重置</el-button></el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="table-toolbar">
        <div><h2>群组与频道</h2><p>筛选、排序和分页均由 Telegram 服务端执行。</p></div>
        <div><el-button type="primary" plain @click="downloadCSV">导出 CSV</el-button><el-button type="danger" :disabled="!selected.length" @click="leaveSelected">退出选中项</el-button></div>
      </div>
      <el-table v-loading="loading" :data="rows" row-key="id" @selection-change="selected = $event" @sort-change="changeSort">
        <el-table-column type="selection" width="48" />
        <el-table-column prop="name" label="名称" min-width="220" sortable="custom">
          <template #default="scope"><a v-if="scope.row.telegram_url" :href="scope.row.telegram_url" target="_blank" rel="noopener" class="dialog-link">{{ scope.row.name }}</a><span v-else>{{ scope.row.name }}</span><div><el-tag v-for="tag in scope.row.tags" :key="tag" size="small" effect="plain" class="mr-1">{{ tag }}</el-tag></div></template>
        </el-table-column>
        <el-table-column prop="dialog_type" label="类型" width="110" sortable="custom" />
        <el-table-column prop="member_count" label="成员" width="105" sortable="custom" />
        <el-table-column prop="last_message_at" label="最后消息" width="180" sortable="custom"><template #default="scope">{{ formatTime(scope.row.last_message_at) }}</template></el-table-column>
        <el-table-column prop="unread_count" label="未读" width="90" sortable="custom" />
        <el-table-column label="属性" min-width="170"><template #default="scope"><el-tag v-if="scope.row.is_creator" size="small" type="success">我创建</el-tag><el-tag v-if="scope.row.is_muted" size="small" type="info" class="ml-1">静音</el-tag><el-tag v-if="scope.row.is_archived" size="small" type="warning" class="ml-1">归档</el-tag></template></el-table-column>
        <el-table-column prop="status" label="状态" width="100" sortable="custom"><template #default="scope"><el-tag :type="scope.row.status === 'normal' ? 'success' : 'danger'">{{ scope.row.status }}</el-tag></template></el-table-column>
      </el-table>
      <div class="gva-pagination"><el-pagination v-model:current-page="filters.page" v-model:page-size="filters.page_size" :total="total" :page-sizes="[20, 30, 50, 100]" layout="total, sizes, prev, pager, next" @current-change="load" @size-change="search" /></div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { exportTelegramDialogs, getTelegramDialogs, leaveTelegramDialogs } from '@/plugin/xy/api/telegram-resource'

defineOptions({ name: 'XyTelegramDialogs' })

const filterOptions = [
  { label: '群组', value: 'groups' }, { label: '频道', value: 'channels' }, { label: '我创建', value: 'creator' },
  { label: '正常', value: 'normal' }, { label: '失效', value: 'invalid' }, { label: '静音', value: 'muted' }, { label: '归档', value: 'archived' }
]
const filters = reactive({ page: 1, page_size: 30, q: '', filter: '', members_min: undefined, members_max: undefined, sort: 'last_message', order: 'desc' })
const rows = ref([])
const total = ref(0)
const selected = ref([])
const loading = ref(false)

const load = async () => {
  loading.value = true
  try {
    const res = await getTelegramDialogs(filters)
    if (res.code === 0) { rows.value = res.data.list || []; total.value = res.data.total || 0 }
  } finally { loading.value = false }
}
const search = () => { filters.page = 1; load() }
const reset = () => { Object.assign(filters, { page: 1, page_size: 30, q: '', filter: '', members_min: undefined, members_max: undefined, sort: 'last_message', order: 'desc' }); load() }
const changeSort = ({ prop, order }) => {
  const fields = { name: 'name', dialog_type: 'type', member_count: 'members', last_message_at: 'last_message', unread_count: 'unread', status: 'status' }
  filters.sort = fields[prop] || 'last_message'
  filters.order = order === 'ascending' ? 'asc' : 'desc'
  search()
}
const leaveSelected = async () => {
  await ElMessageBox.prompt(`将退出 ${selected.value.length} 个群组或频道，此操作无法由本系统撤销。请输入“退出”确认。`, '退出会话', { confirmButtonText: '确认退出', cancelButtonText: '取消', inputPattern: /^退出$/, inputErrorMessage: '请输入“退出”' })
  const res = await leaveTelegramDialogs(selected.value.map(row => row.id))
  if (res.code === 0) { ElMessage.success(`退出成功 ${res.data.success} 项，失败 ${res.data.failed} 项`); await load() }
}
const downloadCSV = async () => {
  const res = await exportTelegramDialogs({ q: filters.q, filter: filters.filter, members_min: filters.members_min, members_max: filters.members_max })
  const url = URL.createObjectURL(res.data)
  const link = document.createElement('a')
  link.href = url; link.download = `telegram-dialogs-${Date.now()}.csv`; link.click()
  URL.revokeObjectURL(url)
}
const formatTime = (value) => value ? new Date(value).toLocaleString() : '-'

onMounted(load)
</script>

<style scoped>
.range-separator { margin: 0 8px; color: var(--el-text-color-placeholder); }
.table-toolbar { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; }
.table-toolbar h2 { margin: 0; font-size: 18px; }
.table-toolbar p { margin: 5px 0 0; color: var(--el-text-color-secondary); font-size: 13px; }
.dialog-link { color: var(--el-color-primary); font-weight: 550; text-decoration: none; }
@media (max-width: 768px) { .table-toolbar { flex-direction: column; } }
</style>
