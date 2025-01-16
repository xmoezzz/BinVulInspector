import { App, createApp } from 'vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import router from './router'
import Entrance from './App.vue'
import './styles/index.scss'
import './assets/iconfont/iconfont.css'

const initializeApp = (app: App) => {
	app.use(router)
	for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
		app.component(key, component)
	}
	app.mount('#app')
}

const app = createApp(Entrance)
initializeApp(app)
