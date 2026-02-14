<template>
  <div class="home-container">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <h3>Go AI Copilot</h3>
        <el-button type="primary" size="small" @click="handleNewChat">新建会话</el-button>
      </div>

      <div class="session-list">
        <div
          v-for="session in chatStore.sessions"
          :key="session.id"
          :class="['session-item', { active: session.id === chatStore.currentSessionId }]"
          @click="handleSelectSession(session.id)"
        >
          <span class="session-title">{{ session.title }}</span>
          <el-icon class="delete-btn" @click.stop="handleDeleteSession(session.id)">
            <Delete />
          </el-icon>
        </div>
      </div>

      <div class="sidebar-footer">
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-icon><User /></el-icon>
            {{ userStore.userInfo?.username }}
          </span>
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
          <h1>欢迎使用 Go AI Copilot</h1>
          <p>选择一个会话或创建新会话开始对话</p>
          <div class="quick-actions">
            <el-button @click="handleNewChat">开始新会话</el-button>
          </div>
        </div>

        <div v-else class="messages">
          <div
            v-for="(msg, index) in chatStore.messages"
            :key="index"
            :class="['message', msg.role]"
          >
            <div class="message-avatar">
              <el-icon v-if="msg.role === 'user'"><User /></el-icon>
              <el-icon v-else><ChatDotRound /></el-icon>
            </div>
            <div class="message-content">
              <div v-if="msg.role === 'assistant'" class="markdown-body" v-html="renderMarkdown(msg.content)"></div>
              <div v-else>{{ msg.content }}</div>
            </div>
          </div>

          <!-- 正在回复 -->
          <div v-if="isStreaming" class="message assistant">
            <div class="message-avatar">
              <el-icon><Robot /></el-icon>
            </div>
            <div class="message-content">
              <span class="typing">正在输入...</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="input-area">
        <div class="mode-select">
          <el-select v-model="chatMode" placeholder="选择模式" style="width: 150px">
            <el-option label="通用对话" value="chat" />
            <el-option label="代码生成" value="code_generate" />
            <el-option label="代码解释" value="code_explain" />
            <el-option label="代码优化" value="code_optimize" />
            <el-option label="漏洞检测" value="code_vuln" />
            <el-option label="单元测试" value="code_test" />
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
          />
          <el-button
            type="primary"
            :loading="isStreaming"
            @click="handleSend"
          >
            发送
          </el-button>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, User, ChatDotRound } from '@element-plus/icons-vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import { useUserStore } from '../stores/user'
import { useChatStore } from '../stores/chat'
import { chat, streamChat, chatWithMode } from '../api/chat'

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
      // 普通对话
      const res = await chat({
        message,
        session_id: chatStore.currentSessionId!
      })
      chatStore.addMessage('assistant', res.data.reply)
    } else {
      // 带模式对话
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
.home-container {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 260px;
  background: #2d2d2d;
  color: #fff;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 16px;
}

.session-list {
  flex: 1;
  overflow-y: auto;
}

.session-item {
  padding: 12px 20px;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: background 0.2s;
}

.session-item:hover {
  background: #3d3d3d;
}

.session-item.active {
  background: #409eff;
}

.session-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.delete-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.sidebar-footer {
  padding: 15px;
  border-top: 1px solid #3d3d3d;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #fff;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}

.chat-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.welcome {
  text-align: center;
  margin-top: 100px;
}

.welcome h1 {
  color: #333;
}

.quick-actions {
  margin-top: 20px;
}

.messages {
  max-width: 800px;
  margin: 0 auto;
}

.message {
  display: flex;
  margin-bottom: 20px;
}

.message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #409eff;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.message.assistant .message-avatar {
  background: #67c23a;
}

.message-content {
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 8px;
  background: #fff;
  margin: 0 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.message.user .message-content {
  background: #409eff;
  color: #fff;
}

.typing {
  color: #999;
}

.input-area {
  background: #fff;
  padding: 20px;
  border-top: 1px solid #e4e4e4;
}

.mode-select {
  margin-bottom: 10px;
}

.input-wrapper {
  display: flex;
  gap: 10px;
}

.input-wrapper .el-textarea {
  flex: 1;
}

/* markdown样式 */
.markdown-body {
  line-height: 1.6;
}

.markdown-body code {
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 4px;
}

.markdown-body pre {
  background: #282c34;
  color: #abb2bf;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
}

.markdown-body pre code {
  background: none;
  padding: 0;
}
</style>
