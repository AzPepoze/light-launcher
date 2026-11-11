<script lang="ts">
	import type { Snippet } from "svelte";
	import { collectDroppedFiles, filesFromInput } from "../../lib/features/pluginEntries";
	import type { UploadFile } from "../../lib/features/pluginUpload";
	import {
		CHOOSE_FILE_LABEL,
		CHOOSE_FOLDER_LABEL,
		DROP_HINT,
		DROP_OVERLAY_TEXT,
	} from "./pluginCopy";

	let {
		onFiles,
		children,
	}: { onFiles: (files: UploadFile[]) => void; children: Snippet } = $props();

	let dragging = $state(false);
	let fileInput: HTMLInputElement | undefined = $state();
	let folderInput: HTMLInputElement | undefined = $state();
	/** Counts nested dragenter and dragleave events, so the highlight does not flicker over children. */
	let dragDepth = 0;

	function carriesFiles(event: DragEvent): boolean {
		return event.dataTransfer?.types.includes("Files") ?? false;
	}

	function onDragEnter(event: DragEvent) {
		if (!carriesFiles(event)) return;
		event.preventDefault();
		dragDepth += 1;
		dragging = true;
	}

	function onDragOver(event: DragEvent) {
		if (!carriesFiles(event)) return;
		event.preventDefault();
	}

	function onDragLeave(event: DragEvent) {
		if (!carriesFiles(event)) return;
		dragDepth -= 1;
		if (dragDepth <= 0) stopDragging();
	}

	function stopDragging() {
		dragDepth = 0;
		dragging = false;
	}

	async function onDrop(event: DragEvent) {
		if (!carriesFiles(event) || !event.dataTransfer) return;
		event.preventDefault();
		stopDragging();
		onFiles(await collectDroppedFiles(event.dataTransfer));
	}

	/** Passes the picked files on and clears the input, so picking the same item again still fires. */
	function onPicked(input: HTMLInputElement) {
		const picked = Array.from(input.files ?? []);
		input.value = "";
		onFiles(filesFromInput(picked));
	}
</script>

<div
	class="drop-zone"
	class:dragging
	ondragenter={onDragEnter}
	ondragover={onDragOver}
	ondragleave={onDragLeave}
	ondrop={onDrop}
	role="region"
	aria-label="Plugin drop zone"
>
	<div class="bar">
		<p class="hint">{DROP_HINT}</p>
		<div class="actions">
			<button class="btn sm" type="button" onclick={() => fileInput?.click()}>
				{CHOOSE_FILE_LABEL}
			</button>
			<button class="btn sm" type="button" onclick={() => folderInput?.click()}>
				{CHOOSE_FOLDER_LABEL}
			</button>
		</div>
		<input
			bind:this={fileInput}
			class="hidden-input"
			type="file"
			accept=".llplugin,.zip"
			multiple
			onchange={() => fileInput && onPicked(fileInput)}
		/>
		<input
			bind:this={folderInput}
			class="hidden-input"
			type="file"
			webkitdirectory
			onchange={() => folderInput && onPicked(folderInput)}
		/>
	</div>

	{@render children()}

	{#if dragging}
		<div class="overlay" aria-hidden="true">
			<p>{DROP_OVERLAY_TEXT}</p>
		</div>
	{/if}
</div>

<style lang="scss">
	.drop-zone {
		position: relative;
		min-height: 100%;
	}

	.bar {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		margin-bottom: 16px;
	}

	.hint {
		flex: 1;
		min-width: 0;
		margin: 0;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.actions {
		display: flex;
		gap: 8px;
	}

	.hidden-input {
		display: none;
	}

	.overlay {
		position: fixed;
		inset: 0;
		z-index: 50;
		display: flex;
		align-items: center;
		justify-content: center;
		background: color-mix(in srgb, var(--bg-main) 70%, transparent);
		border: 2px dashed var(--glass-border-bright);
		pointer-events: none;

		p {
			margin: 0;
			font-size: 1.1rem;
			font-weight: 700;
			color: var(--text-main);
		}
	}
</style>
