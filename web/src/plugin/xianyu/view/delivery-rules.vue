<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="商品">
          <el-select v-model="searchInfo.productId" clearable filterable placeholder="全部商品" class="w-64">
            <el-option v-for="item in products" :key="item.ID" :label="item.title" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="模式">
          <el-select v-model="searchInfo.mode" clearable placeholder="全部模式" class="w-36">
            <el-option label="文本发货" value="text" />
            <el-option label="卡密发货" value="card" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchInfo.keyword" clearable placeholder="SKU 或模板内容" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openCreate">新增规则</el-button>
      </div>
      <el-alert title="启用商品的自动发货开关后，新订单会按规则真实发货；请先确认模板和卡密库存。" type="warning" :closable="false" class="mb-4" />
      <el-table v-loading="loading" :data="tableData" row-key="ID" show-overflow-tooltip>
        <el-table-column label="商品" prop="productTitle" min-width="180" />
        <el-table-column label="SKU" min-width="150">
          <template #default="scope">
            <div>{{ scope.row.skuName || '商品通用' }}</div>
            <div v-if="scope.row.skuKey" class="text-xs text-gray-400">{{ scope.row.skuKey }}</div>
          </template>
        </el-table-column>
        <el-table-column label="模式" width="100">
          <template #default="scope">{{ scope.row.mode === 'card' ? '卡密发货' : '文本发货' }}</template>
        </el-table-column>
        <el-table-column label="模板" prop="template" min-width="260" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'info'">{{ scope.row.enabled ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="scope">
            <el-button link type="primary" @click="openEdit(scope.row)">编辑</el-button>
            <el-button v-if="scope.row.mode === 'card'" link type="primary" @click="openImport(scope.row)">导入卡密</el-button>
            <el-button v-if="scope.row.mode === 'card'" link type="primary" @click="showInventory(scope.row)">库存</el-button>
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

    <el-dialog v-model="dialogVisible" :title="form.ID ? '编辑发货规则' : '新增发货规则'" width="620px">
      <el-form label-width="100px">
        <el-form-item label="商品" required>
          <el-select v-model="form.productId" filterable placeholder="请选择商品" class="w-full">
            <el-option v-for="item in products" :key="item.ID" :label="item.title" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="SKU 标识">
          <el-input v-model="form.skuKey" maxlength="128" placeholder="留空表示商品通用规则" />
        </el-form-item>
        <el-form-item label="SKU 名称">
          <el-input v-model="form.skuName" maxlength="255" placeholder="仅用于展示" />
        </el-form-item>
        <el-form-item label="发货模式" required>
          <el-radio-group v-model="form.mode">
            <el-radio-button value="text">文本发货</el-radio-button>
            <el-radio-button value="card">卡密发货</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="发货模板" required>
          <el-input v-model="form.template" type="textarea" :rows="7" maxlength="4000" show-word-limit />
          <div v-if="form.mode === 'card'" class="text-xs text-gray-500">卡密模式必须使用 <code>{card}</code> 表示卡密内容。</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importVisible" title="导入卡密" width="560px">
      <el-alert title="每行一条；卡密会加密保存，重复内容自动忽略。" type="info" :closable="false" class="mb-4" />
      <el-input v-model="resourceText" type="textarea" :rows="10" placeholder="CARD-001&#10;CARD-002" />
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" @click="importResources">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { onMounted, reactive, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { getXianyuProductList } from '@/plugin/xianyu/api/product'
  import {
    deleteXianyuDeliveryRule,
    getXianyuDeliveryInventory,
    getXianyuDeliveryRuleList,
    importXianyuDeliveryResources,
    saveXianyuDeliveryRule
  } from '@/plugin/xianyu/api/delivery-rule'

  defineOptions({ name: 'XianyuDeliveryRules' })

  const route = useRoute()
  const router = useRouter()
  const routeProductId = Number(route.query.productId) || undefined
  const loading = ref(false)
  const tableData = ref([])
  const products = ref([])
  const total = ref(0)
  const dialogVisible = ref(false)
  const importVisible = ref(false)
  const importRuleId = ref(0)
  const resourceText = ref('')
  const searchInfo = reactive({ page: 1, pageSize: 10, keyword: '', productId: routeProductId, mode: '' })
  const form = reactive({ ID: 0, productId: undefined, skuKey: '', skuName: '', mode: 'text', template: '', enabled: true })

  const load = async () => {
    loading.value = true
    try {
      const res = await getXianyuDeliveryRuleList(searchInfo)
      if (res.code === 0) {
        tableData.value = res.data.list || []
        total.value = res.data.total
      }
    } finally {
      loading.value = false
    }
  }
  const loadProducts = async () => {
    const res = await getXianyuProductList({ page: 1, pageSize: 100 })
    if (res.code === 0) products.value = res.data.list || []
  }
  const search = () => {
    searchInfo.page = 1
    load()
  }
  const resetSearch = () => {
    Object.assign(searchInfo, { page: 1, pageSize: 10, keyword: '', productId: undefined, mode: '' })
    load()
  }
  const openCreate = () => {
    Object.assign(form, { ID: 0, productId: searchInfo.productId, skuKey: '', skuName: '', mode: 'text', template: '', enabled: true })
    dialogVisible.value = true
  }
  const openEdit = (row) => {
    Object.assign(form, row)
    dialogVisible.value = true
  }
  const save = async () => {
    if (!form.productId) return ElMessage.warning('请选择商品')
    if (!form.template.trim()) return ElMessage.warning('请输入发货模板')
    if (form.mode === 'card' && !form.template.includes('{card}')) return ElMessage.warning('卡密模板必须包含 {card}')
    const res = await saveXianyuDeliveryRule(form)
    if (res.code === 0) {
      ElMessage.success('发货规则已保存')
      dialogVisible.value = false
      load()
    }
  }
  const remove = async (row) => {
    const res = await deleteXianyuDeliveryRule({ ID: row.ID })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      load()
    }
  }
  const openImport = (row) => {
    importRuleId.value = row.ID
    resourceText.value = ''
    importVisible.value = true
  }
  const importResources = async () => {
    const values = resourceText.value.split(/\r?\n/).map(item => item.trim()).filter(Boolean)
    if (!values.length) return ElMessage.warning('请输入卡密')
    const res = await importXianyuDeliveryResources({ ruleId: importRuleId.value, values })
    if (res.code === 0) {
      ElMessage.success(`成功导入 ${res.data.count} 条卡密`)
      importVisible.value = false
    }
  }
  const showInventory = async (row) => {
    const res = await getXianyuDeliveryInventory({ ID: row.ID })
    if (res.code === 0) ElMessage.info(`可用 ${res.data.available}，已占用 ${res.data.reserved}，已消耗 ${res.data.consumed}`)
  }
  onMounted(async () => {
    await loadProducts()
    await load()
    if (route.query.editor === 'auto' && routeProductId) {
      const commonRule = tableData.value.find(item => !item.skuKey)
      if (commonRule) openEdit(commonRule)
      else openCreate()
      router.replace({ query: { productId: routeProductId } })
    }
  })
</script>
