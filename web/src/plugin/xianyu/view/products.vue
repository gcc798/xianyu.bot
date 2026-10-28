<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="账号">
          <el-select v-model="searchInfo.accountId" clearable placeholder="全部账号" class="w-52">
            <el-option v-for="item in accounts" :key="item.ID" :label="item.label" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题或商品 ID">
          <el-input v-model="searchInfo.titleOrId" clearable placeholder="输入标题或闲鱼商品 ID" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable placeholder="全部状态" class="w-36">
            <el-option label="在售" value="on_sale" />
            <el-option label="已下架" value="offline" />
            <el-option label="已售出" value="sold" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" :loading="syncing" @click="submitSync">同步闲鱼商品</el-button>
      </div>
      <el-table v-loading="loading" :data="tableData" row-key="ID" show-overflow-tooltip>
        <el-table-column label="图片" width="110" align="center">
          <template #default="scope">
            <el-carousel
              v-if="scope.row.imageUrls?.length > 1"
              class="product-image-carousel"
              height="64px"
              :interval="0"
              :autoplay="false"
              arrow="always"
              indicator-position="none"
            >
              <el-carousel-item v-for="image in scope.row.imageUrls" :key="image">
                <el-image :src="image" :preview-src-list="scope.row.imageUrls" :preview-teleported="true" fit="contain" class="product-image" />
              </el-carousel-item>
            </el-carousel>
            <el-image
              v-else-if="scope.row.imageUrls?.length"
              :src="scope.row.imageUrls[0]"
              :preview-src-list="scope.row.imageUrls"
              :preview-teleported="true"
              fit="contain"
              class="product-image"
            />
            <span v-else class="text-gray-400">暂无图片</span>
          </template>
        </el-table-column>
        <el-table-column label="商品" min-width="240">
          <template #default="scope">
            <div>{{ scope.row.title }}</div>
            <div class="text-xs text-gray-400">{{ scope.row.externalProductId }}</div>
          </template>
        </el-table-column>
        <el-table-column label="账号" prop="accountLabel" min-width="130" />
        <el-table-column label="价格" width="110">
          <template #default="scope">¥ {{ (scope.row.priceCent / 100).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="scope">
            <el-tag :type="productStatusType(scope.row.status)">{{ productStatusLabel(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="自动回复" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.autoReplyEnabled ? 'success' : 'info'">
              {{ scope.row.autoReplyEnabled ? '已开启' : '已关闭' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="自动发货" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.autoDeliveryEnabled ? 'success' : 'info'">
              {{ scope.row.autoDeliveryEnabled ? '已开启' : '已关闭' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="自动续售" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.autoRelistEnabled ? 'success' : 'info'">
              {{ scope.row.autoRelistEnabled ? '已开启' : '已关闭' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="scope">
            <el-button v-if="!['offline', 'sold'].includes(scope.row.status)" link type="primary" @click="openSetting(scope.row)">业务设置</el-button>
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

    <el-dialog v-model="settingDialog" title="商品业务设置" width="500px">
      <el-form label-width="110px">
        <el-form-item label="自动回复">
          <el-switch v-model="setting.autoReplyEnabled" />
          <el-button v-if="setting.autoReplyEnabled" link type="primary" class="ml-3" @click="openRuleSetting('XianyuReplyRules')">配置规则</el-button>
        </el-form-item>
        <el-form-item label="自动发货">
          <el-switch v-model="setting.autoDeliveryEnabled" />
          <el-button v-if="setting.autoDeliveryEnabled" link type="primary" class="ml-3" @click="openRuleSetting('XianyuDeliveryRules')">配置规则</el-button>
        </el-form-item>
        <el-form-item label="自动续售">
          <el-switch v-model="setting.autoRelistEnabled" />
        </el-form-item>
        <el-form-item label="回复延迟">
          <el-input-number v-model="setting.replyDelaySeconds" :min="0" :max="3600" />
          <span class="ml-2 text-gray-500">秒</span>
        </el-form-item>
        <el-form-item v-if="setting.autoRelistEnabled" label="续售延迟">
          <el-input-number v-model="setting.relistDelaySeconds" :min="0" :max="86400" />
          <span class="ml-2 text-gray-500">秒</span>
        </el-form-item>
        <el-form-item v-if="setting.autoRelistEnabled" label="续售描述">
          <el-input v-model="setting.relistDescription" type="textarea" :rows="4" maxlength="5000" show-word-limit />
        </el-form-item>
        <el-form-item v-if="setting.autoRelistEnabled" label="发布所在地" required>
          <el-input v-model="setting.relistAddress" placeholder="请输入，例如：杭州市西湖区" />
        </el-form-item>
        <el-alert v-if="setting.autoRelistEnabled" title="自动续售发布时需要所在地；虚拟商品也需要填写，但不需要邮寄地址。" type="info" :closable="false" class="mb-4" />
        <el-form-item v-if="setting.autoRelistEnabled" label="续售图片">
          <el-input v-model="setting.relistImageUrls" type="textarea" :rows="3" placeholder="每行一个图片 URL，留空则使用同步到的原商品图片" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="settingDialog = false">取消</el-button>
        <el-button type="primary" @click="saveSetting">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { getXianyuAccountList } from '@/plugin/xianyu/api/account'
  import {
    getXianyuProductList,
    syncXianyuProducts,
    updateXianyuProductSetting
  } from '@/plugin/xianyu/api/product'

  defineOptions({ name: 'XianyuProducts' })

  const router = useRouter()
  const loading = ref(false)
  const syncing = ref(false)
  const tableData = ref([])
  const accounts = ref([])
  const total = ref(0)
  const settingDialog = ref(false)
  const searchInfo = reactive({ page: 1, pageSize: 10, titleOrId: '', accountId: undefined, status: '' })
  const setting = reactive({
    productId: 0,
    autoReplyEnabled: false,
    autoDeliveryEnabled: false,
    autoRelistEnabled: false,
    replyDelaySeconds: 0,
    relistDelaySeconds: 60,
    relistDescription: '',
    relistImageUrls: '',
    relistAddress: ''
  })

  const load = async () => {
    loading.value = true
    try {
      const res = await getXianyuProductList({
        page: searchInfo.page,
        pageSize: searchInfo.pageSize,
        titleOrId: searchInfo.titleOrId.trim(),
        accountId: searchInfo.accountId,
        status: searchInfo.status
      })
      if (res.code === 0) {
        tableData.value = res.data.list || []
        total.value = res.data.total
      }
    } finally {
      loading.value = false
    }
  }
  const loadAccounts = async () => {
    const res = await getXianyuAccountList({ page: 1, pageSize: 100 })
    if (res.code === 0) accounts.value = res.data.list || []
  }
  const search = () => {
    searchInfo.page = 1
    load()
  }
  const resetSearch = () => {
    Object.assign(searchInfo, { page: 1, pageSize: 10, titleOrId: '', accountId: undefined, status: '' })
    load()
  }
  const productStatusLabel = (status) => ({ on_sale: '在售', offline: '已下架', sold: '已售出' })[status] || '未知状态'
  const productStatusType = (status) => ({ on_sale: 'success', offline: 'info', sold: 'warning' })[status] || 'info'
  const submitSync = async () => {
    const accountId = searchInfo.accountId || (accounts.value.length === 1 ? accounts.value[0].ID : undefined)
    if (!accountId) return ElMessage.warning('请先在账号筛选中选择要同步的账号')
    syncing.value = true
    try {
      const res = await syncXianyuProducts({ accountId })
      if (res.code === 0) {
        ElMessage.success(`已从闲鱼同步 ${res.data.count} 个在售商品`)
        searchInfo.accountId = accountId
        search()
      }
    } finally {
      syncing.value = false
    }
  }
  const openSetting = (row) => {
    Object.assign(setting, {
      productId: row.ID,
      autoReplyEnabled: row.autoReplyEnabled,
      autoDeliveryEnabled: row.autoDeliveryEnabled,
      autoRelistEnabled: row.autoRelistEnabled,
      replyDelaySeconds: row.replyDelaySeconds,
      relistDelaySeconds: row.relistDelaySeconds,
      relistDescription: row.relistDescription || row.title,
      relistImageUrls: row.relistImageUrls || '',
      relistAddress: row.relistAddress || ''
    })
    settingDialog.value = true
  }
  const saveSetting = async () => {
    if (setting.autoRelistEnabled && !setting.relistAddress.trim()) {
      ElMessage.warning('开启自动续售时请填写发布所在地')
      return false
    }
    const res = await updateXianyuProductSetting(setting)
    if (res.code === 0) {
      ElMessage.success('商品设置已保存')
      settingDialog.value = false
      load()
      return true
    }
    return false
  }
  const openRuleSetting = async (name) => {
    if (!await saveSetting()) return
    router.push({ name, query: { productId: setting.productId, editor: 'auto' } })
  }

  onMounted(() => {
    loadAccounts()
    load()
  })
</script>

<style scoped>
.product-image {
  width: 64px;
  height: 64px;
}

.product-image-carousel {
  width: 96px;
  margin: 0 auto;
}

.product-image-carousel :deep(.el-carousel__arrow) {
  width: 20px;
  height: 20px;
}
</style>
