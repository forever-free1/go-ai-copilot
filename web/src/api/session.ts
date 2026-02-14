import request from './request'

export const getSessions = () => {
  return request.get<any, any>('/api/v1/session/list')
}

export const createSession = (title: string) => {
  return request.post<any, any>('/api/v1/session', { title })
}

export const getSession = (id: number) => {
  return request.get<any, any>(`/api/v1/session/${id}`)
}

export const updateSession = (id: number, title: string) => {
  return request.put<any, any>(`/api/v1/session/${id}`, { title })
}

export const deleteSession = (id: number) => {
  return request.delete<any, any>(`/api/v1/session/${id}`)
}

export const getHistory = (id: number) => {
  return request.get<any, any>(`/api/v1/session/${id}/history`)
}
