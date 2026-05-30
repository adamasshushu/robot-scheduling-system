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
    <!-- 背景装饰 -->
    <div class="bg-orb orb-1"></div>
    <div class="bg-orb orb-2"></div>
    <div class="bg-orb orb-3"></div>

    <div class="login-card">
      <div class="login-header">
        <div class="logo-ring">
          <span class="logo">🤖</span>
        </div>
        <h2>机器人调度系统</h2>
        <p>Robot Scheduling System</p>
      </div>

      <el-form @submit.prevent="handleLogin" size="large">
        <el-form-item v-if="error">
          <el-alert :title="error" type="error" show-icon :closable="false" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" :prefix-icon="User" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" :prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item>
          <el-button native-type="submit" :loading="loading" class="glass-btn" style="width: 100%; height: 44px; font-size: 16px;">
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
  background: linear-gradient(135deg, #0f0c29 0%, #302b63 50%, #24243e 100%);
  overflow: hidden;
  position: relative;
}

/* 背景光晕 */
.bg-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.3;
  pointer-events: none;
}
.orb-1 { width: 400px; height: 400px; background: #6C5CE7; top: -100px; left: -100px; animation: float 8s ease-in-out infinite; }
.orb-2 { width: 350px; height: 350px; background: #a29bfe; bottom: -80px; right: -80px; animation: float 10s ease-in-out infinite reverse; }
.orb-3 { width: 200px; height: 200px; background: #fd79a8; top: 50%; left: 50%; transform: translate(-50%, -50%); animation: float 12s ease-in-out infinite; }
@keyframes float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(30px, -30px) scale(1.1); }
}

.login-card {
  width: 400px;
  padding: 44px 40px;
  background: rgba(255, 255, 255, 0.06);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  position: relative;
  z-index: 1;
}
.login-header { text-align: center; margin-bottom: 36px; }
.logo-ring {
  width: 80px; height: 80px; border-radius: 24px;
  background: linear-gradient(135deg, #6C5CE7, #a29bfe);
  display: flex; align-items: center; justify-content: center;
  margin: 0 auto 16px;
  box-shadow: 0 8px 30px rgba(108, 92, 231, 0.4);
}
.logo { font-size: 40px; }
.login-header h2 { margin: 0 0 6px; font-size: 24px; color: #fff; font-weight: 700; }
.login-header p { color: rgba(255,255,255,0.5); font-size: 13px; letter-spacing: 2px; }
.login-footer { text-align: center; margin-top: 20px; font-size: 12px; color: rgba(255,255,255,0.35); }

/* 覆盖 element-plus 输入框在深色背景下的样式 */
:deep(.el-input__wrapper) {
  background: rgba(255,255,255,0.08) !important;
  border: 1px solid rgba(255,255,255,0.12) !important;
  box-shadow: none !important;
}
:deep(.el-input__inner) { color: #fff !important; }
:deep(.el-input__inner::placeholder) { color: rgba(255,255,255,0.4) !important; }
</style>
