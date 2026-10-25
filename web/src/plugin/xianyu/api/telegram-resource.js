import service from '@/utils/request'

export const getTelegramResourceOptions = () => service({ url: '/xianyu/telegram/options', method: 'get' })
export const createTelegramResourceTask = (data) => service({ url: '/xianyu/telegram/tasks', method: 'post', data })
export const getTelegramResourceTasks = (params) => service({ url: '/xianyu/telegram/tasks', method: 'get', params, donNotShowLoading: true })
export const cancelTelegramResourceTask = (id) => service({ url: `/xianyu/telegram/tasks/${id}/cancel`, method: 'post' })
export const retryTelegramResourceTask = (id) => service({ url: `/xianyu/telegram/tasks/${id}/retry`, method: 'post' })
export const getTelegramResourceFiles = (params) => service({ url: '/xianyu/telegram/files', method: 'get', params, donNotShowLoading: true })
export const getTelegramResourceFile = (id, action) => service({ url: `/xianyu/telegram/files/${id}/${action}`, method: 'get', responseType: 'blob' })
