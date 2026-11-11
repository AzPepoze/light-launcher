import type { IpcFieldError, IpcResponse } from "@shared";
import { createLogger } from "@lib/logger";

const log = createLogger("Transport");

const INVOKE_PATH = "/api/invoke";
const EVENTS_PATH = "/api/events";
const READY_PATH = "/api/ready";
const TOKEN_HEADER = "X-LL-Token";
const TOKEN_FRAGMENT_KEY = "t";

const MIN_RECONNECT_MS = 500;
const MAX_RECONNECT_MS = 10_000;

export interface LlShell {
	pickFolder?: () => Promise<string>;
	pickFile?: () => Promise<string>;
}

declare global {
	interface Window {
		llShell?: LlShell;
	}
}

export type EventCallback = (data: any) => void;

interface EventEnvelope {
	channel: string;
	data: unknown;
}

let token: string | undefined;
let streamStarted = false;
let reconnectDelay = MIN_RECONNECT_MS;
const listeners = new Map<string, Set<EventCallback>>();

/** Reads the token from the URL fragment once, then removes it from the address bar. */
function readTokenFromFragment(): string {
	const params = new URLSearchParams(window.location.hash.slice(1));
	const value = params.get(TOKEN_FRAGMENT_KEY) ?? "";
	window.history.replaceState(null, "", window.location.pathname + window.location.search);
	return value;
}

function getToken(): string {
	if (token === undefined) {
		token = readTokenFromFragment();
	}
	return token;
}

function authHeaders(extra: Record<string, string> = {}): Record<string, string> {
	return { ...extra, [TOKEN_HEADER]: getToken() };
}

async function readEnvelope<T>(response: Response, method: string): Promise<IpcResponse<T>> {
	const body = (await response.json().catch(() => undefined)) as IpcResponse<T> | undefined;
	if (!body) {
		throw new Error(`HTTP ${response.status} calling ${method}`);
	}
	return body;
}

/** A call the server refused. fields lists the inputs it rejected, when it said which ones. */
export class InvokeError extends Error {
	readonly fields: IpcFieldError[];

	constructor(message: string, fields: IpcFieldError[] = []) {
		super(message);
		this.name = "InvokeError";
		this.fields = fields;
	}
}

async function unwrapEnvelope<T>(response: Response, label: string): Promise<T> {
	const body = await readEnvelope<T>(response, label);
	if (!body.success) {
		throw new InvokeError(body.error || `IPC error calling ${label}`, body.fields);
	}
	return body.data as T;
}

/** Calls a Go method through POST /api/invoke. Throws with the server error text on failure. */
export async function invoke<T>(method: string, payload: unknown = {}): Promise<T> {
	const response = await fetch(INVOKE_PATH, {
		method: "POST",
		headers: authHeaders({ "Content-Type": "application/json" }),
		body: JSON.stringify({ method, payload })
	});
	return unwrapEnvelope<T>(response, method);
}

/** Downloads a server route as a blob, with the session token. Throws when the server refuses. */
export async function fetchBlob(path: string): Promise<Blob> {
	const response = await fetch(path, { headers: authHeaders() });
	if (!response.ok) {
		throw new Error(`HTTP ${response.status} loading ${path}`);
	}
	return response.blob();
}

/** Posts multipart form data to a server route. Throws with the server error text on failure. */
export async function postForm<T>(path: string, form: FormData): Promise<T> {
	const response = await fetch(path, {
		method: "POST",
		headers: authHeaders(),
		body: form
	});
	return unwrapEnvelope<T>(response, path);
}

/** Uses the native picker from window.llShell when the desktop shell provides one. */
export function pickWithShell(
	name: keyof LlShell,
	fallback: () => Promise<string>
): Promise<string> {
	const shell = window.llShell;
	const pick = shell?.[name];
	return pick ? pick.call(shell) : fallback();
}

/** Extracts the data lines of one SSE block. Returns undefined when the block has none. */
export function parseSseData(block: string): string | undefined {
	const lines = block
		.split("\n")
		.filter((line) => line.startsWith("data:"))
		.map((line) => line.slice("data:".length).replace(/^ /, ""));
	return lines.length > 0 ? lines.join("\n") : undefined;
}

function dispatchEvent(block: string): void {
	const raw = parseSseData(block);
	if (raw === undefined) return;

	let envelope: EventEnvelope;
	try {
		envelope = JSON.parse(raw) as EventEnvelope;
	} catch {
		log.warn("Dropped malformed event", raw);
		return;
	}

	const channelListeners = listeners.get(envelope.channel);
	channelListeners?.forEach((listener) => listener(envelope.data));
}

async function readEventStream(): Promise<void> {
	const response = await fetch(EVENTS_PATH, {
		headers: authHeaders({ Accept: "text/event-stream" })
	});
	if (!response.ok || !response.body) {
		throw new Error(`HTTP ${response.status} opening event stream`);
	}

	reconnectDelay = MIN_RECONNECT_MS;
	const reader = response.body.getReader();
	const decoder = new TextDecoder();
	let buffer = "";

	for (;;) {
		const { done, value } = await reader.read();
		if (done) return;

		buffer += decoder.decode(value, { stream: true });
		const blocks = buffer.split("\n\n");
		buffer = blocks.pop() ?? "";
		blocks.forEach(dispatchEvent);
	}
}

function sleep(ms: number): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, ms));
}

async function runEventStream(): Promise<void> {
	for (;;) {
		try {
			await readEventStream();
		} catch (error) {
			log.warn("Event stream failed, reconnecting", error);
		}
		await sleep(reconnectDelay);
		reconnectDelay = Math.min(reconnectDelay * 2, MAX_RECONNECT_MS);
	}
}

/** Subscribes to a server event channel. Returns a function that unsubscribes. */
export function onEvent(channel: string, callback: EventCallback): () => void {
	const channelListeners = listeners.get(channel) ?? new Set<EventCallback>();
	listeners.set(channel, channelListeners);
	channelListeners.add(callback);

	if (!streamStarted) {
		streamStarted = true;
		void runEventStream();
	}

	return () => {
		channelListeners.delete(callback);
	};
}

/** Tells the launcher the UI has loaded. */
export function signalReady(): void {
	const send = () => {
		fetch(READY_PATH, { method: "POST", headers: authHeaders() }).catch((error) =>
			log.warn("Ready signal failed", error)
		);
	};
	if (document.readyState === "complete") {
		send();
	} else {
		window.addEventListener("load", send, { once: true });
	}
}
