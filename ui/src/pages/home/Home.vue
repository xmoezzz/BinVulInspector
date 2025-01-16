<template>
	<search-bar @search="handleSearch"></search-bar>
	<task-table
		:total="total"
		:loading="loading"
		:tasks="tasks"
		:pager="pager"
		@search="handleSearch"
		@size-change="handleSizeChange"
		@page-change="handlePageChange"
	></task-table>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import SearchBar from './components/SearchBar.vue'
import TaskTable from './components/TaskTable.vue'
import { getTaskAPI } from '@/http/business/task'
import { TASK_STATUS } from '@/domain/const/task'

const loading = ref(false)
const tasks = ref<ITask[]>([])
const total = ref(0)
const pager = ref({
	page: 1,
	page_size: 10
})
const loopTimer = ref()
const formData = ref({
	date: '',
	name: '',
	statuses: [] as string[],
	source: '',
	risk_level: '',
	username: ''
})

const getTasks = async () => {
	const { date, name, statuses, source, risk_level, username } = formData.value
	const params = {
		page: pager.value.page,
		page_size: pager.value.page_size
	}
	if (date && date.length > 0) {
		Object.assign(params, {
			start_at: dayjs(`${date[0]} 00:00:00`).format(),
			ends_at: dayjs(`${date[1]} 23:59:59`).format()
		})
	}
	if (name) {
		Object.assign(params, { name })
	}
	if (statuses.length) {
		Object.assign(params, { statuses: statuses.join(',') })
	}
	if (source) {
		Object.assign(params, { source })
	}
	if (risk_level) {
		Object.assign(params, { risk_level })
	}
	if (username) {
		Object.assign(params, { username })
	}
	if (loopTimer.value) {
		clearInterval(loopTimer.value)
	}
	const loopRequestTask = async () => {
		const { count, list } = await getTaskAPI(params)
		total.value = count
		tasks.value = list
		const hasRunning = list.some(
			(item) =>
				item.status === TASK_STATUS.queuing ||
				item.status === TASK_STATUS.processing
		)
		if (hasRunning) {
			if (loopTimer.value) {
				clearInterval(loopTimer.value)
			}
			loopTimer.value = setInterval(loopRequestTask, 2000)
		}
	}

	loopRequestTask()
}

const handleSearch = (searchParam = {}) => {
	Object.assign(formData.value, searchParam)
	getTasks()
}

const handleSizeChange = (size: number) => {
	pager.value.page_size = size
	getTasks()
}
const handlePageChange = (page: number) => {
	pager.value.page = page
	getTasks()
}
onMounted(() => {
	getTasks()
})
onBeforeUnmount(() => {
	if (loopTimer.value) {
		clearInterval(loopTimer.value)
	}
})
</script>

<style></style>
