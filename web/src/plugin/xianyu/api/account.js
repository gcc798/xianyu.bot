import service from '@/utils/request'

export const getXianyuAccountList = (params) => service({
  url: '/xianyu/account/list',
  method: 'get',
  params
})

export const createXianyuAccount = (data) => service({
  url: '/xianyu/account/create',
  method: 'post',
  data
})

export const updateXianyuAccount = (data) => service({
  url: '/xianyu/account/update',
  method: 'put',
  data
})

export const deleteXianyuAccount = (params) => service({
  url: '/xianyu/account/delete',
  method: 'delete',
  params
})

export const saveXianyuCredential = (data) => service({
  url: '/xianyu/account/credential',
  method: 'put',
  data
})

export const checkXianyuLogin = (data) => service({
  url: '/xianyu/account/check-login',
  method: 'post',
  data
})

export const startXianyuQRLogin = (data) => service({
  url: '/xianyu/qr-login/start',
  method: 'post',
  data
})

export const getXianyuQRLoginStatus = (params) => service({
  url: '/xianyu/qr-login/status',
  method: 'get',
  params
})

export const cancelXianyuQRLogin = (params) => service({
  url: '/xianyu/qr-login/cancel',
  method: 'delete',
  params
})

export const startXianyuWebSocket = (data) => service({
  url: '/xianyu/websocket/start',
  method: 'post',
  data
})

export const stopXianyuWebSocket = (data) => service({
  url: '/xianyu/websocket/stop',
  method: 'post',
  data
})

export const getXianyuWebSocketStatus = (params) => service({
  url: '/xianyu/websocket/status',
  method: 'get',
  params
})
