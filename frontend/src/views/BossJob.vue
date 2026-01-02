<template>
  <div class="boss-job-page">
    <!-- 头部筛选区域 -->
    <div class="header-section">
      <div class="title-section">
        <h1>文档管理(专为Boss直聘信息使用)</h1>
        <p class="subtitle">选择文档索引以查看相关文档列表</p>
      </div>
      
      <div class="filter-section">
        <div class="index-selector">
          <label for="index-select">文档索引：</label>
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
        </div>
        
        <div class="pagination-controls">
          <div class="page-info">
            第 {{ page }} 页，每页 {{ size }} 条
          </div>
          <div class="page-buttons">
            <button @click="prevPage" :disabled="page <= 1" class="page-btn">
              上一页
            </button>
            <button @click="nextPage" :disabled="!hasMoreData" class="page-btn">
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 文档统计 -->
    <div class="stats-section" v-if="docs.length > 0">
      <div class="stat-card">
        <div class="stat-value">{{ mapIndexCount[selectedIndex] || 0 }}</div>
        <div class="stat-label">总文档数</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{(page - 1) * size + 1}} - {{(page - 1) * size  + docs.length}}</div>
        <div class="stat-label">当前文档</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ page }}</div>
        <div class="stat-label">当前页数</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ selectedIndex }}</div>
        <div class="stat-label">当前索引</div>
      </div>
    </div>

    <!-- 文档列表 -->
    <div class="document-list-container">
      <div v-if="docs.length === 0" class="empty-state">
        <div class="empty-icon">📄</div>
        <div class="empty-title">暂无文档数据</div>
        <div class="empty-desc">
          {{ selectedIndex ? '当前索引下暂无文档数据' : '请先选择文档索引' }}
        </div>
      </div>

      <div v-else class="document-grid">
        <div 
          v-for="(doc, index) in docs" 
          :key="doc.encryptJobId || index" 
          class="document-card"
        >
          <!-- 文档头部 -->
          <div class="card-header">
            <div class="job-title-section">
              <h3 class="job-title">
                {{ doc.jobName || '未命名文档' }}
                <span v-if="doc.salaryDesc" class="salary-tag">
                  {{ doc.salaryDesc }}
                </span>
              </h3>
              <div class="company-info">
                <span class="company-name">{{ doc.brandName || '未知公司' }}</span>
                <span class="company-size">{{ doc.brandScaleName }}</span>
              </div>
            </div>
            <div class="location-info">
              <span class="location-icon">📍</span>
              <span class="location-text">
                {{ doc.cityName || '未知城市' }}
                <span v-if="doc.areaDistrict">- {{ doc.areaDistrict }}</span>
              </span>
            </div>
          </div>

          <!-- 标签区域 -->
          <div class="tags-section" v-if="doc.jobLabels && doc.jobLabels.length">
            <div class="section-label">标签：</div>
            <div class="tags-list">
              <span 
                v-for="label in doc.jobLabels" 
                :key="label" 
                class="tag"
                :class="getTagClass(label)"
              >
                {{ label }}
              </span>
            </div>
          </div>

          <!-- 技能要求 -->
          <div class="skills-section" v-if="doc.skills && doc.skills.length">
            <div class="section-label">技能要求：</div>
            <div class="skills-list">
              <span 
                v-for="skill in doc.skills" 
                :key="skill" 
                class="skill-tag"
              >
                {{ skill }}
              </span>
            </div>
          </div>

          <!-- 福利待遇 -->
          <div class="welfare-section" v-if="doc.welfareList && doc.welfareList.length">
            <div class="section-label">福利待遇：</div>
            <div class="welfare-list">
              <span 
                v-for="welfare in doc.welfareList.slice(0, 4)" 
                :key="welfare" 
                class="welfare-tag"
              >
                {{ welfare }}
              </span>
              <span 
                v-if="doc.welfareList.length > 4" 
                class="more-tag"
                @click="toggleWelfare(index)"
              >
                +{{ doc.welfareList.length - 4 }} 项福利
              </span>
            </div>
            <!-- 展开的福利列表 -->
            <div 
              v-if="expandedWelfare.includes(index)" 
              class="welfare-expanded"
            >
              <div class="expanded-list">
                <span 
                  v-for="welfare in doc.welfareList" 
                  :key="welfare" 
                  class="welfare-tag"
                >
                  {{ welfare }}
                </span>
              </div>
            </div>
          </div>

          <!-- 其他信息 -->
          <div class="other-info">
            <div class="info-row">
              <span class="info-label">经验要求：</span>
              <span class="info-value">{{ doc.jobExperience || '不限' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">学历要求：</span>
              <span class="info-value">{{ doc.jobDegree || '不限' }}</span>
            </div>
            <div v-if="doc.businessDistrict" class="info-row">
              <span class="info-label">商圈：</span>
              <span class="info-value">{{ doc.businessDistrict }}</span>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="card-footer">
            <a 
              :href="doc.detailAddress" 
              target="_blank" 
              class="detail-btn"
              v-if="doc.detailAddress"
            >
              查看详情
            </a>
            <button class="copy-btn" @click="copyDocumentInfo(doc)">
              复制信息
            </button>
          </div>
        </div>
      </div>

      <!-- 底部分页控制 -->
      <div v-if="docs.length > 0" class="pagination-bottom">
        <div class="page-info">
          第 {{ page }} 页，每页 {{ size }} 条，共 {{ mapIndexCount[selectedIndex] || 0 }} 条
        </div>
        <div class="page-buttons">
          <button @click="prevPage" :disabled="page <= 1" class="page-btn">
            上一页
          </button>
          <button @click="nextPage" :disabled="!hasMoreData" class="page-btn">
            下一页
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch, computed } from 'vue'
import request from '../api/request'

// 响应式数据
const isOpen = ref(false)
const selectedIndex = ref('')
const mapIndexCount = ref<Record<string, string>>({})
const docs = ref<any[]>([])
const page = ref(1)
const size = ref(10)
const loading = ref(false)
const expandedWelfare = ref<number[]>([])
const totalDocs = ref(0) // 总文档数

// 计算属性：判断是否还有更多数据
const hasMoreData = computed(() => {
  return totalDocs.value > page.value * size.value
})

// 方法
const toggleDropdown = () => {
  isOpen.value = !isOpen.value
}

const selectOption = (index: string) => {
  selectedIndex.value = index
  isOpen.value = false
  page.value = 1 // 重置页码
  docs.value = [] // 清空当前文档
  expandedWelfare.value = [] // 清空展开状态
  totalDocs.value = 0 // 重置总文档数
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  const container = document.querySelector('.dropdown-container')
  
  if (container && !container.contains(target)) {
    isOpen.value = false
  }
}

const getTagClass = (label: string) => {
  const classes: Record<string, string> = {
    '在校/应届': 'tag-intern',
    '本科': 'tag-degree',
    '硕士': 'tag-master'
  }
  return classes[label] || 'tag-default'
}

const toggleWelfare = (index: number) => {
  const idx = expandedWelfare.value.indexOf(index)
  if (idx > -1) {
    expandedWelfare.value.splice(idx, 1)
  } else {
    expandedWelfare.value.push(index)
  }
}

const copyDocumentInfo = async (doc: any) => {
  const info = `
职位名称: ${doc.jobName || '无'}
公司: ${doc.brandName || '无'}
薪资: ${doc.salaryDesc || '面议'}
地点: ${doc.cityName || '无'}${doc.areaDistrict ? ` - ${doc.areaDistrict}` : ''}
经验要求: ${doc.jobExperience || '不限'}
学历要求: ${doc.jobDegree || '不限'}
技能要求: ${doc.skills?.join('、') || '无'}
  `.trim()
  
  try {
    await navigator.clipboard.writeText(info)
    alert('职位信息已复制到剪贴板！')
  } catch (err) {
    console.error('复制失败:', err)
    alert('复制失败，请手动复制')
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

// 获取文档列表
const getDocumentList = async (pageNum: number, pageSize: number) => {
  if (!selectedIndex.value) {
    docs.value = []
    totalDocs.value = 0
    return
  }
  
  loading.value = true
  try {
    const response = await request({
      url: `/api/documents/${selectedIndex.value}`,
      method: 'GET',
      params: {
        page: pageNum,
        size: pageSize,
      }
    })
    
    if (response.code === 200 && Array.isArray(response.data)) {
      // 直接更新文档列表，而不是追加
      docs.value = response.data
      // 如果有分页信息，更新总文档数
      if (response.total !== undefined) {
        totalDocs.value = response.total
      } else {
        // 如果API没有返回总数，假设还有更多数据
        totalDocs.value = response.data.length < pageSize ? 
          (pageNum - 1) * pageSize + response.data.length : 
          pageNum * pageSize + 1
      }
    } else {
      console.error('返回数据格式错误:', response)
      docs.value = []
      totalDocs.value = 0
    }
  } catch (error) {
    console.error('获取文档列表失败:', error)
    docs.value = []
    totalDocs.value = 0
  } finally {
    loading.value = false
  }
}

const prevPage = () => {
  if (page.value > 1) {
    page.value -= 1
    getDocumentList(page.value, size.value)
  }
}

const nextPage = () => {
  if (hasMoreData.value) {
    page.value += 1
    getDocumentList(page.value, size.value)
  }
}

// 监听器
watch(selectedIndex, (newIndex) => {
  if (newIndex) {
    getDocumentList(page.value, size.value)
  } else {
    docs.value = []
    totalDocs.value = 0
  }
})

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
/* 添加底部分页样式 */
.pagination-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #eee;
  flex-wrap: wrap;
  gap: 16px;
}

.pagination-bottom .page-info {
  color: #666;
  font-size: 14px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 6px;
}

.pagination-bottom .page-buttons {
  display: flex;
  gap: 8px;
}

.pagination-bottom .page-btn {
  padding: 8px 16px;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
}

.pagination-bottom .page-btn:hover:not(:disabled) {
  border-color: #4a6cf7;
  color: #4a6cf7;
}

.pagination-bottom .page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 原有CSS保持不变 */
.boss-job-page {
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

.filter-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.index-selector {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 300px;
}

.index-selector label {
  font-weight: 500;
  color: #333;
  white-space: nowrap;
}

.dropdown-container {
  position: relative;
  flex: 1;
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

/* 分页控件 */
.pagination-controls {
  display: flex;
  align-items: center;
  gap: 20px;
}

.page-info {
  color: #666;
  font-size: 14px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 6px;
}

.page-buttons {
  display: flex;
  gap: 8px;
}

.page-btn {
  padding: 8px 16px;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
}

.page-btn:hover:not(:disabled) {
  border-color: #4a6cf7;
  color: #4a6cf7;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 统计区域 */
.stats-section {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  flex: 1;
  background: white;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #4a6cf7;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

/* 文档列表 */
.document-list-container {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-title {
  font-size: 18px;
  color: #333;
  margin-bottom: 8px;
  font-weight: 500;
}

.empty-desc {
  color: #999;
  font-size: 14px;
}

/* 文档网格 */
.document-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.document-card {
  background: white;
  border: 1px solid #e8e8e8;
  border-radius: 12px;
  padding: 20px;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.document-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  border-color: #4a6cf7;
}

/* 卡片头部 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.job-title-section {
  flex: 1;
}

.job-title {
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.salary-tag {
  background: linear-gradient(135deg, #4a6cf7, #6a8cff);
  color: white;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.company-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.company-name {
  color: #333;
  font-weight: 500;
}

.company-size {
  color: #666;
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.location-info {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #666;
  font-size: 14px;
  white-space: nowrap;
}

.location-icon {
  font-size: 12px;
}

/* 标签区域 */
.section-label {
  font-size: 14px;
  color: #666;
  font-weight: 500;
  margin-bottom: 8px;
}

.tags-section,
.skills-section,
.welfare-section {
  margin-top: 4px;
}

.tags-list,
.skills-list,
.welfare-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.tag-intern {
  background: #e6f7ff;
  color: #1890ff;
  border: 1px solid #91d5ff;
}

.tag-degree {
  background: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

.tag-master {
  background: #fff7e6;
  color: #fa8c16;
  border: 1px solid #ffd591;
}

.tag-default {
  background: #f5f5f5;
  color: #666;
  border: 1px solid #d9d9d9;
}

.skill-tag {
  padding: 4px 10px;
  background: #f0f7ff;
  color: #4a6cf7;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid #d6e4ff;
}

.welfare-tag {
  padding: 4px 10px;
  background: #fff0f6;
  color: #eb2f96;
  border-radius: 6px;
  font-size: 12px;
  border: 1px solid #ffadd2;
}

.more-tag {
  padding: 4px 10px;
  background: #f5f5f5;
  color: #666;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  border: 1px solid #d9d9d9;
  transition: all 0.2s;
}

.more-tag:hover {
  background: #e8e8e8;
}

.welfare-expanded {
  margin-top: 8px;
}

.expanded-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding-top: 8px;
  border-top: 1px dashed #eee;
}

/* 其他信息 */
.other-info {
  background: #fafafa;
  border-radius: 8px;
  padding: 12px;
  margin-top: 4px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  font-size: 14px;
}

.info-label {
  color: #666;
}

.info-value {
  color: #333;
  font-weight: 500;
}

/* 卡片底部 */
.card-footer {
  display: flex;
  gap: 12px;
  margin-top: auto;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

.detail-btn {
  flex: 1;
  padding: 10px 16px;
  background: #4a6cf7;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  text-align: center;
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.detail-btn:hover {
  background: #3a5ce5;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(74, 108, 247, 0.3);
}

.copy-btn {
  flex: 1;
  padding: 10px 16px;
  background: white;
  color: #4a6cf7;
  border: 1px solid #4a6cf7;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.copy-btn:hover {
  background: #f0f7ff;
  transform: translateY(-1px);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .boss-job-page {
    padding: 16px;
  }
  
  .filter-section {
    flex-direction: column;
    align-items: stretch;
  }
  
  .index-selector {
    min-width: 100%;
  }
  
  .document-grid {
    grid-template-columns: 1fr;
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .card-footer {
    flex-direction: column;
  }
  
  .pagination-bottom {
    flex-direction: column;
    align-items: stretch;
    text-align: center;
  }
  
  .stats-section {
    flex-direction: column;
  }
}
</style>