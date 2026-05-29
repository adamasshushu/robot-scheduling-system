<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { getRobots, deleteRobot, sendCommand } from '@/api'

const router = useRouter()
const robots = ref<any[]>([])
const loading = ref(false)
const filters = reactive({ status: '', model: '', battery_min: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

async function fetchRobots() {
  loading.value = true
  try {
    const { data } = await getRobots({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      model: filters.model || undefined,
      battery_min: filters.battery_min || undefined,
    })
    robots.value = data.data?.list || []
    pagination.total = data.data?.pagination?.total || 0
  } finally { loading.value = false }
}

async function handleCommand(robot: any, cmd: string) {
  try {
    await sendCommand(robot.id, cmd)
    ElMessage.success(`已向 ${robot.name} 下发 ${cmd} 指令`)
  } catch { ElMessage.error('指令下发失败') }
}

onMounted(fetchRobots)
</script>

<template>
  <div class="robots-page">
    <!-- 筛选栏 -->
    <el-card class="filter-bar">
      <el-form :model="filters" inline>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部" @change="fetchRobots">
            <el-option label="待机" value="standby" />
            <el-option label="运行中" value="running" />
            <el-option label="充电中" value="charging" />
            <el-option label="故障" value="fault" />
          </el-select>
        </el-form-item>
        <el-form-item label="最低电量">
          <el-input-number v-model="filters.battery_min" :min="0" :max="100" @change="fetchRobots" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchRobots">查询</el-button>
          <el-button @click="router.push('/tasks')">创建任务</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 机器人卡片网格 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="8" v-for="r in robots" :key="r.id" style="margin-bottom: 16px;">
        <el-card shadow="hover" class="robot-card" @click="router.push(`/robots/${r.id}`)">
          <template #header>
            <div class="robot-card-header">
              <span class="robot-name">{{ r.name }}</span>
              <el-tag :type="r.status === 'running' ? 'success' : r.status === 'fault' ? 'danger' : r.status === 'charging' ? 'warning' : 'info'" size="small">
                {{ { standby: '待机', running: '运行中', charging: '充电中', fault: '故障', manual: '手动' }[r.status] || r.status }}
              </el-tag>
            </div>
          </template>
          <div class="robot-info">
            <p>🔋 电量: {{ r.battery_pct }}%</p>
            <p>📍 位置: {{ r.map_zone || `${r.location_x},${r.location_y}` }}</p>
            <p>🚀 速度: {{ r.speed }} m/s</p>
            <p>💡 灯光: <span :style="{ color: r.light_color }">{{ r.light_mode }}</span></p>
            <div class="robot-actions">
              <el-button size="small" type="success" @click.stop="handleCommand(r, 'start')">启动</el-button>
              <el-button size="small" type="warning" @click.stop="handleCommand(r, 'pause')">暂停</el-button>
              <el-button size="small" type="danger" @click.stop="handleCommand(r, 'stop')">急停</el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 分页 -->
    <el-pagination
      v-if="pagination.total > pagination.page_size"
      v-model:current-page="pagination.page"
      :page-size="pagination.page_size"
      :total="pagination.total"
      layout="total, prev, pager, next"
      @change="fetchRobots"
      style="margin-top: 16px; justify-content: center;"
    />

    <el-empty v-if="!loading && !robots.length" description="暂无机器人" />
  </div>
</template>

<style scoped>
.robot-card { cursor: pointer; transition: transform 0.2s; }
.robot-card:hover { transform: translateY(-2px); }
.robot-card-header { display: flex; justify-content: space-between; align-items: center; }
.robot-name { font-size: 16px; font-weight: 600; }
.robot-info p { margin: 6px 0; font-size: 13px; }
.robot-actions { margin-top: 12px; display: flex; gap: 4px; }
</style>
