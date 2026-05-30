<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { getRobots, deleteRobot, sendCommand, createRobot, sortRobots } from '@/api'

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

// ── 添加机器人 ──
const addRobotDialog = ref(false)
const addRobotForm = reactive({ name: '', robot_code: '', robot_model: '', status: 'standby' })

async function handleAddRobot() {
  await createRobot(addRobotForm)
  ElMessage.success('机器人已添加')
  addRobotDialog.value = false
  Object.assign(addRobotForm, { name: '', robot_code: '', robot_model: '', status: 'standby' })
  fetchRobots()
}

// ── 删除机器人 ──
async function handleDeleteRobot(robot: any) {
  await ElMessageBox.confirm(`确定删除机器人 ${robot.name}？`, '警告', { type: 'warning' })
  await deleteRobot(robot.id)
  ElMessage.success('已删除')
  fetchRobots()
}

// ── 排序 ──
async function moveRobot(robot: any, direction: 'up' | 'down') {
  const idx = robots.value.indexOf(robot)
  if (direction === 'up' && idx > 0) {
    [robots.value[idx], robots.value[idx-1]] = [robots.value[idx-1], robots.value[idx]]
  } else if (direction === 'down' && idx < robots.value.length - 1) {
    [robots.value[idx], robots.value[idx+1]] = [robots.value[idx+1], robots.value[idx]]
  }
  const items = robots.value.map((r, i) => ({ id: r.id, sort_order: i }))
  await sortRobots(items)
}

onMounted(fetchRobots)
</script>

<template>
  <div class="robots-page">
    <!-- 筛选栏 -->
    <el-card class="filter-bar glass-panel">
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
          <el-button type="primary" class="glass-btn" @click="fetchRobots">查询</el-button>
          <el-button class="glass-btn" @click="addRobotDialog = true">+ 添加机器人</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 机器人卡片网格 -->
    <el-row :gutter="16" style="margin-top: 16px;">
      <el-col :span="8" v-for="r in robots" :key="r.id" style="margin-bottom: 16px;">
        <el-card shadow="hover" class="robot-card glass-card" @click="router.push(`/robots/${r.id}`)">
          <template #header>
            <div class="robot-card-header">
              <div style="display:flex;align-items:center;gap:4px">
                <span class="robot-name">{{ r.name }}</span>
                <el-button size="small" circle @click.stop="moveRobot(r, 'up')" :disabled="robots.indexOf(r) === 0">↑</el-button>
                <el-button size="small" circle @click.stop="moveRobot(r, 'down')" :disabled="robots.indexOf(r) === robots.length - 1">↓</el-button>
              </div>
              <el-tag :type="r.status === 'running' ? 'success' : r.status === 'fault' ? 'danger' : r.status === 'charging' ? 'warning' : 'info'" size="small" effect="dark">
                {{ { standby: '待机', running: '运行中', charging: '充电中', fault: '故障', manual: '手动' }[r.status] || r.status }}
              </el-tag>
            </div>
          </template>
          <div class="robot-info">
            <div class="robot-stat">
              <span class="stat-icon">🔋</span>
              <span>{{ r.battery_pct }}%</span>
              <el-progress :percentage="r.battery_pct || 0" :stroke-width="5" :color="r.battery_pct > 20 ? '#67c23a' : '#f56c6c'" style="flex:1;margin-left:8px;" />
            </div>
            <div class="robot-stat"><span class="stat-icon">📍</span><span>{{ r.map_zone || `${r.location_x},${r.location_y}` }}</span></div>
            <div class="robot-stat"><span class="stat-icon">🚀</span><span>{{ r.speed }} m/s</span></div>
            <div class="robot-stat"><span class="stat-icon">💡</span><span :style="{ color: r.light_color }">{{ r.light_mode }}</span></div>
            <div class="robot-actions">
              <el-button size="small" type="success" class="glass-btn" @click.stop="handleCommand(r, 'start')">启动</el-button>
              <el-button size="small" type="warning" @click.stop="handleCommand(r, 'pause')">暂停</el-button>
              <el-button size="small" type="danger" @click.stop="handleCommand(r, 'stop')">急停</el-button>
              <el-button size="small" type="danger" @click.stop="handleDeleteRobot(r)">删除</el-button>
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

    <!-- 添加机器人弹窗 -->
    <el-dialog v-model="addRobotDialog" title="添加机器人" width="420px">
      <el-form :model="addRobotForm" label-width="80px">
        <el-form-item label="名称"><el-input v-model="addRobotForm.name" placeholder="如: 物流A-1" /></el-form-item>
        <el-form-item label="编号"><el-input v-model="addRobotForm.robot_code" placeholder="如: RB-A001" /></el-form-item>
        <el-form-item label="型号"><el-input v-model="addRobotForm.robot_model" placeholder="如: T6-Pro" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="addRobotForm.status" style="width:100%">
            <el-option label="待机" value="standby" />
            <el-option label="运行中" value="running" />
            <el-option label="充电中" value="charging" />
            <el-option label="故障" value="fault" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addRobotDialog = false">取消</el-button>
        <el-button type="primary" class="glass-btn" @click="handleAddRobot">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.robot-card { cursor: pointer; transition: transform 0.3s, box-shadow 0.3s; }
.robot-card:hover { transform: translateY(-4px); }
.robot-card-header { display: flex; justify-content: space-between; align-items: center; }
.robot-name { font-size: 16px; font-weight: 700; }
.robot-info { display: flex; flex-direction: column; gap: 8px; }
.robot-stat { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--el-text-color-regular); }
.stat-icon { font-size: 15px; width: 20px; text-align: center; }
.robot-actions { margin-top: 8px; display: flex; gap: 6px; flex-wrap: wrap; }
.filter-bar { margin-bottom: 0; }
</style>
