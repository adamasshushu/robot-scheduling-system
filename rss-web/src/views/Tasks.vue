<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { getTasks, createTask, assignTask, cancelTask, updateTask, deleteTask } from '@/api'
import { getRobots } from '@/api'

const tasks = ref<any[]>([])
const robots = ref<any[]>([])
const loading = ref(false)
const filters = reactive({ status: '', robot_id: '' })
const dialogVisible = ref(false)
const editingTask = ref<any>(null)
const editDialogVisible = ref(false)
const editForm = reactive({ task_type: 'transport', target_location: '', description: '', priority: 5 })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const newTask = ref({
  task_type: 'transport',
  target_location: '',
  description: '',
  priority: 5,
  robot_id: null as number | null,
})

async function fetch() {
  loading.value = true
  try {
    const { data } = await getTasks({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      robot_id: filters.robot_id || undefined,
    })
    tasks.value = data.data?.list || []
    pagination.total = data.data?.pagination?.total || 0
  } finally { loading.value = false }
}

async function handleCreate() {
  const resp = await createTask(newTask.value)
  if (newTask.value.robot_id) {
    const taskId = resp.data?.data?.id
    if (taskId) await assignTask(taskId, newTask.value.robot_id, false)
  }
  ElMessage.success('任务已创建')
  dialogVisible.value = false
  newTask.value = { task_type: 'transport', target_location: '', description: '', priority: 5, robot_id: null }
  fetch()
}

async function handleAssign(task: any) {
  await assignTask(task.id, undefined, true)
  ElMessage.success('智能指派成功')
  fetch()
}

function handleEdit(task: any) {
  editingTask.value = task
  Object.assign(editForm, { task_type: task.task_type, target_location: task.target_location, description: task.description, priority: task.priority })
  editDialogVisible.value = true
}
async function handleSaveEdit() {
  await updateTask(editingTask.value.id, editForm)
  ElMessage.success('任务已更新')
  editDialogVisible.value = false
  fetch()
}
async function handleDeleteTask(task: any) {
  await ElMessageBox.confirm(`确定删除任务 ${task.task_code}？`, '警告', { type: 'warning' })
  await deleteTask(task.id)
  ElMessage.success('已删除')
  fetch()
}

onMounted(async () => {
  await fetch()
  const { data } = await getRobots()
  robots.value = data.data?.list || []
})
</script>

<template>
  <div class="tasks-page">
    <!-- 筛选 + 创建 -->
    <el-card class="filter-bar glass-panel">
      <el-row justify="space-between" align="middle">
        <el-form :model="filters" inline>
          <el-form-item label="状态">
            <el-select v-model="filters.status" clearable placeholder="全部" @change="fetch">
              <el-option label="待执行" value="pending" />
              <el-option label="已指派" value="assigned" />
              <el-option label="执行中" value="running" />
              <el-option label="已完成" value="completed" />
            </el-select>
          </el-form-item>
        </el-form>
        <el-button type="primary" class="glass-btn" @click="dialogVisible = true">+ 创建任务</el-button>
      </el-row>
    </el-card>

    <!-- 任务列表 -->
    <el-card class="glass-card" style="margin-top: 16px;">
      <el-table :data="tasks" v-loading="loading" stripe>
        <el-table-column prop="task_code" label="任务编号" width="180" />
        <el-table-column prop="task_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" effect="dark">{{ { transport: '搬运', patrol: '巡逻', charge: '充电' }[row.task_type] || row.task_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="优先级" width="80">
          <template #default="{ row }">
            <el-tag :type="row.priority <= 3 ? 'danger' : row.priority <= 6 ? 'warning' : ''" size="small" effect="dark">
              {{ '🔥'.repeat(Math.min(row.priority, 3)) || row.priority }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'completed' ? 'success' : row.status === 'running' ? '' : 'info'" size="small" effect="dark">
              {{ { pending: '待执行', assigned: '已指派', running: '执行中', completed: '已完成', cancelled: '已取消', failed: '失败' }[row.status] || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="机器人" width="120">
          <template #default="{ row }">
            {{ row.robot?.name || '未指派' }}
          </template>
        </el-table-column>
        <el-table-column prop="target_location" label="目标位置" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="进度" width="120">
          <template #default="{ row }">
            <el-progress :percentage="row.progress_pct" :stroke-width="6" :color="row.progress_pct === 100 ? '#67c23a' : '#6C5CE7'" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'pending'" size="small" class="glass-btn" @click="handleAssign(row)">智能指派</el-button>
            <el-button v-if="['pending','assigned','running'].includes(row.status)" size="small" type="danger" @click="cancelTask(row.id).then(fetch)">取消</el-button>
            <el-button size="small" class="glass-btn" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDeleteTask(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-pagination
      v-if="pagination.total > pagination.page_size"
      v-model:current-page="pagination.page"
      :total="pagination.total"
      :page-size="pagination.page_size"
      layout="total, prev, pager, next"
      @change="fetch"
      style="margin-top: 16px; justify-content: center;"
    />

    <!-- 创建任务弹窗 -->
    <el-dialog v-model="dialogVisible" title="创建任务" width="500px">
      <el-form :model="newTask" label-width="80px">
        <el-form-item label="类型">
          <el-select v-model="newTask.task_type" style="width: 100%;">
            <el-option label="搬运" value="transport" />
            <el-option label="巡逻" value="patrol" />
            <el-option label="充电" value="charge" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标位置">
          <el-input v-model="newTask.target_location" placeholder="如: B区-装配线" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-slider v-model="newTask.priority" :min="1" :max="10" show-input />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newTask.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="机器人">
          <el-select v-model="newTask.robot_id" clearable placeholder="可选(留空=不指派)" style="width:100%">
            <el-option v-for="r in robots" :key="r.id" :label="`${r.name} (${r.status})`" :value="r.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" class="glass-btn" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 编辑任务弹窗 -->
    <el-dialog v-model="editDialogVisible" title="编辑任务" width="500px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="类型">
          <el-select v-model="editForm.task_type" style="width:100%">
            <el-option label="搬运" value="transport" />
            <el-option label="巡逻" value="patrol" />
            <el-option label="充电" value="charge" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标位置"><el-input v-model="editForm.target_location" /></el-form-item>
        <el-form-item label="优先级"><el-slider v-model="editForm.priority" :min="1" :max="10" show-input /></el-form-item>
        <el-form-item label="描述"><el-input v-model="editForm.description" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" class="glass-btn" @click="handleSaveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
