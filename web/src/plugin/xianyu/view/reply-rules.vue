<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="商品">
          <el-select v-model="searchInfo.productId" clearable filterable placeholder="选择商品" class="w-72" @change="search">
            <el-option v-for="item in products" :key="item.ID" :label="item.title" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchInfo.type" clearable placeholder="全部类型" class="w-40">
            <el-option v-for="item in ruleTypes" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词或回复">
          <el-input v-model="searchInfo.keyword" clearable placeholder="筛选关键词或回复内容" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="primary" @click="openCreate">新增规则</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list flex items-center justify-between">
        <div class="min-w-0 truncate font-medium">{{ selectedProduct?.title || '全部商品回复规则' }}</div>
      </div>
      <el-table v-loading="loading" :data="tableData" row-key="ID" show-overflow-tooltip>
        <el-table-column label="商品" prop="productTitle" min-width="200" />
        <el-table-column label="类型" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.type === 'fallback' ? 'warning' : 'primary'">{{ typeLabel(scope.row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="触发关键词" prop="keyword" min-width="180">
          <template #default="scope">{{ scope.row.type === 'fallback' ? '未匹配到其他规则' : scope.row.keyword }}</template>
        </el-table-column>
        <el-table-column label="回复内容" prop="content" min-width="280" />
        <el-table-column label="优先级" width="150">
          <template #default="scope">
            <el-input-number
              v-if="scope.row.type !== 'fallback'"
              :model-value="scope.row.priority"
              :min="0"
              :max="10000"
              controls-position="right"
              @change="updatePriority(scope.row, $event)"
            />
            <span v-else class="text-gray-500">最后兜底</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'info'">{{ scope.row.enabled ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="scope">
            <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
            <el-popconfirm title="确定删除这条规则吗？" @confirm="remove(scope.row)">
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
    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑回复规则' : '新增回复规则'" width="600px">
      <el-form label-width="100px">
        <el-form-item label="商品">
          <el-select v-model="form.productId" filterable placeholder="选择商品" class="w-full">
            <el-option
              v-for="item in ruleProducts"
              :key="item.ID"
              :label="item.status === 'on_sale' ? item.title : `${item.title}（${item.status === 'sold' ? '已售出' : '已下架'}）`"
              :value="item.ID"
              :disabled="item.status !== 'on_sale'"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="规则类型" required>
          <el-radio-group v-model="form.type" @change="handleTypeChange">
            <el-radio-button v-for="item in ruleTypes" :key="item.value" :value="item.value">{{ item.label }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.type === 'fuzzy'" label="触发关键词" required>
          <el-input v-model="form.keyword" maxlength="255" show-word-limit />
        </el-form-item>
        <el-form-item v-if="form.type === 'fuzzy'" label="优先级">
          <el-input-number v-model="form.priority" :min="0" :max="10000" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="回复内容" required>
          <el-input v-model="form.content" type="textarea" :rows="6" maxlength="4000" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getXianyuProductList } from '@/plugin/xianyu/api/product'
import {
  deleteXianyuReplyRule,
  getXianyuReplyRuleList,
  saveXianyuReplyRule
} from '@/plugin/xianyu/api/reply-rule'

defineOptions({ name: 'XianyuReplyRules' })

const ruleTypes = [
  { label: '模糊匹配', value: 'fuzzy' },
  { label: '兜底回复', value: 'fallback' }
]
const loading = ref(false)
const tableData = ref([])
const products = ref([])
const onSaleProducts = ref([])
const total = ref(0)
const dialogVisible = ref(false)
const searchInfo = reactive({ page: 1, pageSize: 10, keyword: '', productId: undefined, type: '' })
const form = reactive({ ID: 0, productId: undefined, type: 'fuzzy', keyword: '', priority: 100, enabled: true, content: '', legacyExact: false })

const selectedProduct = computed(() => products.value.find(item => item.ID === searchInfo.productId))
const ruleProducts = computed(() => {
  const current = products.value.find(item => item.ID === form.productId)
  if (!form.ID || !current || current.status === 'on_sale') return onSaleProducts.value
  return [current, ...onSaleProducts.value]
})
const typeLabel = (type) => type === 'fallback' ? '兜底回复' : '模糊匹配'

const load = async () => {
  loading.value = true
  try {
    const res = await getXianyuReplyRuleList(searchInfo)
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total
    }
  } finally {
    loading.value = false
  }
}

const loadProducts = async () => {
  const [allRes, onSaleRes] = await Promise.all([
    getXianyuProductList({ page: 1, pageSize: 100 }),
    getXianyuProductList({ page: 1, pageSize: 100, status: 'on_sale' })
  ])
  if (allRes.code === 0) products.value = allRes.data.list || []
  if (onSaleRes.code === 0) onSaleProducts.value = onSaleRes.data.list || []
}

const search = () => {
  searchInfo.page = 1
  load()
}

const resetSearch = () => {
  Object.assign(searchInfo, { page: 1, pageSize: 10, keyword: '', productId: undefined, type: '' })
  load()
}

const openCreate = () => {
  const priorities = tableData.value.filter(item => item.type !== 'fallback').map(item => item.priority)
  const priority = Math.min(10000, Math.max(0, ...priorities) + 10)
  const productId = selectedProduct.value?.status === 'on_sale' ? searchInfo.productId : undefined
  Object.assign(form, { ID: 0, productId, productTitle: '', type: 'fuzzy', keyword: '', priority, enabled: true, content: '', legacyExact: false })
  dialogVisible.value = true
}

const openEdit = (row) => {
  Object.assign(form, { ...row, productTitle: row.productTitle, type: row.type === 'exact' ? 'fuzzy' : row.type, legacyExact: row.type === 'exact' })
  dialogVisible.value = true
}

const handleTypeChange = (type) => {
  if (type === 'fallback') form.keyword = ''
  form.legacyExact = false
}

const save = async () => {
  if (!form.productId) return ElMessage.warning('请选择商品')
  if (form.type === 'fallback') {
    const fallbackRes = await getXianyuReplyRuleList({ page: 1, pageSize: 1, productId: form.productId, type: 'fallback' })
    const existingFallbackID = fallbackRes.code === 0 ? fallbackRes.data.list?.[0]?.ID : 0
    if (existingFallbackID && existingFallbackID !== form.ID) return ElMessage.warning('该商品已经配置兜底回复')
  }
  if (form.type === 'fuzzy' && !form.keyword.trim()) return ElMessage.warning('请输入触发关键词')
  if (!form.content.trim()) return ElMessage.warning('请输入回复内容')
  const data = {
    ...form,
    productId: form.productId,
    type: form.legacyExact && form.type === 'fuzzy' ? 'exact' : form.type,
    keyword: form.type === 'fallback' ? '' : form.keyword.trim()
  }
  const res = await saveXianyuReplyRule(data)
  if (res.code === 0) {
    ElMessage.success('回复规则已保存')
    dialogVisible.value = false
    load()
  }
}

const updatePriority = async (row, priority) => {
  if (priority === null || priority === undefined || priority === row.priority) return
  const res = await saveXianyuReplyRule({ ...row, priority })
  if (res.code === 0) {
    ElMessage.success('优先级已更新')
    load()
  }
}

const remove = async (row) => {
  const res = await deleteXianyuReplyRule({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('删除成功')
    load()
  }
}

onMounted(async () => {
  await loadProducts()
  await load()
})
</script>
