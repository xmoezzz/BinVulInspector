<template>
	<div class="flex justify-between">
		<el-form inline ref="searchFormRef" :model="formData" @submit.prevent>
			<el-form-item prop="name">
				<el-input
					v-model="formData.name"
					placeholder="Input task name"
					@keyup.enter="handleSearchClick"
					style="width: 220px"
				></el-input>
			</el-form-item>
			<!-- <el-form-item prop="username">
				<el-input
					v-model="formData.username"
					placeholder="创建者账号名称"
					@keyup.enter="handleSearchClick"
				></el-input>
			</el-form-item> -->
			<!-- <el-form-item prop="source">
				<el-select
					v-model="formData.source"
					placeholder="请选择任务来源"
					clearable
					@change="handleSearchClick"
					style="width: 160px"
				>
					<el-option label="IDE" value="plugin"></el-option>
					<el-option label="手动上传" value="web"></el-option>
					<el-option label="CI/CD" value="cli"></el-option>
					<el-option label="Git" value="git"></el-option>
					<el-option label="SVN" value="svn"></el-option>
				</el-select>
			</el-form-item> -->
			<el-form-item prop="date">
				<el-date-picker
					v-model="formData.date"
					type="daterange"
					style="box-sizing: border-box; width: 320px"
					format="YYYY-MM-DD"
					value-format="YYYY-MM-DD"
					range-separator="To"
					start-placeholder="Start Date"
					end-placeholder="End Date"
				>
				</el-date-picker>
			</el-form-item>
			<el-form-item>
				<div class="flex flex-nowrap">
					<el-button type="primary" @click="handleSearchClick"
						><el-icon><Search /></el-icon> Search</el-button
					>
					<el-button @click="handleResetClick(searchFormRef)"
						><el-icon><RefreshRight /></el-icon> Reset</el-button
					>
				</div>
			</el-form-item>
		</el-form>

		<el-button type="primary" @click="handleCreateBtnClick"
			><i class="iconfont icon-create-task"></i>&nbsp; Create a new
			task</el-button
		>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import type { FormInstance } from 'element-plus'
import { ROUTE_NAMES } from '@/router'

const router = useRouter()

const emits = defineEmits(['search'])

const formData = ref({
	date: '',
	name: '',
	statuses: [],
	source: '',
	risk_level: '',
	username: ''
})
const searchFormRef = ref<FormInstance | null>(null)

const handleSearchClick = () => {
	emits('search', formData.value)
}
const handleResetClick = (formInstance: FormInstance | null) => {
	// 重置
	if (formInstance) {
		formInstance.resetFields()
		handleSearchClick()
	}
}

const handleCreateBtnClick = () => {
	router.push({
		name: ROUTE_NAMES.taskCreation
	})
}
</script>

<style></style>
