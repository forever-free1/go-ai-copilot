<template>
  <div class="app-container">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="logo-area">
          <span class="logo-icon">✨</span>
          <span class="logo-text">Go AI</span>
        </div>
      </div>

      <!-- 模式切换标签 -->
      <div class="mode-tabs">
        <div
          :class="['tab-item', { active: currentView === 'chat' }]"
          @click="currentView = 'chat'"
        >
          <span class="tab-icon">💬</span>
          <span>对话</span>
        </div>
        <div
          :class="['tab-item', { active: currentView === 'knowledge' }]"
          @click="currentView = 'knowledge'"
        >
          <span class="tab-icon">📚</span>
          <span>知识库</span>
        </div>
      </div>

      <!-- 对话模式内容 -->
      <div v-if="currentView === 'chat'" class="sidebar-content">
        <div class="section-title">历史会话</div>
        <div class="session-list">
          <div
            v-for="session in chatSessions"
            :key="session.id"
            :class="['session-item', { active: session.id === chatStore.currentSessionId }]"
            @click="handleSelectSession(session.id)"
          >
            <div class="session-icon">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <div class="session-info">
              <span class="session-title">{{ session.title }}</span>
            </div>
            <el-icon class="delete-btn" @click.stop="handleDeleteSession(session.id)">
              <Delete />
            </el-icon>
          </div>
        </div>
      </div>

      <!-- 知识库模式内容 -->
      <div v-else class="sidebar-content">
        <div class="section-title">我的知识库</div>
        <div class="document-list">
          <div
            v-for="doc in documents"
            :key="doc.id"
            :class="['doc-item', { active: doc.id === selectedDocId }]"
            @click="handleSelectDoc(doc)"
          >
            <div class="doc-icon">📄</div>
            <div class="doc-info">
              <span class="doc-name">{{ doc.file_name }}</span>
              <span class="doc-date">{{ formatDate(doc.created_at) }}</span>
            </div>
            <el-icon class="delete-btn" @click.stop="handleDeleteDoc(doc.id)">
              <Delete />
            </el-icon>
          </div>
        </div>

        <!-- 上传区域 -->
        <div class="upload-section">
          <el-upload
            class="upload-area"
            :action="uploadUrl"
            :headers="uploadHeaders"
            :show-file-list="false"
            :on-success="handleFileUpload"
            :on-error="handleUploadError"
            :before-upload="beforeUpload"
            drag
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <span>拖拽文件到这里</span>
              <span class="upload-hint">或点击上传</span>
            </div>
          </el-upload>
        </div>
      </div>

      <!-- 新建按钮 -->
      <div class="new-btn-area">
        <div v-if="currentView === 'chat'" class="new-btn" @click="handleNewChat">
          <el-icon><Plus /></el-icon>
          <span>新建会话</span>
        </div>
      </div>

      <!-- 用户信息 -->
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
      <!-- 对话模式 -->
      <div v-if="currentView === 'chat'" class="chat-view">
        <div class="chat-area" ref="chatAreaRef">
          <!-- 欢迎页 -->
          <div v-if="!chatStore.currentSessionId" class="welcome">
            <div class="welcome-content">
              <div class="welcome-icon">
                <span>🤖</span>
              </div>
              <h1>你好！我是 Go AI</h1>
              <p>有什么我可以帮助你的吗？</p>
              <div class="suggestions">
                <div class="suggestion-chip" @click="handleSuggestion('帮我写一个快速排序算法')">
                  帮我写一个快速排序算法
                </div>
                <div class="suggestion-chip" @click="handleSuggestion('解释什么是闭包')">
                  解释什么是闭包
                </div>
                <div class="suggestion-chip" @click="handleSuggestion('优化这段代码')">
                  优化这段代码
                </div>
              </div>
            </div>
          </div>

          <!-- 消息列表 -->
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
          <div class="input-wrapper">
            <el-input
              v-model="inputMessage"
              type="textarea"
              :rows="2"
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

      <!-- 知识库模式 -->
      <div v-else class="knowledge-view">
        <div class="knowledge-header">
          <h2>知识库问答</h2>
          <p>基于您上传的文档进行智能问答</p>
        </div>

        <!-- 文档概览 -->
        <div v-if="selectedDoc" class="doc-overview">
          <div class="doc-card">
            <div class="doc-icon-large">📄</div>
            <div class="doc-details">
              <h3>{{ selectedDoc.file_name }}</h3>
              <p>{{ selectedDoc.chunk_count || 0 }} 个文档片段</p>
            </div>
          </div>
        </div>

        <!-- 问答区域 -->
        <div class="qa-area" ref="qaAreaRef">
          <div v-if="!ragMessages.length" class="qa-empty">
            <span>📚</span>
            <p>选择或上传知识库文档，然后开始提问</p>
          </div>
          <div v-else class="qa-messages">
            <div
              v-for="(msg, index) in ragMessages"
              :key="index"
              :class="['qa-message', msg.role]"
            >
              <div class="qa-content">
                <div v-if="msg.role === 'assistant'" class="markdown-body" v-html="renderMarkdown(msg.content)"></div>
                <div v-else>{{ msg.content }}</div>
              </div>
            </div>
            <div v-if="isStreaming" class="qa-message assistant">
              <div class="qa-content">
                <span class="typing">
                  <span class="dot"></span>
                  <span class="dot"></span>
                  <span class="dot"></span>
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 知识库输入 -->
        <div class="input-area">
          <div class="input-wrapper knowledge-input">
            <el-input
              v-model="ragInput"
              type="textarea"
              :rows="2"
              placeholder="基于知识库提问..."
              @keydown.enter.exact.prevent="handleRagSend"
              :disabled="isStreaming || !selectedDoc"
              class="message-input"
            />
            <el-button
              type="primary"
              :loading="isStreaming"
              @click="handleRagSend"
              class="send-btn"
              :disabled="!selectedDoc"
            >
              <el-icon v-if="!isStreaming"><Promotion /></el-icon>
            </el-button>
          </div>
          <p v-if="!selectedDoc" class="input-hint">请先在左侧选择或上传知识库文档</p>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, User, ChatDotRound, Promotion, Service, UploadFilled } from '@element-plus/icons-vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import { useUserStore } from '../stores/user'
import { useChatStore } from '../stores/chat'
import { chat, ragChat, getDocuments, deleteDocument } from '../api/rag'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()

const currentView = ref<'chat' | 'knowledge'>('chat')
const inputMessage = ref('')
const ragInput = ref('')
const isStreaming = ref(false)
const chatAreaRef = ref<HTMLElement>()
const qaAreaRef = ref<HTMLElement>()

// 知识库相关
const documents = ref<any[]>([])
const selectedDocId = ref<number | null>(null)
const selectedDoc = computed(() => documents.value.find(d => d.id === selectedDocId.value))
const ragMessages = ref<{ role: string; content: string }[]>([])

// 上传配置
const uploadUrl = '/api/v1/rag/upload'
const uploadHeaders = {
  Authorization: `Bearer ${userStore.token}`
}

// 过滤出非RAG模式的会话
const chatSessions = computed(() => {
  return chatStore.sessions.filter(s => s.mode !== 'rag')
})

// marked 配置
marked.setOptions({
  highlight: (code: string) => hljs.highlightAuto(code).value
})

const renderMarkdown = (content: string) => marked(content)

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

const truncate = (str: string, len: number): string => {
  if (!str) return ''
  return str.length > len ? str.slice(0, len) + '...' : str
}

// 获取知识库文档列表
const fetchDocuments = async () => {
  try {
    const res = await getDocuments()
    documents.value = res.data || []
    if (documents.value.length > 0 && !selectedDocId.value) {
      selectedDocId.value = documents.value[0].id
    }
  } catch (error) {
    console.error('获取文档失败:', error)
  }
}

// 选择文档
const handleSelectDoc = (doc: any) => {
  selectedDocId.value = doc.id
  ragMessages.value = []
}

// 删除文档
const handleDeleteDoc = async (id: number) => {
  await ElMessageBox.confirm('确定删除该文档吗？', '提示', { type: 'warning' })
  await deleteDocument(id)
  ElMessage.success('删除成功')
  if (selectedDocId.value === id) {
    selectedDocId.value = null
    ragMessages.value = []
  }
  await fetchDocuments()
}

// 文件上传
const beforeUpload = (file: File) => {
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    ElMessage.error('文件大小不能超过 10MB')
  }
  return isLt10M
}

const handleFileUpload = async (response: any) => {
  ElMessage.success('文件上传成功')
  await fetchDocuments()
}

const handleUploadError = () => {
  ElMessage.error('文件上传失败')
}

// 新建会话
const handleNewChat = async () => {
  chatStore.selectSession(0)
  inputMessage.value = ''
}

// 快捷建议
const handleSuggestion = (text: string) => {
  inputMessage.value = text
  handleSend()
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

// 发送聊天消息
const handleSend = async () => {
  if (!inputMessage.value.trim() || isStreaming.value) return

  if (!chatStore.currentSessionId) {
    const summary = truncate(inputMessage.value, 8)
    const session = await chatStore.createNewSession(summary, 'chat')
    await chatStore.selectSession(session.id)
  }

  const message = inputMessage.value.trim()
  inputMessage.value = ''
  isStreaming.value = true

  chatStore.addMessage('user', message)
  await nextTick()
  scrollToBottom()

  try {
    const res = await chat({ message, session_id: chatStore.currentSessionId! })
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

// 发送知识库问答
const handleRagSend = async () => {
  if (!ragInput.value.trim() || isStreaming.value || !selectedDocId.value) return

  const message = ragInput.value.trim()
  ragInput.value = ''
  isStreaming.value = true

  ragMessages.value.push({ role: 'user', content: message })
  await nextTick()
  scrollToQaBottom()

  try {
    const res = await ragChat({ message, document_id: selectedDocId.value })
    ragMessages.value.push({ role: 'assistant', content: res.data.reply })
  } catch (error: any) {
    console.error('发送失败:', error)
    ElMessage.error(error?.response?.data?.message || '发送失败')
  } finally {
    isStreaming.value = false
    await nextTick()
    scrollToQaBottom()
  }
}

const scrollToBottom = () => {
  if (chatAreaRef.value) {
    chatAreaRef.value.scrollTop = chatAreaRef.value.scrollHeight
  }
}

const scrollToQaBottom = () => {
  if (qaAreaRef.value) {
    qaAreaRef.value.scrollTop = qaAreaRef.value.scrollHeight
  }
}

const handleLogout = async () => {
  await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' })
  userStore.logout()
  router.push('/login')
}

onMounted(async () => {
  await chatStore.fetchSessions()
  await fetchDocuments()
})
</script>

<style scoped>
/* 引入圆润字体 */
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@400;500;600;700;800&display=swap');

* {
  box-sizing: border-box;
}

.app-container {
  display: flex;
  height: 100vh;
  background: #fafbfc;
  font-family: 'Nunito', 'Comic Neue', sans-serif;
}

/* 侧边栏 */
.sidebar {
  width: 280px;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #e8eaf0;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.02);
}

.sidebar-header {
  padding: 20px;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logo-icon {
  font-size: 24px;
}

.logo-text {
  font-size: 22px;
  font-weight: 800;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 模式切换标签 */
.mode-tabs {
  display: flex;
  padding: 0 12px;
  gap: 8px;
  margin-bottom: 16px;
}

.tab-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 12px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  color: #6b7280;
  background: #f3f4f6;
}

.tab-item:hover {
  background: #e5e7eb;
}

.tab-item.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.tab-icon {
  font-size: 16px;
}

/* 侧边栏内容 */
.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 0 12px;
}

.section-title {
  font-size: 12px;
  font-weight: 700;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 8px 8px 12px;
}

/* 会话列表 */
.session-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  cursor: pointer;
  border-radius: 14px;
  transition: all 0.2s;
  margin-bottom: 6px;
}

.session-item:hover {
  background: #f3f4f6;
}

.session-item.active {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border: 1px solid rgba(102, 126, 234, 0.2);
}

.session-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  flex-shrink: 0;
  font-size: 18px;
}

.session-item.active .session-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.session-info {
  flex: 1;
  min-width: 0;
}

.session-title {
  color: #374151;
  font-size: 14px;
  font-weight: 600;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.delete-btn {
  color: #9ca3af;
  opacity: 0;
  transition: opacity 0.2s;
  cursor: pointer;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  color: #ef4444;
}

/* 文档列表 */
.doc-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  cursor: pointer;
  border-radius: 14px;
  transition: all 0.2s;
  margin-bottom: 6px;
  background: #f9fafb;
}

.doc-item:hover {
  background: #f3f4f6;
}

.doc-item.active {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.1) 0%, rgba(5, 150, 105, 0.1) 100%);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.doc-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: #ecfdf5;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.doc-info {
  flex: 1;
  min-width: 0;
}

.doc-name {
  color: #374151;
  font-size: 14px;
  font-weight: 600;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.doc-date {
  color: #9ca3af;
  font-size: 12px;
}

/* 上传区域 */
.upload-section {
  padding: 12px 0;
}

.upload-area {
  width: 100%;
}

.upload-area :deep(.el-upload-dragger) {
  border: 2px dashed #d1d5db;
  border-radius: 14px;
  background: #f9fafb;
  padding: 20px;
  transition: all 0.2s;
}

.upload-area :deep(.el-upload-dragger:hover) {
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.05);
}

.upload-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  font-size: 14px;
  font-weight: 600;
}

.upload-icon {
  font-size: 28px;
  color: #9ca3af;
}

.upload-hint {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 400;
}

/* 新建按钮 */
.new-btn-area {
  padding: 12px;
}

.new-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  border-radius: 14px;
  cursor: pointer;
  font-weight: 700;
  font-size: 15px;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.new-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
}

/* 用户信息 */
.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #e8eaf0;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 10px;
  border-radius: 12px;
  transition: background 0.2s;
}

.user-info:hover {
  background: #f3f4f6;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
}

.username {
  color: #374151;
  font-size: 14px;
  font-weight: 600;
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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
  padding: 24px 40px;
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
  max-width: 500px;
}

.welcome-icon {
  width: 100px;
  height: 100px;
  margin: 0 auto 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 50px;
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.3);
}

.welcome h1 {
  font-size: 36px;
  font-weight: 800;
  color: #1f2937;
  margin: 0 0 12px;
  letter-spacing: -0.5px;
}

.welcome p {
  color: #6b7280;
  font-size: 18px;
  margin: 0 0 32px;
}

.suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: center;
}

.suggestion-chip {
  padding: 10px 18px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 20px;
  color: #4b5563;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.suggestion-chip:hover {
  background: #667eea;
  border-color: #667eea;
  color: #ffffff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
}

/* 消息 */
.messages-container {
  max-width: 800px;
  margin: 0 auto;
}

.messages {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.message {
  display: flex;
  gap: 14px;
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
  border-radius: 14px;
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
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff;
}

.message-content {
  max-width: 70%;
  padding: 16px 20px;
  border-radius: 20px;
  background: #ffffff;
  color: #1f2937;
  line-height: 1.6;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.message.user .message-content {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
}

.text-content {
  white-space: pre-wrap;
}

/* 打字动画 */
.typing {
  display: flex;
  gap: 4px;
  padding: 8px 0;
}

.dot {
  width: 8px;
  height: 8px;
  background: #9ca3af;
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
  padding: 20px 40px 28px;
  background: #ffffff;
  border-top: 1px solid #e8eaf0;
}

.input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
  background: #f9fafb;
  border: 2px solid #e5e7eb;
  border-radius: 20px;
  padding: 14px 16px;
  transition: all 0.2s;
  max-width: 900px;
  margin: 0 auto;
}

.input-wrapper:focus-within {
  border-color: #667eea;
  box-shadow: 0 0 0 4px rgba(102, 126, 234, 0.1);
}

.message-input {
  flex: 1;
}

.message-input :deep(.el-textarea__inner) {
  background: transparent;
  border: none;
  box-shadow: none;
  color: #1f2937;
  font-size: 16px;
  line-height: 1.6;
  resize: none;
  font-family: 'Nunito', sans-serif;
}

.message-input :deep(.el-textarea__inner::placeholder) {
  color: #9ca3af;
}

.send-btn {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 14px;
  color: #fff;
  font-size: 20px;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.send-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
}

/* 知识库视图 */
.knowledge-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fafbfc;
}

.knowledge-header {
  padding: 28px 40px 20px;
  background: #ffffff;
  border-bottom: 1px solid #e8eaf0;
}

.knowledge-header h2 {
  font-size: 24px;
  font-weight: 800;
  color: #1f2937;
  margin: 0 0 8px;
}

.knowledge-header p {
  color: #6b7280;
  font-size: 15px;
  margin: 0;
}

/* 文档概览 */
.doc-overview {
  padding: 20px 40px;
  background: #ffffff;
  border-bottom: 1px solid #e8eaf0;
}

.doc-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.05) 0%, rgba(5, 150, 105, 0.05) 100%);
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: 16px;
}

.doc-icon-large {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  background: linear-gradient(135deg, #10b981, #059669);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.doc-details h3 {
  font-size: 16px;
  font-weight: 700;
  color: #1f2937;
  margin: 0 0 4px;
}

.doc-details p {
  font-size: 14px;
  color: #6b7280;
  margin: 0;
}

/* 问答区域 */
.qa-area {
  flex: 1;
  overflow-y: auto;
  padding: 24px 40px;
}

.qa-empty {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  gap: 12px;
}

.qa-empty span {
  font-size: 48px;
}

.qa-empty p {
  font-size: 16px;
  margin: 0;
}

.qa-messages {
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.qa-message {
  display: flex;
}

.qa-message.user {
  justify-content: flex-end;
}

.qa-content {
  max-width: 70%;
  padding: 14px 18px;
  border-radius: 18px;
  background: #ffffff;
  color: #1f2937;
  line-height: 1.6;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.qa-message.user .qa-content {
  background: linear-gradient(135deg, #10b981, #059669);
  color: #ffffff;
}

/* 知识库输入 */
.knowledge-input {
  background: #ffffff;
}

.input-hint {
  text-align: center;
  color: #9ca3af;
  font-size: 13px;
  margin: 10px 0 0;
}

/* Markdown样式 */
.markdown-body {
  line-height: 1.7;
}

.markdown-body code {
  background: rgba(0, 0, 0, 0.06);
  padding: 2px 6px;
  border-radius: 6px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
}

.message.user .markdown-body code {
  background: rgba(255, 255, 255, 0.2);
}

.markdown-body pre {
  background: #f3f4f6;
  padding: 14px;
  border-radius: 12px;
  overflow-x: auto;
  margin: 10px 0;
}

.markdown-body pre code {
  background: none;
  padding: 0;
}
</style>
