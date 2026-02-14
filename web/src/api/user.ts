import request from './request'

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  nickname?: string
}

export const login = (username: string, password: string) => {
  return request.post<any, any>('/api/v1/user/login', { username, password })
}

export const register = (username: string, password: string, nickname?: string) => {
  return request.post<any, any>('/api/v1/user/register', { username, password, nickname })
}

export const getUserInfo = () => {
  return request.get<any, any>('/api/v1/user/info')
}

export const updateUserInfo = (data: { nickname?: string; email?: string }) => {
  return request.put<any, any>('/api/v1/user/info', data)
}

export const changePassword = (oldPassword: string, newPassword: string) => {
  return request.put<any, any>('/api/v1/user/password', { old_password: oldPassword, new_password: newPassword })
}
