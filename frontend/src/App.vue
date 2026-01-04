<script setup lang="ts">
</script>

<template>
  <div class="top-nav">
    <router-link :to="{ name: 'Home' }" class="nav-item">首页</router-link>
    <router-link :to="{ name: 'BossJob' }" class="nav-item">Boss直聘岗位</router-link>
    <router-link :to="{ name: 'SearchAgentSetting' }" class="nav-item">搜索智能体设置</router-link>
    <router-link :to="{ name: 'SearchAgent' }" class="nav-item">搜索智能体</router-link>
  </div>
  <div class="router-view-container">
    <router-view></router-view>
  </div>
</template>

<style scoped>
.top-nav {
  display: flex;
  width: 100%;
  justify-content: space-between;
  padding: 0.625rem;
  background-color: #f5f5f5;
}
.nav-item {
  cursor: pointer;
}
.router-view-container {
  padding: 20px;
}
</style>