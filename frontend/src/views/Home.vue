<template>
  <div class="home-page">
    <div class="container">
      <p>文档索引列表:</p>

      <div class="dropdown-container">
        <button @click="toggleDropdown">
          {{ selectedOption || '请选择' }}
          <span>{{ isOpen ? '▲' : '▼' }}</span>
        </button>

        <!-- 下拉列表 -->
        <div v-show="isOpen" class="dropdown-list">
          <div
            v-for="index in indices"
            @click="selectOption(index)"
            class="dropdown-item"
          >
            {{ index }}
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted,onUnmounted} from 'vue'
import request from '../api/request'
import { ref } from 'vue'

const isOpen = ref(false)
const selectedOption = ref('')
const indices = ref<string[]>([])

// 获取文档索引列表
const getDocumentIndexList = async () => {
  try {
    const response = await request(`/api/documents/indices`, 'GET')
    console.log('response:', response)
    indices.value = response.data
    console.log('indices:', indices.value)
  } catch (error) {
    console.error('获取文档索引列表失败:', error)
  }
}

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
}

const selectOption = (index: string) => {
  selectedOption.value = index
  isOpen.value = false
}

// 点击外部关闭
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  const container = document.querySelector('.container')
  
  if (container && !container.contains(target)) {
    isOpen.value = false
  }
}

onMounted(() => {
  getDocumentIndexList()
})

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.container {
  display: flex;
  position: relative;
  /*控制 Flex 子元素在「主轴」上的对齐方式*/
  justify-content: flex-end;
  /*控制 Flex 子元素在「交叉轴」上的对齐方式*/ 
  align-items: center;
  gap: 1rem;
}

button {
  width: 200px;
  padding: 8px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dropdown-list {
  position: absolute;
  top: 100%;
  margin-top: 4px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background: white;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 1000;
}

.dropdown-item {
  padding: 8px 12px;
  cursor: pointer;
}

.dropdown-item:hover {
  background-color: #f5f7fa;
}
</style>