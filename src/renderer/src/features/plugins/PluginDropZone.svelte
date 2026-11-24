<script lang="ts">
	import type { Snippet } from "svelte";
	import Button from "@components/shared/Button.svelte";
	import DropOverlay from "@components/shared/DropOverlay.svelte";
	import { carriesFiles, DragDepth } from "../../lib/fileDrag";
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
	const dragDepth = new DragDepth();

	function onDragEnter(event: DragEvent) {
		if (!carriesFiles(event)) return;
		event.preventDefault();
		dragDepth.enter();
		dragging = true;
	}

	function onDragOver(event: DragEvent) {
		if (!carriesFiles(event)) return;
		event.preventDefault();
	}

	function onDragLeave(event: DragEvent) {
		if (!carriesFiles(event)) return;
		if (dragDepth.leave()) dragging = false;
	}

	async function onDrop(event: DragEvent) {
		if (!carriesFiles(event) || !event.dataTransfer) return;
		event.preventDefault();
		dragDepth.reset();
		dragging = false;
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
			<Button size="sm" onclick={() => fileInput?.click()}>{CHOOSE_FILE_LABEL}</Button>
			<Button size="sm" onclick={() => folderInput?.click()}>{CHOOSE_FOLDER_LABEL}</Button>
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
		<DropOverlay title={DROP_OVERLAY_TEXT} />
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
		font-size: 0.9375rem;
		color: var(--text-muted);
	}

	.actions {
		display: flex;
		gap: 8px;
	}

	.hidden-input {
		display: none;
	}
</style>
