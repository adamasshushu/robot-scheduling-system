<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { getRobots, getTasks } from '@/api'
import { useWebSocket } from '@/composables/useWebSocket'

const stats = ref({ robots: 0, online: 0, tasks: 0, completed: 0, alerts: 0 })
const runningTasks = ref<any[]>([])
const wsConnected = ref(false)
const wsAlertCount = ref(0)

const { connected, alerts: wsAlerts } = useWebSocket()

watch(connected, (v) => { wsConnected.value = v })
watch(wsAlerts, (a) => {
  wsAlertCount.value = a.length
  stats.value.alerts = a.filter((x: any) => x.status === 'unack').length
}, { deep: true })

onMounted(async () => {
  try {
    const [robotsRes, tasksRes] = await Promise.all([
      getRobots(),
      getTasks({ page_size: 100 }),
    ])
    const robots = robotsRes.data.data?.list || []
    const tasks = tasksRes.data.data?.list || []
    stats.value = {
      robots: robots.length,
      online: robots.filter((r: any) => r.comm_status === 'online').length,
      tasks: tasks.length,
      completed: tasks.filter((t: any) => t.status === 'completed').length,
      alerts: wsAlertCount.value,
    }
    runningTasks.value = tasks.filter((t: any) => t.status === 'running' || t.status === 'assigned')
  } catch (e) { /* silent */ }
})
</script>

<template>
  <div class="dashboard">
    <!-- 连接状态条 -->
    <div class="ws-status" :class="{ connected: wsConnected }">
      <span class="ws-dot"></span>
      {{ wsConnected ? '🔗 实时连接已建立' : '⏳ 正在连接实时推送...' }}
    </div>

    <!-- KPI 卡片 -->
    <el-row :gutter="16" class="kpi-row" style="margin-top: 8px;">
      <el-col :span="6" v-for="kpi in [
        { label: '机器人在线', value: `${stats.online}/${stats.robots}`, icon: 'Monitor', color: '#67c23a', gradient: '135deg, #11998e, #38ef7d' },
        { label: '今日任务', value: stats.tasks, icon: 'List', color: '#409eff', gradient: '135deg, #4facfe, #00f2fe' },
        { label: '任务完成率', value: stats.tasks ? `${Math.round(stats.completed/stats.tasks*100)}%` : '0%', icon: 'CircleCheck', color: '#e6a23c', gradient: '135deg, #f093fb, #f5576c' },
        { label: '未处理告警', value: stats.alerts, icon: 'Bell', color: '#f56c6c', gradient: '135deg, #fa709a, #fee140' },
      ]" :key="kpi.label">
        <el-card shadow="hover" class="kpi-card glass-card">
          <div class="kpi-content">
            <div>
              <p class="kpi-label">{{ kpi.label }}</p>
              <p class="kpi-value kpi-animate">{{ kpi.value }}</p>
            </div>
            <div class="kpi-icon-circle" :style="{ background: `linear-gradient(${kpi.gradient})` }">
              <el-icon :size="28" color="#fff"><component :is="kpi.icon" /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 实时告警推送 -->
    <el-card v-if="wsAlerts.length" class="glass-card" style="margin-top: 16px;">
      <template #header>🔔 实时告警推送</template>
      <div class="alert-stream">
        <div v-for="a in wsAlerts.slice(0, 5)" :key="a.id" class="alert-row"
          :class="{ critical: a.severity === 'critical' }">
          <el-tag :type="a.severity === 'critical' ? 'danger' : 'warning'" size="small" effect="dark">
            {{ a.severity }}
          </el-tag>
          <span class="alert-title">{{ a.title }}</span>
          <span class="alert-time">{{ a.created_at?.slice(11, 19) }}</span>
        </div>
      </div>
    </el-card>

    <!-- 地图 + 侧栏 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="16">
        <el-card class="glass-card">
          <template #header>📍 机器人实时位置</template>
          <div class="map-placeholder">
            <span class="map-icon">🗺️</span>
            <p>Leaflet 地图组件</p>
            <span class="map-hint">等待配置地图服务后加载</span>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="glass-card">
          <template #header>⚡ 执行中任务</template>
          <el-empty v-if="!runningTasks.length" description="暂无任务" :image-size="80" />
          <div v-else>
            <div v-for="t in runningTasks" :key="t.id" class="task-item">
              <el-tag size="small" :type="t.status === 'running' ? 'success' : 'warning'">
                {{ t.status === 'running' ? '执行中' : '已指派' }}
              </el-tag>
              <span class="task-code">{{ t.task_code }}</span>
              <span class="task-desc">{{ t.description || t.target_location }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.ws-status {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 16px; border-radius: 10px;
  background: rgba(245,108,108,0.1); color: #f56c6c;
  font-size: 13px; transition: all 0.3s;
}
.ws-status.connected {
  background: rgba(103,194,58,0.1); color: #67c23a;
}
.ws-dot {
  width: 8px; height: 8px; border-radius: 50%;
  background: currentColor; animation: pulse 2s infinite;
}
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }
.kpi-card { cursor: pointer; }
.kpi-content { display: flex; justify-content: space-between; align-items: center; }
.kpi-label { font-size: 13px; color: var(--el-text-color-secondary); margin: 0 0 8px; }
.kpi-value { font-size: 28px; font-weight: 700; margin: 0; color: var(--el-text-color-primary); }
.kpi-icon-circle {
  width: 56px; height: 56px; border-radius: 14px;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 4px 15px rgba(0,0,0,0.15);
}
.alert-stream { display: flex; flex-direction: column; gap: 8px; }
.alert-row {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 12px; border-radius: 8px;
  background: rgba(255,255,255,0.03); border-left: 3px solid #e6a23c;
  transition: background 0.3s;
}
.alert-row.critical { border-left-color: #f56c6c; background: rgba(245,108,108,0.06); }
.alert-title { flex: 1; font-size: 13px; font-weight: 500; }
.alert-time { font-size: 11px; color: var(--el-text-color-placeholder); }
.map-placeholder {
  height: 400px; border-radius: 12px;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  color: var(--el-text-color-secondary);
  background: rgba(255,255,255,0.03);
  border: 1px dashed rgba(255,255,255,0.1);
}
.map-icon { font-size: 56px; margin-bottom: 8px; }
.map-hint { font-size: 12px; opacity: 0.6; }
.task-item { display: flex; align-items: center; gap: 8px; padding: 10px 0; border-bottom: 1px solid var(--el-border-color-lighter); }
.task-item:last-child { border-bottom: none; }
.task-code { font-weight: 600; font-size: 13px; color: var(--el-color-primary); }
.task-desc { flex: 1; font-size: 12px; color: var(--el-text-color-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
