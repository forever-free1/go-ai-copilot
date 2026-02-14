import request from './request'

// 上传文档
export const uploadDocument = (formData: FormData) => {
  return request({
    url: '/api/v1/rag/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 获取文档列表
export const getDocuments = () => {
  return request({
    url: '/api/v1/rag/list',
    method: 'get'
  })
}

// 获取文档详情
export const getDocument = (id: number) => {
  return request({
    url: `/api/v1/rag/${id}`,
    method: 'get'
  })
}

// 删除文档
export const deleteDocument = (id: number) => {
  return request({
    url: `/api/v1/rag/${id}`,
    method: 'delete'
  })
}

// RAG 搜索
export const ragSearch = (query: string) => {
  return request({
    url: '/api/v1/rag/search',
    method: 'post',
    data: { query }
  })
}

// RAG 对话
export const ragChat = (message: string, session_id: number) => {
  return request({
    url: '/api/v1/rag/chat',
    method: 'post',
    data: { message, session_id }
  })
}
