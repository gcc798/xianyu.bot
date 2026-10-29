import service from '@/utils/request'

export const getXyProductList = (params) => service({
  url: '/xy/product/list',
  method: 'get',
  params
})

export const syncXyProducts = (data) => service({
  url: '/xy/product/sync',
  method: 'post',
  data
})

export const updateXyProductSetting = (data) => service({
  url: '/xy/product/setting',
  method: 'put',
  data
})
