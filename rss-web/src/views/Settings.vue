<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { getUsers, createUser, updateUser, deleteUser, getModels, createModel, deleteModel, getSystemConfig, updateSystemConfig, getFirmwares, uploadFirmware, toggleAutoUpgrade, upgradeRobot, getRobots } from '@/api'

// ── Tab ──
const activeTab = ref('users')

// ── 用户管理 ──
const users = ref<any[]>([])
const userDialog = ref(false)
const userForm = reactive({ username: '', password: '', role: 'operator' })
const editingUserId = ref<number | null>(null)

async function fetchUsers() {
  const { data } = await getUsers({ page_size: 100 })
  users.value = data.data?.list || []
}
async function handleSaveUser() {
  if (editingUserId.value) {
    await updateUser(editingUserId.value, userForm)
  } else {
    await createUser(userForm)
  }
  ElMessage.success(editingUserId.value ? '已更新' : '已创建')
  userDialog.value = false
  Object.assign(userForm, { username: '', password: '', role: 'operator' })
  editingUserId.value = null
  fetchUsers()
}
function handleEditUser(u: any) {
  editingUserId.value = u.id
  Object.assign(userForm, { username: u.username, password: '', role: u.role })
  userDialog.value = true
}
async function handleDeleteUser(id: number) {
  await ElMessageBox.confirm('确定删除该用户？', '警告', { type: 'warning' })
  await deleteUser(id)
  ElMessage.success('已删除')
  fetchUsers()
}

// ── 型号管理 ──
const models = ref<any[]>([])
const modelDialog = ref(false)
const modelForm = reactive({ name: '', maker: '', max_payload: 0, max_speed: 0, specs: '' })

async function fetchModels() {
  const { data } = await getModels()
  models.value = data.data || []
}
async function handleSaveModel() {
  await createModel(modelForm)
  ElMessage.success('型号已创建')
  modelDialog.value = false
  Object.assign(modelForm, { name: '', maker: '', max_payload: 0, max_speed: 0, specs: '' })
  fetchModels()
}
async function handleDeleteModel(id: number) {
  await ElMessageBox.confirm('确定删除？', '警告', { type: 'warning' })
  await deleteModel(id)
  fetchModels()
}

// ── 系统配置 ──
const sysConfig = reactive({ site_name: 'RSS 机器人调度系统', site_logo: '🤖' })
async function fetchSysConfig() {
  try {
    const { data } = await getSystemConfig()
    const cfg = data.data || {}
    if (cfg.site_name) sysConfig.site_name = cfg.site_name
    if (cfg.site_logo) sysConfig.site_logo = cfg.site_logo
  } catch(e) { /* */ }
}
async function saveSysConfig() {
  await updateSystemConfig({ site_name: sysConfig.site_name, site_logo: sysConfig.site_logo })
  ElMessage.success('系统配置已保存')
}

// ── OTA 固件管理 ──
const firmwares = ref<any[]>([])
const robots = ref<any[]>([])
const otaVisible = ref(false)
const otaForm = reactive({ module: 'firmware', robot_id: null as number | null, mode: 'manual' })

async function fetchFirmwares() {
  const { data } = await getFirmwares()
  firmwares.value = data.data || []
}
async function fetchRobotsForOTA() {
  const { data } = await getRobots()
  robots.value = data.data?.list || []
}
async function handleOTAFile() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.bin,.gz,.zip,.tar,.img'
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) return
    const formData = new FormData()
    formData.append('file', file)
    formData.append('module', otaForm.module)
    formData.append('version', prompt('输入版本号 (如 v1.2.0)') || 'v1.0.0')
    formData.append('change_log', prompt('更新日志') || '')
    formData.append('auto_upgrade', confirm('自动推送升级?') ? 'true' : 'false')
    await uploadFirmware(formData)
    ElMessage.success('固件上传成功')
    fetchFirmwares()
  }
  input.click()
}
async function handleUpgrade(fw: any) {
  if (!otaForm.robot_id) { ElMessage.warning('请先选择目标机器人'); return }
  await ElMessageBox.confirm(
    `确认向机器人在线升级 ${fw.module} 到 ${fw.version}？`,
    'OTA升级确认',
    { confirmButtonText: '确认升级', cancelButtonText: '取消', type: 'warning' }
  )
  await upgradeRobot(otaForm.robot_id, fw.id, otaForm.mode)
  ElMessage.success('升级指令已下发')
}
async function handleToggleAuto(fw: any) {
  await toggleAutoUpgrade(fw.id, !fw.auto_upgrade)
  fetchFirmwares()
}

function handleLogoUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.onchange = () => {
    const file = input.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => { sysConfig.site_logo = reader.result as string }
    reader.readAsDataURL(file)
  }
  input.click()
}

onMounted(() => { fetchUsers(); fetchModels(); fetchSysConfig(); fetchFirmwares(); fetchRobotsForOTA() })
</script>

<template>
  <div class="settings-page">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="👥 用户管理" name="users">
        <el-card class="glass-card">
          <template #header>
            <div class="card-header-split">
              <span>用户列表</span>
              <el-button size="small" type="primary" class="glass-btn" @click="userDialog = true; editingUserId = null; Object.assign(userForm, { username: '', password: '', role: 'operator' })">+ 添加用户</el-button>
            </div>
          </template>
          <el-table :data="users" stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="username" label="用户名" min-width="120" />
            <el-table-column label="角色" width="100">
              <template #default="{ row }">
                <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small" effect="dark">
                  {{ { admin: '管理员', operator: '操作员', viewer: '观察者' }[row.role] || row.role }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="handleEditUser(row)">编辑</el-button>
                <el-button v-if="row.id !== 1" size="small" type="danger" @click="handleDeleteUser(row.id)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="🤖 机器人型号" name="models">
        <el-card class="glass-card">
          <template #header>
            <div class="card-header-split">
              <span>型号列表</span>
              <el-button size="small" type="primary" class="glass-btn" @click="modelDialog = true">+ 添加型号</el-button>
            </div>
          </template>
          <el-table :data="models" stripe>
            <el-table-column prop="name" label="型号" />
            <el-table-column prop="maker" label="厂商" />
            <el-table-column prop="max_payload" label="载重(kg)" width="90" />
            <el-table-column prop="max_speed" label="速度(m/s)" width="90" />
            <el-table-column label="操作" width="70" fixed="right">
              <template #default="{ row }"><el-button size="small" type="danger" @click="handleDeleteModel(row.id)">删除</el-button></template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!models.length" description="暂无型号，点击添加" :image-size="60" />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="⚙️ 系统配置" name="config">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-card class="glass-card" header="🎨 品牌设置">
              <el-form :model="sysConfig" label-width="80px">
                <el-form-item label="系统名称">
                  <el-input v-model="sysConfig.site_name" placeholder="RSS 机器人调度系统" />
                </el-form-item>
                <el-form-item label="Logo">
                  <div style="display:flex;align-items:center;gap:8px">
                    <img v-if="sysConfig.site_logo && sysConfig.site_logo.startsWith('data:')"
                         :src="sysConfig.site_logo" style="width:40px;height:40px;border-radius:6px;object-fit:contain" />
                    <span v-else style="font-size:32px">{{ sysConfig.site_logo || '🤖' }}</span>
                    <el-button size="small" @click="handleLogoUpload">上传图片</el-button>
                  </div>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" class="glass-btn" @click="saveSysConfig">保存配置</el-button>
                </el-form-item>
              </el-form>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card class="glass-card" header="📡 OTA 远程升级">
              <el-form label-width="80px">
                <el-form-item label="目标机器人">
                  <el-select v-model="otaForm.robot_id" clearable placeholder="选择机器人" style="width:100%">
                    <el-option v-for="r in robots" :key="r.id" :label="`${r.name} (${r.robot_code})`" :value="r.id" />
                  </el-select>
                </el-form-item>
                <el-form-item label="升级模式">
                  <el-radio-group v-model="otaForm.mode">
                    <el-radio value="manual">手动</el-radio>
                    <el-radio value="auto">自动</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item label="模块">
                  <el-select v-model="otaForm.module" style="width:120px">
                    <el-option label="下位机" value="firmware" />
                    <el-option label="前端" value="frontend" />
                    <el-option label="后端" value="backend" />
                    <el-option label="导航" value="nav" />
                  </el-select>
                </el-form-item>
              </el-form>
            </el-card>
          </el-col>
        </el-row>

        <el-card class="glass-card" style="margin-top:12px" header="📦 固件版本">
          <div style="margin-bottom:12px">
            <el-button size="small" type="primary" class="glass-btn" @click="handleOTAFile">+ 上传固件</el-button>
          </div>
          <el-table :data="firmwares" stripe size="small">
            <el-table-column prop="module" label="模块" width="90" />
            <el-table-column prop="version" label="版本" width="100" />
            <el-table-column prop="file_name" label="文件名" />
            <el-table-column prop="file_size" label="大小" width="80">
              <template #default="{ row }">{{ (row.file_size / 1024).toFixed(0) }} KB</template>
            </el-table-column>
            <el-table-column label="自动" width="70">
              <template #default="{ row }">
                <el-switch v-model="row.auto_upgrade" size="small" @change="handleToggleAuto(row)" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button size="small" type="success" @click="handleUpgrade(row)">升级</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!firmwares.length" description="暂无固件" :image-size="60" />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 用户弹窗 -->
    <el-dialog v-model="userDialog" :title="editingUserId ? '编辑用户' : '添加用户'" width="420px">
      <el-form :model="userForm" label-width="80px">
        <el-form-item label="用户名"><el-input v-model="userForm.username" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="userForm.password" type="password" show-password :placeholder="editingUserId ? '留空不修改' : ''" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="userForm.role" style="width:100%">
            <el-option label="管理员" value="admin" />
            <el-option label="操作员" value="operator" />
            <el-option label="观察者" value="viewer" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="userDialog = false">取消</el-button>
        <el-button type="primary" class="glass-btn" @click="handleSaveUser">保存</el-button>
      </template>
    </el-dialog>

    <!-- 型号弹窗 -->
    <el-dialog v-model="modelDialog" title="添加型号" width="420px">
      <el-form :model="modelForm" label-width="80px">
        <el-form-item label="型号名"><el-input v-model="modelForm.name" /></el-form-item>
        <el-form-item label="厂商"><el-input v-model="modelForm.maker" /></el-form-item>
        <el-form-item label="最大载重"><el-input-number v-model="modelForm.max_payload" :min="0" :step="10" /> kg</el-form-item>
        <el-form-item label="最大速度"><el-input-number v-model="modelForm.max_speed" :min="0" :precision="1" :step="0.5" /> m/s</el-form-item>
        <el-form-item label="规格"><el-input v-model="modelForm.specs" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modelDialog = false">取消</el-button>
        <el-button type="primary" class="glass-btn" @click="handleSaveModel">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.card-header-split { display: flex; justify-content: space-between; align-items: center; }
</style>
