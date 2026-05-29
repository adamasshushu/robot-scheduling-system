<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getRobots, getTasks, getAlerts } from '@/api'

const stats = ref({ robots: 0, online: 0, tasks: 0, completed: 0, alerts: 0 })
const recentAlerts = ref<any[]>([])
const runningTasks = ref<any[]>([])

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
      alerts: 0,
    }
    runningTasks.value = tasks.filter((t: any) => t.status === 'running' || t.status === 'assigned')
  } catch (e) { /* silent */ }
})
</script>

<template>
  <div class="dashboard">
    <!-- KPI 卡片 -->
    <el-row :gutter="16" class="kpi-row">
      <el-col :span="6" v-for="kpi in [
        { label: '机器人在线', value: `${stats.online}/${stats.robots}`, icon: 'Monitor', color: '#67c23a' },
        { label: '今日任务', value: stats.tasks, icon: 'List', color: '#409eff' },
        { label: '任务完成率', value: stats.tasks ? `${Math.round(stats.completed/stats.tasks*100)}%` : '0%', icon: 'CircleCheck', color: '#e6a23c' },
        { label: '未处理告警', value: stats.alerts, icon: 'Bell', color: '#f56c6c' },
      ]" :key="kpi.label">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-content">
            <div>
              <p class="kpi-label">{{ kpi.label }}</p>
              <p class="kpi-value">{{ kpi.value }}</p>
            </div>
            <el-icon :size="40" :color="kpi.color"><component :is="kpi.icon" /></el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 地图 + 侧栏 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="16">
        <el-card>
          <template #header>📍 机器人实时位置</template>
          <div style="height: 400px; background: var(--el-fill-color-light); border-radius: 8px;
            display: flex; align-items: center; justify-content: center; color: var(--el-text-color-secondary);">
            🗺️ 地图组件加载中 (Leaflet)
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
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
.kpi-card { cursor: default; }
.kpi-content { display: flex; justify-content: space-between; align-items: center; }
.kpi-label { font-size: 13px; color: var(--el-text-color-secondary); margin: 0 0 8px; }
.kpi-value { font-size: 28px; font-weight: 700; margin: 0; color: var(--el-text-color-primary); }
.task-item { display: flex; align-items: center; gap: 8px; padding: 8px 0; border-bottom: 1px solid var(--el-border-color-lighter); }
.task-code { font-weight: 600; font-size: 13px; }
.task-desc { font-size: 12px; color: var(--el-text-color-secondary); flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
