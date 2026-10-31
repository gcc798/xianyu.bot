import service from '@/utils/request'

export const getXyBusinessEventList = (params) => service({
  url: '/xy/event/list',
  method: 'get',
  params
})
