<template>
  <div class="rag-container">
    <!-- 侧边栏 - 始终展开显示历史会话 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <h1 class="logo">Go AI</h1>
      </div>

      <!-- 会话列表 -->
      <div class="session-header">
        <span>历史会话</span>
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
          <div class="session-info">
            <span class="session-title">{{ session.title }}</span>
            <span :class="['session-mode', `mode-${session.mode || 'chat'}`]">
              {{ getModeLabel(session.mode) }}
            </span>
          </div>
          <el-icon class="delete-btn" @click.stop="handleDeleteSession(session.id)">
            <Delete />
          </el-icon>
        </div>
      </div>

      <!-- 新建会话按钮 -->
      <div class="new-session-btn" @click="handleNewChat">
        <el-icon><Plus /></el-icon>
        <span>新建会话</span>
      </div>

      <div class="sidebar-footer">
        <div class="user-info" @click="handleLogout">
          <div class="user-avatar">
            <el-icon><User /></el-icon>
          </div>
          <span class="username">{{ userStore.userInfo?.username }}</span>
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
          <!-- 模式选择器 -->
          <div class="mode-bar">
            <el-select v-model="chatMode" placeholder="选择模式" class="mode-selector">
              <el-option label="💬 通用对话" value="chat" />
              <el-option label="💻 代码生成" value="code_generate" />
              <el-option label="🔍 代码解释" value="code_explain" />
              <el-option label="⚡ 代码优化" value="code_optimize" />
              <el-option label="🛡️ 漏洞检测" value="code_vuln" />
              <el-option label="🧪 单元测试" value="code_test" />
              <el-option label="📚 知识库问答" value="rag" />
            </el-select>

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
import { ref, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, User, ChatDotRound, Promotion, Service, FolderAdd } from '@element-plus/icons-vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import { useUserStore } from '../stores/user'
import { useChatStore } from '../stores/chat'
import { chat, chatWithMode, ragChat } from '../api/chat'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()

const inputMessage = ref('')
const chatMode = ref('chat')
const isStreaming = ref(false)
const chatAreaRef = ref<HTMLElement>()

// 上传配置
const uploadUrl = '/api/v1/rag/upload'
const uploadHeaders = {
  Authorization: `Bearer ${userStore.token}`
}

// 获取模式标签
const getModeLabel = (mode: string | undefined) => {
  const labels: Record<string, string> = {
    chat: '对话',
    code_generate: '代码',
    code_explain: '解释',
    code_optimize: '优化',
    code_vuln: '漏洞',
    code_test: '测试',
    rag: '知识库'
  }
  return labels[mode || 'chat'] || '对话'
}

// marked 配置
marked.setOptions({
  highlight: (code: string) => hljs.highlightAuto(code).value
})

const renderMarkdown = (content: string) => marked(content)

const truncate = (str: string, len: number): string => {
  if (!str) return ''
  return str.length > len ? str.slice(0, len) + '...' : str
}

const generateSummary = (message: string): string => {
  const cleaned = message.replace(/\s+/g, '').trim()
  return truncate(cleaned, 8)
}

// 新建会话
const handleNewChat = async () => {
  chatStore.selectSession(0)
  inputMessage.value = ''
}

// 选择会话
const handleSelectSession = async (id: number) => {
  await chatStore.selectSession(id)
}

// 删除会话
const handleDeleteSession = async (id: number) => {
  await ElMessageBox.confirm('确定删除该会话吗？', '提示', { type: 'warning' })
  await chatStore.removeSession(id)
  ElMessage.success('删除成功')
}

// 文件上传
const beforeUpload = (file: File) => {
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    ElMessage.error('文件大小不能超过 10MB')
  }
  return isLt10M
}

const handleFileUpload = async (response: any, file: File) => {
  ElMessage.success('文件上传成功')
}

const handleUploadError = () => {
  ElMessage.error('文件上传失败')
}

// 发送消息
const handleSend = async () => {
  if (!inputMessage.value.trim() || isStreaming.value) return

  // 如果没有会话，先创建
  if (!chatStore.currentSessionId) {
    const summary = generateSummary(inputMessage.value)
    const session = await chatStore.createNewSession(summary, chatMode.value)
    chatStore.selectSession(session.id)
  }

  const message = inputMessage.value.trim()
  inputMessage.value = ''
  isStreaming.value = true

  chatStore.addMessage('user', message)
  await nextTick()
  scrollToBottom()

  try {
    let res
    if (chatMode.value === 'rag') {
      res = await ragChat({ message, session_id: chatStore.currentSessionId! })
    } else if (chatMode.value === 'chat') {
      res = await chat({ message, session_id: chatStore.currentSessionId! })
    } else {
      res = await chatWithMode({ message, session_id: chatStore.currentSessionId!, mode: chatMode.value })
    }
    chatStore.addMessage('assistant', res.data.reply)
  } catch (error: any) {
    console.error('发送失败:', error)
    ElMessage.error(error?.response?.data?.message || '发送失败')
  } finally {
    isStreaming.value = false
    await nextTick()
    scrollToBottom()
  }
}

const scrollToBottom = () => {
  if (chatAreaRef.value) {
    chatAreaRef.value.scrollTop = chatAreaRef.value.scrollHeight
  }
}

const handleLogout = async () => {
  await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' })
  userStore.logout()
  router.push('/login')
}

onMounted(async () => {
  await chatStore.fetchSessions()
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&family=DM+Serif+Display&display=swap');

.rag-container {
  display: flex;
  height: 100vh;
  background: #0a0a0f;
  font-family: 'DM Sans', sans-serif;
}

/* 侧边栏 */
.sidebar {
  width: 280px;
  background: #14141a;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #27272a;
}

.sidebar-header {
  padding: 20px;
  display: flex;
  justify-content: center;
}

.logo {
  font-family: 'DM Serif Display', serif;
  font-size: 22px;
  color: #f59e0b;
  margin: 0;
}

.session-header {
  padding: 12px 20px;
  font-size: 12px;
  color: #71717a;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 8px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  cursor: pointer;
  border-radius: 10px;
  transition: all 0.2s;
  margin-bottom: 4px;
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

.session-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.session-title {
  color: #d4d4d8;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-item.active .session-title {
  color: #fafafa;
}

.session-mode {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  width: fit-content;
}

.mode-chat { background: #3b82f6; color: #fff; }
.mode-code_generate { background: #10b981; color: #fff; }
.mode-code_explain { background: #8b5cf6; color: #fff; }
.mode-code_optimize { background: #f59e0b; color: #fff; }
.mode-code_vuln { background: #ef4444; color: #fff; }
.mode-code_test { background: #06b6d4; color: #fff; }
.mode-rag { background: #ec4899; color: #fff; }

.delete-btn {
  color: #71717a;
  opacity: 0;
  transition: opacity 0.2s;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  color: #ef4444;
}

.new-session-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin: 12px 16px;
  padding: 12px;
  background: #f59e0b;
  color: #0a0a0f;
  border-radius: 10px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.new-session-btn:hover {
  background: #d97706;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #27272a;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 8px;
  border-radius: 10px;
}

.user-info:hover {
  background: #1c1c24;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0a0a0f;
}

.username {
  color: #fafafa;
  font-size: 14px;
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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

.mode-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.mode-selector {
  width: 160px;
}

.mode-selector :deep(.el-input__wrapper) {
  background: #14141a;
  border: 1px solid #27272a;
  box-shadow: none;
  border-radius: 10px;
}

.mode-selector :deep(.el-input__inner) {
  color: #fafafa;
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
}

.tool-btn:hover {
  background: #3f3f46;
  color: #fafafa;
}

.input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
  background: #14141a;
  border: 1px solid #27272a;
  border-radius: 16px;
  padding: 12px;
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

.send-btn {
  width: 48px;
  height: 48px;
  background: #f59e0b;
  border: none;
  border-radius: 12px;
  color: #0a0a0f;
  font-size: 20px;
}

.send-btn:hover {
  background: #d97706;
}

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

.markdown-body h1, .markdown-body h2, .markdown-body h3 {
  color: #fafafa;
  margin: 16px 0 8px;
}

.markdown-body p {
  margin: 8px 0;
}

.markdown-body ul, .markdown-body ol {
  padding-left: 20px;
}

.markdown-body li {
  margin: 4px 0;
}
</style>
