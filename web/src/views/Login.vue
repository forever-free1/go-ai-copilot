<template>
  <div class="login-container">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
      <div class="blob blob-3"></div>
    </div>

    <div class="login-wrapper">
      <!-- 左侧品牌区 -->
      <div class="brand-section">
        <div class="brand-content">
          <h1 class="brand-title">Go AI<br/>Copilot</h1>
          <p class="brand-tagline">智能代码助手，让编程更高效</p>
          <div class="brand-features">
            <div class="feature-item">
              <span class="feature-icon">✨</span>
              <span>智能代码生成</span>
            </div>
            <div class="feature-item">
              <span class="feature-icon">🔍</span>
              <span>代码解释优化</span>
            </div>
            <div class="feature-item">
              <span class="feature-icon">🛡️</span>
              <span>漏洞检测</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧登录区 -->
      <div class="form-section">
        <div class="form-card">
          <div class="form-header">
            <h2>{{ activeTab === 'login' ? '欢迎回来' : '创建账号' }}</h2>
            <p>{{ activeTab === 'login' ? '登录您的账户' : '开始智能编程之旅' }}</p>
          </div>

          <el-tabs v-model="activeTab" class="auth-tabs">
            <el-tab-pane label="登录" name="login">
              <el-form :model="loginForm" :rules="loginRules" ref="loginFormRef" class="auth-form">
                <el-form-item prop="username">
                  <el-input
                    v-model="loginForm.username"
                    placeholder="用户名"
                    prefix-icon="User"
                    size="large"
                  />
                </el-form-item>
                <el-form-item prop="password">
                  <el-input
                    v-model="loginForm.password"
                    type="password"
                    placeholder="密码"
                    prefix-icon="Lock"
                    show-password
                    size="large"
                  />
                </el-form-item>
                <el-form-item>
                  <el-button
                    type="primary"
                    size="large"
                    :loading="loading"
                    @click="handleLogin"
                    class="submit-btn"
                  >
                    登录
                  </el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <el-tab-pane label="注册" name="register">
              <el-form :model="registerForm" :rules="registerRules" ref="registerFormRef" class="auth-form">
                <el-form-item prop="username">
                  <el-input
                    v-model="registerForm.username"
                    placeholder="用户名"
                    prefix-icon="User"
                    size="large"
                  />
                </el-form-item>
                <el-form-item prop="nickname">
                  <el-input
                    v-model="registerForm.nickname"
                    placeholder="昵称（可选）"
                    prefix-icon="User"
                    size="large"
                  />
                </el-form-item>
                <el-form-item prop="password">
                  <el-input
                    v-model="registerForm.password"
                    type="password"
                    placeholder="密码"
                    prefix-icon="Lock"
                    show-password
                    size="large"
                  />
                </el-form-item>
                <el-form-item prop="confirmPassword">
                  <el-input
                    v-model="registerForm.confirmPassword"
                    type="password"
                    placeholder="确认密码"
                    prefix-icon="Lock"
                    show-password
                    size="large"
                  />
                </el-form-item>
                <el-form-item>
                  <el-button
                    type="primary"
                    size="large"
                    :loading="loading"
                    @click="handleRegister"
                    class="submit-btn"
                  >
                    注册
                  </el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const activeTab = ref('login')
const loading = ref(false)
const loginFormRef = ref<FormInstance>()
const registerFormRef = ref<FormInstance>()

const loginForm = reactive({
  username: '',
  password: ''
})

const registerForm = reactive({
  username: '',
  nickname: '',
  password: '',
  confirmPassword: ''
})

const loginRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const validateConfirmPassword = (rule: any, value: string, callback: any) => {
  if (value !== registerForm.password) {
    callback(new Error('两次输入密码不一致'))
  } else {
    callback()
  }
}

const registerRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在3-50个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在6-20个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await userStore.loginAction(loginForm.username, loginForm.password)
        ElMessage.success('登录成功')
        router.push('/')
      } finally {
        loading.value = false
      }
    }
  })
}

const handleRegister = async () => {
  if (!registerFormRef.value) return
  await registerFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await userStore.registerAction(registerForm.username, registerForm.password, registerForm.nickname)
        ElMessage.success('注册成功，请登录')
        activeTab.value = 'login'
        loginForm.username = registerForm.username
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
/* CSS Variables */
:root {
  --color-bg: #0a0a0f;
  --color-surface: #14141a;
  --color-surface-hover: #1c1c24;
  --color-primary: #f59e0b;
  --color-primary-hover: #d97706;
  --color-text: #fafafa;
  --color-text-muted: #a1a1aa;
  --color-border: #27272a;
  --font-display: 'DM Serif Display', serif;
  --font-body: 'DM Sans', sans-serif;
}

/* 字体引入 */
@import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&family=DM+Serif+Display&display=swap');

.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0a0a0f;
  position: relative;
  overflow: hidden;
  font-family: 'DM Sans', sans-serif;
}

/* 背景装饰 */
.bg-decoration {
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
}

.blob-1 {
  width: 600px;
  height: 600px;
  background: #f59e0b;
  top: -200px;
  left: -100px;
  animation: float 20s ease-in-out infinite;
}

.blob-2 {
  width: 500px;
  height: 500px;
  background: #dc2626;
  bottom: -150px;
  right: -100px;
  animation: float 25s ease-in-out infinite reverse;
}

.blob-3 {
  width: 300px;
  height: 300px;
  background: #2563eb;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation: pulse 15s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(30px, -30px); }
}

@keyframes pulse {
  0%, 100% { transform: translate(-50%, -50%) scale(1); opacity: 0.3; }
  50% { transform: translate(-50%, -50%) scale(1.2); opacity: 0.5; }
}

/* 布局 */
.login-wrapper {
  display: flex;
  width: 900px;
  max-width: 95vw;
  min-height: 600px;
  background: rgba(20, 20, 26, 0.8);
  backdrop-filter: blur(20px);
  border-radius: 24px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  position: relative;
  z-index: 1;
}

/* 左侧品牌区 */
.brand-section {
  flex: 1;
  background: linear-gradient(135deg, #1a1a20 0%, #14141a 100%);
  padding: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.brand-section::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(245, 158, 11, 0.5), transparent);
}

.brand-content {
  color: #fafafa;
}

.brand-title {
  font-family: 'DM Serif Display', serif;
  font-size: 56px;
  line-height: 1.1;
  margin: 0 0 16px;
  background: linear-gradient(135deg, #fafafa 0%, #a1a1aa 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.brand-tagline {
  font-size: 16px;
  color: #a1a1aa;
  margin: 0 0 48px;
  letter-spacing: 0.5px;
}

.brand-features {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  color: #d4d4d8;
}

.feature-icon {
  font-size: 18px;
}

/* 右侧表单区 */
.form-section {
  flex: 1;
  padding: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.form-card {
  width: 100%;
  max-width: 360px;
}

.form-header {
  text-align: center;
  margin-bottom: 32px;
}

.form-header h2 {
  font-family: 'DM Serif Display', serif;
  font-size: 28px;
  color: #fafafa;
  margin: 0 0 8px;
}

.form-header p {
  font-size: 14px;
  color: #71717a;
  margin: 0;
}

/* 表单样式 */
.auth-tabs :deep(.el-tabs__nav-wrap::after) {
  background: #27272a;
}

.auth-tabs :deep(.el-tabs__item) {
  color: #71717a;
  font-size: 15px;
  padding: 0 24px;
}

.auth-tabs :deep(.el-tabs__item.is-active) {
  color: #f59e0b;
}

.auth-tabs :deep(.el-tabs__active-bar) {
  background: #f59e0b;
}

.auth-form {
  margin-top: 24px;
}

.auth-form :deep(.el-input__wrapper) {
  background: #1c1c24;
  border: 1px solid #27272a;
  box-shadow: none;
  border-radius: 12px;
  padding: 4px 12px;
}

.auth-form :deep(.el-input__wrapper:hover) {
  border-color: #3f3f46;
}

.auth-form :deep(.el-input__wrapper.is-focus) {
  border-color: #f59e0b;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.1);
}

.auth-form :deep(.el-input__inner) {
  color: #fafafa;
  height: 40px;
}

.auth-form :deep(.el-input__inner::placeholder) {
  color: #52525b;
}

.auth-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.submit-btn {
  width: 100%;
  height: 48px;
  background: #f59e0b;
  border: none;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 600;
  color: #0a0a0f;
  transition: all 0.3s ease;
  margin-top: 8px;
}

.submit-btn:hover {
  background: #d97706;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(245, 158, 11, 0.3);
}

.submit-btn:active {
  transform: translateY(0);
}

/* 响应式 */
@media (max-width: 768px) {
  .login-wrapper {
    flex-direction: column;
    width: 100%;
    max-width: 100%;
    min-height: auto;
    border-radius: 0;
  }

  .brand-section {
    padding: 32px;
  }

  .brand-title {
    font-size: 36px;
  }

  .form-section {
    padding: 32px 24px;
  }
}
</style>
