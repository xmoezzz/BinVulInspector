export const TASK_STATUS = {
	queuing: 'queuing',
	processing: 'processing',
	finished: 'finished',
	failed: 'failed',
	terminated: 'terminated'
}

export const TASK_STATUS_OPTION = [
	{ label: 'Queuing', value: TASK_STATUS.queuing },
	{ label: 'Processing', value: TASK_STATUS.processing },
	{ label: 'Finished', value: TASK_STATUS.finished },
	{ label: 'Failed', value: TASK_STATUS.failed },
	{ label: 'Terminated', value: TASK_STATUS.terminated }
]
export const TASK_STATUS_LABEL_MAP = {
	[TASK_STATUS.queuing]: 'Queuing',
	[TASK_STATUS.processing]: 'Processing',
	[TASK_STATUS.finished]: 'Finished',
	[TASK_STATUS.failed]: 'Failed',
	[TASK_STATUS.terminated]: 'Terminated'
}
