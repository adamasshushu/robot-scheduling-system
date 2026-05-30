import { ref, onUnmounted } from 'vue'

export interface WSMessage {
  type: string  // telemetry | alert | task_status | robot_status | connected
  data: any
}

export function useWebSocket() {
  const connected = ref(false)
  const lastMessage = ref<WSMessage | null>(null)
  const alerts = ref<any[]>([])
  const robotUpdates = ref<Record<string, any>>({})

  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let pingTimer: ReturnType<typeof setInterval> | null = null

  function connect() {
    if (ws?.readyState === WebSocket.OPEN) return

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = import.meta.env.VITE_WS_URL || `${protocol}//${location.hostname}:8000`
    const url = `${host}/api/v1/ws`

    try {
      ws = new WebSocket(url)
    } catch (e) {
      console.warn('[WS] Connection failed:', e)
      scheduleReconnect()
      return
    }

    ws.onopen = () => {
      console.log('[WS] ✅ Connected')
      connected.value = true
      // 心跳 ping 每 25 秒
      pingTimer = setInterval(() => {
        if (ws?.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'ping' }))
        }
      }, 25000)
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        lastMessage.value = msg

        switch (msg.type) {
          case 'alert':
            alerts.value.unshift(msg.data)
            if (alerts.value.length > 50) alerts.value.pop()
            break
          case 'telemetry':
          case 'robot_status':
            if (msg.data?.robot_id) {
              robotUpdates.value = { ...robotUpdates.value, [msg.data.robot_id]: msg.data }
            }
            break
        }
      } catch (e) {
        console.warn('[WS] Parse error:', e)
      }
    }

    ws.onclose = (event) => {
      console.log('[WS] Disconnected:', event.code, event.reason)
      connected.value = false
      if (pingTimer) { clearInterval(pingTimer); pingTimer = null }
      scheduleReconnect()
    }

    ws.onerror = (e) => {
      console.error('[WS] Error:', e)
      ws?.close()
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) return
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      console.log('[WS] Reconnecting...')
      connect()
    }, 3000)
  }

  function disconnect() {
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
    if (pingTimer) { clearInterval(pingTimer); pingTimer = null }
    ws?.close()
    ws = null
    connected.value = false
  }

  // 自动连接 + 清理
  connect()
  onUnmounted(disconnect)

  return { connected, lastMessage, alerts, robotUpdates, disconnect, reconnect: connect }
}
