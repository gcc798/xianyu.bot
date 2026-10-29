import service from '@/utils/request'

// @Summary 获取闲鱼助手基础服务状态
// @Router /xianyu/status [get]
export const getXianyuStatus = () => {
  return service({
    url: '/xianyu/status',
    method: 'get'
  })
}

export const getXianyuDashboardOverview = (params) => {
  return service({
    url: '/xianyu/dashboard/overview',
    method: 'get',
    params
  })
}
