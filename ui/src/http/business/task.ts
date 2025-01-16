// 任务相关
import axios from 'axios'
import request from '@/http/axios'

/**
 * 任务列表
 * @param params
 * @returns 任务列表ITask[]
 */
export const getTaskAPI = (params: TaskSearchParam) =>
	request<ListResponse<ITask>>({
		url: '/scs/api/v1/tasks',
		method: 'get',
		params
	})

export const deleteTaskAPI = (id: string) =>
	request({
		url: `/scs/api/v1/tasks/${id}`,
		method: 'delete'
	})

export const terminateTaskAPI = (id: string) =>
	request({
		url: `/scs/api/v1/tasks/${id}/terminate`,
		method: 'post'
	})

export const createTaskAPI = (data: any) =>
	request({
		url: '/scs/api/v1/tasks',
		method: 'post',
		data
	})

export const getTaskDetailAPI = (id: string) =>
	request<ITask>({
		url: `/scs/api/v1/tasks/${id}`,
		method: 'get'
	})

export const getModelAPI = (params: GetModelParam) =>
	request<ListResponse<IModel>>({
		url: '/scs/api/v1/bha/model',
		method: 'get',
		params
	})

export const getDetectFuncAPI = (id: string, params: GetFuncParam) =>
	request<ListResponse<IFunc>>({
		url: `/scs/api/v1/bha/task/${id}/file/funcs`,
		method: 'get',
		params
	})

export const getFuncResultAPI = (id: string, params: GetFuncResultParam) =>
	request<ListResponse<IFuncResult>>({
		url: `/scs/api/v1/bha/task/${id}/file/func_results`,
		method: 'get',
		params
	})

export const getTaskLogAPI = (id: string) =>
	axios({
		url: `/scs/api/v1/tasks/${id}/bha/log`,
		method: 'get'
	})

export const postModelAPI = () =>
	request({
		url: '/scs/api/v1/bha/model',
		method: 'post'
	})

export const getCompileLogAPI = (id: string) =>
	axios({
		url: `/scs/api/v1/tasks/${id}/bha/asm_file`,
		method: 'get'
	})
