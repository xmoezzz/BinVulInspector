<template>
	<el-dialog
		v-model="dialogFormVisible"
		title="Model Upload"
		width="800"
		:show-close="false"
	>
		<el-form ref="formRef" :model="modelForm" :rules="rules" class="px-6">
			<el-form-item label="Model Name" prop="name">
				<el-input
					v-model="modelForm.name"
					placeholder="Input Model Name"
					style="width: 400px"
				></el-input>
			</el-form-item>
			<el-form-item label="Model Category" prop="type">
				<el-select v-model="modelForm.type" style="width: 400px">
					<el-option label="SSFS" value="SSFS"></el-option>
					<el-option label="BSD" value="BSD"></el-option>
				</el-select>
			</el-form-item>
			<el-form-item label="Upload File" prop="file">
				<el-upload
					drag
					ref="uploadRef"
					:data="uploadData"
					action="/scs/api/v1/bha/model"
					:limit="1"
					name="upload_file"
					:auto-upload="false"
					v-model:file-list="fileList"
					:on-change="onFileChange"
					:before-upload="beforeUpload"
					:on-success="uploadSuccess"
					:on-error="uploadError"
					:show-file-list="true"
					:on-remove="handleRemove"
					class="upload-wrapper w-[600px]"
				>
					<template #default>
						<div class="flex flex-col items-center">
							<div class="flex w-[120px] items-center justify-center">
								<img src="@/assets/images/file.png" />
							</div>
							<div class="el-upload__text">
								<h4 class="upload-text">
									Drag file to this area or click to upload
								</h4>
								<p class="upload-tips">File limitation: 2G</p>
								<!-- <div class="upload-tips">
									<div>支持的格式：<strong>PE, ELF</strong> 二进制文件</div>
								</div> -->
							</div>
						</div>
					</template>
				</el-upload>
			</el-form-item>
		</el-form>
		<template #footer>
			<div class="dialog-footer">
				<el-button @click="handleCancelClick">Cancel</el-button>
				<el-button type="primary" @click="handleConfirmClick">
					Confirm
				</el-button>
			</div>
		</template>
	</el-dialog>
</template>

<script setup lang="ts">
import {
	FormInstance,
	UploadInstance,
	ElMessage,
	type UploadFile
} from 'element-plus'
import { ref } from 'vue'

const dialogFormVisible = defineModel()

const emits = defineEmits(['modelCreated'])

const formRef = ref<FormInstance>()
const uploadRef = ref<UploadInstance>()
const fileList = ref([])
const modelForm = ref<{ name: string; type: string; file: UploadFile | null }>({
	name: '',
	type: 'SSFS',
	file: null
})
const uploadData = ref()
const rules = {
	name: [
		{ required: true, message: 'Please input model name', trigger: 'blur' }
	],
	type: [
		{ required: true, message: 'Please select model type', trigger: 'change' }
	],
	file: [
		{ required: true, message: 'Please select model file', trigger: 'change' }
	]
}

const handleRemove = () => {
	modelForm.value.file = null
	fileList.value = []
}
const onFileChange = (file: UploadFile) => {
	modelForm.value.file = file
}
const beforeUpload = (file: any) => {
	let errMsg = ''
	const isLt100M = file.size / 1024 / 1024 / 1024 < 2
	if (!isLt100M) {
		errMsg = 'File size must less than 2G。'
	}
	if (errMsg) {
		ElMessage.warning(errMsg)
		return false
	}
	return true
}
const uploadSuccess = (res: any) => {
	if (res.code === 0) {
		formRef?.value?.resetFields()
		handleRemove()
		dialogFormVisible.value = false
		emits('modelCreated')
		ElMessage.success('Upload Success')
	} else {
		modelForm.value.file = null
		ElMessage.error(res.err_message || 'Upload Fail')
	}
}
const uploadError = (err: Error) => {
	ElMessage.error(err.message || 'Upload Fail')
}

const handleCancelClick = () => {
	formRef.value?.resetFields()
	dialogFormVisible.value = false
}
const handleConfirmClick = () => {
	formRef.value?.validate(async (valid) => {
		if (valid) {
			uploadData.value = {
				name: modelForm.value.name,
				type: modelForm.value.type
			}
			uploadRef.value!.submit()
		}
	})
}
</script>

<style></style>
