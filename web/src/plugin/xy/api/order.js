import service from '@/utils/request'

export const getXyOrderList = (params) => service({
  url: '/xy/order/list',
  method: 'get',
  params
})

export const syncXyPendingOrders = (data) => service({
  url: '/xy/order/sync-pending',
  method: 'post',
  data
})
