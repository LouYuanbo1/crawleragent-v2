// src/router.js
import { createRouter, createWebHistory } from 'vue-router'
// 导入页面组件（从 views/ 目录中）
import Home from './views/Home.vue'
import BossJob from './views/BossJob.vue'
import SearchAgentSetting from './views/SearchAgentSetting.vue'

// 定义路由规则：路径 ↔ 页面组件
const routes = [
  // 首页路由
  { path: '/', name: 'Home', component: Home },
  // 设置页路由
  { path: '/bossjob', name: 'BossJob', component: BossJob },
  // 搜索智能体路由
  { path: '/searchagent/setting', name: 'SearchAgentSetting', component: SearchAgentSetting },
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(), // HTML5 历史模式（无 # 号，更美观）
  routes // 注入路由规则
})

export default router