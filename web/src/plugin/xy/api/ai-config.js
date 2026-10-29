import service from '@/utils/request'

export const getAIModels = () => service({ url: '/xy/ai/models', method: 'get' })
export const createAIModel = (data) => service({ url: '/xy/ai/models', method: 'post', data })
export const updateAIModel = (id, data) => service({ url: `/xy/ai/models/${id}`, method: 'put', data })
export const deleteAIModel = (id) => service({ url: `/xy/ai/models/${id}`, method: 'delete' })
export const setDefaultAIModel = (id) => service({ url: `/xy/ai/models/${id}/default`, method: 'put' })

export const getMCPServers = () => service({ url: '/xy/ai/mcp-servers', method: 'get' })
export const createMCPServer = (data) => service({ url: '/xy/ai/mcp-servers', method: 'post', data })
export const updateMCPServer = (id, data) => service({ url: `/xy/ai/mcp-servers/${id}`, method: 'put', data })
export const deleteMCPServer = (id) => service({ url: `/xy/ai/mcp-servers/${id}`, method: 'delete' })

export const getAgentConfig = () => service({ url: '/xy/ai/agent-config', method: 'get' })
export const updateAgentConfig = (data) => service({ url: '/xy/ai/agent-config', method: 'put', data })
