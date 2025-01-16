<template>
	<div
		class="tbold flex h-14 items-center justify-between bg-[var(--info-background-color)] px-4"
	>
		<el-breadcrumb separator="/">
			<el-breadcrumb-item :to="{ path: '/' }">Tasks</el-breadcrumb-item>
			<el-breadcrumb-item>Task Report</el-breadcrumb-item>
		</el-breadcrumb>
		<el-button type="primary" size="small" @click="getLog"
			><el-icon><Download /></el-icon>&nbsp;Logs</el-button
		>
	</div>
	<div class="infos mt-4 bg-[var(--info-background-color)] p-4">
		<el-descriptions title="Basic Information">
			<el-descriptions-item label="Task Name:">{{
				task?.name
			}}</el-descriptions-item>
			<el-descriptions-item label="File Size:">
				<el-tag size="small"
					>{{ ((task?.file_size || 0) / 1024 / 1024).toFixed(2) }} M</el-tag
				>
			</el-descriptions-item>
			<el-descriptions-item label="Creation Date:">{{
				dayjs(task?.created_at).format('YYYY-MM-DD HH:mm:ss')
			}}</el-descriptions-item>
			<el-descriptions-item label="Inspect Mode:">
				<span>{{ route.query.detectWay }}</span>
			</el-descriptions-item>
			<el-descriptions-item label="Algorithm:">
				<span>{{ route.query.algorithm }}</span>
			</el-descriptions-item>
			<el-descriptions-item label="Model Name:">
				<span>{{ route.query.model }}</span>
			</el-descriptions-item>
			<el-descriptions-item label="Model Category:">
				<span>{{ route.query.modelType }}</span>
			</el-descriptions-item>
			<el-descriptions-item label="Detection Duration:">
				{{ formatDuration(task?.modified_at, task?.created_at) }}
			</el-descriptions-item>
			<el-descriptions-item label="Task Description:">
				<span>{{ task?.desc }}</span>
			</el-descriptions-item>
		</el-descriptions>
	</div>

	<el-tabs
		v-model="activeName"
		@tab-click="handleClick"
		class="flex flex-1 flex-col overflow-hidden"
	>
		<el-tab-pane label="Matched Vulnerabilities" name="vuln">
			<vuln-result
				:task-id="taskId"
				:type="route.query.type as string"
			></vuln-result>
		</el-tab-pane>
		<el-tab-pane label="Disassemble Result" name="compile">
			<compile-result :task-id="taskId"></compile-result>
		</el-tab-pane>
	</el-tabs>
</template>
<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import { useRoute } from 'vue-router'
import { TabsPaneContext } from 'element-plus'
import dayjs from 'dayjs'
import { saveAs } from 'file-saver'
import VulnResult from './components/VulnResult.vue'
import CompileResult from './components/CompileResult.vue'
import {
	getTaskDetailAPI,
	getDetectFuncAPI,
	getTaskLogAPI
} from '@/http/business/task'
import { MAXIMUM_PAGE_SIZE } from '@/domain/const'

const route = useRoute()
const activeName = ref('vuln')
const props = defineProps<{ taskId: string }>()
const task = ref<ITask>()
const funcTotal = ref(0)
const funcs = ref<IFunc[]>([])

const handleClick = (tab: TabsPaneContext, event: Event) => {
	console.log(tab, event)
}
const formatDuration = (modifyAt?: string, createdAt?: string) => {
	if (!modifyAt || !createdAt) return ''
	const modifyDate = new Date(modifyAt)
	const createdDate = new Date(createdAt)
	const duration = modifyDate.getTime() - createdDate.getTime()

	const hours = Math.floor(duration / (1000 * 60 * 60))
	const minutes = Math.floor((duration % (1000 * 60 * 60)) / (1000 * 60))
	const seconds = Math.floor((duration % (1000 * 60)) / 1000)

	// 不足一小时时，不显示小时
	if (hours > 0) {
		return `${hours}h${minutes}m${seconds}s`
	}
	if (minutes > 0) {
		return `${minutes}m${seconds}s`
	}
	return `${seconds}s`
}

const getLog = async () => {
	const content = await getTaskLogAPI(props.taskId)
	const blob = new Blob([content?.data], { type: 'application/json' })
	saveAs(blob, `${task.value?.name}.log`)
}

watchEffect(async () => {
	task.value = await getTaskDetailAPI(props.taskId)
	const { count, list } = await getDetectFuncAPI(props.taskId, {
		page: 1,
		page_size: MAXIMUM_PAGE_SIZE
	})
	funcTotal.value = count
	funcs.value = list
})
</script>

<style scoped lang="scss">
:deep(.el-descriptions__body) {
	background-color: var(--info-background-color);
}
:deep(.el-tab-pane) {
	height: 100%;
}
</style>
