// 获取任务列表接口参数
type TaskSearchParam = {
	page: number
	page_size: number
	type?: 'sca' | 'sast'
	task_ids?: string[]
	project?: string
	source?: 'web' | 'cli' | 'plugin' | 'CI'
	statuses?: string[]
	start_at?: string
	ends_at?: string
	asset_id?: string
	username?: string
}
type TaskCreatePayload = {
	mode: number // 0 上传扫描
	types: string[] // 任务模式
	name?: string
	desc?: string
	source?: string
	upload_file?: File
	extra?: string
}
type GetModelParam = {
	page: number
	page_size: number
	name?: string
	types: string
}
type GetFuncParam = {
	page: number
	page_size: number
	q?: string
}
type GetFuncResultParam = {
	page: number
	page_size: number
	func_id: string
	top_n?: number
	q?: string
}

type StatisticType = {
	type: string
	critical: number
	high: number
	low: number
	medium: number
	unknown: number
	secure: number
}
// 任务类型接口
interface ITask {
	asset_id: string
	branch: string
	created_at: string
	desc: string
	detail: {
		bha: {
			algorithm: string
			detection_method: string
			model_id: string
		}
	}
	file_hash: string
	file_path: string
	file_size: number
	modified_at: string
	name: string // 任务名
	source: string // 来源
	statistics: StatisticType[]
	status: string
	task_id: string
	types: string[]
	uploaded_by: string
	user_id: string
	version: string
}
interface IModel {
	created_at: string
	id: string
	modify_at: string
	name: string
	path: string
	type: string
}
interface IFunc {
	id: string
	fname: string
	task_id: string
	addr: string
	file_id: string
	file_path: string
}
interface IFuncResult {
	id: string
	arch: string
	cve: string
	fname: string
	func_id: string
	optlevel: string
	purl: string
	refs: string[]
	sim: string
	task_id: string
	version: string
}
