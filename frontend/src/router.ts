// src/router.js
import { createRouter, createWebHistory } from 'vue-router'
// 导入页面组件（从 views/ 目录中）
import Home from './views/Home.vue'

// 定义路由规则：路径 ↔ 页面组件
const routes = [
  // 首页路由
  { path: '/', name: 'Home', component: Home },
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(), // HTML5 历史模式（无 # 号，更美观）
  routes // 注入路由规则
})

export default router