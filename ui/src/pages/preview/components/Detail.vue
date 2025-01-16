<template>
	<el-dialog
		v-model="dialogFormVisible"
		title="Matched Result"
		width="1200"
		style="max-height: 600px; overflow: auto"
	>
		<div class="mb-2 flex items-center justify-center" v-if="type !== 'fast'">
			<h3>Similarity Ratio TOP-</h3>
			<el-select
				style="width: 100px"
				class="ml-2"
				v-model="topK"
				@change="handleTopKChange"
			>
				<el-option :value="10" label="10"></el-option>
				<el-option :value="20" label="20"></el-option>
				<el-option :value="50" label="50"></el-option>
				<el-option :value="100" label="100"></el-option>
			</el-select>
		</div>
		<el-table :data="data" max-height="400px">
			<el-table-column
				prop="fname"
				label="Function Name"
				min-width="120px"
			></el-table-column>
			<el-table-column prop="cve" label="CVE"></el-table-column>
			<el-table-column prop="arch" label="Arch"></el-table-column>
			<el-table-column
				prop="optlevel"
				label="Optimization Level"
			></el-table-column>
			<el-table-column prop="refs" label="Source" min-width="200px">
				<template #default="{ row }">
					<div
						v-for="link of row.refs"
						:key="link"
						class="line-clamp-1 overflow-hidden text-ellipsis"
					>
						<span
							class="inline-block h-full cursor-pointer text-[var(--primary-color)]"
							@click="navigateTo(link)"
							>{{ link }}</span
						>
					</div>
				</template>
			</el-table-column>
			<el-table-column prop="sim" label="Similarity Ratio"></el-table-column>
		</el-table>

		<template #footer>
			<div class="dialog-footer">
				<el-button type="primary" @click="dialogFormVisible = false">
					Close
				</el-button>
			</div>
		</template>
	</el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ data: IFuncResult[]; type: string }>()
const emits = defineEmits(['topChange'])

const dialogFormVisible = defineModel()
const topK = ref(10)

const handleTopKChange = (top: number) => {
	emits('topChange', top)
}
const navigateTo = (url: string) => {
	window.open(url, '__blank')
}
</script>

<style></style>
