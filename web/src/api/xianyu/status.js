import service from '@/utils/request'

// @Summary 获取闲鱼助手基础服务状态
// @Router /xy/status [get]
export const getXyStatus = () => {
  return service({
    url: '/xy/status',
    method: 'get'
  })
}

export const getXyDashboardOverview = (params) => {
  return service({
    url: '/xy/dashboard/overview',
    method: 'get',
    params
  })
}
