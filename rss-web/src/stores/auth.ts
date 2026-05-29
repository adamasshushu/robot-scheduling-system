// ── 认证 Store ──
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { loginApi } from '@/api'
import router from '@/router'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('rss_token') || '')
  const username = ref(localStorage.getItem('rss_username') || '')
  const role = ref(localStorage.getItem('rss_role') || '')

  const isLoggedIn = () => !!token.value

  async function login(user: string, pass: string) {
    const { data } = await loginApi(user, pass)
    if (data.code === 200) {
      token.value = data.data.token
      username.value = data.data.username
      role.value = data.data.role
      localStorage.setItem('rss_token', token.value)
      localStorage.setItem('rss_username', username.value)
      localStorage.setItem('rss_role', role.value)
      router.push('/')
    }
    return data
  }

  function logout() {
    token.value = ''
    username.value = ''
    role.value = ''
    localStorage.clear()
    router.push('/login')
  }

  return { token, username, role, isLoggedIn, login, logout }
})
