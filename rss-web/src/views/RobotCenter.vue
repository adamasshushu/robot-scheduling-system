<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getRobot } from '@/api'

const route = useRoute()
const robot = ref<any>({})
const activeTab = ref('info')

const robotId = Number(route.params.id)

onMounted(async () => {
  const { data } = await getRobot(robotId)
  robot.value = data.data || {}
})

// API基础URL
const API = '/api/v1'
const token = () => localStorage.getItem('rss_token') || ''

async function apiGet(url: string) {
  const res = await fetch(`${API}${url}`, { headers: { Authorization: `Bearer ${token()}` } })
  return res.json()
}
async function apiPost(url: string, body?: any, method: 'POST' | 'PUT' = 'POST') {
  try {
    const opts: any = {
      method,
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token()}` },
    }
    if (body) opts.body = JSON.stringify(body)
    const res = await fetch(`${API}${url}`, opts)
    const data = await res.json()
    if (data.code !== 200) throw new Error(data.message)
    ElMessage.success('操作成功')
    return data
  } catch (e: any) { ElMessage.error(e.message || '操作失败') }
}

// ── 模块2: 地图 ──
const maps = ref<any[]>([])
const robotMap = ref<any>({})
const mapForm = ref({ map_id: 0, map_name: '', is_auto: true })
async function loadMaps() {
  const r = await apiGet('/maps')
  maps.value = r.data || []
  const rm = await apiGet(`/robots/${robotId}/map-binding`)
  robotMap.value = rm.data || {}
}
async function bindMap() { await apiPost(`/robots/${robotId}/map-binding`, mapForm.value); loadMaps(); ElMessage.success('已绑定') }

// ── 模块3: 定时任务 ──
const schedules = ref<any[]>([])
const schedForm = ref({ name: '', cron_expr: '0 8 * * *', task_type: 'transport', params: '{}', enabled: true })
async function loadSchedules() { const r = await apiGet(`/robots/${robotId}/schedules`); schedules.value = r.data || [] }
async function addSchedule() { await apiPost(`/robots/${robotId}/schedules`, schedForm.value); loadSchedules(); ElMessage.success('已添加') }
async function toggleSchedule(s: any) { await apiPost(`/robots/${robotId}/schedules/${s.id}`, { enabled: !s.enabled }, 'PUT'); loadSchedules() }
async function delSchedule(id: number) { await fetch(`${API}/robots/${robotId}/schedules/${id}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token()}` } }); loadSchedules() }

// ── 模块4: 工作记录 ──
const workRecords = ref<any[]>([])
async function loadWorkRecords() { const r = await apiGet(`/robots/${robotId}/work-records`); workRecords.value = r.data || [] }

// ── 模块5: 日志 ──
const opLogs = ref<any[]>([])
const logFilter = ref('')
async function loadOpLogs() { const r = await apiGet(`/robots/${robotId}/op-logs?type=${logFilter.value}`); opLogs.value = r.data || [] }

// ── 模块6: 重定位 ──
const relocForm = ref({ mode: 'auto', pos_x: 0, pos_y: 0, theta: 0 })
async function relocate() { await apiPost(`/robots/${robotId}/relocate`, relocForm.value); ElMessage.success('已下发') }

// ── 模块7: 设置 ──
const settingsForm = ref<any>({})
async function loadSettings() {
  const r = await apiGet(`/robots/${robotId}/full`)
  const rob = r.data || {}
  settingsForm.value = {
    voice_volume: rob.voice_volume || 80,
    max_speed_setting: rob.max_speed_setting || rob.max_speed || 0,
    battery_alert_pct: rob.battery_alert_pct || 20,
    auto_charge_pct: rob.auto_charge_pct || 15,
    locked: rob.locked || false,
    screen_lock_pwd: '',
  }
}
async function saveSettings() {
  await apiPost(`/robots/${robotId}/settings`, settingsForm.value, 'PUT')
  ElMessage.success('设置已保存')
}

// 初始化
loadMaps(); loadSchedules(); loadWorkRecords(); loadOpLogs(); loadSettings()
</script>

<template>
  <div class="robot-center">
    <el-page-header @back="$router.back()" :content="robot.name || '机器人管理中心'" />

    <el-card style="margin-top:12px" class="glass-card">
      <el-descriptions :column="4" border size="small">
        <el-descriptions-item label="编号">{{ robot.robot_code }}</el-descriptions-item>
        <el-descriptions-item label="SN">{{ robot.serial_number || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="robot.status === 'running' ? 'success' : 'info'" effect="dark" size="small">{{ robot.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="电量">{{ robot.battery_pct }}%</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-tabs v-model="activeTab" style="margin-top:12px">
      <!-- 模块1: 机器人信息 -->
      <el-tab-pane label="📋 机器人信息" name="info">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-card class="glass-card" header="💻 软件版本">
              <el-descriptions :column="2" border size="small">
                <el-descriptions-item label="前端">{{ robot.sw_frontend_ver || '-' }}</el-descriptions-item>
                <el-descriptions-item label="后端">{{ robot.sw_backend_ver || '-' }}</el-descriptions-item>
                <el-descriptions-item label="下位机">{{ robot.sw_firmware_ver || '-' }}</el-descriptions-item>
                <el-descriptions-item label="导航系统">{{ robot.sw_nav_ver || '-' }}</el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card class="glass-card" header="🔧 硬件信息">
              <el-descriptions :column="2" border size="small">
                <el-descriptions-item label="型号">{{ robot.hw_model || robot.model }}</el-descriptions-item>
                <el-descriptions-item label="尺寸">{{ robot.hw_dimensions || '-' }}</el-descriptions-item>
                <el-descriptions-item label="重量">{{ robot.hw_weight || '-' }} kg</el-descriptions-item>
                <el-descriptions-item label="载重">{{ robot.hw_payload || robot.max_payload }} kg</el-descriptions-item>
                <el-descriptions-item label="工作温度">{{ robot.hw_temp_range || '-' }}</el-descriptions-item>
                <el-descriptions-item label="充电方式">{{ robot.hw_charge_method || '-' }}</el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>

      <!-- 模块2: 地图 -->
      <el-tab-pane label="🗺️ 地图管理" name="map">
        <el-card class="glass-card">
          <p style="color:var(--el-text-color-secondary);margin-bottom:12px">当前绑定: {{ robotMap.map_name || '未绑定' }} {{ robotMap.is_auto ? '(自动定位)' : '(手动)' }}</p>
          <el-form :model="mapForm" inline>
            <el-form-item label="选择地图">
              <el-select v-model="mapForm.map_id" placeholder="选择地图" style="width:200px">
                <el-option v-for="m in maps" :key="m.id" :label="m.name" :value="m.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="定位方式">
              <el-switch v-model="mapForm.is_auto" active-text="自动" inactive-text="手动" />
            </el-form-item>
            <el-form-item><el-button type="primary" class="glass-btn" @click="bindMap">绑 定</el-button></el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 模块3: 定时任务 -->
      <el-tab-pane label="📋 任务管理" name="schedule">
        <el-card class="glass-card">
          <el-form :model="schedForm" inline>
            <el-form-item label="名称"><el-input v-model="schedForm.name" size="small" style="width:120px" /></el-form-item>
            <el-form-item label="Cron"><el-input v-model="schedForm.cron_expr" size="small" style="width:100px" /></el-form-item>
            <el-form-item label="类型">
              <el-select v-model="schedForm.task_type" size="small" style="width:100px">
                <el-option label="运输" value="transport" /><el-option label="巡逻" value="patrol" /><el-option label="充电" value="charge" />
              </el-select>
            </el-form-item>
            <el-form-item><el-button type="primary" class="glass-btn" size="small" @click="addSchedule">添加</el-button></el-form-item>
          </el-form>

          <el-table :data="schedules" stripe size="small" style="margin-top:12px">
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="cron_expr" label="Cron" width="120" />
            <el-table-column prop="task_type" label="类型" width="80" />
            <el-table-column label="开关" width="80">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" size="small" @change="toggleSchedule(row)" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="60">
              <template #default="{ row }"><el-button size="small" type="danger" @click="delSchedule(row.id)">删</el-button></template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 模块4: 工作记录 -->
      <el-tab-pane label="📜 工作记录" name="work">
        <el-table :data="workRecords" stripe size="small" empty-text="暂无记录">
          <el-table-column prop="task_code" label="任务编号" width="120" />
          <el-table-column prop="task_desc" label="描述" />
          <el-table-column prop="start_time" label="开始" width="170" />
          <el-table-column prop="end_time" label="结束" width="170" />
        </el-table>
      </el-tab-pane>

      <!-- 模块5: 日志 -->
      <el-tab-pane label="📝 日志" name="log">
        <el-radio-group v-model="logFilter" size="small" @change="loadOpLogs" style="margin-bottom:12px">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="manual">手动操作</el-radio-button>
          <el-radio-button value="remote">远程操作</el-radio-button>
          <el-radio-button value="alert">异常报警</el-radio-button>
        </el-radio-group>
        <el-table :data="opLogs" stripe size="small" empty-text="暂无日志">
          <el-table-column label="类型" width="90">
            <template #default="{ row }"><el-tag size="small" effect="dark">{{ row.log_type }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="operator" label="操作人" width="100" />
          <el-table-column prop="action" label="操作" />
          <el-table-column prop="created_at" label="时间" width="170" />
        </el-table>
      </el-tab-pane>

      <!-- 模块6: 重定位 -->
      <el-tab-pane label="📍 重定位" name="reloc">
        <el-card class="glass-card">
          <el-form :model="relocForm" label-width="80px">
            <el-form-item label="模式">
              <el-radio-group v-model="relocForm.mode">
                <el-radio value="auto">自动寻位</el-radio>
                <el-radio value="manual">手动校准</el-radio>
              </el-radio-group>
            </el-form-item>
            <template v-if="relocForm.mode === 'manual'">
              <el-form-item label="X坐标"><el-input-number v-model="relocForm.pos_x" /></el-form-item>
              <el-form-item label="Y坐标"><el-input-number v-model="relocForm.pos_y" /></el-form-item>
              <el-form-item label="角度"><el-input-number v-model="relocForm.theta" :min="0" :max="360" /></el-form-item>
            </template>
            <el-form-item><el-button type="primary" class="glass-btn" @click="relocate">执行重定位</el-button></el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 模块7: 其他设置 -->
      <el-tab-pane label="⚙️ 其他设置" name="settings">
        <el-card class="glass-card">
          <el-form :model="settingsForm" label-width="140px">
            <el-form-item label="语音音量"><el-slider v-model="settingsForm.voice_volume" :min="0" :max="100" show-input /></el-form-item>
            <el-form-item label="最大速度"><el-input-number v-model="settingsForm.max_speed_setting" :min="0" :step="0.1" /> m/s</el-form-item>
            <el-form-item label="电量报警阈值"><el-slider v-model="settingsForm.battery_alert_pct" :min="5" :max="50" show-input /> %</el-form-item>
            <el-form-item label="自动回充阈值"><el-slider v-model="settingsForm.auto_charge_pct" :min="5" :max="30" show-input /> %</el-form-item>
            <el-form-item label="锁屏密码"><el-input v-model="settingsForm.screen_lock_pwd" type="password" show-password style="width:200px" /></el-form-item>
            <el-form-item label="锁屏"><el-switch v-model="settingsForm.locked" /></el-form-item>
            <el-form-item><el-button type="primary" class="glass-btn" @click="saveSettings">保存设置</el-button></el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.robot-center { padding: 0; }
</style>
