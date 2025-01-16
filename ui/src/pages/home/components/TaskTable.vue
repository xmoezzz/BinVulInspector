<template>
	<div class="flex flex-1 flex-col overflow-hidden">
		<el-table
			ref="tableCot"
			v-loading="loading"
			:data="tasks"
			style="width: 100%"
		>
			<el-table-column prop="project" label="Task Name " min-width="200">
				<template #default="{ row }">
					<el-tooltip
						v-if="row.name.length > 20"
						class="box-item"
						effect="dark"
						:content="row.name"
						placement="top-start"
					>
						<div
							class="task-type text-overflow line-clamp-2 text-ellipsis break-all"
							:class="row.source"
						>
							{{ row.name }}
						</div>
					</el-tooltip>
					<div
						v-else
						class="task-type text-overflow line-clamp-2 text-ellipsis break-all"
						:class="row.source"
					>
						{{ row.name }}
					</div>
				</template>
			</el-table-column>
			<el-table-column min-width="150">
				<template #header>
					<div style="display: flex; align-items: center">
						<span>Task State</span>
						<el-dropdown trigger="click" :hide-on-click="false">
							<el-icon
								style="margin-left: 5px; cursor: pointer"
								:color="
									statuses.length > 0
										? 'rgb(48, 75, 218)'
										: 'rgba(0, 0, 0, 0.6)'
								"
								><Filter
							/></el-icon>

							<template #dropdown>
								<el-dropdown-menu>
									<el-checkbox-group
										v-model="statuses"
										@change="handleStatusChange"
										style="
											display: flex;
											flex-direction: column;
											padding: 5px 10px;
										"
									>
										<el-checkbox
											v-for="option in TASK_STATUS_OPTION"
											:label="option.label"
											:value="option.value"
											:key="option.value"
										></el-checkbox>
									</el-checkbox-group>
								</el-dropdown-menu>
							</template>
						</el-dropdown>
					</div>
				</template>
				<template #default="{ row }">
					<span class="task-status" :class="row.status">{{
						TASK_STATUS_LABEL_MAP[row.status]
					}}</span>
				</template>
			</el-table-column>
			<el-table-column label="Inspect Mode">
				<template #default="{ row }">
					{{
						row.detail.bha.detection_method === 'intelligent' ? 'ML' : 'FAST'
					}}
				</template>
			</el-table-column>
			<el-table-column label="Algorithm">
				<template #default="{ row }">
					{{
						row.detail.bha.detection_method === 'intelligent'
							? row.detail.bha.algorithm
							: '-'
					}}
				</template>
			</el-table-column>
			<el-table-column label="Model Name">
				<template #default="{ row }">{{ getModelName(row) }}</template>
			</el-table-column>
			<el-table-column label="Model Category">
				<template #default="{ row }">
					{{
						row.detail.bha.detection_method === 'intelligent'
							? row.detail.bha.algorithm.toUpperCase()
							: '-'
					}}</template
				>
			</el-table-column>
			<!-- <el-table-column prop="status" label="检测结果" width="240">
			<template #default="{ row }">
				<div class="status-row">
					<taskResult :statistics="row.statistics" :row="row"></taskResult>
				</div>
			</template>
		</el-table-column>
		<el-table-column
			prop="uploaded_by"
			label="创建者"
			width="120"
		></el-table-column> -->
			<el-table-column label="Creation Time" min-width="110">
				<template #default="{ row }">
					{{
						row.status !== 'queuing'
							? dayjs(row.created_at).format('YYYY-MM-DD HH:mm:ss')
							: ''
					}}
				</template>
			</el-table-column>
			<el-table-column prop="modified_at" label="End Time" min-width="110">
				<template #default="{ row }">
					{{
						row.status !== 'queuing' && row.status !== 'processing'
							? dayjs(row.modified_at).format('YYYY-MM-DD HH:mm:ss')
							: ''
					}}
				</template>
			</el-table-column>
			<el-table-column label="Operation" min-width="230" fixed="right">
				<template #default="scope">
					<el-button
						link
						type="primary"
						size="small"
						@click.prevent="viewReport(scope.row)"
						><el-icon><Monitor /></el-icon>View Report</el-button
					>
					<el-popconfirm
						v-if="
							scope.row.status === 'finished' ||
							scope.row.status === 'failed' ||
							scope.row.status === 'terminated'
						"
						title="Are you sure to DELETE this?"
						@confirm="deleteTask(scope.row)"
					>
						<template #reference>
							<el-button link type="danger" size="small"
								><el-icon><Delete /></el-icon>Delete Task</el-button
							>
						</template>
					</el-popconfirm>
					<el-button
						v-if="
							scope.row.status !== 'finished' &&
							scope.row.status !== 'failed' &&
							scope.row.status !== 'terminated'
						"
						link
						size="small"
						type="primary"
						@click.prevent="terminateTask(scope.row)"
						><el-icon><SwitchButton /></el-icon>Terminate Task</el-button
					>
				</template>
			</el-table-column>
		</el-table>
		<el-pagination
			class="mt-4 justify-end"
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
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { MAXIMUM_PAGE_SIZE, PAGE_SIZES } from '@/domain/const'
import { TASK_STATUS_OPTION, TASK_STATUS_LABEL_MAP } from '@/domain/const/task'
import {
	deleteTaskAPI,
	getModelAPI,
	terminateTaskAPI
} from '@/http/business/task'
import { ROUTE_NAMES } from '@/router'

type PropType = {
	loading: boolean
	total: number
	tasks: ITask[]
	pager: { page: number; page_size: number }
}
const props = defineProps<PropType>()
const emits = defineEmits(['pageChange', 'sizeChange', 'search'])
const models = ref<IModel[]>([])

const router = useRouter()
const getModels = async (type: string) => {
	const params = {
		page: 1,
		page_size: MAXIMUM_PAGE_SIZE,
		types: type
	}
	const { list } = await getModelAPI(params)
	return list
}
const statuses = ref([])

const getModelName = (row: ITask) => {
	const model = models.value.find((item) => item.id === row.detail.bha.model_id)
	return model?.name || '-'
}

const handleSizeChange = (size: number) => {
	emits('sizeChange', size)
}
const handlePageChange = (page: number) => {
	emits('pageChange', page)
}
const handleStatusChange = () => {
	emits('search', { statuses: statuses.value })
}
const viewReport = (task: ITask) => {
	const { href } = router.resolve({
		name: ROUTE_NAMES.preview,
		params: {
			taskId: task.task_id
		},
		query: {
			type: task.detail.bha.detection_method,
			algorithm:
				task.detail.bha.detection_method === 'intelligent'
					? task.detail.bha.algorithm
					: '-',
			model: getModelName(task),
			modelType:
				task.detail.bha.detection_method === 'intelligent'
					? task.detail.bha.algorithm.toUpperCase()
					: '-',
			detectWay:
				task.detail.bha.detection_method === 'intelligent' ? 'ML' : 'FAST'
		}
	})
	window.open(href, '_blank')
}
const deleteTask = async (task: ITask) => {
	try {
		await deleteTaskAPI(task.task_id)
		ElMessage.success('删除成功')
		emits('search')
	} catch (err: any) {
		ElMessage.error(err.err_message)
	}
}
const terminateTask = async (task: ITask) => {
	try {
		await terminateTaskAPI(task.task_id)
		ElMessage.success('操作成功')
		emits('search')
	} catch (err: any) {
		ElMessage.error(err.err_message)
	}
}
onMounted(async () => {
	const bsdModels = await getModels('BSD')
	const ssfsModels = await getModels('SSFS')
	models.value = bsdModels.concat(ssfsModels)
})
</script>

<style scoped lang="scss">
.task-type {
	background-size: 18px;
	background-repeat: no-repeat;
	background-position: left center;
	padding-left: 24px;
	margin: 10px 0;
	background-image: url('@/assets/images/task/web.png');
	&.web {
		background-image: url('@/assets/images/task/web.png');
	}
	&.plugin {
		background-image: url('@/assets/images/task/plugin.png');
	}
	&.cli {
		background-image: url('@/assets/images/task/cli.png');
	}
	&.git {
		background-image: url('@/assets/images/task/git.png');
	}
	&.svn {
		background-image: url('@/assets/images/task/svn.png');
	}
}
.task-status {
	background-size: 18px;
	background-repeat: no-repeat;
	background-position: left center;
	padding: 10px 0 10px 24px;
	&.finished {
		background-image: url('@/assets/images/task/task-finished.png');
	}
	&.queuing {
		background-image: url('@/assets/images/task/task-queuing.png');
	}
	&.processing {
		background-image: url('@/assets/images/task/task-processing.png');
	}
	&.failed {
		background-image: url('@/assets/images/task/task-failed.png');
	}
	&.terminated {
		background-image: url('@/assets/images/task/task-terminated.png');
		background-size: 17px;
	}
}
</style>
