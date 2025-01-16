import { createRouter, createWebHistory } from 'vue-router'
// import registerRouteGuard from './Interceptor'
import routes from './routes'

const router = createRouter({
	history: createWebHistory(import.meta.env.VITE_APP_PUBLIC_PATH as string),
	routes
})

// 注册路由守卫
// registerRouteGuard(router)

export { ROUTE_NAMES } from './routes'
export default router
