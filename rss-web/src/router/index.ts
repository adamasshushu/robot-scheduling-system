// ── 路由配置 ──
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { title: '登录' },
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: { title: '仪表盘', icon: 'Odometer' },
        },
        {
          path: 'robots',
          name: 'Robots',
          component: () => import('@/views/Robots.vue'),
          meta: { title: '机器人管理', icon: 'Monitor' },
        },
        {
          path: 'robots/:id',
          name: 'RobotDetail',
          component: () => import('@/views/RobotDetail.vue'),
          meta: { title: '机器人详情', hidden: true },
        },
        {
          path: 'tasks',
          name: 'Tasks',
          component: () => import('@/views/Tasks.vue'),
          meta: { title: '任务调度', icon: 'List' },
        },
        {
          path: 'monitor',
          name: 'Monitor',
          component: () => import('@/views/Monitor.vue'),
          meta: { title: '实时监控', icon: 'VideoCamera' },
        },
        {
          path: 'alerts',
          name: 'Alerts',
          component: () => import('@/views/Alerts.vue'),
          meta: { title: '告警中心', icon: 'Bell' },
        },
        {
          path: 'reports',
          name: 'Reports',
          component: () => import('@/views/Reports.vue'),
          meta: { title: '报表分析', icon: 'DataAnalysis' },
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/Settings.vue'),
          meta: { title: '系统设置', icon: 'Setting' },
        },
      ],
    },
  ],
})

// 路由守卫 — 未登录跳转
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('rss_token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
