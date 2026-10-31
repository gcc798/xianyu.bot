import service from '@/utils/request'

export const getXyDeliveryRuleList = (params) => service({
  url: '/xy/delivery-rule/list',
  method: 'get',
  params
})

export const saveXyDeliveryRule = (data) => service({
  url: '/xy/delivery-rule/save',
  method: 'post',
  data
})

export const deleteXyDeliveryRule = (params) => service({
  url: '/xy/delivery-rule/delete',
  method: 'delete',
  params
})

export const importXyDeliveryResources = (data) => service({
  url: '/xy/delivery-rule/resource-import',
  method: 'post',
  data
})

export const getXyDeliveryInventory = (params) => service({
  url: '/xy/delivery-rule/inventory',
  method: 'get',
  params
})

export const getXyDeliveryTaskList = (params) => service({
  url: '/xy/delivery-rule/task-list',
  method: 'get',
  params
})

export const retryXyDeliveryTask = (data) => service({
  url: '/xy/delivery-rule/task-retry',
  method: 'post',
  data
})
