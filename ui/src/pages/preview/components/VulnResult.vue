<template>
	<div class="flex h-full flex-col">
		<el-table :data="funcs" class="overflow-auto">
			<el-table-column label="Function Name">
				<template #default="{ row }">
					<el-button @click="showDetail(row)" link type="primary">
						{{ row.fname }}</el-button
					>
				</template>
			</el-table-column>
			<el-table-column label="Path" prop="file_path">
				<template #default="{ row }"> {{ row.file_path }}1 </template>
			</el-table-column>
			<el-table-column label="Address" prop="addr"></el-table-column>
		</el-table>
		<el-pagination
			class="pagination mt-4 justify-end"
			size="small"
			background
			:current-page="pager.page"
			:page-size="pager.page_size"
			:page-sizes="PAGE_SIZES"
			layout="total, sizes, prev, pager, next, jumper, "
			:total="total"
			@size-change="handleSizeChange"
			@current-change="handlePageChange"
		>
		</el-pagination>
	</div>
	<result-detail
		:data="result"
		v-model="resultVisible"
		:type="type"
		@top-change="handleTopChange"
	></result-detail>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { PAGE_SIZES } from '@/domain/const'
import { getDetectFuncAPI, getFuncResultAPI } from '@/http/business/task'
import ResultDetail from './Detail.vue'

const props = defineProps<{ taskId: string; type: string }>()
const funcs = ref<IFunc[]>([])
const pager = ref({ page: 1, page_size: 10 })
const total = ref(0)
const result = ref<IFuncResult[]>([])
const getFuncs = async () => {
	const { count, list } = await getDetectFuncAPI(props.taskId, {
		page: pager.value.page,
		page_size: pager.value.page_size
	})
	total.value = count
	funcs.value = list
}
const resultVisible = ref(false)

const curFunc = ref<IFunc>()
const showDetail = async (func: IFunc, size = 10) => {
	const { list, count } = await getFuncResultAPI(props.taskId, {
		page: 1,
		page_size: size,
		func_id: func.id
	})
	curFunc.value = func
	result.value = list

	resultVisible.value = true
}
const handleTopChange = (size: number) => {
	showDetail(curFunc.value!, size)
}

const handleSizeChange = (size: number) => {
	pager.value.page_size = size
	getFuncs()
}
const handlePageChange = (page: number) => {
	pager.value.page = page
	getFuncs()
}

watch(
	() => props.taskId,
	() => {
		if (props.taskId) {
			getFuncs()
		}
	},
	{ immediate: true }
)
</script>

<style></style>
