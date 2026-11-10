/** Reply body of POST /api/invoke. */
export interface IpcResponse<T = unknown> {
	success: boolean;
	data?: T;
	error?: string;
}
