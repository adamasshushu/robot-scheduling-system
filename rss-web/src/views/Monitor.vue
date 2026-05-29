<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getMapRobots } from '@/api'

const robots = ref<any[]>([])

onMounted(async () => {
  try {
    const { data } = await getMapRobots()
    robots.value = data.data || []
  } catch { /* use robot list fallback */ }
})
</script>

<template>
  <div class="monitor-page">
    <el-row :gutter="16">
      <el-col :span="18">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>🗺️ 地图监控</span>
              <el-radio-group size="small" v-model="mapMode">
                <el-radio-button value="2d">2D 平面</el-radio-button>
                <el-radio-button value="3d">3D 场景</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div id="map-container" style="height: 600px; background: #1a1a2e; border-radius: 8px;
            display: flex; align-items: center; justify-content: center; color: #888;">
            <div style="text-align: center;">
              <p style="font-size: 48px; margin: 0;">🗺️</p>
              <p>Leaflet/Cesium 地图组件</p>
              <p style="font-size: 12px;">等待配置地图服务后加载</p>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card>
          <template #header>📍 在线机器人</template>
          <div v-for="r in robots" :key="r.id" class="robot-dot">
            <span class="dot" :style="{ background: r.status === 'running' ? '#67c23a' : r.status === 'fault' ? '#f56c6c' : '#909399' }"></span>
            <span class="r-name">{{ r.name }}</span>
            <el-tag size="small">{{ r.status }}</el-tag>
          </div>
          <el-empty v-if="!robots.length" description="无在线机器人" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script lang="ts">
const mapMode = ref('2d')
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.robot-dot { display: flex; align-items: center; gap: 8px; padding: 8px 0; border-bottom: 1px solid var(--el-border-color-lighter); }
.dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
.r-name { flex: 1; font-size: 13px; }
</style>
