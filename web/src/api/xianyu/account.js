import service from '@/utils/request'

export const getXyAccountList = (params) => service({
  url: '/xy/account/list',
  method: 'get',
  params
})

export const createXyAccount = (data) => service({
  url: '/xy/account/create',
  method: 'post',
  data
})

export const updateXyAccount = (data) => service({
  url: '/xy/account/update',
  method: 'put',
  data
})

export const deleteXyAccount = (params) => service({
  url: '/xy/account/delete',
  method: 'delete',
  params
})

export const saveXyCredential = (data) => service({
  url: '/xy/account/credential',
  method: 'put',
  data
})

export const checkXyLogin = (data) => service({
  url: '/xy/account/check-login',
  method: 'post',
  data
})

export const startXyQRLogin = (data) => service({
  url: '/xy/qr-login/start',
  method: 'post',
  data
})

export const getXyQRLoginStatus = (params) => service({
  url: '/xy/qr-login/status',
  method: 'get',
  params
})

export const cancelXyQRLogin = (params) => service({
  url: '/xy/qr-login/cancel',
  method: 'delete',
  params
})

export const startXyWebSocket = (data) => service({
  url: '/xy/websocket/start',
  method: 'post',
  data
})

export const stopXyWebSocket = (data) => service({
  url: '/xy/websocket/stop',
  method: 'post',
  data
})

export const getXyWebSocketStatus = (params) => service({
  url: '/xy/websocket/status',
  method: 'get',
  params
})
