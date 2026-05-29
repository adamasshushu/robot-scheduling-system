<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDark, useToggle } from '@vueuse/core'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isDark = useDark()
const toggleDark = useToggle(isDark)

const isCollapse = ref(false)

const menuItems = router.getRoutes()
  .filter(r => r.meta.icon && !r.meta.hidden)
  .map(r => ({
    path: r.path,
    title: r.meta.title as string,
    icon: r.meta.icon as string,
  }))

const currentTitle = computed(() => route.meta.title || '机器人调度系统')
</script>

<template>
  <el-container class="layout">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '220px'" class="aside">
      <div class="logo">
        <span v-if="!isCollapse">🤖 RSS</span>
        <span v-else>🤖</span>
      </div>
      <el-menu
        :default-active="route.path"
        :collapse="isCollapse"
        router
        :collapse-transition="false"
      >
        <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主体 -->
    <el-container>
      <!-- 顶栏 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <span class="page-title">{{ currentTitle }}</span>
        </div>
        <div class="header-right">
          <el-switch
            v-model="isDark"
            inline-prompt
            :active-icon="Moon"
            :inactive-icon="Sunny"
            @change="toggleDark()"
          />
          <el-dropdown>
            <span class="user-info">
              {{ auth.username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="auth.logout()">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区 -->
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout { height: 100vh; }
.aside { background: var(--el-menu-bg-color); border-right: 1px solid var(--el-border-color-light); transition: width 0.3s; }
.logo { height: 60px; display: flex; align-items: center; justify-content: center; font-size: 20px; font-weight: bold; color: var(--el-color-primary); border-bottom: 1px solid var(--el-border-color-light); }
.header { display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--el-border-color-light); padding: 0 20px; }
.header-left { display: flex; align-items: center; gap: 12px; }
.collapse-btn { font-size: 20px; cursor: pointer; }
.page-title { font-size: 18px; font-weight: 600; }
.header-right { display: flex; align-items: center; gap: 16px; }
.user-info { cursor: pointer; display: flex; align-items: center; gap: 4px; }
</style>
