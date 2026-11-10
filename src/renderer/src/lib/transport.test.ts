import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

type Transport = typeof import("./transport");

const fakeWindow = () => ({
	location: { hash: "#t=secret", pathname: "/", search: "" },
	history: { replaceState: vi.fn() },
	addEventListener: vi.fn(),
	llShell: undefined as undefined | Record<string, unknown>
});

/** Loads a fresh transport module so the token and stream state start empty. */
async function loadTransport(): Promise<Transport> {
	vi.resetModules();
	return import("./transport");
}

function jsonResponse(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { "Content-Type": "application/json" }
	});
}

/** A stream that delivers the given text chunks and then stays open. */
function openStream(chunks: string[]): Response {
	const encoder = new TextEncoder();
	const body = new ReadableStream<Uint8Array>({
		start(controller) {
			for (const chunk of chunks) controller.enqueue(encoder.encode(chunk));
		}
	});
	return new Response(body, { status: 200 });
}

/** A stream that delivers the given text chunks and then ends. */
function closedStream(chunks: string[]): Response {
	const encoder = new TextEncoder();
	const body = new ReadableStream<Uint8Array>({
		start(controller) {
			for (const chunk of chunks) controller.enqueue(encoder.encode(chunk));
			controller.close();
		}
	});
	return new Response(body, { status: 200 });
}

function eventBlock(channel: string, data: unknown): string {
	return `data: ${JSON.stringify({ channel, data })}\n\n`;
}

function headersOf(call: unknown[]): Record<string, string> {
	const init = call[1] as RequestInit;
	return init.headers as Record<string, string>;
}

/** Lets pending promises and stream reads settle. */
async function flush(): Promise<void> {
	for (let i = 0; i < 10; i++) {
		await new Promise<void>((resolve) => setImmediate(resolve));
	}
}

describe("transport", () => {
	let win: ReturnType<typeof fakeWindow>;

	beforeEach(() => {
		win = fakeWindow();
		vi.stubGlobal("window", win);
		vi.stubGlobal("document", { readyState: "complete" });
	});

	afterEach(() => {
		vi.useRealTimers();
		vi.unstubAllGlobals();
	});

	describe("token", () => {
		it("reads the token from the URL fragment and clears the fragment", async () => {
			vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ success: true, data: 1 })));
			const { invoke } = await loadTransport();

			await invoke("GetPlatform");

			expect(win.history.replaceState).toHaveBeenCalledWith(null, "", "/");
			expect(headersOf(vi.mocked(fetch).mock.calls[0])["X-LL-Token"]).toBe("secret");
		});

		it("keeps the path and query when clearing the fragment", async () => {
			win.location.pathname = "/index.html";
			win.location.search = "?dev=1";
			vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ success: true, data: 1 })));
			const { invoke } = await loadTransport();

			await invoke("GetPlatform");

			expect(win.history.replaceState).toHaveBeenCalledWith(null, "", "/index.html?dev=1");
		});

		it("reads the fragment only once", async () => {
			vi.stubGlobal(
				"fetch",
				vi.fn().mockImplementation(async () => jsonResponse({ success: true, data: 1 }))
			);
			const { invoke } = await loadTransport();

			await invoke("First");
			win.location.hash = "#t=other";
			await invoke("Second");

			expect(win.history.replaceState).toHaveBeenCalledTimes(1);
			expect(headersOf(vi.mocked(fetch).mock.calls[1])["X-LL-Token"]).toBe("secret");
		});

		it("sends an empty token when the fragment has none", async () => {
			win.location.hash = "";
			vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ success: true, data: 1 })));
			const { invoke } = await loadTransport();

			await invoke("GetPlatform");

			expect(headersOf(vi.mocked(fetch).mock.calls[0])["X-LL-Token"]).toBe("");
		});
	});

	describe("invoke", () => {
		it("posts the method and payload as JSON and returns the data", async () => {
			const fetchMock = vi
				.fn()
				.mockResolvedValue(jsonResponse({ success: true, data: { os: "linux" } }));
			vi.stubGlobal("fetch", fetchMock);
			const { invoke } = await loadTransport();

			const result = await invoke<{ os: string }>("GetPlatform", { a: 1 });

			expect(result).toEqual({ os: "linux" });
			const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
			expect(url).toBe("/api/invoke");
			expect(init.method).toBe("POST");
			expect(headersOf(fetchMock.mock.calls[0])["Content-Type"]).toBe("application/json");
			expect(JSON.parse(init.body as string)).toEqual({ method: "GetPlatform", payload: { a: 1 } });
		});

		it("sends an empty object payload when none is given", async () => {
			const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ success: true, data: null }));
			vi.stubGlobal("fetch", fetchMock);
			const { invoke } = await loadTransport();

			await invoke("GetAllGames");

			const init = fetchMock.mock.calls[0][1] as RequestInit;
			expect(JSON.parse(init.body as string)).toEqual({ method: "GetAllGames", payload: {} });
		});

		it("throws the server error text when the call fails", async () => {
			vi.stubGlobal(
				"fetch",
				vi.fn().mockResolvedValue(jsonResponse({ success: false, error: "prefix exists" }, 500))
			);
			const { invoke } = await loadTransport();

			await expect(invoke("CreatePrefix", { name: "x" })).rejects.toThrow("prefix exists");
		});

		it("throws a fallback message when the failure has no error text", async () => {
			vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ success: false })));
			const { invoke } = await loadTransport();

			await expect(invoke("RunGame")).rejects.toThrow("IPC error calling RunGame");
		});

		it("throws an HTTP error when the body is not JSON", async () => {
			vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("nope", { status: 404 })));
			const { invoke } = await loadTransport();

			await expect(invoke("GetAllGames")).rejects.toThrow("HTTP 404 calling GetAllGames");
		});
	});

	describe("pickers", () => {
		it("uses the native picker from window.llShell when it exists", async () => {
			const fetchMock = vi.fn();
			vi.stubGlobal("fetch", fetchMock);
			const pickFolder = vi.fn().mockResolvedValue("/games");
			win.llShell = { pickFolder };
			const { pickWithShell } = await loadTransport();

			const result = await pickWithShell("pickFolder", () => Promise.resolve("fallback"));

			expect(result).toBe("/games");
			expect(pickFolder).toHaveBeenCalledOnce();
			expect(fetchMock).not.toHaveBeenCalled();
		});

		it("falls back when window.llShell has no such picker", async () => {
			win.llShell = {};
			const { pickWithShell } = await loadTransport();

			const result = await pickWithShell("pickFile", () => Promise.resolve("fallback"));

			expect(result).toBe("fallback");
		});

		it("falls back when window.llShell is missing", async () => {
			const { pickWithShell } = await loadTransport();

			const result = await pickWithShell("pickFolder", () => Promise.resolve("fallback"));

			expect(result).toBe("fallback");
		});
	});

	describe("events", () => {
		it("parses SSE data lines, including multi-line data and comments", async () => {
			const { parseSseData } = await loadTransport();

			expect(parseSseData(': keepalive\ndata: {"a":1}')).toBe('{"a":1}');
			expect(parseSseData("data: one\ndata: two")).toBe("one\ntwo");
			expect(parseSseData(": only a comment")).toBeUndefined();
		});

		it("delivers events only to the subscribed channel", async () => {
			vi.stubGlobal(
				"fetch",
				vi
					.fn()
					.mockResolvedValue(
						openStream([eventBlock("FilesDropped", ["a"]), eventBlock("Other", 2)])
					)
			);
			const { onEvent } = await loadTransport();
			const callback = vi.fn();

			onEvent("FilesDropped", callback);
			await flush();

			expect(callback).toHaveBeenCalledOnce();
			expect(callback).toHaveBeenCalledWith(["a"]);
		});

		it("handles an event split across two chunks", async () => {
			const block = eventBlock("Progress", { pct: 50 });
			const middle = Math.floor(block.length / 2);
			vi.stubGlobal(
				"fetch",
				vi.fn().mockResolvedValue(openStream([block.slice(0, middle), block.slice(middle)]))
			);
			const { onEvent } = await loadTransport();
			const callback = vi.fn();

			onEvent("Progress", callback);
			await flush();

			expect(callback).toHaveBeenCalledWith({ pct: 50 });
		});

		it("opens one shared stream for all subscriptions", async () => {
			const fetchMock = vi.fn().mockResolvedValue(openStream([]));
			vi.stubGlobal("fetch", fetchMock);
			const { onEvent } = await loadTransport();

			onEvent("A", () => {});
			onEvent("B", () => {});
			await flush();

			expect(fetchMock).toHaveBeenCalledOnce();
			const [url] = fetchMock.mock.calls[0] as [string];
			expect(url).toBe("/api/events");
			expect(headersOf(fetchMock.mock.calls[0])["X-LL-Token"]).toBe("secret");
		});

		it("stops delivering to a listener after it unsubscribes", async () => {
			vi.stubGlobal("fetch", vi.fn().mockResolvedValue(openStream([eventBlock("Tick", 1)])));
			const { onEvent } = await loadTransport();
			const kept = vi.fn();
			const removed = vi.fn();

			onEvent("Tick", kept);
			const unsubscribe = onEvent("Tick", removed);
			unsubscribe();
			await flush();

			expect(removed).not.toHaveBeenCalled();
			expect(kept).toHaveBeenCalledWith(1);
		});

		it("skips malformed event data without stopping the stream", async () => {
			vi.stubGlobal(
				"fetch",
				vi.fn().mockResolvedValue(openStream(["data: {bad json\n\n", eventBlock("Tick", 2)]))
			);
			const { onEvent } = await loadTransport();
			const callback = vi.fn();

			onEvent("Tick", callback);
			await flush();

			expect(callback).toHaveBeenCalledOnce();
			expect(callback).toHaveBeenCalledWith(2);
		});

		it("reconnects with backoff after the stream closes", async () => {
			vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
			const fetchMock = vi
				.fn()
				.mockResolvedValueOnce(closedStream([]))
				.mockResolvedValueOnce(openStream([eventBlock("Tick", 7)]));
			vi.stubGlobal("fetch", fetchMock);
			const { onEvent } = await loadTransport();
			const callback = vi.fn();

			onEvent("Tick", callback);
			await flush();
			expect(fetchMock).toHaveBeenCalledTimes(1);

			vi.advanceTimersByTime(500);
			await flush();

			expect(fetchMock).toHaveBeenCalledTimes(2);
			expect(callback).toHaveBeenCalledWith(7);
		});

		it("doubles the reconnect delay after each failure", async () => {
			vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
			const fetchMock = vi.fn().mockRejectedValue(new TypeError("network down"));
			vi.stubGlobal("fetch", fetchMock);
			const { onEvent } = await loadTransport();

			onEvent("Tick", () => {});
			await flush();
			expect(fetchMock).toHaveBeenCalledTimes(1);

			vi.advanceTimersByTime(499);
			await flush();
			expect(fetchMock).toHaveBeenCalledTimes(1);

			vi.advanceTimersByTime(1);
			await flush();
			expect(fetchMock).toHaveBeenCalledTimes(2);

			vi.advanceTimersByTime(999);
			await flush();
			expect(fetchMock).toHaveBeenCalledTimes(2);

			vi.advanceTimersByTime(1);
			await flush();
			expect(fetchMock).toHaveBeenCalledTimes(3);
		});
	});

	describe("signalReady", () => {
		it("posts to /api/ready with the token", async () => {
			const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
			vi.stubGlobal("fetch", fetchMock);
			const { signalReady } = await loadTransport();

			signalReady();

			const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
			expect(url).toBe("/api/ready");
			expect(init.method).toBe("POST");
			expect(headersOf(fetchMock.mock.calls[0])["X-LL-Token"]).toBe("secret");
		});

		it("waits for the load event when the page is still loading", async () => {
			const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
			vi.stubGlobal("fetch", fetchMock);
			vi.stubGlobal("document", { readyState: "loading" });
			const { signalReady } = await loadTransport();

			signalReady();
			expect(fetchMock).not.toHaveBeenCalled();

			const [, onLoad] = win.addEventListener.mock.calls[0] as [string, () => void];
			onLoad();
			expect(fetchMock).toHaveBeenCalledOnce();
		});
	});
});
