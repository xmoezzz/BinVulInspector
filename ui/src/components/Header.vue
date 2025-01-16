<template>
	<header
		class="page-header bg-red flex h-[70px] items-center justify-center text-white"
	>
		<a @click="navTo('/home')" class="ml-10 cursor-pointer">
			BinVulInspector
		</a>

		<!-- <nav class="navs mr-[200px]">
			<a
				class="nav-item"
				:class="{ active: activeRoute === 'home' }"
				@click="navTo('/home')"
				>检测任务</a
			>
		</nav> -->
	</header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const activeRoute = ref('home')

const navTo = (path: string) => {
	router.push({
		path
	})
}
</script>

<style scoped lang="scss">
.page-header {
	// background: url('@/assets/images/header/header-bg.png') no-repeat;
	background-color: var(--el-color-primary);
	background-size: 100%;
	.logo {
		height: 40px;
		background-size: 32px;
		padding-left: 36px;
		line-height: 40px;
		background-position: left center;
		font-weight: bold;
	}
	.navs {
		.nav-item {
			color: #ffffff;
			font-size: 14px;
			text-decoration: none;
			cursor: pointer;
			position: relative;
			padding: 0 1vw;

			&.active,
			&:hover {
				&::before {
					content: '';
					position: absolute;
					bottom: -11px;
					left: 0;
					right: 0;
					margin: 0 auto;
					width: 48px;
					height: 2px;
					background: #ffffff;
				}
			}
			&:hover {
				&::before {
					background: #d4caca87;
				}
			}
		}
	}
}
</style>
