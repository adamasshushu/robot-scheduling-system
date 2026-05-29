<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const form = ref({ username: 'admin', password: 'admin123' })
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(form.value.username, form.value.password)
  } catch (e: any) {
    error.value = e.response?.data?.message || '登录失败，请检查网络连接'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <span class="logo">🤖</span>
        <h2>机器人调度系统</h2>
        <p>Robot Scheduling System</p>
      </div>

      <el-form @submit.prevent="handleLogin" size="large">
        <el-form-item v-if="error">
          <el-alert :title="error" type="error" show-icon :closable="false" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%">
            {{ loading ? '登录中...' : '登 录' }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <span>默认账号: admin / admin123</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 400px;
  padding: 40px;
  background: var(--el-bg-color);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}
.login-header { text-align: center; margin-bottom: 32px; }
.logo { font-size: 48px; }
.login-header h2 { margin: 12px 0 4px; font-size: 24px; }
.login-header p { color: var(--el-text-color-secondary); font-size: 13px; }
.login-footer { text-align: center; margin-top: 16px; font-size: 12px; color: var(--el-text-color-placeholder); }
</style>
