import { onEvent } from "@lib/api";
// `invoke` is not exported from lib/api yet; the transport owns it. Only this file imports it.
import { invoke } from "@lib/transport";

export function rpc<T>(method: string, payload: object = {}): Promise<T> {
	return invoke<T>(method, payload);
}

export function listen<T>(channel: string, handler: (payload: T) => void): () => void {
	return onEvent(channel, handler);
}

export function errorMessage(err: unknown): string {
	if (err instanceof Error && err.message) return err.message;
	return "Something went wrong.";
}
