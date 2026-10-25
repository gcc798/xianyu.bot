import service from '@/utils/request'

export const getXianyuDeliveryRuleList = (params) => service({
  url: '/xianyu/delivery-rule/list',
  method: 'get',
  params
})

export const saveXianyuDeliveryRule = (data) => service({
  url: '/xianyu/delivery-rule/save',
  method: 'post',
  data
})

export const deleteXianyuDeliveryRule = (params) => service({
  url: '/xianyu/delivery-rule/delete',
  method: 'delete',
  params
})

export const importXianyuDeliveryResources = (data) => service({
  url: '/xianyu/delivery-rule/resource-import',
  method: 'post',
  data
})

export const getXianyuDeliveryInventory = (params) => service({
  url: '/xianyu/delivery-rule/inventory',
  method: 'get',
  params
})

export const getXianyuDeliveryTaskList = (params) => service({
  url: '/xianyu/delivery-rule/task-list',
  method: 'get',
  params
})

export const retryXianyuDeliveryTask = (data) => service({
  url: '/xianyu/delivery-rule/task-retry',
  method: 'post',
  data
})
