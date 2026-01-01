<template>
  <div class="home-page">
    <h1>首页</h1>
    <p>这是路由式多页面的首页</p>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import request from '../api/request'

onMounted(async () => {
  const resp = await request('/api/docs', 'GET')
  console.log(resp)
})
</script>

<style scoped>
.home-page {
  padding: 20px;
}
</style>