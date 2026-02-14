import request from './request'

export interface ChatRequest {
  message: string
  session_id?: number
  api_key?: string
  model?: string
  temperature?: number
}

// 普通对话
export const chat = (data: ChatRequest) => {
  return request.post<any, any>('/api/v1/chat', data)
}

// 流式对话
export const streamChat = (data: ChatRequest, onMessage: (content: string) => void, onDone: () => void, onError: (error: string) => void) => {
  const token = localStorage.getItem('token')
  const eventSource = new EventSource(`/api/v1/chat/stream?message=${encodeURIComponent(data.message)}&session_id=${data.session_id || 0}`, {
    withCredentials: true,
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })

  eventSource.onmessage = (event) => {
    if (event.data === '[DONE]') {
      onDone()
      eventSource.close()
    } else {
      onMessage(event.data)
    }
  }

  eventSource.onerror = (error) => {
    onError('连接错误')
    eventSource.close()
  }

  return eventSource
}

// 带模式的对话
export const chatWithMode = (data: ChatRequest & { mode: string }) => {
  return request.post<any, any>('/api/v1/chat/mode', data, {
    params: { mode: data.mode }
  })
}

// RAG对话
export const ragChat = (data: ChatRequest) => {
  return request.post<any, any>('/api/v1/rag/chat', data)
}
