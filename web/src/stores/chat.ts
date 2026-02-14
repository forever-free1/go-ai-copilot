import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getSessions, createSession, deleteSession, getHistory } from '../api/session'

export interface Session {
  id: number
  title: string
  created_at: string
  updated_at: string
}

export interface Message {
  id: number
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

export const useChatStore = defineStore('chat', () => {
  const sessions = ref<Session[]>([])
  const currentSessionId = ref<number | null>(null)
  const messages = ref<Message[]>([])
  const loading = ref(false)

  // 获取会话列表
  const fetchSessions = async () => {
    const res = await getSessions()
    sessions.value = res.data
    return res.data
  }

  // 创建会话
  const createNewSession = async (title: string) => {
    const res = await createSession(title)
    sessions.value.unshift(res.data)
    return res.data
  }

  // 删除会话
  const removeSession = async (id: number) => {
    await deleteSession(id)
    sessions.value = sessions.value.filter(s => s.id !== id)
    if (currentSessionId.value === id) {
      currentSessionId.value = null
      messages.value = []
    }
  }

  // 选择会话
  const selectSession = async (id: number) => {
    currentSessionId.value = id
    loading.value = true
    try {
      const res = await getHistory(id)
      messages.value = res.data
    } finally {
      loading.value = false
    }
  }

  // 添加消息
  const addMessage = (role: 'user' | 'assistant', content: string) => {
    messages.value.push({
      id: Date.now(),
      role,
      content,
      created_at: new Date().toISOString()
    })
  }

  // 清空当前消息
  const clearMessages = () => {
    messages.value = []
    currentSessionId.value = null
  }

  return {
    sessions,
    currentSessionId,
    messages,
    loading,
    fetchSessions,
    createNewSession,
    removeSession,
    selectSession,
    addMessage,
    clearMessages
  }
})
