import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login, register, getUserInfo } from '../api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>('')
  const userInfo = ref<{
    id: number
    username: string
    nickname: string
    email: string
  } | null>(null)

  // 初始化时从localStorage恢复
  const init = () => {
    const savedToken = localStorage.getItem('token')
    if (savedToken) {
      token.value = savedToken
      fetchUserInfo()
    }
  }

  // 登录
  const loginAction = async (username: string, password: string) => {
    const res = await login(username, password)
    token.value = res.data.token
    userInfo.value = res.data.user
    localStorage.setItem('token', res.data.token)
    return res
  }

  // 注册
  const registerAction = async (username: string, password: string, nickname?: string) => {
    const res = await register(username, password, nickname)
    return res
  }

  // 获取用户信息
  const fetchUserInfo = async () => {
    if (!token.value) return
    try {
      const res = await getUserInfo()
      userInfo.value = res.data
    } catch (error) {
      logout()
    }
  }

  // 退出登录
  const logout = () => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }

  // 初始化
  init()

  return {
    token,
    userInfo,
    loginAction,
    registerAction,
    fetchUserInfo,
    logout
  }
})
