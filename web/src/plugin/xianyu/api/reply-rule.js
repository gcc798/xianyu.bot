import service from '@/utils/request'

export const getXianyuReplyRuleList = (params) => service({
  url: '/xianyu/reply-rule/list',
  method: 'get',
  params
})

export const saveXianyuReplyRule = (data) => service({
  url: '/xianyu/reply-rule/save',
  method: 'post',
  data
})

export const deleteXianyuReplyRule = (params) => service({
  url: '/xianyu/reply-rule/delete',
  method: 'delete',
  params
})
