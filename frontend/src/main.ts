import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router.ts' // 导入路由实例

createApp(App).use(router).mount('#app')
