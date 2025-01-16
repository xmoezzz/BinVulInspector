<template>
	<header>
		<el-breadcrumb separator="/">
			<el-breadcrumb-item :to="{ path: '/' }">Tasks</el-breadcrumb-item>
			<el-breadcrumb-item>Task Creation</el-breadcrumb-item>
		</el-breadcrumb>
	</header>

	<el-form
		:model="formData"
		label-width="160px"
		label-position="right"
		ref="ruleFormRef"
		:rules="rules"
		class="mt-10 px-14"
		style="display: flex; flex-direction: column"
	>
		<h3>Task Option</h3>
		<el-divider />
		<div class="flex flex-col items-center justify-center">
			<el-form-item prop="detectWay" label="Inspect Mode" class="w-[1000px]">
				<el-radio-group
					v-model="formData.detectWay"
					@change="handleDetectWayChange"
				>
					<el-radio
						v-for="way of WAYS"
						:label="way.label"
						:value="way.value"
						:key="way.value"
					></el-radio>
				</el-radio-group>
			</el-form-item>
			<el-form-item
				v-if="formData.detectWay === DETECT_WAY.intelligent"
				prop="algorithm"
				label="Algorithm"
				class="w-[1000px]"
			>
				<el-radio-group
					v-model="formData.algorithm"
					@change="handleAlgorithmChange"
				>
					<el-radio label="BSD" value="BSD"></el-radio>
					<el-radio label="SSFS" value="SSFS"></el-radio>
				</el-radio-group>
			</el-form-item>
			<el-form-item
				v-if="formData.detectWay === DETECT_WAY.intelligent"
				prop="model_id"
				label="Model Selection"
				class="w-[1000px]"
			>
				<el-select style="width: 400px" v-model="formData.model_id">
					<el-option
						v-for="model of models"
						:key="model.id"
						:label="model.name"
						:value="model.id"
					></el-option>
				</el-select>
				<el-button type="primary" link class="ml-2" @click="handleUploadClick"
					>Upload Model</el-button
				>
			</el-form-item>
		</div>

		<h3>File Selection</h3>
		<el-divider />
		<div class="flex flex-col items-center justify-center">
			<el-form-item
				prop="uploadFile"
				class="updata-box form-item w-[1000px]"
				label="Upload File"
			>
				<el-upload
					drag
					ref="uploadRef"
					:data="uploadData"
					action="/scs/api/v1/tasks"
					:limit="1"
					name="upload_file"
					:auto-upload="false"
					v-model:file-list="fileList"
					:on-change="onFileChange"
					:before-upload="beforeUpload"
					:on-success="UploadSuccess"
					:on-error="UploadError"
					:show-file-list="true"
					:on-exceed="handleExceed"
					:on-remove="handleRemove"
					class="upload-wrapper w-[600px]"
				>
					<template #default>
						<div class="flex flex-col items-center">
							<div class="flex w-[120px] items-center justify-center">
								<img src="@/assets/images/file.png" />
							</div>
							<div class="el-upload__text">
								<h4 class="upload-text">Drag or click this area to upload</h4>
								<p class="upload-tips">File size limitation: 100M</p>
								<div class="upload-tips">
									<div>
										Supported format:<strong
											>PE, ELF, Mach-O, binary executable</strong
										>
									</div>
								</div>
							</div>
						</div>
					</template>
				</el-upload>
			</el-form-item>
		</div>

		<h3>Other Information</h3>
		<el-divider />
		<div class="flex flex-col items-center justify-center">
			<el-form-item class="form-item w-[1000px]" label="Task Name" prop="name">
				<el-input
					v-model="formData.name"
					placeholder="Input task name"
					maxlength="64"
					style="width: 400px"
				/>
			</el-form-item>
			<el-form-item class="form-item w-[1000px]" label="Task Description">
				<el-input
					v-model="formData.desc"
					placeholder="Input task description"
					style="width: 400px"
				/>
			</el-form-item>

			<el-form-item class="form-item w-[1000px]">
				<el-button @click="router.push('/home')">Cancel</el-button>
				<el-button
					type="primary"
					@click="confirmClick(ruleFormRef, uploadRef)"
					:loading="creating"
					>Submit</el-button
				>
			</el-form-item>
		</div>
	</el-form>
	<model-upload
		v-model="dialogFormVisible"
		@modelCreated="getModels"
	></model-upload>
</template>

<script setup lang="ts">
import {
	ElMessage,
	FormInstance,
	FormRules,
	genFileId,
	UploadFile,
	UploadInstance,
	UploadProps,
	UploadRawFile
} from 'element-plus'
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getModelAPI } from '@/http/business/task'
import { MAXIMUM_PAGE_SIZE } from '@/domain/const'
import ModelUpload from './components/ModelUpload.vue'

const router = useRouter()

const DETECT_WAY = {
	fast: 'fast',
	intelligent: 'intelligent'
}
const WAYS = [
	{
		value: DETECT_WAY.fast,
		label: 'FAST'
	},
	{
		value: DETECT_WAY.intelligent,
		label: 'ML'
	}
]
const formLabelWidth = 120

const dialogFormVisible = ref(false)
const formData = ref<{
	uploadFile: UploadFile | null
	detectWay: string
	name: string
	desc: string
	model_id: string
	algorithm: string
}>({
	uploadFile: null,
	detectWay: DETECT_WAY.fast,
	name: '',
	desc: '',
	model_id: '',
	algorithm: 'BSD'
})
const creating = ref(false)
const ruleFormRef = ref<FormInstance | null>(null)
const uploadRef = ref<UploadInstance | null>(null)
const uploadData = ref({
	name: '',
	desc: '',
	types: ['bha'],
	extra: ''
})
const fileList = ref([])
const rules = reactive<FormRules>({
	name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
	model_id: [{ required: true, message: '请选择检测算法', trigger: 'change' }],
	algorithm: [{ required: true }],
	uploadFile: [
		{
			required: true,
			message: '请上传文件!',
			trigger: 'change'
		}
	],
	detectWay: [{ required: true, message: '', trigger: 'change' }]
})
const models = ref<IModel[]>([])

const handleUploadClick = () => {
	dialogFormVisible.value = true
}

const getModels = async () => {
	const params = {
		page: 1,
		page_size: MAXIMUM_PAGE_SIZE,
		types: formData.value.algorithm
	}
	const { list } = await getModelAPI(params)
	models.value = list
}

const handleDetectWayChange = async (way: string) => {
	if (way === DETECT_WAY.intelligent) {
		getModels()
	}
}
const handleAlgorithmChange = async () => {
	formData.value.model_id = ''
	getModels()
}

const onFileChange = (file: UploadFile) => {
	formData.value.uploadFile = file
	formData.value.name = file.name
}
const beforeUpload = (file: any) => {
	let errMsg = ''
	console.log(file.size)
	const isLt2g = file.size / 1024 / 1024 <= 100
	if (!isLt2g) {
		errMsg = '单个文件大小不能超过100M哦。'
	}
	if (errMsg) {
		ElMessage.warning(errMsg)
		return false
	}
	return true
}
const UploadSuccess = (res: any) => {
	console.log(res)
	if (res.code === 0) {
		ruleFormRef?.value?.resetFields()
		ElMessage.success('上传成功')
		fileList.value = []
		router.push('/home')
	} else {
		formData.value.uploadFile = null
		ElMessage.warning(res.err_message || '上传失败')
		fileList.value = []
	}
}
const UploadError = (err: any) => {
	const match = String(err).match(/\{.*\}/)
	if (match) {
		// 解析JSON字符串
		const jsonString = match[0]
		const errObj = JSON.parse(jsonString)
		const errMessage = errObj.err_message || '上传失败'
		ElMessage.warning(errMessage)
	} else {
		ElMessage.warning('上传失败')
	}
}
const handleExceed: UploadProps['onExceed'] = (files: any) => {
	uploadRef.value?.clearFiles()
	const file = files[0] as UploadRawFile
	file.uid = genFileId()
	uploadRef.value?.handleStart(file)
}
const handleRemove = () => {
	formData.value.uploadFile = null
}
const confirmClick = async (
	formInstance: FormInstance | null,
	uploadInstance: UploadInstance | null
) => {
	if (!formInstance || !uploadInstance) return
	await formInstance.validate(async (valid, fields) => {
		if (valid) {
			const extra = {
				bha: {
					detection_method: formData.value.detectWay,
					algorithm:
						formData.value.detectWay === DETECT_WAY.fast
							? 'SFS'
							: formData.value.algorithm
				}
			}
			if (formData.value.detectWay === DETECT_WAY.intelligent) {
				Object.assign(extra.bha, { model_id: formData.value.model_id })
			}
			uploadData.value = {
				name: formData.value.name,
				desc: formData.value.desc,
				types: ['bha'],
				extra: JSON.stringify(extra)
			}
			uploadInstance.submit()
		}
	})
}
</script>

<style></style>
