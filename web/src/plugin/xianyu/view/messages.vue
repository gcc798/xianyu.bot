<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="账号">
          <el-select v-model="searchInfo.accountId" clearable placeholder="全部账号" style="width: 180px">
            <el-option v-for="item in accounts" :key="item.ID" :label="item.label" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="searchInfo.timeRange"
            type="datetimerange"
            value-format="x"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            style="width: 360px"
          />
        </el-form-item>
        <el-form-item label="方向">
          <el-select v-model="searchInfo.direction" clearable placeholder="全部方向" style="width: 120px">
            <el-option label="买家发来" value="inbound" />
            <el-option label="卖家发出" value="outbound" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品 ID">
          <el-input v-model="searchInfo.externalProductId" clearable placeholder="闲鱼商品 ID" style="width: 180px" />
        </el-form-item>
        <el-form-item label="买家名称">
          <el-input v-model="searchInfo.buyerName" clearable placeholder="按名称筛选" style="width: 160px" />
        </el-form-item>
        <el-form-item label="消息内容">
          <el-input v-model="searchInfo.content" clearable placeholder="按内容筛选" style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <el-table v-loading="loading" :data="tableData" row-key="ID" show-overflow-tooltip>
        <el-table-column label="时间" width="180">
          <template #default="scope">{{ messageTime(scope.row) }}</template>
        </el-table-column>
        <el-table-column label="账号" prop="accountLabel" width="140" />
        <el-table-column label="发送方" min-width="170">
          <template #default="scope">
            <el-tag v-if="isPlatformMessage(scope.row)" type="warning">平台提示</el-tag>
            <template v-else>
              <div>{{ scope.row.direction === 'outbound' ? '卖家' : '买家' }} · {{ scope.row.senderUserName || '-' }}</div>
              <div class="subtle">{{ scope.row.senderUserId || '-' }}</div>
            </template>
          </template>
        </el-table-column>
        <el-table-column label="方向" width="90">
          <template #default="scope">
            <el-tag :type="scope.row.direction === 'outbound' ? 'success' : 'info'">
              {{ scope.row.direction === 'outbound' ? '发出' : '收到' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="消息内容" prop="content" min-width="320" show-overflow-tooltip>
          <template #default="scope">{{ scope.row.content || `[类型 ${scope.row.contentType}]` }}</template>
        </el-table-column>
        <el-table-column label="商品 ID" prop="externalProductId" min-width="150">
          <template #default="scope">{{ scope.row.externalProductId || '-' }}</template>
        </el-table-column>
        <el-table-column label="订单 ID" prop="externalOrderId" min-width="150">
          <template #default="scope">{{ scope.row.externalOrderId || '-' }}</template>
        </el-table-column>
        <el-table-column label="会话 ID" prop="sessionId" min-width="220" show-overflow-tooltip />
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="scope">
            <el-button v-if="scope.row.direction !== 'outbound' && !isPlatformMessage(scope.row)" link type="primary" @click="openReply(scope.row)">回复</el-button>
            <el-button v-if="!isPlatformMessage(scope.row)" link :type="scope.row.takeoverActive ? 'warning' : 'primary'" @click="toggleTakeover(scope.row)">
              {{ scope.row.takeoverActive ? '释放接管' : '人工接管' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="gva-pagination">
        <el-pagination
          v-model:current-page="searchInfo.page"
          v-model:page-size="searchInfo.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="load"
          @size-change="search"
        />
      </div>
    </div>

    <el-dialog v-model="replyDialog" title="人工回复" width="520px">
      <el-alert title="发送成功后，该会话会进入人工接管，自动回复将暂停。" type="warning" :closable="false" class="mb-4" />
      <el-input v-model="replyForm.text" type="textarea" :rows="5" maxlength="4000" show-word-limit placeholder="请输入回复内容" />
      <template #footer>
        <el-button @click="replyDialog = false">取消</el-button>
        <el-button type="primary" :loading="sending" @click="submitReply">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { onMounted, reactive, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { getXianyuAccountList } from '@/plugin/xianyu/api/account'
  import { getXianyuMessageList, replyXianyuMessage, updateXianyuTakeover } from '@/plugin/xianyu/api/message'

  defineOptions({ name: 'XianyuMessages' })

  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const accounts = ref([])
  const replyDialog = ref(false)
  const sending = ref(false)
  const searchInfo = reactive({
    page: 1, pageSize: 10, accountId: undefined, timeRange: [], direction: '',
    externalProductId: '', buyerName: '', content: ''
  })
  const replyForm = reactive({ messageId: 0, text: '' })

  const load = async () => {
    loading.value = true
    try {
      const [startTime, endTime] = searchInfo.timeRange || []
      const res = await getXianyuMessageList({
        page: searchInfo.page,
        pageSize: searchInfo.pageSize,
        accountId: searchInfo.accountId,
        direction: searchInfo.direction,
        externalProductId: searchInfo.externalProductId.trim(),
        buyerName: searchInfo.buyerName.trim(),
        content: searchInfo.content.trim(),
        startTime: startTime ? Number(startTime) : undefined,
        endTime: endTime ? Number(endTime) : undefined
      })
      if (res.code === 0) {
        tableData.value = res.data.list || []
        total.value = res.data.total
      }
    } finally {
      loading.value = false
    }
  }
  const search = () => { searchInfo.page = 1; load() }
  const resetSearch = () => {
    Object.assign(searchInfo, { page: 1, pageSize: 10, accountId: undefined, timeRange: [], direction: '', externalProductId: '', buyerName: '', content: '' })
    load()
  }
  const isPlatformMessage = (row) => row.contentType === 25 || row.content === '快给ta一个评价吧～'
  const messageTime = (row) => {
    const value = row.messageTime || new Date(row.CreatedAt).getTime()
    return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'
  }
  const openReply = (row) => {
    Object.assign(replyForm, { messageId: row.ID, text: '' })
    replyDialog.value = true
  }
  const submitReply = async () => {
    if (!replyForm.text.trim()) return ElMessage.warning('请输入回复内容')
    sending.value = true
    try {
      const res = await replyXianyuMessage(replyForm)
      if (res.code === 0) {
        ElMessage.success('消息已发送，会话已转为人工接管')
        replyDialog.value = false
        load()
      }
    } finally {
      sending.value = false
    }
  }
  const toggleTakeover = async (row) => {
    const active = !row.takeoverActive
    try {
      await ElMessageBox.confirm(
        active ? '接管后该会话将暂停自动回复。' : '释放后该会话可以再次触发自动回复。',
        active ? '确认人工接管' : '确认释放接管',
        { type: active ? 'warning' : 'info' }
      )
    } catch {
      return
    }
    const res = await updateXianyuTakeover({ messageId: row.ID, active })
    if (res.code === 0) {
      ElMessage.success(active ? '已人工接管' : '已释放接管')
      load()
    }
  }

  onMounted(async () => {
    const accountRes = await getXianyuAccountList({ page: 1, pageSize: 100 })
    if (accountRes.code === 0) accounts.value = accountRes.data.list || []
    load()
  })
</script>

<style scoped>
  .subtle {
    margin-top: 2px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
</style>
