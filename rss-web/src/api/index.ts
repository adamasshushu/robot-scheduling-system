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

export const sortRobots = (items: { id: number, sort_order: number }[]) =>
  api.put('/robots/sort', items)

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

export const updateTask = (id: number, data: any) =>
  api.put(`/tasks/${id}`, data)

export const deleteTask = (id: number) =>
  api.delete(`/tasks/${id}`)

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

// ── Reports ──
export const getTaskStats = (days?: number) =>
  api.get('/reports/tasks', { params: { days } })

export const getBatteryTrend = (hours?: number) =>
  api.get('/reports/battery', { params: { hours } })

export const getTaskTypes = () =>
  api.get('/reports/task-types')

export const getRobotUsage = () =>
  api.get('/reports/robot-usage')

// ── Settings ──
export const getUsers = (params?: any) =>
  api.get('/settings/users', { params })

export const createUser = (data: any) =>
  api.post('/settings/users', data)

export const updateUser = (id: number, data: any) =>
  api.put(`/settings/users/${id}`, data)

export const deleteUser = (id: number) =>
  api.delete(`/settings/users/${id}`)

export const getModels = () =>
  api.get('/settings/models')

export const createModel = (data: any) =>
  api.post('/settings/models', data)

export const deleteModel = (id: number) =>
  api.delete(`/settings/models/${id}`)

// ── Maps ──
export const uploadMap = (formData: FormData) =>
  api.post('/maps/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })

export const getMaps = () =>
  api.get('/maps')

export const getActiveMap = () =>
  api.get('/maps/active')

export const calibrateMap = (id: number, data: any) =>
  api.put(`/maps/${id}/calibrate`, data)

export const setActiveMap = (id: number) =>
  api.post(`/maps/${id}/active`)

export const deleteMap = (id: number) =>
  api.delete(`/maps/${id}`)

// ── Robot Center (模块1-7) ──
export const getRobotFull = (id: number) =>
  api.get(`/robots/${id}/full`)

export const updateRobotInfo = (id: number, data: any) =>
  api.put(`/robots/${id}/info`, data)

export const getRobotMapBinding = (id: number) =>
  api.get(`/robots/${id}/map-binding`)

export const bindRobotMap = (id: number, data: any) =>
  api.post(`/robots/${id}/map-binding`, data)

export const getSchedules = (robotId: number) =>
  api.get(`/robots/${robotId}/schedules`)

export const createSchedule = (robotId: number, data: any) =>
  api.post(`/robots/${robotId}/schedules`, data)

export const updateSchedule = (robotId: number, scheduleId: number, data: any) =>
  api.put(`/robots/${robotId}/schedules/${scheduleId}`, data)

export const deleteSchedule = (robotId: number, scheduleId: number) =>
  api.delete(`/robots/${robotId}/schedules/${scheduleId}`)

export const getWorkRecords = (robotId: number) =>
  api.get(`/robots/${robotId}/work-records`)

export const getOpLogs = (robotId: number, logType?: string) =>
  api.get(`/robots/${robotId}/op-logs`, { params: { type: logType } })

export const relocateRobot = (robotId: number, data: any) =>
  api.post(`/robots/${robotId}/relocate`, data)

export const updateRobotSettings = (robotId: number, data: any) =>
  api.put(`/robots/${robotId}/settings`, data)

// ── OTA ──
export const uploadFirmware = (formData: FormData) =>
  api.post('/ota/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })

export const getFirmwares = (module?: string) =>
  api.get('/ota/firmwares', { params: { module } })

export const toggleAutoUpgrade = (id: number, autoUpgrade: boolean) =>
  api.put(`/ota/firmwares/${id}`, { auto_upgrade: autoUpgrade })

export const upgradeRobot = (robotId: number, firmwareId: number, mode?: string) =>
  api.post(`/ota/upgrade/${robotId}`, { firmware_id: firmwareId, mode: mode || 'manual' })

// ── System Config ──
export const getSystemConfig = () =>
  api.get('/system/config')

export const updateSystemConfig = (data: Record<string, string>) =>
  api.put('/system/config', data)
