<template>
  <div class="ai-chat-container">
    <!-- 头部筛选区域 -->
    <div class="header-section">
      <div class="title-section">
        <h1>AI交互助手</h1>
        <p class="subtitle">选择文档索引和提示词以开始智能对话</p>
      </div>
      
      <div class="filter-section">
        <!-- Index选择栏 -->
        <div class="index-selector">
          <label for="index-select">文档索引：</label>
          <div class="dropdown-container">
            <button @click="toggleIndexDropdown" class="dropdown-btn">
              <span class="selected-text">{{ selectedIndex || '请选择文档索引' }}</span>
              <span class="dropdown-icon">{{ isIndexOpen ? '▲' : '▼' }}</span>
            </button>
            
            <div v-show="isIndexOpen" class="dropdown-list">
              <div
                v-for="index in Object.keys(mapIndexCount)"
                :key="index"
                @click="selectIndex(index)"
                class="dropdown-item"
                :class="{ active: selectedIndex === index }"
              >
                <span class="item-text">{{ index }}</span>
                <span v-if="selectedIndex === index" class="check-mark">✓</span>
              </div>
              <div v-if="Object.keys(mapIndexCount).length === 0" class="dropdown-empty">
                暂无文档索引
              </div>
            </div>
          </div>
        </div>
        
        <!-- Prompt选择栏 -->
        <div class="index-selector">
          <label for="prompt-select">提示词模板：</label>
          <div class="dropdown-container">
            <button @click="togglePromptDropdown" class="dropdown-btn">
              <span class="selected-text">{{ selectedPrompt || '请选择提示词模板' }}</span>
              <span class="dropdown-icon">{{ isPromptOpen ? '▲' : '▼' }}</span>
            </button>
            
            <div v-show="isPromptOpen" class="dropdown-list">
              <div
                v-for="prompt in availablePrompts"
                @click="selectPrompt(prompt)"
                class="dropdown-item"
                :class="{ active: selectedPrompt === prompt }"
              >
                <span class="item-text">{{ prompt }}</span>
                <span v-if="selectedPrompt === prompt" class="check-mark">✓</span>
              </div>
              <div v-if="availablePrompts.length === 0" class="dropdown-empty">
                暂无提示词模板
              </div>
            </div>
          </div>
        </div>
        
        <!-- 操作按钮 -->
        <div class="action-buttons">
          <button @click="clearConversation" class="action-btn secondary">
            <span class="btn-icon">🗑️</span>
            清空对话
          </button>
        </div>
      </div>
    </div>

    <!-- AI对话区域 -->
    <div class="chat-section">
      <!-- 对话历史 -->
      <div class="chat-history" ref="chatHistoryRef">
        <div v-if="conversation.length === 0" class="empty-conversation">
          <div class="empty-icon">🤖</div>
          <div class="empty-title">开始对话吧！</div>
          <div class="empty-desc">
            选择文档索引和提示词模板，然后在下方的输入框中输入您的问题
          </div>
        </div>
        
        <div v-else class="message-list">
          <div 
            v-for="(message, index) in conversation" 
            :key="index" 
            class="message-item"
            :class="message.role"
          >
            <div class="message-avatar">
              <div v-if="message.role === 'user'" class="avatar user-avatar">👤</div>
              <div v-else class="avatar assistant-avatar">🤖</div>
            </div>
            
            <div class="message-content">
              <div class="message-header">
                <span class="message-role">
                  {{ message.role === 'user' ? '您' : 'AI助手' }}
                </span>
                <span class="message-time">{{ formatTime(message.timestamp) }}</span>
              </div>
              
              <div class="message-body">
                <!-- 用户消息 -->
                <div v-if="message.role === 'user'" class="user-message">
                  {{ message.content }}
                </div>
                
                <!-- AI回复 -->
                <div v-else class="assistant-message">
                  <div v-if="message" class="streaming-response">
                    <div class="streaming-text">{{ message.content }}</div>
                    <div class="streaming-cursor"></div>
                  </div>
                </div>
              </div>
              
              <div class="message-actions">
                <button @click="copyMessage(message.content)" class="message-action-btn">
                  <span class="icon">📋</span> 复制
                </button>
              </div>
            </div>
          </div>
          
          <!-- 加载指示器 -->
          <div v-if="isLoading" class="loading-indicator">
            <div class="loading-dots">
              <span></span>
              <span></span>
              <span></span>
            </div>
            <div class="loading-text">AI正在思考中...</div>
          </div>
        </div>
      </div>
      
      <!-- 输入区域 -->
      <div class="input-section">
        <div class="input-container">
          <textarea
            v-model="userInput"
            @keydown.enter.exact.prevent="sendMessage"
            @keydown.enter.shift.exact.prevent="userInput += '\n'"
            placeholder="输入您的问题，按Enter发送，Shift+Enter换行"
            rows="3"
            :disabled="isLoading"
            class="message-input"
            ref="inputRef"
          ></textarea>
          
          <div class="input-actions">
            <div class="send-controls">
              <button @click="clearInput" class="action-btn secondary" :disabled="!userInput.trim()">
                清空
              </button>
              <button 
                @click="sendMessage" 
                class="action-btn primary send-btn"
                :disabled="isLoading || !userInput.trim() || !selectedIndex"
              >
                <span v-if="isLoading" class="sending-spinner"></span>
                <span v-else class="btn-icon">🚀</span>
                {{ isLoading ? '发送中...' : '发送' }}
              </button>
            </div>
          </div>
        </div>
        
        <div class="input-hints">
          <span class="hint-text">💡 提示：选择文档索引后,使用查询模式时,AI将基于索引中的文档内容回答问题</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import  request  from '../api/request'

interface message { 
    role: string,
    content: string,
    timestamp: Date,
}

// 响应式数据
const isIndexOpen = ref(false)
const isPromptOpen = ref(false)
const selectedIndex = ref('')
const selectedPrompt = ref<string>('')
const userInput = ref('')
const isLoading = ref(false)
const conversation = ref<message[]>([])

// 可用选项
const mapIndexCount = ref<Record<string, string>>({})
const availablePrompts = ref<string[]>([])

// 方法：下拉菜单控制
const toggleIndexDropdown = () => {
  isIndexOpen.value = !isIndexOpen.value
  isPromptOpen.value = false
}

const togglePromptDropdown = () => {
  isPromptOpen.value = !isPromptOpen.value
  isIndexOpen.value = false
}

// 获取文档索引列表
const getDocumentIndexList = async () => {
  try {
    const response = await request({
      url: '/api/documents/indices',
      method: 'GET',
    })
    mapIndexCount.value = response.data || {}
  } catch (error) {
    console.error('获取文档索引列表失败:', error)
    mapIndexCount.value = {}
  }
}


const selectIndex = (index: string) => {
  selectedIndex.value = index
  isIndexOpen.value = false
  // 当选择索引时，可以加载相关的提示词
  loadPromptsForIndex(index)
}


const loadPromptsForIndex = async (index: string) => {
  try {
    const response = await request({
      url: '/api/searchagent/setting',
      method: 'GET',
      params: { index }
    })
    availablePrompts.value = response.data || []
  } catch (error) {
    console.error('获取提示词列表失败:', error)
    availablePrompts.value = []
  }
}


const selectPrompt = (prompt: any) => {
  selectedPrompt.value = prompt
  isPromptOpen.value = false
}


// 方法：辅助功能
const clearConversation = () => {
  if (conversation.value.length > 0 && confirm('确定要清空对话历史吗？')) {
    conversation.value = []
  }
}

const clearInput = () => {
  userInput.value = ''
}

const copyMessage = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    alert('已复制到剪贴板')
  })
}

const formatTime = (date: Date) => {
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}


const sendMessage = async () => {
  if (!userInput.value.trim() || !selectedIndex.value || !selectedPrompt.value) {
    alert('请选择索引、提示词和输入问题')
    return
  }

  isLoading.value = true
  const message = {
    role: 'user',
    content: userInput.value,
    timestamp: new Date(),
  }
  conversation.value.push(message)

  const response = await request({
    url: '/api/searchagent',
    method: 'POST',
    data: {
      query: userInput.value,
      setting: selectedPrompt.value,
    }
  })
  if (response.data) {
    const message = {
      role: 'assistant',
      content: response.data,
      timestamp: new Date(),
    }
    conversation.value.push(message)
  }
  userInput.value = ''
  isLoading.value = false
}

onMounted(() => {
  getDocumentIndexList()
})

</script>

<style scoped>
.ai-chat-container {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
  background-color: #f8f9fa;
  min-height: 100vh;
}

/* 头部样式 - 复用主页样式 */
.header-section {
  background: white;
  border-radius: 12px;
  padding: 16px; /* 减小头部内边距 */
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  margin-bottom: 16px; /* 减小底部外边距 */
}

.title-section h1 {
  margin: 0 0 4px 0; /* 减小标题底部外边距 */
  color: #1a1a1a;
  font-size: 24px; /* 稍微减小标题字体大小 */
  font-weight: 600;
}

.subtitle {
  color: #666;
  margin: 0;
  font-size: 13px; /* 减小副标题字体大小 */
}

.filter-section {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-top: 12px; /* 减小筛选区域顶部外边距 */
  flex-wrap: wrap;
  gap: 12px; /* 减小筛选项之间的间隙 */
}

.index-selector {
  display: flex;
  flex-direction: column;
  gap: 6px; /* 减小索引选择器内部间隙 */
  flex: 1;
  min-width: 240px; /* 稍微减小最小宽度 */
}

.index-selector label {
  font-weight: 500;
  color: #333;
  font-size: 13px; /* 减小标签字体大小 */
}

.dropdown-container {
  position: relative;
}

.dropdown-btn {
  width: 100%;
  padding: 8px 12px; /* 大幅减小下拉按钮内边距 */
  border: 1px solid #e1e5e9; /* 减小边框宽度 */
  border-radius: 6px; /* 稍微减小圆角 */
  background: white;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px; /* 减小字体大小 */
  transition: all 0.3s ease;
  min-height: 36px; /* 减小最小高度 */
}

.dropdown-btn:hover {
  border-color: #4a6cf7;
  box-shadow: 0 2px 6px rgba(74, 108, 247, 0.1); /* 减小阴影范围 */
}

.selected-text {
  color: #1a1a1a;
  flex: 1;
  text-align: left;
}

.dropdown-icon {
  color: #666;
  font-size: 11px; /* 减小图标字体大小 */
}

.dropdown-list {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 4px; /* 减小下拉列表顶部外边距 */
  border: 1px solid #e1e5e9; /* 减小边框宽度 */
  border-radius: 6px; /* 减小圆角 */
  background: white;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1); /* 减小阴影范围 */
  z-index: 1000;
  max-height: 200px; /* 减小最大高度 */
  overflow-y: auto;
}

.dropdown-item {
  padding: 8px 12px; /* 减小下拉项内边距 */
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: background-color 0.2s;
  flex-direction: column;
  align-items: flex-start;
}

.dropdown-item:hover {
  background-color: #f5f7fa;
}

.dropdown-item.active {
  background-color: #f0f7ff;
  color: #4a6cf7;
}

.item-text {
  font-weight: 500;
  margin-bottom: 2px; /* 减小文本底部外边距 */
}

.item-desc {
  font-size: 11px; /* 减小描述字体大小 */
  color: #888;
  font-weight: normal;
}

.check-mark {
  color: #4a6cf7;
  font-weight: bold;
  position: absolute;
  right: 12px; /* 调整复选标记位置 */
}

.dropdown-empty {
  padding: 12px; /* 减小空状态内边距 */
  text-align: center;
  color: #999;
  font-size: 13px; /* 减小字体大小 */
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 8px; /* 减小按钮间隙 */
  align-items: center;
  flex-wrap: wrap;
}

.action-btn {
  padding: 6px 12px; /* 减小按钮内边距 */
  border-radius: 6px; /* 减小圆角 */
  cursor: pointer;
  font-size: 13px; /* 减小字体大小 */
  font-weight: 500;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 4px; /* 减小图标间隙 */
  border: none;
  min-height: 32px; /* 减小最小高度 */
}

.action-btn.primary {
  background: #4a6cf7;
  color: white;
}

.action-btn.primary:hover {
  background: #3a5ce5;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(74, 108, 247, 0.3); /* 减小阴影范围 */
}

.action-btn.secondary {
  background: white;
  color: #4a6cf7;
  border: 1px solid #4a6cf7;
}

.action-btn.secondary:hover {
  background: #f0f7ff;
  transform: translateY(-1px);
}

.btn-icon {
  font-size: 14px; /* 减小图标大小 */
}

/* 对话区域 - 大幅增加高度 */
.chat-section {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 220px); /* 大幅增加高度，从300px减到220px */
  min-height: 600px; /* 增加最小高度 */
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.chat-history {
  flex: 1;
  overflow-y: auto;
  padding: 16px; /* 适当减小内边距以增加内容区域 */
  display: flex;
  flex-direction: column;
  gap: 16px; /* 适当减小消息间隙 */
}

.empty-conversation {
  text-align: center;
  padding: 40px 20px; /* 适当减小空状态内边距 */
  color: #999;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.empty-icon {
  font-size: 48px; /* 适当减小图标大小 */
  margin-bottom: 12px;
  opacity: 0.5;
}

.empty-title {
  font-size: 18px; /* 适当减小标题大小 */
  color: #666;
  margin-bottom: 6px;
  font-weight: 500;
}

.empty-desc {
  color: #999;
  font-size: 13px; /* 适当减小描述字体大小 */
  max-width: 400px;
  line-height: 1.5;
}

/* 消息列表 */
.message-list {
  display: flex;
  flex-direction: column;
  gap: 16px; /* 适当减小消息间隙 */
}

.message-item {
  display: flex;
  gap: 12px; /* 适当减小头像间隙 */
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.message-item.user {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
}

.avatar {
  width: 32px; /* 适当减小头像大小 */
  height: 32px; /* 适当减小头像大小 */
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px; /* 适当减小头像内文字大小 */
}

.user-avatar {
  background: #4a6cf7;
  color: white;
}

.assistant-avatar {
  background: #f0f7ff;
  color: #4a6cf7;
}

.message-content {
  flex: 1;
  max-width: calc(100% - 44px); /* 适当调整最大宽度 */
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px; /* 减小头部底部外边距 */
}

.message-role {
  font-weight: 600;
  color: #333;
  font-size: 13px; /* 减小角色字体大小 */
}

.message-time {
  font-size: 11px; /* 减小时间字体大小 */
  color: #999;
}

.message-body {
  margin-bottom: 8px; /* 减小消息体底部外边距 */
}

/* 用户消息 */
.user-message {
  background: #4a6cf7;
  color: white;
  padding: 8px 12px; /* 减小内边距 */
  border-radius: 10px 10px 4px 10px; /* 适当减小圆角 */
  line-height: 1.4; /* 适当减小行高 */
  word-break: break-word;
}

/* AI回复 */
.assistant-message {
  background: #f8f9fa;
  color: #333;
  padding: 8px 12px; /* 减小内边距 */
  border-radius: 10px 10px 10px 4px; /* 适当减小圆角 */
  line-height: 1.4; /* 适当减小行高 */
  word-break: break-word;
}

.streaming-response {
  display: flex;
  align-items: center;
}

.streaming-text {
  flex: 1;
}

.streaming-cursor {
  width: 6px; /* 减小光标宽度 */
  height: 12px; /* 减小光标高度 */
  background: #4a6cf7;
  margin-left: 3px; /* 减小光标左边距 */
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* 加载指示器 */
.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px; /* 减小加载指示器间隙 */
  padding: 12px; /* 减小内边距 */
}

.loading-dots {
  display: flex;
  gap: 4px; /* 减小点之间间隙 */
}

.loading-dots span {
  width: 6px; /* 减小点大小 */
  height: 6px; /* 减小点大小 */
  border-radius: 50%;
  background: #4a6cf7;
  animation: bounce 1.4s infinite ease-in-out both;
}

.loading-dots span:nth-child(1) { animation-delay: -0.32s; }
.loading-dots span:nth-child(2) { animation-delay: -0.16s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1); }
}

.loading-text {
  color: #666;
  font-size: 13px; /* 减小字体大小 */
}

/* 输入区域 - 大幅降低高度 */
.input-section {
  border-top: 1px solid #eee;
  padding: 8px 16px; /* 大幅减小输入区域内边距 */
  background: #fafafa;
}

.input-container {
  background: white;
  border: 1px solid #e1e5e9; /* 减小边框宽度 */
  border-radius: 8px; /* 减小圆角 */
  padding: 8px; /* 大幅减小内边距 */
  transition: all 0.3s ease;
}

.input-container:focus-within {
  border-color: #4a6cf7;
  box-shadow: 0 0 0 2px rgba(74, 108, 247, 0.1); /* 减小阴影范围 */
}

.message-input {
  width: 100%;
  border: none;
  outline: none;
  font-size: 14px; /* 减小字体大小 */
  line-height: 1.4; /* 减小行高 */
  resize: none;
  font-family: inherit;
  background: transparent;
  min-height: 24px; /* 设置最小高度 */
}

.message-input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.input-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px; /* 减小操作区域顶部外边距 */
  flex-wrap: wrap;
  gap: 8px; /* 减小操作按钮间隙 */
}

.input-controls {
  display: flex;
  gap: 6px; /* 减小控制按钮间隙 */
}

.control-btn {
  padding: 4px 8px; /* 减小控制按钮内边距 */
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px; /* 减小圆角 */
  cursor: pointer;
  font-size: 12px; /* 减小字体大小 */
  color: #666;
  display: flex;
  align-items: center;
  gap: 3px; /* 减小图标间隙 */
  transition: all 0.2s;
}

.control-btn:hover {
  background: #f8f9fa;
  color: #4a6cf7;
  border-color: #4a6cf7;
}

.send-controls {
  display: flex;
  gap: 6px; /* 减小发送按钮间隙 */
  align-items: center;
}

.send-btn {
  padding: 6px 16px; /* 减小发送按钮内边距 */
  font-size: 14px; /* 减小字体大小 */
  min-height: 28px; /* 减小最小高度 */
}

.sending-spinner {
  width: 14px; /* 减小加载图标大小 */
  height: 14px; /* 减小加载图标大小 */
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.input-hints {
  margin-top: 6px; /* 减小提示文本顶部外边距 */
  text-align: center;
}

.hint-text {
  font-size: 12px; /* 减小字体大小 */
  color: #999;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .filter-section {
    flex-direction: column;
  }
  
  .index-selector {
    min-width: 100%;
  }
  
  .action-buttons {
    width: 100%;
    justify-content: center;
  }
  
  /* 响应式时调整聊天区域高度 */
  .chat-section {
    height: calc(100vh - 250px);
    min-height: 550px;
  }
}

@media (max-width: 768px) {
  .ai-chat-container {
    padding: 12px;
  }
  
  .message-item {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .message-item.user {
    flex-direction: column;
    align-items: flex-end;
  }
  
  .message-content {
    max-width: 100%;
  }
  
  .input-actions {
    flex-direction: column;
    align-items: stretch;
  }
  
  .input-controls,
  .send-controls {
    width: 100%;
    justify-content: center;
  }
  
  /* 移动端时进一步调整聊天区域高度 */
  .chat-section {
    height: calc(100vh - 280px);
    min-height: 450px;
  }
}
</style>