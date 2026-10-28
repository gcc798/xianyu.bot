import service from '@/utils/request'

const telegramJSON = (config) => service(config).then(response => ({ code: 0, data: response.data }))
const proxyURL = (path) => `/xianyu/proxy/telegram${path}`

export const getTelegramAdminStatus = () => telegramJSON({ url: proxyURL('/api/admin/status'), method: 'get', donNotShowLoading: true })
export const startTelegramLogin = () => telegramJSON({ url: proxyURL('/api/admin/login/qr'), method: 'post' })
export const getTelegramLoginStatus = () => telegramJSON({ url: proxyURL('/api/admin/login/status'), method: 'get', donNotShowLoading: true })
export const submitTelegramPassword = (password) => telegramJSON({ url: proxyURL('/api/admin/login/2fa'), method: 'post', data: { password } })
export const startTelegramScan = () => telegramJSON({ url: proxyURL('/api/admin/scan'), method: 'post' })
export const getTelegramScanStatus = () => telegramJSON({ url: proxyURL('/api/admin/scan/status'), method: 'get', donNotShowLoading: true })
export const getTelegramDialogs = (params) => telegramJSON({ url: proxyURL('/api/admin/dialogs'), method: 'get', params })
export const leaveTelegramDialogs = (dialogIds) => telegramJSON({ url: proxyURL('/api/admin/dialogs/leave'), method: 'post', data: { dialog_ids: dialogIds, confirm: '退出' } })
export const exportTelegramDialogs = (params) => service({ url: proxyURL('/api/admin/dialogs/export'), method: 'get', params, responseType: 'blob' })
export const getTelegramSettings = () => telegramJSON({ url: proxyURL('/api/admin/settings'), method: 'get' })
export const getTelegramResourceOptions = () => telegramJSON({ url: proxyURL('/api/resources/options'), method: 'get' })
export const createTelegramResourceTask = (data) => telegramJSON({ url: proxyURL('/api/resources/tasks'), method: 'post', data })
export const getTelegramResourceTasks = (params) => telegramJSON({ url: proxyURL('/api/resources/tasks'), method: 'get', params, donNotShowLoading: true })
export const cancelTelegramResourceTask = (id) => telegramJSON({ url: proxyURL(`/api/resources/tasks/${id}/cancel`), method: 'post' })
export const retryTelegramResourceTask = (id) => telegramJSON({ url: proxyURL(`/api/resources/tasks/${id}/retry`), method: 'post' })
export const getTelegramResourceFiles = (params) => telegramJSON({ url: proxyURL('/api/resources/files'), method: 'get', params, donNotShowLoading: true })
export const getTelegramResourceFile = (id, action) => service({ url: proxyURL(`/api/resources/files/${id}/${action}`), method: 'get', responseType: 'blob' })
