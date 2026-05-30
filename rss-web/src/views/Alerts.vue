<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAlerts, ackAlert, resolveAlert } from '@/api'

const alerts = ref<any[]>([])
const loading = ref(false)

async function fetch() {
  loading.value = true
  try {
    const { data } = await getAlerts()
    alerts.value = data.data?.list || data.data || []
  } catch(e) { /* ignore */ }
  loading.value = false
}

onMounted(fetch)
</script>

<template>
  <div class="alerts-page">
    <el-card class="glass-card">
      <template #header>
        <span>🚨 告警中心</span>
        <el-tag size="small" style="margin-left:8px;" effect="dark">{{ alerts.length }} 条</el-tag>
      </template>
      <el-table :data="alerts" v-loading="loading" stripe>
        <el-table-column label="级别" width="90">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'critical' ? 'danger' : row.severity === 'warning' ? 'warning' : 'info'" size="small" effect="dark">
              {{ row.severity }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="dark">{{ { battery_low: '电量低', task_timeout: '任务超时', comm_lost: '通信中断', fault: '故障', collision: '碰撞' }[row.alert_type] || row.alert_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column prop="content" label="内容" min-width="250" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'unack' ? 'danger' : row.status === 'acknowledged' ? 'warning' : 'success'" size="small" effect="dark">
              {{ { unack: '未处理', acknowledged: '已确认', resolved: '已解决' }[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="170" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'unack'" size="small" class="glass-btn" @click="ackAlert(row.id).then(fetch)">确认</el-button>
            <el-button v-if="row.status !== 'resolved'" size="small" type="success" @click="resolveAlert(row.id).then(fetch)">解决</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && !alerts.length" description="暂无告警 🎉" />
    </el-card>
  </div>
</template>
