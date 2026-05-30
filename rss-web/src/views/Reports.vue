<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import * as echarts from 'echarts'
import { getTaskStats, getBatteryTrend, getTaskTypes, getRobotUsage, getRobots } from '@/api'

const taskChart = ref<HTMLDivElement>()
const batteryChart = ref<HTMLDivElement>()
const typeChart = ref<HTMLDivElement>()
const usageChart = ref<HTMLDivElement>()

const robots = ref<any[]>([])
const filters = ref({ robot_id: '' })

let chartInstances: echarts.ECharts[] = []

async function loadCharts() {
  // 销毁旧图表
  chartInstances.forEach(c => c.dispose())
  chartInstances = []

  const robotFilter = filters.value.robot_id ? `&robot_id=${filters.value.robot_id}` : ''

  // 任务趋势
  const taskRes = await getTaskStats(14)
  const taskData = taskRes.data.data || []
  if (taskChart.value) {
    const c = echarts.init(taskChart.value); chartInstances.push(c)
    c.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['总任务', '已完成'], textStyle: { color: '#aaa' } },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: taskData.map((d: any) => d.date.slice(5)), axisLabel: { color: '#999' } },
      yAxis: { type: 'value', axisLabel: { color: '#999' } },
      series: [
        { name: '总任务', type: 'bar', data: taskData.map((d: any) => d.total), itemStyle: { borderRadius: [6,6,0,0], color: new echarts.graphic.LinearGradient(0,0,0,1,[{offset:0,color:'#6C5CE7'},{offset:1,color:'#a29bfe'}]) } },
        { name: '已完成', type: 'line', data: taskData.map((d: any) => d.completed), smooth: true, lineStyle: { color: '#67c23a', width: 2 }, itemStyle: { color: '#67c23a' }, symbol: 'circle', symbolSize: 6 },
      ],
    })
  }

  // 电量趋势
  const batRes = await getBatteryTrend(24)
  const batData = batRes.data.data || []
  const robotNames = [...new Set(batData.map((d: any) => d.robot_name))]
  if (batteryChart.value) {
    const bc = echarts.init(batteryChart.value); chartInstances.push(bc)
    bc.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: robotNames as string[], textStyle: { color: '#aaa' } },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: [...new Set(batData.map((d: any) => d.time?.slice(11,16)))].slice(0,50), axisLabel: { color: '#999' } },
      yAxis: { type: 'value', max: 100, axisLabel: { color: '#999', formatter: '{value}%' } },
      series: robotNames.map(name => ({
        name, type: 'line', smooth: true, symbol: 'none', lineStyle: { width: 2 },
        data: batData.filter((d: any) => d.robot_name === name).map((d: any) => d.battery_pct),
      })),
    })
  }

  // 任务类型分布
  const typeRes = await getTaskTypes()
  const typeData = typeRes.data.data || []
  const typeLabels: Record<string, string> = { transport: '搬运', patrol: '巡逻', charge: '充电' }
  if (typeChart.value) {
    const tc = echarts.init(typeChart.value); chartInstances.push(tc)
    tc.setOption({
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie', radius: ['50%', '75%'], center: ['50%', '50%'], label: { color: '#999' },
        data: typeData.map((d: any) => ({ name: typeLabels[d.task_type] || d.task_type, value: d.count })),
        emphasis: { itemStyle: { shadowBlur: 20, shadowColor: 'rgba(0,0,0,0.3)' } },
      }],
    })
  }

  // 机器人利用率
  const usageRes = await getRobotUsage()
  const usageData = usageRes.data.data || []
  if (usageChart.value) {
    const uc = echarts.init(usageChart.value); chartInstances.push(uc)
    uc.setOption({
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'value', axisLabel: { color: '#999' } },
      yAxis: { type: 'category', data: usageData.map((d: any) => d.robot_name).reverse(), axisLabel: { color: '#999' } },
      series: [{
        type: 'bar',
        data: usageData.map((d: any) => d.task_count).reverse(),
        itemStyle: { borderRadius: [0,6,6,0], color: new echarts.graphic.LinearGradient(0,0,1,0,[{offset:0,color:'#6C5CE7'},{offset:1,color:'#a29bfe'}]) },
        label: { show: true, position: 'right', color: '#999' },
      }],
    })
  }
}

onMounted(async () => {
  const { data } = await getRobots()
  robots.value = data.data?.list || []
  loadCharts()
})

watch(() => filters.value.robot_id, loadCharts)
</script>

<template>
  <div class="reports-page">
    <el-card class="glass-card" style="margin-bottom:12px">
      <el-form inline>
        <el-form-item label="筛选机器人">
          <el-select v-model="filters.robot_id" clearable placeholder="全部机器人" style="width:200px">
            <el-option v-for="r in robots" :key="r.id" :label="`${r.name} (${r.robot_code})`" :value="r.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="区域">
          <el-input placeholder="如 A区-仓库" style="width:150px" />
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="16">
      <el-col :span="12">
        <el-card class="glass-card">
          <template #header>📈 任务趋势 (近14天)</template>
          <div ref="taskChart" style="height:300px;"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="glass-card">
          <template #header>🔋 电量趋势 (近24h)</template>
          <div ref="batteryChart" style="height:300px;"></div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top:16px;">
      <el-col :span="12">
        <el-card class="glass-card">
          <template #header>🍩 任务类型分布 (近30天)</template>
          <div ref="typeChart" style="height:300px;"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="glass-card">
          <template #header>🏆 机器人利用率 (近30天)</template>
          <div ref="usageChart" style="height:300px;"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
