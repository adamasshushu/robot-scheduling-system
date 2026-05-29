// ── API 封装层 ──
import axios from 'axios'
import router from '@/router'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// 请求拦截器 — 自动带 Token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('rss_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器 — 401 跳转登录
api.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('rss_token')
      router.push('/login')
    }
    return Promise.reject(err)
  }
)

// ── Auth ──
export const loginApi = (username: string, password: string) =>
  api.post('/login', { username, password })

// ── Robots ──
export const getRobots = (params?: Record<string, any>) =>
  api.get('/robots', { params })

export const getRobot = (id: number) =>
  api.get(`/robots/${id}`)

export const createRobot = (data: any) =>
  api.post('/robots', data)

export const updateRobot = (id: number, data: any) =>
  api.put(`/robots/${id}`, data)

export const deleteRobot = (id: number) =>
  api.delete(`/robots/${id}`)

export const sendCommand = (id: number, command: string, params?: any) =>
  api.post(`/robots/${id}/commands`, { command, params })

// ── Tasks ──
export const getTasks = (params?: Record<string, any>) =>
  api.get('/tasks', { params })

export const getTask = (id: number) =>
  api.get(`/tasks/${id}`)

export const createTask = (data: any) =>
  api.post('/tasks', data)

export const assignTask = (id: number, robotId?: number, auto?: boolean) =>
  api.post(`/tasks/${id}/assign`, { robot_id: robotId, auto })

export const cancelTask = (id: number) =>
  api.post(`/tasks/${id}/cancel`)

// ── Alerts ──
export const getAlerts = (params?: Record<string, any>) =>
  api.get('/alerts', { params })

export const ackAlert = (id: number) =>
  api.post(`/alerts/${id}/ack`)

export const resolveAlert = (id: number) =>
  api.post(`/alerts/${id}/resolve`)

// ── Monitor ──
export const getDashboard = () =>
  api.get('/monitor/dashboard')

export const getOnlineRobots = () =>
  api.get('/monitor/robots/online')

export const getMapRobots = () =>
  api.get('/monitor/map/robots')
