import service from '@/utils/request'

export const getXianyuMessageList = (params) => service({
  url: '/xianyu/message/list',
  method: 'get',
  params
})

export const replyXianyuMessage = (data) => service({
  url: '/xianyu/message/reply',
  method: 'post',
  data
})

export const updateXianyuTakeover = (data) => service({
  url: '/xianyu/message/takeover',
  method: 'put',
  data
})
