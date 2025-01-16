import path from 'path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import ElementPlus from 'unplugin-element-plus/vite'
import eslint from 'vite-plugin-eslint'
// element-plus auto import
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

const DEV_ENV = 'http://10.240.17.35:8899'
// https://vitejs.dev/config/
export default defineConfig({
	resolve: {
		alias: {
			'~': path.resolve(__dirname, './src'),
			'@': path.resolve(__dirname, './src')
		}
	},
	plugins: [
		vue(),
		AutoImport({
			resolvers: [ElementPlusResolver()]
		}),

		Components({
			resolvers: [
				ElementPlusResolver({
					importStyle: 'sass' // 必须加，自定义颜色才生效
				})
			]
		}),
		ElementPlus({
			defaultLocale: 'zh-cn',
			useSource: true
		}),
		eslint({
			fix: true
		})
	],
	css: {
		preprocessorOptions: {
			scss: {
				additionalData: `@use "~/styles/element/index.scss" as *;`
			}
		}
	},

	server: {
		port: 8807,
		proxy: {
			'/scs/api': {
				target: DEV_ENV,
				secure: false // 忽略https证书校验
			}
		}
	}
})
