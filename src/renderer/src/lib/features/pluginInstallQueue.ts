import { errorMessage } from "./rpc";
import type { UploadFile } from "./pluginUpload";

/**
 * idle: nothing queued. uploading: the next plugin is being sent to the server.
 * review: the install dialog is open for the plugin. failed: an upload failed and waits for
 * the user to dismiss it. done: every plugin was handled.
 */
export type QueuePhase = "idle" | "uploading" | "review" | "failed" | "done";

export interface QueueView {
	phase: QueuePhase;
	/** 1-based number of the plugin being handled, 0 when idle. */
	position: number;
	total: number;
	installed: number;
	/** Server path of the plugin under review. */
	source: string;
	failure: string;
}

export type UploadPlugin = (files: UploadFile[]) => Promise<string>;

const BUSY_PHASES: QueuePhase[] = ["uploading", "review", "failed"];

const IDLE_VIEW: QueueView = {
	phase: "idle",
	position: 0,
	total: 0,
	installed: 0,
	source: "",
	failure: ""
};

/**
 * Installs dropped plugins one after another. Each plugin is uploaded only when its turn
 * comes, and each one goes through its own consent dialog.
 */
export class PluginInstallQueue {
	readonly #upload: UploadPlugin;
	#groups: UploadFile[][] = [];
	#nextIndex = 0;
	#view: QueueView = IDLE_VIEW;
	readonly #listeners = new Set<(view: QueueView) => void>();

	constructor(upload: UploadPlugin) {
		this.#upload = upload;
	}

	get view(): QueueView {
		return this.#view;
	}

	/** True while a batch still has plugins to handle. A new batch must wait for it. */
	get busy(): boolean {
		return BUSY_PHASES.includes(this.#view.phase);
	}

	/** Calls the listener with each new view. Returns a function that stops it. */
	subscribe(listener: (view: QueueView) => void): () => void {
		this.#listeners.add(listener);
		return () => {
			this.#listeners.delete(listener);
		};
	}

	/** Starts a batch. Does nothing while another batch is still busy. */
	start(groups: UploadFile[][]): Promise<void> {
		if (this.busy || groups.length === 0) return Promise.resolve();
		this.#groups = groups;
		this.#nextIndex = 0;
		this.#setView({ ...IDLE_VIEW, total: groups.length });
		return this.#advance();
	}

	/** The user cancelled the current plugin. The queue goes on with the next one. */
	skip(): Promise<void> {
		if (this.#view.phase !== "review") return Promise.resolve();
		return this.#advance();
	}

	markInstalled(): Promise<void> {
		if (this.#view.phase !== "review") return Promise.resolve();
		this.#update({ installed: this.#view.installed + 1 });
		return this.#advance();
	}

	dismissFailure(): Promise<void> {
		if (this.#view.phase !== "failed") return Promise.resolve();
		return this.#advance();
	}

	async #advance(): Promise<void> {
		if (this.#nextIndex >= this.#groups.length) {
			this.#update({ phase: "done", source: "", failure: "" });
			return;
		}
		const files = this.#groups[this.#nextIndex];
		this.#nextIndex += 1;
		this.#update({ phase: "uploading", position: this.#nextIndex, source: "", failure: "" });
		await this.#uploadCurrent(files);
	}

	async #uploadCurrent(files: UploadFile[]): Promise<void> {
		try {
			const source = await this.#upload(files);
			this.#update({ phase: "review", source });
		} catch (err) {
			this.#update({ phase: "failed", failure: errorMessage(err) });
		}
	}

	#update(changes: Partial<QueueView>) {
		this.#setView({ ...this.#view, ...changes });
	}

	#setView(view: QueueView) {
		this.#view = view;
		for (const listener of this.#listeners) listener(view);
	}
}
