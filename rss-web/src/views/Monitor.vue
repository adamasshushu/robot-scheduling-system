<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { getMapRobots, getActiveMap, uploadMap, getMaps, setActiveMap, deleteMap, calibrateMap } from '@/api'
import { useWebSocket } from '@/composables/useWebSocket'
import L from 'leaflet'
import * as THREE from 'three'

const mapMode = ref<'2d' | '3d'>('2d')
const mapRef = ref<HTMLDivElement>()
const threeRef = ref<HTMLDivElement>()

// 地图数据
const mapConfig = ref<any>(null)
const maps = ref<any[]>([])
const robots = ref<any[]>([])

// 上传
const uploadDialog = ref(false)
const mapName = ref('')
const uploading = ref(false)

// WebSocket 机器人位置
const { robotUpdates } = useWebSocket()

// ── 2D Leaflet ──
let leafletMap: L.Map | null = null
let imageOverlay: L.ImageOverlay | null = null
let robotMarkers: L.Marker[] = []

function initLeaflet(mapData: any) {
  if (!mapRef.value) return

  // 默认 Dublin 坐标
  const center: [number, number] = [53.3498, -6.2603]

  if (leafletMap) {
    leafletMap.remove()
    robotMarkers = []
  }

  leafletMap = L.map(mapRef.value, {
    center,
    zoom: 5,
    attributionControl: false,
  })

  L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
    maxZoom: 19,
  }).addTo(leafletMap)

  // 有校准数据时叠加平面图
  if (mapData?.image_url && mapData.calib_point1_x) {
    const bounds: L.LatLngBoundsExpression = [[0, 0], [mapData.real_height || 10, mapData.real_width || 10]]
    imageOverlay = L.imageOverlay(mapData.image_url, bounds, { opacity: 0.85 })
    imageOverlay.addTo(leafletMap)
    leafletMap.fitBounds(bounds)
  }

  updateRobotMarkers()
}

function updateRobotMarkers() {
  if (!leafletMap) return
  robotMarkers.forEach(m => m.remove())
  robotMarkers = []

  const merged = robots.value.map(r => {
    const update = robotUpdates.value[r.id]
    return update ? { ...r, ...update } : r
  })

  merged.forEach(r => {
    if (r.location_x == null) return
    const color = r.status === 'running' ? '#67c23a' : r.status === 'fault' ? '#f56c6c' : '#409eff'
    const icon = L.divIcon({
      className: 'robot-marker',
      html: `<div style="
        width:14px;height:14px;border-radius:50%;background:${color};
        border:2px solid #fff;box-shadow:0 0 8px ${color};transform:rotate(${r.location_theta || 0}deg)
      "><div style="width:4px;height:4px;background:#fff;border-radius:50%;margin:2px auto 0"></div></div>`,
      iconSize: [18, 18],
      iconAnchor: [9, 9],
    })
    const marker = L.marker([r.location_y, r.location_x], { icon })
      .bindPopup(`<b>${r.name}</b><br>电量: ${r.battery_pct}%<br>状态: ${r.status}`)
      .addTo(leafletMap!)
    robotMarkers.push(marker)
  })
}

// ── 3D Three.js ──
let threeScene: any = null
let threeCamera: any = null
let threeRenderer: any = null
let threeRobotMeshes: Map<number, any> = new Map()
let animFrame: number | null = null

function initThreeJS(mapData: any) {
  if (!threeRef.value) return
  cleanupThree()

  try {
    const w = threeRef.value.clientWidth
    const h = threeRef.value.clientHeight

  threeScene = new THREE.Scene()
  threeScene.background = new THREE.Color(0x1a1a2e)

  threeCamera = new THREE.PerspectiveCamera(60, w / h, 0.5, 1000)
  threeCamera.position.set(15, 15, 15)
  threeCamera.lookAt(5, 0, 5)

  threeRenderer = new THREE.WebGLRenderer({ antialias: true })
  threeRenderer.setSize(w, h)
  threeRenderer.shadowMap.enabled = true
  threeRef.value.appendChild(threeRenderer.domElement)

  // 鼠标旋转

  // 网格地面
  const grid = new THREE.GridHelper(20, 20, 0x444466, 0x222244)
  threeScene.add(grid)

  // 光照
  const ambient = new THREE.AmbientLight(0x404060, 1)
  threeScene.add(ambient)
  const dir = new THREE.DirectionalLight(0xffffff, 0.8)
  dir.position.set(10, 20, 5)
  threeScene.add(dir)

  // 平面图纹理
  if (mapData?.image_url) {
    const loader = new THREE.TextureLoader()
    loader.load(mapData.image_url, (texture) => {
      const planeW = mapData.real_width || 10
      const planeH = mapData.real_height || 10
      const geometry = new THREE.PlaneGeometry(planeW, planeH)
      const material = new THREE.MeshStandardMaterial({ map: texture, side: THREE.DoubleSide, transparent: true, opacity: 0.8 })
      const plane = new THREE.Mesh(geometry, material)
      plane.rotation.x = -Math.PI / 2
      plane.position.set(planeW / 2, 0.01, planeH / 2)
      threeScene!.add(plane)
    })
  }

  createRobotModels()
  animateThree()
  } catch (err) {
    console.error('Three.js init failed:', err)
  }
}

function createRobotModels() {
  if (!threeScene || !mapConfig.value) return
  threeRobotMeshes.forEach(m => threeScene!.remove(m))
  threeRobotMeshes.clear()

  const geo = new THREE.BoxGeometry(0.3, 0.3, 0.3)
  const mat = new THREE.MeshStandardMaterial({ color: 0x67c23a, emissive: 0x224422 })

  robots.value.forEach(r => {
    if (r.location_x == null) return
    const mesh = new THREE.Mesh(geo, mat.clone())
    mesh.position.set(r.location_x, 0.3, r.location_y)
    mesh.castShadow = true

    // 方向指示
    if (r.location_theta) {
      const arrowGeo = new THREE.ConeGeometry(0.1, 0.2, 4)
      const arrowMat = new THREE.MeshStandardMaterial({ color: 0xff0000 })
      const arrow = new THREE.Mesh(arrowGeo, arrowMat)
      arrow.position.set(0, 0.25, 0)
      arrow.rotation.z = -r.location_theta * Math.PI / 180
      mesh.add(arrow)
    }

    threeScene!.add(mesh)
    threeRobotMeshes.set(r.id, mesh)
  })
}

function animateThree() {
  animFrame = requestAnimationFrame(animateThree)
  threeRenderer?.render(threeScene!, threeCamera!)

  // 实时更新机器人位置
  if (Object.keys(robotUpdates.value).length > 0) {
    threeRobotMeshes.forEach((mesh, id) => {
      const update = robotUpdates.value[id]
      if (update?.location_x != null) {
        mesh.position.set(update.location_x, 0.3, update.location_y)
      }
    })
  }
}

function cleanupThree() {
  if (animFrame) cancelAnimationFrame(animFrame)
  threeRenderer?.dispose()
  if (threeRef.value) threeRef.value.innerHTML = ''
  threeScene = null; threeCamera = null; threeRenderer = null
  threeRobotMeshes.clear()
}

// ── 上传 ──
async function handleUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) return
    uploading.value = true
    const formData = new FormData()
    formData.append('file', file)
    if (mapName.value) formData.append('name', mapName.value)
    await uploadMap(formData)
    ElMessage.success('地图上传成功')
    uploadDialog.value = false
    uploading.value = false
    loadMapData()
  }
  input.click()
}

// ── 数据加载 ──
async function loadMapData() {
  try {
    const [mapRes, mapsRes, robotsRes] = await Promise.all([
      getActiveMap(),
      getMaps(),
      getMapRobots(),
    ])
    mapConfig.value = mapRes.data.data
    maps.value = mapsRes.data.data || []
    robots.value = robotsRes.data.data || []
    if (mapMode.value === '2d') initLeaflet(mapConfig.value)
    else initThreeJS(mapConfig.value)
  } catch (e) { /* ignore */ }
}

// ── 切换地图 ──
watch(mapMode, (mode) => {
  if (mode === '2d') {
    cleanupThree()
    nextTick(() => initLeaflet(mapConfig.value))
  } else {
    if (leafletMap) { leafletMap.remove(); leafletMap = null }
    nextTick(() => initThreeJS(mapConfig.value))
  }
})

// WebSocket 位置更新
watch(robotUpdates, () => {
  updateRobotMarkers()
}, { deep: true })

onMounted(loadMapData)
onUnmounted(() => {
  if (leafletMap) leafletMap.remove()
  cleanupThree()
})
</script>

<template>
  <div class="monitor-page">
    <el-row :gutter="16">
      <el-col :span="18">
        <el-card class="glass-card">
          <template #header>
            <div class="card-header">
              <span>🗺️ 地图监控</span>
              <div style="display:flex;gap:8px;">
                <el-radio-group size="small" v-model="mapMode">
                  <el-radio-button value="2d">2D 平面</el-radio-button>
                  <el-radio-button value="3d">3D 场景</el-radio-button>
                </el-radio-group>
                <el-button size="small" class="glass-btn" @click="uploadDialog = true">📤 上传地图</el-button>
              </div>
            </div>
          </template>
          <div v-show="mapMode === '2d'" ref="mapRef" style="height:600px;border-radius:8px;"></div>
          <div v-show="mapMode === '3d'" ref="threeRef" style="height:600px;border-radius:8px;overflow:hidden;"></div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <!-- 在线机器人 -->
        <el-card class="glass-card" style="margin-bottom:12px;">
          <template #header>📍 在线机器人</template>
          <div v-for="r in robots" :key="r.id" class="robot-dot">
            <span class="dot" :style="{ background: r.status === 'running' ? '#67c23a' : r.status === 'fault' ? '#f56c6c' : '#909399' }"></span>
            <span class="r-name">{{ r.name }}</span>
            <el-tag size="small" effect="dark">{{ r.status }}</el-tag>
          </div>
          <el-empty v-if="!robots.length" description="无在线机器人" :image-size="60" />
        </el-card>

        <!-- 地图列表 -->
        <el-card class="glass-card">
          <template #header>📁 已导入地图</template>
          <div v-for="m in maps" :key="m.id" class="map-item" :class="{ active: m.active }">
            <span class="map-name">{{ m.name }}</span>
            <el-button v-if="!m.active" size="small" @click="setActiveMap(m.id).then(loadMapData)">启用</el-button>
            <el-tag v-else size="small" type="success" effect="dark">当前</el-tag>
            <el-button size="small" type="danger" @click="deleteMap(m.id).then(loadMapData)">删除</el-button>
          </div>
          <el-empty v-if="!maps.length" description="暂无地图" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 上传弹窗 -->
    <el-dialog v-model="uploadDialog" title="上传工厂平面图" width="400px">
      <el-form>
        <el-form-item label="名称"><el-input v-model="mapName" placeholder="如：A车间-1F" /></el-form-item>
        <el-form-item>
          <el-button class="glass-btn" :loading="uploading" @click="handleUpload" style="width:100%">
            {{ uploading ? '上传中...' : '选择文件并上传' }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-dialog>
  </div>
</template>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.robot-dot { display: flex; align-items: center; gap: 8px; padding: 10px 0; border-bottom: 1px solid var(--el-border-color-lighter); }
.robot-dot:last-child { border-bottom: none; }
.dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; box-shadow: 0 0 6px currentColor; }
.r-name { flex: 1; font-size: 13px; font-weight: 500; }
.map-item { display: flex; align-items: center; gap: 6px; padding: 8px 0; border-bottom: 1px solid var(--el-border-color-lighter); }
.map-item:last-child { border-bottom: none; }
.map-item.active { background: rgba(108,92,231,0.08); margin: 0 -12px; padding: 8px 12px; border-radius: 6px; }
.map-name { flex: 1; font-size: 13px; }
</style>
