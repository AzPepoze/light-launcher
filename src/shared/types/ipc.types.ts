/** One input the server rejected, and why. */
export interface IpcFieldError {
	field: string;
	message: string;
}

/** Reply body of POST /api/invoke. fields is set when the server rejected some inputs. */
export interface IpcResponse<T = unknown> {
	success: boolean;
	data?: T;
	error?: string;
	/** Stable machine-readable reason, such as method_not_found or auth_required. */
	code?: string;
	fields?: IpcFieldError[];
}
