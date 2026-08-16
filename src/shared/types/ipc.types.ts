export interface IpcResponse<T = unknown> {
	success: boolean;
	data?: T;
	error?: string;
}

export type IpcHandler = (payload?: any) => Promise<any> | any;
