<template>
  <div class="prompt-config-page">
    <!-- 头部区域 -->
    <div class="header-section">
      <div class="title-section">
        <h1>提示词配置</h1>
        <p class="subtitle">配置文档索引和提示词模板</p>
      </div>
    </div>

    <!-- 表单内容区域 -->
    <div class="form-container">
      <!-- 文档索引选择 -->
      <div class="form-section">
        <div class="section-header">
          <h2>文档索引设置</h2>
          <p class="section-desc">选择要关联的文档索引</p>
        </div>
        
        <div class="index-selector-container">
          <div class="form-group">
            <label class="form-label">
              文档索引：
              <span class="required">*</span>
            </label>
            <div class="dropdown-container">
              <button @click="toggleDropdown" class="dropdown-btn">
                <span class="selected-text">{{ selectedIndex || '请选择文档索引' }}</span>
                <span class="dropdown-icon">{{ isOpen ? '▲' : '▼' }}</span>
              </button>
              
              <!-- 下拉列表 -->
              <div v-show="isOpen" class="dropdown-list">
                <div
                  v-for="index in Object.keys(mapIndexCount)"
                  :key="index"
                  @click="selectOption(index)"
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
            <div class="form-hint">选择文档索引后，提示词将与该索引中的文档关联</div>
          </div>
        </div>
      </div>

      <!-- 提示词配置区域 -->
      <div class="form-section">
        <div class="section-header">
          <h2>提示词配置</h2>
          <p class="section-desc">为不同模式配置提示词模板</p>
        </div>

        <!-- EsRAG模式提示词 -->
        <div class="form-group">
          <label class="form-label">
            EsRAG模式提示词：
            <span class="required">*</span>
          </label>
          <div class="input-container">
            <textarea
              v-model="formData.promptEsRagMode"
              class="form-textarea"
              placeholder="请输入EsRAG模式下的提示词，例如：请根据以下文档内容回答问题："
              rows="5"
            ></textarea>
            <!--
            <div class="input-footer">
              <span class="char-count">{{ formData.promptEsRagMode.length }}/{{ maxPromptLength }}</span>
            </div>
            -->
          </div>
          <div class="form-hint">此提示词将用于文档检索增强生成模式</div>
        </div>

        <!-- Chat模式提示词 -->
        <div class="form-group">
          <label class="form-label">
            Chat模式提示词：
            <span class="required">*</span>
          </label>
          <div class="input-container">
            <textarea
              v-model="formData.promptChatMode"
              class="form-textarea"
              placeholder="请输入Chat模式下的提示词，例如：请以友好的方式回答以下问题："
              rows="5"
            ></textarea>
            <!--
            <div class="input-footer">
              <span class="char-count">{{ formData.promptChatMode.length }}/{{ maxPromptLength }}</span>
            </div>
            -->
          </div>
          <div class="form-hint">此提示词将用于普通对话模式</div>
        </div>

        <!-- 用户输入示例 -->
        <div class="form-group">
          <label class="form-label">
            用户输入示例(使用EsRAG模式时请在输入时加入"查询模式")：
          </label>
          <div class="input-container">
            <textarea
              v-model="formData.query"
              class="form-textarea"
              placeholder="请输入示例用户查询，用于测试提示词效果"
              rows="3"
            ></textarea>
            <!--
            <div class="input-footer">
              <span class="char-count">{{ formData.query.length }}/500</span>
            </div>
            -->
          </div>
          <div class="form-hint">可选的示例查询，用于测试提示词效果</div>
        </div>
      </div>

      <!-- 操作按钮区域 -->
      <div class="form-actions">
        <button class="btn-secondary" @click="resetForm">
          重置
        </button>
        <button class="btn-outline" @click="testConfiguration">
          测试配置
        </button>
        <button class="btn-primary" @click="submitForm" :disabled="!isFormValid">
          保存配置
        </button>
      </div>

      <!-- 测试结果展示区域 -->
      <div v-if="testResult" class="test-results">
        <div class="results-header">
          <h3>配置测试结果</h3>
          <button class="close-results" @click="testResult = ''">×</button>
        </div>
        <div class="results-content">
          <div class="result-item">
            <div class="result-label">输出结果：</div>
            <div class="result-value">{{ testResult }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import request from '../api/request'
import router from '../router'

// 响应式数据
const isOpen = ref(false)
const selectedIndex = ref('')
const mapIndexCount = ref<Record<string, string>>({})
const testResult = ref("")

// 表单数据
const formData = ref({
  promptEsRagMode: '',
  promptChatMode: '',
  query: ''
})

// 计算属性：表单验证
const isFormValid = computed(() => {
  return selectedIndex.value && 
         formData.value.promptEsRagMode.trim().length > 0 && 
         formData.value.promptChatMode.trim().length > 0
})

// 方法
const toggleDropdown = () => {
  isOpen.value = !isOpen.value
}

const selectOption = (index: string) => {
  selectedIndex.value = index
  isOpen.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  const container = document.querySelector('.dropdown-container')
  
  if (container && !container.contains(target)) {
    isOpen.value = false
  }
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

// 表单提交
const submitForm = async () => {
  if (!isFormValid.value) {
    alert('请填写所有必填字段')
    return
  }
  
  try {
    const response = await request({
      url: '/api/searchagent/setting',
      method: 'POST',
      data: {
      index: selectedIndex.value,
      promptEsRagMode: formData.value.promptEsRagMode,
      promptChatMode: formData.value.promptChatMode,
      }
    })
    
    if (response.code === 200) {
      alert('配置保存成功,三秒后跳转至搜索智能体页面')

      setTimeout(() => {
        router.push({ name: 'SearchAgent' })
      }, 3000)
      
    } else {
      alert('保存失败：' + response.msg)
    }
  } catch (error) {
    console.error('保存配置失败:', error)
    alert('保存配置失败，请检查网络连接')
  }
}

// 重置表单
const resetForm = () => {
  selectedIndex.value = ''
  formData.value = {
    promptEsRagMode: '',
    promptChatMode: '',
    query: ''
  }
  testResult.value = ""
}

// 测试配置
const testConfiguration = async () => {
  if (!isFormValid.value) {
    alert('请先填写必填字段')
    return
  }

try {
    const response = await request({
      url: '/api/searchagent/test',
      method: 'POST',
      data: {
      index: selectedIndex.value,
      ...formData.value
      }
    })
    
    if (response.code === 200) {
      testResult.value = response.data || ""
    } else {
      alert('测试失败：' + response.msg)
    }
  } catch (error) {
    console.error('测试失败:', error)
    alert('测试失败 ,请检查网络连接或模型连接')
  }
}

// 生命周期
onMounted(() => {
  getDocumentIndexList()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.prompt-config-page {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
  background-color: #f8f9fa;
  min-height: 100vh;
}

/* 头部样式 */
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

/* 表单容器 */
.form-container {
  background: white;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

/* 表单区域 */
.form-section {
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid #eee;
}

.form-section:last-child {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.section-header {
  margin-bottom: 24px;
}

.section-header h2 {
  margin: 0 0 8px 0;
  color: #1a1a1a;
  font-size: 20px;
  font-weight: 600;
}

.section-desc {
  color: #666;
  margin: 0;
  font-size: 14px;
}

/* 索引选择器容器 */
.index-selector-container {
  background: #fafafa;
  border-radius: 8px;
  padding: 20px;
}

/* 表单组 */
.form-group {
  margin-bottom: 24px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #333;
  font-size: 15px;
}

.required {
  color: #ff4757;
  margin-left: 4px;
}

/* 下拉选择器 */
.dropdown-container {
  position: relative;
  margin-bottom: 4px;
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
  margin-top: 4px;
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
}

.check-mark {
  color: #4a6cf7;
  font-weight: bold;
}

.dropdown-empty {
  padding: 20px;
  text-align: center;
  color: #999;
  font-size: 14px;
}

/* 文本输入区域 */
.input-container {
  position: relative;
}

.form-textarea {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e1e5e9;
  border-radius: 8px;
  font-size: 15px;
  font-family: inherit;
  resize: vertical;
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.form-textarea:focus {
  outline: none;
  border-color: #4a6cf7;
  box-shadow: 0 0 0 3px rgba(74, 108, 247, 0.1);
}

.form-textarea::placeholder {
  color: #aaa;
}

.input-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 4px;
}

.char-count {
  font-size: 12px;
  color: #999;
}

.form-hint {
  font-size: 13px;
  color: #777;
  margin-top: 6px;
  line-height: 1.4;
}

/* 表单操作按钮 */
.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 40px;
  padding-top: 24px;
  border-top: 1px solid #eee;
}

.btn-primary {
  padding: 12px 24px;
  background: #4a6cf7;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 500;
  transition: all 0.3s ease;
  min-width: 120px;
}

.btn-primary:hover:not(:disabled) {
  background: #3a5ce5;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(74, 108, 247, 0.3);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none !important;
  box-shadow: none !important;
}

.btn-secondary {
  padding: 12px 24px;
  background: #f8f9fa;
  color: #333;
  border: 1px solid #ddd;
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 500;
  transition: all 0.3s ease;
  min-width: 120px;
}

.btn-secondary:hover {
  background: #e9ecef;
  transform: translateY(-2px);
}

.btn-outline {
  padding: 12px 24px;
  background: white;
  color: #4a6cf7;
  border: 1px solid #4a6cf7;
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 500;
  transition: all 0.3s ease;
  min-width: 120px;
}

.btn-outline:hover {
  background: #f0f7ff;
  transform: translateY(-2px);
}

/* 测试结果区域 */
.test-results {
  margin-top: 32px;
  border: 1px solid #e1e5e9;
  border-radius: 12px;
  overflow: hidden;
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #f8f9fa;
  border-bottom: 1px solid #e1e5e9;
}

.results-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.close-results {
  background: none;
  border: none;
  font-size: 20px;
  color: #999;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-results:hover {
  background: #eee;
  color: #666;
}

.results-content {
  padding: 20px;
}

.result-item {
  display: flex;
  margin-bottom: 16px;
}

.result-item:last-child {
  margin-bottom: 0;
}

.result-label {
  font-weight: 500;
  color: #666;
  min-width: 150px;
  font-size: 14px;
}

.result-value {
  flex: 1;
  color: #333;
  background: #fafafa;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-all;
}

.result-status {
  flex: 1;
  display: flex;
  align-items: center;
}

.status-badge {
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}

.status-badge.success {
  background: #e7f7ef;
  color: #10b981;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .prompt-config-page {
    padding: 16px;
  }
  
  .form-container {
    padding: 20px;
  }
  
  .form-actions {
    flex-direction: column;
    align-items: stretch;
  }
  
  .btn-primary, .btn-secondary, .btn-outline {
    min-width: 100%;
    margin-bottom: 8px;
  }
  
  .btn-primary:last-child, .btn-secondary:last-child, .btn-outline:last-child {
    margin-bottom: 0;
  }
  
  .result-item {
    flex-direction: column;
    gap: 6px;
  }
  
  .result-label {
    min-width: auto;
  }
}
</style>