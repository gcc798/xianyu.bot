import service from '@/utils/request'

export const getXyMessageList = (params) => service({
  url: '/xy/message/list',
  method: 'get',
  params
})

export const replyXyMessage = (data) => service({
  url: '/xy/message/reply',
  method: 'post',
  data
})

export const updateXyTakeover = (data) => service({
  url: '/xy/message/takeover',
  method: 'put',
  data
})
