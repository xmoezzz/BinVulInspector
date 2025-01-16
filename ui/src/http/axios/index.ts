import axiosInstance from './axios'

export async function getAPI<T>(url: string, params?: any): Promise<T> {
	const response = await axiosInstance<T>({
		method: 'get',
		url,
		params
	})
	return response
}
export async function postAPI<T>(url: string, data?: any): Promise<T> {
	const response = await axiosInstance<T>({
		method: 'post',
		url,
		data
	})
	return response
}
export async function putAPI<T>(url: string, data?: any): Promise<T> {
	const response = await axiosInstance<T>({
		method: 'put',
		url,
		data
	})
	return response
}
export async function deleteAPI<T>(url: string, params?: any): Promise<T> {
	const response = await axiosInstance<T>({
		method: 'delete',
		url,
		params
	})
	return response
}

export default axiosInstance
