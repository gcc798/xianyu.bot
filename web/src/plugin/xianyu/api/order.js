import service from '@/utils/request'

export const getXianyuOrderList = (params) => service({
  url: '/xianyu/order/list',
  method: 'get',
  params
})

export const syncXianyuPendingOrders = (data) => service({
  url: '/xianyu/order/sync-pending',
  method: 'post',
  data
})
