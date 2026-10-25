import service from '@/utils/request'

export const getXianyuProductList = (params) => service({
  url: '/xianyu/product/list',
  method: 'get',
  params
})

export const syncXianyuProducts = (data) => service({
  url: '/xianyu/product/sync',
  method: 'post',
  data
})

export const updateXianyuProductSetting = (data) => service({
  url: '/xianyu/product/setting',
  method: 'put',
  data
})
