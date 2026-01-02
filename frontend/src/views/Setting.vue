<template>
  <div class="container">
    <!-- 触发按钮 -->
    <button @click="toggleDropdown">
      {{ selectedOption || '请选择' }}
      <span>{{ isOpen ? '▲' : '▼' }}</span>
    </button>

    <!-- 下拉列表 -->
    <div v-show="isOpen" class="dropdown-list">
      <div
        v-for="option in options"
        :key="option.value"
        @click="selectOption(option)"
        class="dropdown-item"
      >
        {{ option.label }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const isOpen = ref(false)
const selectedOption = ref('')
const options = ref([
  { value: '1', label: '选项1' },
  { value: '2', label: '选项2' },
  { value: '3', label: '选项3' }
])

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
}

const selectOption = (option) => {
  selectedOption.value = option.label
  isOpen.value = false
}

// 点击外部关闭
const handleClickOutside = (event) => {
  if (!event.target.closest('.container')) {
    isOpen.value = false
  }
}

// 添加全局点击事件监听
import { onMounted, onUnmounted } from 'vue'
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.container {
  position: relative;
  width: 200px;
}

button {
  width: 100%;
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
  left: 0;
  right: 0;
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