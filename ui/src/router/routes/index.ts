import { RouteRecordRaw } from 'vue-router'

export const ROUTE_NAMES = {
	home: 'home',
	taskCreation: 'taskCreation',
	preview: 'preview'
}

const HomePage = () => import('../../pages/home/Home.vue')
const TaskCreation = () =>
	import('../../pages/taskCreation/TaskCreationPage.vue')

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		redirect: `/${ROUTE_NAMES.home}`
	},
	{
		path: `/${ROUTE_NAMES.home}`,
		name: ROUTE_NAMES.home,
		component: HomePage
	},
	{
		path: `/${ROUTE_NAMES.taskCreation}`,
		name: ROUTE_NAMES.taskCreation,
		component: TaskCreation
	},
	{
		path: `/${ROUTE_NAMES.preview}/:taskId`,
		name: ROUTE_NAMES.preview,
		props: true,
		component: () => import('../../pages/preview/PreviewPage.vue')
	}
]

export default routes
