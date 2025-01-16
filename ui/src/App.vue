<template>
	<el-config-provider :locale="en">
		<page-header></page-header>
		<div class="main-container flex flex-col px-[40px] py-[20px]">
			<router-view></router-view>
		</div>
	</el-config-provider>
</template>
<script setup lang="ts">
import { onMounted } from 'vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import PageHeader from './components/Header.vue'

// import { getAPI } from './http/axios'

onMounted(() => {
	// getAPI('/scs/api/v1/tasks', {
	// 	page: 1,
	// 	page_size: 10
	// })
})
const onClick = () => {
	console.log(1)
}
</script>

<style scoped lang="scss">
.main-container {
	height: calc(100% - 70px);
}
</style>
