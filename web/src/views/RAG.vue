<template>
  <div class="rag-container">
    <!-- 侧边栏 -->
    <aside
      class="sidebar"
      @mouseenter="sidebarExpanded = true"
      @mouseleave="sidebarExpanded = false"
      :class="{ expanded: sidebarExpanded }"
    >
      <div class="sidebar-header">
        <h1 class="logo">Go AI</h1>
      </div>

      <div class="nav-list">
        <div
          class="nav-item"
          :class="{ active: currentView === 'chat' }"
          @click="goToChat"
        >
          <div class="nav-icon">
            <el-icon><ChatDotRound /></el-icon>
          </div>
          <transition name="fade">
            <span v-if="sidebarExpanded" class="nav-label">对话</span>
          </transition>
        </div>

        <div
          class="nav-item"
          :class="{ active: currentView === 'knowledge' }"
          @click="currentView = 'knowledge'"
        >
          <div class="nav-icon">
            <el-icon><Collection /></el-icon>
          </div>
          <transition name="fade">
            <span v-if="sidebarExpanded" class="nav-label">知识库</span>
          </transition>
        </div>
      </div>

      <div class="sidebar-footer">
        <div v-if="sidebarExpanded" class="user-info" @click="handleLogout">
          <div class="user-avatar">
            <el-icon><User /></el-icon>
          </div>
          <span class="username">{{ userStore.userInfo?.username }}</span>
        </div>
        <div v-else class="collapsed-user" @click="handleLogout">
          <div class="user-avatar">
            <el-icon><User /></el-icon>
          </div>
        </div>
      </div>
    </aside>

    <!-- 主内容区 -->
    <main class="main-content">
      <!-- 知识库管理视图 -->
      <div v-if="currentView === 'knowledge'" class="knowledge-view">
        <div class="page-header">
          <h2>知识库管理</h2>
          <el-button type="primary" @click="triggerUpload">
            <el-icon><Plus /></el-icon>
            上传文档
          </el-button>
        </div>

        <!-- 文档列表 -->
        <div class="document-list" v-loading="loading">
          <el-empty v-if="!loading && documents.length === 0" description="暂无文档，请上传" />

          <div
            v-for="doc in documents"
            :key="doc.id"
            class="document-card"
          >
            <div class="doc-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="doc-info">
              <div class="doc-name">{{ doc.filename }}</div>
              <div class="doc-meta">
                <span>{{ formatDate(doc.created_at) }}</span>
                <span>{{ doc.chunk_count }} 个片段</span>
              </div>
            </div>
            <el-button
              type="danger"
              text
              @click="handleDelete(doc.id)"
              class="delete-btn"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>

      <!-- 对话视图 -->
      <div v-else class="chat-view">
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
                    <el-dropdown-item command="rag" divided>
                      <span class="mode-icon">📚</span> 知识库问答
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
      </div>
    </main>

    <!-- 隐藏的上传组件 -->
    <el-upload
      ref="uploadRef"
      class="hidden-upload"
      :action="uploadUrl"
      :headers="uploadHeaders"
      :show-file-list="false"
      :on-success="handleFileUpload"
      :on-error="handleUploadError"
      :before-upload="beforeUpload"
      :auto-upload="false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, User, ChatDotRound, Promotion, Service, FolderAdd, Operation, ArrowDown, Collection, Document } from '@element-plus/icons-vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import { useUserStore } from '../stores/user'
import { useChatStore } from '../stores/chat'
import { chat, chatWithMode, ragChat } from '../api/chat'
import { getDocuments, deleteDocument } from '../api/rag'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()

// 视图状态
const currentView = ref<'chat' | 'knowledge'>('chat')
const sidebarExpanded = ref(false)

// 知识库相关
const loading = ref(false)
const documents = ref<any[]>([])
const uploadRef = ref()

// 对话相关
const inputMessage = ref('')
const chatMode = ref('chat')
const isStreaming = ref(false)
const chatAreaRef = ref<HTMLElement>()

// 计算属性
const modeLabel = computed(() => {
  const labels: Record<string, string> = {
    chat: '通用对话',
    code_generate: '代码生成',
    code_explain: '代码解释',
    code_optimize: '代码优化',
    code_vuln: '漏洞检测',
    code_test: '单元测试',
    rag: '知识库问答'
  }
  return labels[chatMode.value] || '通用对话'
})

const uploadUrl = '/api/v1/rag/upload'
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${userStore.token}`
}))

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

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}

// 导航
const goToChat = () => {
  currentView.value = 'chat'
}

// 加载文档列表
const loadDocuments = async () => {
  loading.value = true
  try {
    const res = await getDocuments()
    documents.value = res.data.data || []
  } catch (error) {
    ElMessage.error('加载文档列表失败')
  } finally {
    loading.value = false
  }
}

// 删除文档
const handleDelete = async (id: number) => {
  await ElMessageBox.confirm('确定删除该文档吗？', '提示', { type: 'warning' })
  try {
    await deleteDocument(id)
    ElMessage.success('删除成功')
    loadDocuments()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

// 触发上传
const triggerUpload = () => {
  uploadRef.value?.$el?.querySelector('input')?.click()
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
  ElMessage.success('文件上传成功，正在处理...')
  loadDocuments()
}

const handleUploadError = () => {
  ElMessage.error('文件上传失败')
}

// 对话功能
const handleNewChat = async () => {
  chatStore.selectSession(0)
  inputMessage.value = ''
}

const handleModeChange = (mode: string) => {
  chatMode.value = mode
}

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

  chatStore.addMessage('user', message)
  await nextTick()
  scrollToBottom()

  try {
    let res
    if (chatMode.value === 'rag') {
      // RAG 对话
      res = await ragChat({
        message,
        session_id: chatStore.currentSessionId!
      })
    } else if (chatMode.value === 'chat') {
      res = await chat({ message, session_id: chatStore.currentSessionId! })
    } else {
      res = await chatWithMode({ message, session_id: chatStore.currentSessionId!, mode: chatMode.value })
    }
    chatStore.addMessage('assistant', res.data.reply)
  } catch (error) {
    ElMessage.error('发送失败')
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
  await loadDocuments()
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
  width: 64px;
  min-width: 64px;
  background: #14141a;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #27272a;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar.expanded {
  width: 200px;
  min-width: 200px;
}

.sidebar-header {
  padding: 24px 16px;
  display: flex;
  justify-content: center;
}

.sidebar.expanded .sidebar-header {
  justify-content: flex-start;
}

.logo {
  font-family: 'DM Serif Display', serif;
  font-size: 20px;
  color: #f59e0b;
  margin: 0;
}

.nav-list {
  flex: 1;
  padding: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  cursor: pointer;
  border-radius: 12px;
  transition: all 0.2s;
  justify-content: center;
  margin-bottom: 4px;
}

.sidebar.expanded .nav-item {
  justify-content: flex-start;
}

.nav-item:hover {
  background: #1c1c24;
}

.nav-item.active {
  background: rgba(245, 158, 11, 0.15);
}

.nav-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #27272a;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #71717a;
}

.nav-item.active .nav-icon {
  background: #f59e0b;
  color: #0a0a0f;
}

.nav-label {
  color: #d4d4d8;
  font-size: 14px;
}

.nav-item.active .nav-label {
  color: #fafafa;
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
}

.user-info:hover {
  background: #1c1c24;
}

.collapsed-user, .user-avatar {
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

.username {
  color: #fafafa;
  font-size: 14px;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 知识库视图 */
.knowledge-view {
  flex: 1;
  padding: 32px;
  overflow-y: auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.page-header h2 {
  font-family: 'DM Serif Display', serif;
  font-size: 28px;
  color: #fafafa;
  margin: 0;
}

.document-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.document-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #14141a;
  border-radius: 16px;
  border: 1px solid #27272a;
  transition: all 0.2s;
}

.document-card:hover {
  border-color: #3f3f46;
  transform: translateY(-2px);
}

.doc-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 24px;
}

.doc-info {
  flex: 1;
  min-width: 0;
}

.doc-name {
  color: #fafafa;
  font-size: 15px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.doc-meta {
  display: flex;
  gap: 16px;
  color: #71717a;
  font-size: 13px;
  margin-top: 4px;
}

.delete-btn {
  color: #71717a;
}

.delete-btn:hover {
  color: #ef4444;
}

/* 对话视图 */
.chat-view {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chat-area {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

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
}

.mode-trigger:hover {
  background: #3f3f46;
}

.mode-label {
  max-width: 100px;
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

.hidden-upload {
  display: none;
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
