<template>
  <div class="home-container">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <h1 class="logo">Go AI</h1>
        <el-button type="primary" size="small" class="new-chat-btn" @click="handleNewChat">
          <el-icon><Plus /></el-icon>
          新建会话
        </el-button>
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
          <span class="session-title">{{ session.title }}</span>
          <el-icon class="delete-btn" @click.stop="handleDeleteSession(session.id)">
            <Delete />
          </el-icon>
        </div>
      </div>

      <div class="sidebar-footer">
        <el-dropdown @command="handleCommand">
          <div class="user-info">
            <div class="user-avatar">
              <el-icon><User /></el-icon>
            </div>
            <span class="username">{{ userStore.userInfo?.username }}</span>
            <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="settings">个人设置</el-dropdown-item>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
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
                <el-icon><Robot /></el-icon>
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
          <div class="mode-select">
            <el-select v-model="chatMode" placeholder="选择模式" class="mode-selector">
              <el-option label="💬 通用对话" value="chat" />
              <el-option label="💻 代码生成" value="code_generate" />
              <el-option label="🔍 代码解释" value="code_explain" />
              <el-option label="⚡ 代码优化" value="code_optimize" />
              <el-option label="🛡️ 漏洞检测" value="code_vuln" />
              <el-option label="🧪 单元测试" value="code_test" />
            </el-select>
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
import { Plus, Delete, User, ChatDotRound, ArrowDown, Promotion, Service } from '@element-plus/icons-vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import { useUserStore } from '../stores/user'
import { useChatStore } from '../stores/chat'
import { chat, chatWithMode } from '../api/chat'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()

const inputMessage = ref('')
const chatMode = ref('chat')
const isStreaming = ref(false)
const chatAreaRef = ref<HTMLElement>()

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

// 新建会话
const handleNewChat = async () => {
  const title = inputMessage.value.slice(0, 20) || '新会话'
  const session = await chatStore.createNewSession(title)
  chatStore.selectSession(session.id)
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

// 发送消息
const handleSend = async () => {
  if (!inputMessage.value.trim() || isStreaming.value) return

  // 如果没有会话，先创建
  if (!chatStore.currentSessionId) {
    await handleNewChat()
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

/* 侧边栏 */
.sidebar {
  width: 280px;
  background: #14141a;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #27272a;
}

.sidebar-header {
  padding: 24px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  font-family: 'DM Serif Display', serif;
  font-size: 24px;
  color: #f59e0b;
  margin: 0;
}

.new-chat-btn {
  background: #f59e0b;
  border: none;
  color: #0a0a0f;
  font-weight: 600;
}

.new-chat-btn:hover {
  background: #d97706;
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
  padding: 12px 16px;
  cursor: pointer;
  border-radius: 12px;
  transition: all 0.2s ease;
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
  opacity: 0;
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

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0a0a0f;
}

.username {
  flex: 1;
  color: #fafafa;
  font-size: 14px;
  font-weight: 500;
}

.dropdown-icon {
  color: #71717a;
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

.mode-select {
  margin-bottom: 12px;
}

.mode-selector {
  width: 200px;
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
