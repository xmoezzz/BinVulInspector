<template>
	<el-button type="primary" link @click="getLog">
		Obtain the disassemble result</el-button
	>
</template>

<script setup lang="ts">
import { saveAs } from 'file-saver'
import { getCompileLogAPI } from '@/http/business/task'

const props = defineProps<{ taskId: string }>()

const getLog = async () => {
	const content = await getCompileLogAPI(props.taskId)
	const blob = new Blob([content?.data], { type: 'application/json' })
	saveAs(blob, `${props.taskId}_反汇编结果.log`)
}
</script>

<style></style>
