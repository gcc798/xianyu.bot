import service from '@/utils/request'

export const getXianyuBusinessEventList = (params) => service({
  url: '/xianyu/event/list',
  method: 'get',
  params
})
