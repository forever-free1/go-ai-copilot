<template>
  <div class="home-container">
    <!-- 侧边栏 -->
    <aside
      class="sidebar"
      @mouseenter="sidebarExpanded = true"
      @mouseleave="sidebarExpanded = false"
      :class="{ expanded: sidebarExpanded }"
    >
      <div class="sidebar-header">
        <h1 class="logo">Go AI</h1>
        <transition name="fade">
          <el-button v-if="sidebarExpanded" type="primary" size="small" class="new-chat-btn" @click="handleNewChat">
            <el-icon><Plus /></el-icon>
            新建会话
          </el-button>
        </transition>
      </div>

      <!-- 未展开时显示的按钮 -->
      <div v-if="!sidebarExpanded" class="collapsed-new-chat" @click="handleNewChat">
        <el-icon><Plus /></el-icon>
      </div>

      <div class="session-list">
        <div
          v-for="session in chatStore.sessions"
          :key="session.id"
          :class="['session-item', { active: session.id === chatStore.currentSessionId }]"
          @click="handleSelectSession(session.id)"
        >
          <div class="session-icon">
            <el-icon><ChatDotRound /></el-icon>
          </div>
          <transition name="fade">
            <span v-if="sidebarExpanded" class="session-title">{{ session.title }}</span>
          </transition>
          <transition name="fade">
            <el-icon v-if="sidebarExpanded" class="delete-btn" @click.stop="handleDeleteSession(session.id)">
              <Delete />
            </el-icon>
          </transition>
        </div>
      </div>

      <div class="sidebar-footer">
        <div v-if="sidebarExpanded" class="user-info" @click="handleCommand('logout')">
          <div class="user-avatar">
            <el-icon><User /></el-icon>
          </div>
          <span class="username">{{ userStore.userInfo?.username }}</span>
        </div>
        <div v-else class="collapsed-user" @click="handleCommand('logout')">
          <div class="user-avatar">
            <el-icon><User /></el-icon>
          </div>
        </div>
      </div>
    </aside>

    <!-- 主内容区 -->
    <main class="main-content">
      <!-- 对话区域 -->
      <div class="chat-area" ref="chatAreaRef">
        <div v-if="!chatStore.currentSessionId" class="welcome">
          <div class="welcome-content">
            <div class="welcome-icon">
              <el-icon><Promotion /></el-icon>
            </div>
            <h1>Go AI Copilot</h1>
            <p>您的智能代码助手</p>
            <div class="quick-actions">
              <el-button class="action-btn" @click="handleNewChat">
                <el-icon><Plus /></el-icon>
                开始新对话
              </el-button>
            </div>
            <div class="features">
              <div class="feature">
                <span class="feature-icon">💻</span>
                <span>代码生成</span>
              </div>
              <div class="feature">
                <span class="feature-icon">🔍</span>
                <span>代码解释</span>
              </div>
              <div class="feature">
                <span class="feature-icon">⚡</span>
                <span>代码优化</span>
              </div>
              <div class="feature">
                <span class="feature-icon">🛡️</span>
                <span>漏洞检测</span>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="messages-container">
          <div class="messages">
            <div
              v-for="(msg, index) in chatStore.messages"
              :key="index"
              :class="['message', msg.role]"
            >
              <div class="message-avatar">
                <el-icon v-if="msg.role === 'user'"><User /></el-icon>
                <el-icon v-else><Service /></el-icon>
              </div>
              <div class="message-content">
                <div v-if="msg.role === 'assistant'" class="markdown-body" v-html="renderMarkdown(msg.content)"></div>
                <div v-else class="text-content">{{ msg.content }}</div>
              </div>
            </div>

            <!-- 正在回复 -->
            <div v-if="isStreaming" class="message assistant">
              <div class="message-avatar">
                <el-icon><Service /></el-icon>
              </div>
              <div class="message-content">
                <span class="typing">
                  <span class="dot"></span>
                  <span class="dot"></span>
                  <span class="dot"></span>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="input-area">
        <div class="input-container">
          <div class="input-tools">
            <!-- 上传文件按钮 -->
            <el-upload
              class="file-upload"
              :action="uploadUrl"
              :headers="uploadHeaders"
              :show-file-list="false"
              :on-success="handleFileUpload"
              :on-error="handleUploadError"
              :before-upload="beforeUpload"
            >
              <el-button text class="tool-btn">
                <el-icon><FolderAdd /></el-icon>
              </el-button>
            </el-upload>

            <!-- 模式选择器 -->
            <el-dropdown trigger="click" @command="handleModeChange">
              <div class="mode-trigger">
                <el-icon><Operation /></el-icon>
                <span class="mode-label">{{ modeLabel }}</span>
                <el-icon class="arrow"><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu class="mode-dropdown">
                  <el-dropdown-item command="chat">
                    <span class="mode-icon">💬</span> 通用对话
                  </el-dropdown-item>
                  <el-dropdown-item command="code_generate">
                    <span class="mode-icon">💻</span> 代码生成
                  </el-dropdown-item>
                  <el-dropdown-item command="code_explain">
                    <span class="mode-icon">🔍</span> 代码解释
                  </el-dropdown-item>
                  <el-dropdown-item command="code_optimize">
                    <span class="mode-icon">⚡</span> 代码优化
                  </el-dropdown-item>
                  <el-dropdown-item command="code_vuln">
                    <span class="mode-icon">🛡️</span> 漏洞检测
                  </el-dropdown-item>
                  <el-dropdown-item command="code_test">
                    <span class="mode-icon">🧪</span> 单元测试
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>

          <div class="input-wrapper">
            <el-input
              v-model="inputMessage"
              type="textarea"
              :rows="3"
              placeholder="输入消息... (Shift+Enter 换行)"
              @keydown.enter.exact.prevent="handleSend"
              :disabled="isStreaming"
              class="message-input"
            />
            <el-button
              type="primary"
              :loading="isStreaming"
              @click="handleSend"
              class="send-btn"
            >
              <el-icon v-if="!isStreaming"><Promotion /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, User, ChatDotRound, ArrowDown, Promotion, Service, FolderAdd, Operation } from '@element-plus/icons-vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import { useUserStore } from '../stores/user'
import { useChatStore } from '../stores/chat'
import { chat, chatWithMode } from '../api/chat'
import { uploadDocument } from '../api/rag'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()

const inputMessage = ref('')
const chatMode = ref('chat')
const isStreaming = ref(false)
const chatAreaRef = ref<HTMLElement>()
const sidebarExpanded = ref(false)

// 计算模式标签
const modeLabel = computed(() => {
  const labels: Record<string, string> = {
    chat: '通用对话',
    code_generate: '代码生成',
    code_explain: '代码解释',
    code_optimize: '代码优化',
    code_vuln: '漏洞检测',
    code_test: '单元测试'
  }
  return labels[chatMode.value] || '通用对话'
})

// 上传配置
const uploadUrl = '/api/v1/rag/upload'
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${userStore.token}`
}))

// 配置marked
marked.setOptions({
  highlight: (code: string) => {
    return hljs.highlightAuto(code).value
  }
})

// 渲染markdown
const renderMarkdown = (content: string) => {
  return marked(content)
}

// 截取字符串（用于生成会话标题）
const truncate = (str: string, len: number): string => {
  if (!str) return ''
  return str.length > len ? str.slice(0, len) + '...' : str
}

// 生成问题摘要（6-10个字）
const generateSummary = (message: string): string => {
  // 去除多余空白字符
  const cleaned = message.replace(/\s+/g, '').trim()
  // 截取前8个字符作为摘要
  return truncate(cleaned, 8)
}

// 新建会话
const handleNewChat = async () => {
  // 清空当前会话消息，显示欢迎页
  chatStore.selectSession(0)
  inputMessage.value = ''
}

// 选择会话
const handleSelectSession = (id: number) => {
  chatStore.selectSession(id)
}

// 删除会话
const handleDeleteSession = async (id: number) => {
  await ElMessageBox.confirm('确定删除该会话吗？', '提示', {
    type: 'warning'
  })
  await chatStore.removeSession(id)
  ElMessage.success('删除成功')
}

// 模式切换
const handleModeChange = (mode: string) => {
  chatMode.value = mode
}

// 文件上传前检查
const beforeUpload = (file: File) => {
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    ElMessage.error('文件大小不能超过 10MB')
  }
  return isLt10M
}

// 文件上传成功
const handleFileUpload = async (response: any, file: File) => {
  ElMessage.success('文件上传成功，正在处理...')
  // 如果有会话，自动发送文件内容
  if (chatStore.currentSessionId) {
    const fileContent = `请分析以下文件：${file.name}`
    inputMessage.value = fileContent
  } else {
    // 创建新会话
    const title = truncate(file.name.replace(/\.[^/.]+$/, ''), 10)
    const session = await chatStore.createNewSession(title)
    chatStore.selectSession(session.id)
    inputMessage.value = `请分析以下文件：${file.name}`
  }
}

// 文件上传失败
const handleUploadError = () => {
  ElMessage.error('文件上传失败')
}

// 发送消息
const handleSend = async () => {
  if (!inputMessage.value.trim() || isStreaming.value) return

  // 如果没有会话，先创建
  if (!chatStore.currentSessionId) {
    const summary = generateSummary(inputMessage.value)
    const session = await chatStore.createNewSession(summary)
    chatStore.selectSession(session.id)
  }

  const message = inputMessage.value.trim()
  inputMessage.value = ''
  isStreaming.value = true

  // 添加用户消息
  chatStore.addMessage('user', message)

  // 滚动到底部
  await nextTick()
  scrollToBottom()

  try {
    if (chatMode.value === 'chat') {
      const res = await chat({
        message,
        session_id: chatStore.currentSessionId!
      })
      chatStore.addMessage('assistant', res.data.reply)
    } else {
      const res = await chatWithMode({
        message,
        session_id: chatStore.currentSessionId!,
        mode: chatMode.value
      })
      chatStore.addMessage('assistant', res.data.reply)
    }
  } catch (error) {
    ElMessage.error('发送失败')
  } finally {
    isStreaming.value = false
    await nextTick()
    scrollToBottom()
  }
}

// 滚动到底部
const scrollToBottom = () => {
  if (chatAreaRef.value) {
    chatAreaRef.value.scrollTop = chatAreaRef.value.scrollHeight
  }
}

// 用户菜单
const handleCommand = async (command: string) => {
  if (command === 'logout') {
    await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' })
    userStore.logout()
    router.push('/login')
  }
}

onMounted(async () => {
  await chatStore.fetchSessions()
})
</script>

<style scoped>
/* 字体引入 */
@import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&family=DM+Serif+Display&display=swap');

.home-container {
  display: flex;
  height: 100vh;
  background: #0a0a0f;
  font-family: 'DM Sans', sans-serif;
}

/* 侧边栏 - 收拢/展开 */
.sidebar {
  width: 64px;
  min-width: 64px;
  background: #14141a;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #27272a;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1), min-width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.sidebar.expanded {
  width: 280px;
  min-width: 280px;
}

.sidebar-header {
  padding: 24px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.logo {
  font-family: 'DM Serif Display', serif;
  font-size: 20px;
  color: #f59e0b;
  margin: 0;
  white-space: nowrap;
}

.new-chat-btn {
  background: #f59e0b;
  border: none;
  color: #0a0a0f;
  font-weight: 600;
  white-space: nowrap;
}

.new-chat-btn:hover {
  background: #d97706;
}

/* 收拢状态的新建按钮 */
.collapsed-new-chat {
  width: 40px;
  height: 40px;
  margin: 0 auto 16px;
  background: #f59e0b;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #0a0a0f;
  transition: all 0.2s;
}

.collapsed-new-chat:hover {
  background: #d97706;
  transform: scale(1.05);
}

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  cursor: pointer;
  border-radius: 12px;
  transition: all 0.2s ease;
  margin-bottom: 4px;
  justify-content: center;
}

.sidebar.expanded .session-item {
  justify-content: flex-start;
}

.session-item:hover {
  background: #1c1c24;
}

.session-item.active {
  background: rgba(245, 158, 11, 0.15);
}

.session-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #27272a;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #71717a;
  flex-shrink: 0;
}

.session-item.active .session-icon {
  background: #f59e0b;
  color: #0a0a0f;
}

.session-title {
  flex: 1;
  color: #d4d4d8;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-item.active .session-title {
  color: #fafafa;
}

.delete-btn {
  color: #71717a;
  transition: opacity 0.2s;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  color: #ef4444;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #27272a;
  display: flex;
  justify-content: center;
}

.sidebar.expanded .sidebar-footer {
  justify-content: flex-start;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 8px;
  border-radius: 12px;
  transition: background 0.2s;
}

.user-info:hover {
  background: #1c1c24;
}

.collapsed-user {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0a0a0f;
  cursor: pointer;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0a0a0f;
  flex-shrink: 0;
}

.username {
  color: #fafafa;
  font-size: 14px;
  font-weight: 500;
}

/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #0a0a0f;
}

.chat-area {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

/* 欢迎页 */
.welcome {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.welcome-content {
  text-align: center;
}

.welcome-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 24px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  border-radius: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 40px;
  color: #0a0a0f;
}

.welcome h1 {
  font-family: 'DM Serif Display', serif;
  font-size: 48px;
  color: #fafafa;
  margin: 0 0 12px;
}

.welcome p {
  color: #71717a;
  font-size: 18px;
  margin: 0 0 32px;
}

.quick-actions {
  margin-bottom: 48px;
}

.action-btn {
  background: #f59e0b;
  border: none;
  color: #0a0a0f;
  font-weight: 600;
  padding: 12px 32px;
  height: auto;
  border-radius: 12px;
}

.action-btn:hover {
  background: #d97706;
}

.features {
  display: flex;
  justify-content: center;
  gap: 32px;
}

.feature {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #71717a;
  font-size: 14px;
}

.feature-icon {
  font-size: 24px;
}

/* 消息列表 */
.messages-container {
  max-width: 900px;
  margin: 0 auto;
}

.messages {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.message {
  display: flex;
  gap: 16px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 20px;
}

.message.user .message-avatar {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #fff;
}

.message.assistant .message-avatar {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: #0a0a0f;
}

.message-content {
  max-width: 70%;
  padding: 16px 20px;
  border-radius: 16px;
  background: #14141a;
  color: #fafafa;
  line-height: 1.6;
}

.message.user .message-content {
  background: #1c1c24;
}

.text-content {
  white-space: pre-wrap;
}

/* typing动画 */
.typing {
  display: flex;
  gap: 4px;
  padding: 8px 0;
}

.dot {
  width: 8px;
  height: 8px;
  background: #71717a;
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out;
}

.dot:nth-child(1) { animation-delay: -0.32s; }
.dot:nth-child(2) { animation-delay: -0.16s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.5; }
  40% { transform: scale(1); opacity: 1; }
}

/* 输入区域 */
.input-area {
  padding: 20px 24px 24px;
  background: #0a0a0f;
}

.input-container {
  max-width: 900px;
  margin: 0 auto;
}

.input-tools {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.file-upload {
  display: inline-block;
}

.tool-btn {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: #27272a;
  border: none;
  color: #71717a;
  transition: all 0.2s;
}

.tool-btn:hover {
  background: #3f3f46;
  color: #fafafa;
}

/* 模式选择器 */
.mode-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #27272a;
  border-radius: 8px;
  cursor: pointer;
  color: #fafafa;
  font-size: 14px;
  transition: all 0.2s;
}

.mode-trigger:hover {
  background: #3f3f46;
}

.mode-label {
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.arrow {
  font-size: 12px;
  color: #71717a;
}

.mode-dropdown {
  background: #1c1c24;
  border: 1px solid #27272a;
}

.mode-dropdown :deep(.el-dropdown-menu__item) {
  color: #d4d4d8;
}

.mode-dropdown :deep(.el-dropdown-menu__item:hover) {
  background: #27272a;
  color: #fafafa;
}

.mode-icon {
  margin-right: 8px;
}

.input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
  background: #14141a;
  border: 1px solid #27272a;
  border-radius: 16px;
  padding: 12px;
  transition: border-color 0.2s;
}

.input-wrapper:focus-within {
  border-color: #f59e0b;
}

.message-input {
  flex: 1;
}

.message-input :deep(.el-textarea__inner) {
  background: transparent;
  border: none;
  box-shadow: none;
  color: #fafafa;
  font-size: 15px;
  line-height: 1.6;
  resize: none;
}

.message-input :deep(.el-textarea__inner:focus) {
  box-shadow: none;
}

.send-btn {
  width: 48px;
  height: 48px;
  background: #f59e0b;
  border: none;
  border-radius: 12px;
  color: #0a0a0f;
  font-size: 20px;
  flex-shrink: 0;
  transition: all 0.2s;
}

.send-btn:hover {
  background: #d97706;
  transform: scale(1.05);
}

/* markdown样式 */
.markdown-body {
  line-height: 1.7;
}

.markdown-body code {
  background: #27272a;
  padding: 2px 6px;
  border-radius: 6px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
}

.markdown-body pre {
  background: #1c1c24;
  padding: 16px;
  border-radius: 12px;
  overflow-x: auto;
  margin: 12px 0;
}

.markdown-body pre code {
  background: none;
  padding: 0;
}

.markdown-body h1,
.markdown-body h2,
.markdown-body h3 {
  color: #fafafa;
  margin: 16px 0 8px;
}

.markdown-body p {
  margin: 8px 0;
}

.markdown-body ul,
.markdown-body ol {
  padding-left: 20px;
}

.markdown-body li {
  margin: 4px 0;
}
</style>
