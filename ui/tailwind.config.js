/** @type {import('tailwindcss').Config} */
export default {
	content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
	theme: {
		// extend: {},
		extend: {
			colors: {
				critical: '#c9474b', // 致命
				high: '#f46e73', // 高危
				medium: '#fcd595', // 中危
				low: '#68a2eb', // 低危
				unknown: ' #aa8eee', // 未知
				secure: '#73c693' // 安全
			}
		}
	},
	plugins: []
}
