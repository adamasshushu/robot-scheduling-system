<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getRobot, sendCommand } from '@/api'

const route = useRoute()
const robot = ref<any>({})
const lightForm = ref({ mode: 'off', color: 'white', brightness: 100 })

onMounted(async () => {
  const id = Number(route.params.id)
  const { data } = await getRobot(id)
  robot.value = data.data || {}
  lightForm.value = {
    mode: robot.value.light_mode || 'off',
    color: robot.value.light_color || 'white',
    brightness: robot.value.light_brightness || 100,
  }
})

async function handleLight() {
  await sendCommand(robot.value.id, 'light', lightForm.value)
  ElMessage.success('灯光指令已发送')
}
</script>

<template>
  <div class="robot-detail">
    <el-page-header @back="$router.back()" :content="robot.name || '机器人详情'" />

    <el-row :gutter="16" style="margin-top: 16px;">
      <!-- 左：基础信息 -->
      <el-col :span="8">
        <el-card class="glass-card">
          <template #header>📋 基本信息</template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="编号">{{ robot.robot_code }}</el-descriptions-item>
            <el-descriptions-item label="型号">{{ robot.model }}</el-descriptions-item>
            <el-descriptions-item label="固件">{{ robot.firmware_ver || '-' }}</el-descriptions-item>
            <el-descriptions-item label="IP">{{ robot.ip_address || '-' }}</el-descriptions-item>
            <el-descriptions-item label="最大载重">{{ robot.max_payload }} kg</el-descriptions-item>
            <el-descriptions-item label="最大速度">{{ robot.max_speed }} m/s</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 电量卡片 -->
        <el-card class="glass-card" style="margin-top: 16px;">
          <template #header>🔋 电量</template>
          <div style="text-align: center;">
            <el-progress type="dashboard" :percentage="robot.battery_pct || 0" :color="robot.battery_pct > 20 ? '#67c23a' : '#f56c6c'" />
            <p style="margin-top: 8px; font-size: 13px; color: var(--el-text-color-secondary);">
              状态: {{ { normal: '正常', low: '低', critical: '严重', charging: '充电中' }[robot.battery_status] || robot.battery_status }}
            </p>
          </div>
        </el-card>
      </el-col>

      <!-- 中：状态 + 控制 -->
      <el-col :span="8">
        <el-card class="glass-card">
          <template #header>📡 实时状态</template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="运行模式">
              <el-tag :type="robot.status === 'running' ? 'success' : 'info'" effect="dark">
                {{ robot.status }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="通信">
              <el-tag :type="robot.comm_status === 'online' ? 'success' : 'danger'" effect="dark">
                {{ robot.comm_status || 'unknown' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="位置">({{ robot.location_x }}, {{ robot.location_y }})</el-descriptions-item>
            <el-descriptions-item label="区域">{{ robot.map_zone }}</el-descriptions-item>
            <el-descriptions-item label="速度">{{ robot.speed }} m/s</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 控制按钮 -->
        <el-card class="glass-card" style="margin-top: 16px;">
          <template #header>🎮 快速控制</template>
          <el-space direction="vertical" style="width: 100%;">
            <el-button type="success" class="glass-btn" style="width: 100%" @click="sendCommand(robot.id, 'start')">▶ 启动</el-button>
            <el-button type="warning" style="width: 100%" @click="sendCommand(robot.id, 'pause')">⏸ 暂停</el-button>
            <el-button type="info" style="width: 100%" @click="sendCommand(robot.id, 'charge')">🔌 回充</el-button>
            <el-button type="danger" style="width: 100%" @click="sendCommand(robot.id, 'stop')">🛑 急停</el-button>
          </el-space>
        </el-card>
      </el-col>

      <!-- 右：灯光控制 -->
      <el-col :span="8">
        <el-card class="glass-card">
          <template #header>💡 灯光控制</template>
          <el-form :model="lightForm" label-width="80px">
            <el-form-item label="模式">
              <el-select v-model="lightForm.mode" style="width: 100%;">
                <el-option label="关闭" value="off" />
                <el-option label="常亮" value="on" />
                <el-option label="呼吸" value="breathe" />
                <el-option label="闪烁" value="flash" />
              </el-select>
            </el-form-item>
            <el-form-item label="颜色">
              <el-color-picker v-model="lightForm.color" show-alpha />
            </el-form-item>
            <el-form-item label="亮度">
              <el-slider v-model="lightForm.brightness" :min="0" :max="100" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" class="glass-btn" @click="handleLight" style="width: 100%;">发 送</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
