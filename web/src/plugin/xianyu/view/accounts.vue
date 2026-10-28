<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="关键词">
          <el-input v-model="searchInfo.keyword" clearable placeholder="名称、备注或闲鱼用户 ID" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openCreate">新增账号</el-button>
      </div>
      <el-table v-loading="loading" :data="tableData" row-key="ID" show-overflow-tooltip>
        <el-table-column label="名称" prop="label" min-width="150" />
        <el-table-column label="闲鱼用户 ID" prop="externalUserId" min-width="150">
          <template #default="scope">{{ scope.row.externalUserId || '-' }}</template>
        </el-table-column>
        <el-table-column label="凭证" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.hasCredential ? 'success' : 'info'">
              {{ scope.row.hasCredential ? '已配置' : '未配置' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="登录态" width="120">
          <template #default="scope">
            <el-tag :type="statusType(scope.row.status)">{{ statusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="消息连接" min-width="180">
          <template #default="scope">
            <div class="connection-status">
              <el-tag :type="connectionType(connectionFor(scope.row).state)" effect="plain">
                {{ connectionText(connectionFor(scope.row)) }}
              </el-tag>
              <span v-if="connectionFor(scope.row).lastError" class="connection-error" :title="connectionFor(scope.row).lastError">
                {{ connectionFor(scope.row).lastError }}
              </span>
              <el-button
                v-if="scope.row.status === 'verification_required'"
                link
                type="primary"
                @click="openMessageCredential(scope.row)"
              >处理消息验证</el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="90">
          <template #default="scope">
            <el-switch v-model="scope.row.enabled" @change="saveEnabled(scope.row)" />
          </template>
        </el-table-column>
        <el-table-column label="备注" prop="note" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" fixed="right" width="520">
          <template #default="scope">
            <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
            <el-button link type="primary" @click="openCredential(scope.row)">设置凭证</el-button>
            <el-button link type="primary" @click="openQRLogin(scope.row)">扫码登录</el-button>
            <el-button
              link
              type="primary"
              :disabled="!scope.row.hasCredential"
              :loading="checkingId === scope.row.ID"
              @click="checkLogin(scope.row)"
            >检查登录</el-button>
            <el-button
              v-if="!connectionFor(scope.row).running"
              link
              type="success"
              :disabled="scope.row.status !== 'ready' || !scope.row.enabled"
              :loading="connectionLoadingId === scope.row.ID"
              @click="startConnection(scope.row)"
            >启动连接</el-button>
            <el-button
              v-else
              link
              type="warning"
              :loading="connectionLoadingId === scope.row.ID"
              @click="stopConnection(scope.row)"
            >停止连接</el-button>
            <el-popconfirm title="确定删除这个账号吗？" @confirm="remove(scope.row)">
              <template #reference><el-button link type="danger">删除</el-button></template>
            </el-popconfirm>
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

    <el-dialog v-model="accountDialog" :title="form.ID ? '编辑账号' : '新增账号'" width="520px">
      <el-form label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.label" maxlength="80" show-word-limit />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.note" type="textarea" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="accountDialog = false">取消</el-button>
        <el-button type="primary" @click="saveAccount">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="credentialDialog" :title="credentialMode === 'message' ? '处理消息安全验证' : '设置账号凭证'" width="660px" @closed="clearCredential">
      <el-alert
        v-if="credentialMode === 'message'"
        type="warning"
        :closable="false"
        class="mb-4"
      >
        <template #title>扫码登录不能解除消息接口的独立风控，请按下面步骤更新 Cookie。</template>
        <ol class="credential-steps">
          <li><el-link href="https://www.goofish.com/" target="_blank" type="primary">打开闲鱼网页版</el-link>，登录后点击“消息”，完成滑块或安全验证。</li>
          <li>按 F12（Mac：⌥⌘I）打开开发者工具，在 Network 中搜索 <code>login.token</code>。</li>
          <li>打开最新成功的请求，在 Request Headers 中复制完整的 <code>Cookie</code> 值并粘贴到下方。</li>
        </ol>
      </el-alert>
      <el-alert v-else title="凭证只会加密保存，保存后不再回显。" type="warning" :closable="false" class="mb-4" />
      <el-form label-width="80px">
        <el-form-item label="Cookie" required>
          <el-input v-model="credential.cookie" type="textarea" :rows="5" autocomplete="off" />
        </el-form-item>
        <el-form-item v-if="credentialMode !== 'message'" label="Token">
          <el-input v-model="credential.token" type="password" show-password autocomplete="new-password" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="credentialDialog = false">取消</el-button>
        <el-button type="primary" @click="saveCredential">{{ credentialMode === 'message' ? '保存并重新连接' : '加密保存' }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="qrDialog" title="闲鱼扫码登录" width="460px" :close-on-click-modal="false" @closed="closeQRLogin">
      <div class="qr-login-panel">
        <div v-if="qrLoading" class="qr-login-placeholder">正在向闲鱼获取二维码…</div>
        <img v-else-if="qrLogin.qrCodeDataUrl" :src="qrLogin.qrCodeDataUrl" alt="闲鱼登录二维码" class="qr-login-code">
        <el-result
          v-else
          :icon="qrLogin.status === 'confirmed' ? 'success' : qrLogin.status === 'verification_required' ? 'warning' : 'info'"
          :title="qrLogin.message || '二维码会话已结束'"
        />
        <div v-if="qrLogin.qrCodeDataUrl" class="qr-login-status">
          <span class="qr-login-pulse" :class="{ scanned: qrLogin.status === 'scanned' }" />
          {{ qrLogin.message || '请使用闲鱼 App 扫码' }}
        </div>
        <el-alert
          v-if="qrLogin.status === 'verification_required'"
          title="请先在闲鱼 App 按提示完成安全验证，再重新扫码登录。"
          type="warning"
          :closable="false"
          class="mt-4"
        />
      </div>
      <template #footer>
        <el-button @click="qrDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { onMounted, onUnmounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    createXianyuAccount,
    cancelXianyuQRLogin,
    checkXianyuLogin,
    deleteXianyuAccount,
    getXianyuAccountList,
    getXianyuQRLoginStatus,
    saveXianyuCredential,
    getXianyuWebSocketStatus,
    startXianyuQRLogin,
    startXianyuWebSocket,
    stopXianyuWebSocket,
    updateXianyuAccount
  } from '@/plugin/xianyu/api/account'

  defineOptions({ name: 'XianyuAccounts' })

  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const accountDialog = ref(false)
  const credentialDialog = ref(false)
  const credentialMode = ref('normal')
  const checkingId = ref(0)
  const connectionLoadingId = ref(0)
  const connections = reactive({})
  const qrDialog = ref(false)
  const qrLoading = ref(false)
  const qrLogin = reactive({ sessionId: '', accountId: 0, status: '', message: '', qrCodeDataUrl: '', verificationUrl: '' })
  let qrTimer
  let connectionTimer
  let qrRequest = 0
  const searchInfo = reactive({ page: 1, pageSize: 10, keyword: '' })
  const form = reactive({ ID: 0, label: '', note: '', enabled: true })
  const credential = reactive({ ID: 0, cookie: '', token: '' })

  const load = async () => {
    loading.value = true
    try {
      const res = await getXianyuAccountList(searchInfo)
      if (res.code === 0) {
        tableData.value = res.data.list || []
        total.value = res.data.total
        await refreshConnections()
      }
    } finally {
      loading.value = false
    }
  }
  const search = () => { searchInfo.page = 1; load() }
  const resetSearch = () => { searchInfo.keyword = ''; search() }
  const openCreate = () => {
    Object.assign(form, { ID: 0, label: '', note: '', enabled: true })
    accountDialog.value = true
  }
  const openEdit = (row) => {
    Object.assign(form, { ID: row.ID, label: row.label, note: row.note, enabled: row.enabled })
    accountDialog.value = true
  }
  const saveAccount = async () => {
    if (!form.label.trim()) return ElMessage.warning('请输入账号名称')
    const res = form.ID ? await updateXianyuAccount(form) : await createXianyuAccount(form)
    if (res.code === 0) {
      ElMessage.success('保存成功')
      accountDialog.value = false
      load()
    }
  }
  const saveEnabled = async (row) => {
    const res = await updateXianyuAccount({ ID: row.ID, label: row.label, note: row.note, enabled: row.enabled })
    if (res.code !== 0) {
      row.enabled = !row.enabled
      return
    }
    ElMessage.success(row.enabled ? '账号已启用' : '账号已停用')
  }
  const remove = async (row) => {
    const res = await deleteXianyuAccount({ ID: row.ID })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      load()
    }
  }
  const openCredential = (row) => {
    credentialMode.value = 'normal'
    Object.assign(credential, { ID: row.ID, cookie: '', token: '' })
    credentialDialog.value = true
  }
  const openMessageCredential = (row) => {
    credentialMode.value = 'message'
    Object.assign(credential, { ID: row.ID, cookie: '', token: '' })
    credentialDialog.value = true
  }
  const clearCredential = () => {
    credentialMode.value = 'normal'
    Object.assign(credential, { ID: 0, cookie: '', token: '' })
  }
  const statusText = (status) => ({
    ready: '正常',
    expired: '已失效',
    verification_required: '需验证',
    unconfigured: '未配置'
  }[status] || status || '未知')
  const statusType = (status) => ({
    ready: 'success',
    expired: 'danger',
    verification_required: 'warning',
    unconfigured: 'info'
  }[status] || 'info')
  const connectionFor = (row) => connections[row.ID] || { running: false, state: '' }
  const connectionText = (connection) => ({
    connecting: '连接中',
    connected: '已连接',
    reconnecting: '重连中',
    failed: '需处理'
  }[connection.state] || '未连接')
  const connectionType = (state) => ({
    connected: 'success',
    connecting: 'primary',
    reconnecting: 'warning',
    failed: 'danger'
  }[state] || 'info')
  const refreshConnection = async (row) => {
    const res = await getXianyuWebSocketStatus({ ID: row.ID })
    if (res.code === 0) {
      connections[row.ID] = res.data.connection
      if (res.data.connection.accountStatus) row.status = res.data.connection.accountStatus
    }
  }
  const refreshConnections = async () => {
    await Promise.all(tableData.value.map(refreshConnection))
  }
  const startConnection = async (row) => {
    connectionLoadingId.value = row.ID
    try {
      const res = await startXianyuWebSocket({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('消息连接正在启动')
        await refreshConnection(row)
      }
    } finally {
      connectionLoadingId.value = 0
    }
  }
  const stopConnection = async (row) => {
    connectionLoadingId.value = row.ID
    try {
      const res = await stopXianyuWebSocket({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('消息连接已停止')
        await refreshConnection(row)
      }
    } finally {
      connectionLoadingId.value = 0
    }
  }
  const checkLogin = async (row) => {
    checkingId.value = row.ID
    try {
      const res = await checkXianyuLogin({ ID: row.ID })
      if (res.code === 0) {
        row.status = res.data.login.status
        const message = row.status === 'ready' ? '登录状态正常' : row.status === 'verification_required' ? '需要人工验证' : '登录状态已失效'
        ElMessage[row.status === 'ready' ? 'success' : 'warning'](message)
      }
    } finally {
      checkingId.value = 0
    }
  }
  const stopQRPolling = () => {
    if (qrTimer) window.clearInterval(qrTimer)
    qrTimer = undefined
  }
  const pollQRLogin = async () => {
    if (!qrLogin.sessionId) return
    const res = await getXianyuQRLoginStatus({ sessionId: qrLogin.sessionId })
    if (res.code !== 0) return
    Object.assign(qrLogin, res.data.login, { qrCodeDataUrl: qrLogin.qrCodeDataUrl })
    if (['confirmed', 'expired', 'cancelled', 'verification_required', 'error'].includes(qrLogin.status)) {
      stopQRPolling()
      qrLogin.qrCodeDataUrl = ''
      if (qrLogin.status === 'confirmed') {
        ElMessage.success('扫码登录成功，凭证已加密保存')
        load()
      }
    }
  }
  const openQRLogin = async (row) => {
    stopQRPolling()
    const requestID = ++qrRequest
    Object.assign(qrLogin, { sessionId: '', accountId: row.ID, status: '', message: '', qrCodeDataUrl: '', verificationUrl: '' })
    qrDialog.value = true
    qrLoading.value = true
    try {
      const res = await startXianyuQRLogin({ ID: row.ID })
      if (res.code === 0) {
        if (requestID !== qrRequest || !qrDialog.value) {
          await cancelXianyuQRLogin({ sessionId: res.data.login.sessionId })
          return
        }
        Object.assign(qrLogin, res.data.login)
        qrTimer = window.setInterval(pollQRLogin, 1500)
      }
    } finally {
      qrLoading.value = false
    }
  }
  const closeQRLogin = async () => {
    qrRequest++
    stopQRPolling()
    if (qrLogin.sessionId && ['pending', 'scanned'].includes(qrLogin.status)) {
      await cancelXianyuQRLogin({ sessionId: qrLogin.sessionId })
    }
    Object.assign(qrLogin, { sessionId: '', accountId: 0, status: '', message: '', qrCodeDataUrl: '', verificationUrl: '' })
  }
  const saveCredential = async () => {
    if (!credential.cookie.trim()) return ElMessage.warning('请输入 Cookie')
    const reconnect = credentialMode.value === 'message'
    const row = tableData.value.find(item => item.ID === credential.ID)
    const res = await saveXianyuCredential(credential)
    if (res.code === 0) {
      ElMessage.success('凭证已加密保存')
      credentialDialog.value = false
      if (reconnect && row) {
        row.status = 'ready'
        await startConnection(row)
      }
      await load()
    }
  }

  onMounted(() => {
    load()
    connectionTimer = window.setInterval(refreshConnections, 5000)
  })
  onUnmounted(() => {
    stopQRPolling()
    if (connectionTimer) window.clearInterval(connectionTimer)
  })
</script>

<style scoped>
  .credential-steps {
    margin: 8px 0 0;
    padding-left: 20px;
    line-height: 1.8;
  }

  .qr-login-panel {
    min-height: 350px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .qr-login-code {
    width: 320px;
    height: 320px;
    border-radius: 16px;
    box-shadow: 0 18px 50px rgb(15 23 42 / 12%);
  }

  .qr-login-placeholder {
    width: 320px;
    height: 320px;
    display: grid;
    place-items: center;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-light);
    border-radius: 16px;
  }

  .qr-login-status {
    margin-top: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--el-text-color-regular);
  }

  .qr-login-pulse {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--el-color-primary);
    box-shadow: 0 0 0 5px var(--el-color-primary-light-8);
  }

  .qr-login-pulse.scanned {
    background: var(--el-color-warning);
    box-shadow: 0 0 0 5px var(--el-color-warning-light-8);
  }

  .connection-status {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .connection-error {
    max-width: 160px;
    overflow: hidden;
    color: var(--el-color-danger);
    font-size: 12px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
