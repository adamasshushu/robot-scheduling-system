<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { getTasks, createTask, assignTask, cancelTask } from '@/api'
import { getRobots } from '@/api'

const tasks = ref<any[]>([])
const robots = ref<any[]>([])
const loading = ref(false)
const filters = reactive({ status: '', robot_id: '' })
const dialogVisible = ref(false)
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const newTask = ref({
  task_type: 'transport',
  target_location: '',
  description: '',
  priority: 5,
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
  await createTask(newTask.value)
  ElMessage.success('任务已创建')
  dialogVisible.value = false
  newTask.value = { task_type: 'transport', target_location: '', description: '', priority: 5 }
  fetch()
}

async function handleAssign(task: any) {
  await assignTask(task.id, undefined, true)
  ElMessage.success('智能指派成功')
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
    <el-card class="filter-bar">
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
        <el-button type="primary" @click="dialogVisible = true">+ 创建任务</el-button>
      </el-row>
    </el-card>

    <!-- 任务列表 -->
    <el-table :data="tasks" v-loading="loading" style="margin-top: 16px;" stripe>
      <el-table-column prop="task_code" label="任务编号" width="180" />
      <el-table-column prop="task_type" label="类型" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ { transport: '搬运', patrol: '巡逻', charge: '充电' }[row.task_type] || row.task_type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="优先级" width="80">
        <template #default="{ row }">
          <el-tag :type="row.priority <= 3 ? 'danger' : row.priority <= 6 ? 'warning' : ''" size="small">{{ row.priority }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'completed' ? 'success' : row.status === 'running' ? '' : 'info'" size="small">
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
          <el-progress :percentage="row.progress_pct" :stroke-width="6" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status === 'pending'" size="small" type="primary" @click="handleAssign(row)">智能指派</el-button>
          <el-button v-if="['pending','assigned','running'].includes(row.status)" size="small" type="danger" @click="cancelTask(row.id).then(fetch)">取消</el-button>
        </template>
      </el-table-column>
    </el-table>

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
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>
