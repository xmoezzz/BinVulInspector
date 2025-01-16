import axios, { AxiosResponse } from 'axios'

export interface ApiResult<T> {
	code: number
	status: string
	message: string
	data: T
}

const instance = axios.create({
	baseURL: '/',
	timeout: 150000
})

instance.interceptors.response.use((res: AxiosResponse<any, any>) => {
	// if (res.config.responseType && res.config.responseType === 'blob') {
	// 	return res
	// }
	const status = Number(res.status) || 200
	if (status === 401) {
		const { pathname } = window.location
		if (pathname === '/') {
			throw new Error('未登录')
		}
		window.sessionStorage.clear()
		window.localStorage.clear()
		window.location.href = window.location.origin
		return res.data
	}
	if (status === 200) {
		if (res.data.code === 0 || res.data.status === 'success') {
			return Promise.resolve(res.data.data)
		}
		return Promise.reject(res.data)
	}
	return Promise.reject(res.data)
})

const wrappedInstance = <T>(...params: any) => {
	// eslint-disable-next-line
	return instance.apply(null, params) as Promise<T>
}

export default wrappedInstance
