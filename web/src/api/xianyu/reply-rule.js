import service from '@/utils/request'

export const getXyReplyRuleList = (params) => service({
  url: '/xy/reply-rule/list',
  method: 'get',
  params
})

export const saveXyReplyRule = (data) => service({
  url: '/xy/reply-rule/save',
  method: 'post',
  data
})

export const deleteXyReplyRule = (params) => service({
  url: '/xy/reply-rule/delete',
  method: 'delete',
  params
})
