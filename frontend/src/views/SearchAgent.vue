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
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
}

.title-section h1 {
  margin: 0 0 8px 0;
  color: #1a1a1a;
  font-size: 28px;
  font-weight: 600;
}

.subtitle {
  color: #666;
  margin: 0;
  font-size: 14px;
}

.filter-section {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-top: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.index-selector {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
  min-width: 280px;
}

.index-selector label {
  font-weight: 500;
  color: #333;
  font-size: 14px;
}

.dropdown-container {
  position: relative;
}

.dropdown-btn {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e1e5e9;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 15px;
  transition: all 0.3s ease;
  min-height: 44px;
}

.dropdown-btn:hover {
  border-color: #4a6cf7;
  box-shadow: 0 2px 8px rgba(74, 108, 247, 0.1);
}

.selected-text {
  color: #1a1a1a;
  flex: 1;
  text-align: left;
}

.dropdown-icon {
  color: #666;
  font-size: 12px;
}

.dropdown-list {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 8px;
  border: 2px solid #e1e5e9;
  border-radius: 8px;
  background: white;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  max-height: 300px;
  overflow-y: auto;
}

.dropdown-item {
  padding: 12px 16px;
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
  margin-bottom: 4px;
}

.item-desc {
  font-size: 12px;
  color: #888;
  font-weight: normal;
}

.check-mark {
  color: #4a6cf7;
  font-weight: bold;
  position: absolute;
  right: 16px;
}

.dropdown-empty {
  padding: 20px;
  text-align: center;
  color: #999;
  font-size: 14px;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.action-btn {
  padding: 10px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 6px;
  border: none;
  min-height: 44px;
}

.action-btn.primary {
  background: #4a6cf7;
  color: white;
}

.action-btn.primary:hover {
  background: #3a5ce5;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(74, 108, 247, 0.3);
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
  font-size: 16px;
}

/* 提示词编辑器 */
.prompt-editor-section {
  background: white;
  border-radius: 12px;
  padding: 0;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
  overflow: hidden;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #eee;
  background: #f8f9fa;
}

.editor-header h3 {
  margin: 0;
  color: #1a1a1a;
  font-size: 18px;
  font-weight: 600;
}

.close-editor-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #666;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.close-editor-btn:hover {
  background: #eee;
  color: #333;
}

.editor-body {
  padding: 24px;
}

.editor-controls {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.editor-field {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.editor-field.full-width {
  flex: 0 0 100%;
}

.editor-field label {
  font-weight: 500;
  color: #333;
  font-size: 14px;
}

.editor-field input,
.editor-field textarea {
  padding: 12px;
  border: 2px solid #e1e5e9;
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.3s ease;
}

.editor-field input:focus,
.editor-field textarea:focus {
  outline: none;
  border-color: #4a6cf7;
  box-shadow: 0 0 0 3px rgba(74, 108, 247, 0.1);
}

.editor-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}

/* 对话区域 */
.chat-section {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 300px);
  min-height: 500px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.chat-history {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.empty-conversation {
  text-align: center;
  padding: 60px 20px;
  color: #999;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-title {
  font-size: 20px;
  color: #666;
  margin-bottom: 8px;
  font-weight: 500;
}

.empty-desc {
  color: #999;
  font-size: 14px;
  max-width: 400px;
  line-height: 1.5;
}

/* 消息列表 */
.message-list {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.message-item {
  display: flex;
  gap: 16px;
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
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
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
  max-width: calc(100% - 56px);
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.message-role {
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

.message-time {
  font-size: 12px;
  color: #999;
}

.message-body {
  margin-bottom: 12px;
}

/* 用户消息 */
.user-message {
  background: #4a6cf7;
  color: white;
  padding: 12px 16px;
  border-radius: 12px 12px 4px 12px;
  line-height: 1.5;
  word-break: break-word;
}

/* AI回复 */
.assistant-message {
  background: #f8f9fa;
  color: #333;
  padding: 12px 16px;
  border-radius: 12px 12px 12px 4px;
  line-height: 1.5;
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
  width: 8px;
  height: 16px;
  background: #4a6cf7;
  margin-left: 4px;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

.markdown-content {
  line-height: 1.6;
}

.markdown-content h1,
.markdown-content h2,
.markdown-content h3 {
  margin-top: 16px;
  margin-bottom: 8px;
}

.markdown-content p {
  margin-bottom: 12px;
}

.markdown-content code {
  background: #f1f3f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Consolas', monospace;
  font-size: 0.9em;
}

.markdown-content pre {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 12px 0;
}

.markdown-content pre code {
  background: none;
  padding: 0;
}

/* 引用来源 */
.sources-section {
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px dashed #ddd;
}

.sources-title {
  font-size: 12px;
  color: #666;
  font-weight: 600;
  margin-bottom: 8px;
  text-transform: uppercase;
}

.sources-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.source-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: white;
  border-radius: 6px;
  border: 1px solid #e8e8e8;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 13px;
}

.source-item:hover {
  border-color: #4a6cf7;
  background: #f0f7ff;
  transform: translateX(2px);
}

.source-id {
  background: #4a6cf7;
  color: white;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: bold;
}

.source-title {
  flex: 1;
  color: #333;
}

.source-score {
  color: #4a6cf7;
  font-size: 11px;
  font-weight: 600;
  background: #f0f7ff;
  padding: 2px 6px;
  border-radius: 10px;
}

/* 消息操作 */
.message-actions {
  display: flex;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s;
}

.message-item:hover .message-actions {
  opacity: 1;
}

.message-action-btn {
  padding: 6px 10px;
  background: white;
  border: 1px solid #ddd;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  color: #666;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.2s;
}

.message-action-btn:hover {
  background: #f8f9fa;
  color: #4a6cf7;
  border-color: #4a6cf7;
}

/* 加载指示器 */
.loading-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 20px;
}

.loading-dots {
  display: flex;
  gap: 6px;
}

.loading-dots span {
  width: 8px;
  height: 8px;
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
  font-size: 14px;
}

/* 输入区域 */
.input-section {
  border-top: 1px solid #eee;
  padding: 16px 24px;
  background: #fafafa;
}

.input-container {
  background: white;
  border: 2px solid #e1e5e9;
  border-radius: 12px;
  padding: 16px;
  transition: all 0.3s ease;
}

.input-container:focus-within {
  border-color: #4a6cf7;
  box-shadow: 0 0 0 3px rgba(74, 108, 247, 0.1);
}

.message-input {
  width: 100%;
  border: none;
  outline: none;
  font-size: 15px;
  line-height: 1.5;
  resize: none;
  font-family: inherit;
  background: transparent;
}

.message-input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.input-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  flex-wrap: wrap;
  gap: 12px;
}

.input-controls {
  display: flex;
  gap: 8px;
}

.control-btn {
  padding: 8px 12px;
  background: white;
  border: 1px solid #ddd;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: #666;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.2s;
}

.control-btn:hover {
  background: #f8f9fa;
  color: #4a6cf7;
  border-color: #4a6cf7;
}

.send-controls {
  display: flex;
  gap: 8px;
  align-items: center;
}

.send-btn {
  padding: 10px 24px;
  font-size: 15px;
}

.sending-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.input-hints {
  margin-top: 12px;
  text-align: center;
}

.hint-text {
  font-size: 13px;
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
}

@media (max-width: 768px) {
  .ai-chat-container {
    padding: 16px;
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
  
  .chat-section {
    height: calc(100vh - 350px);
  }
}
</style>