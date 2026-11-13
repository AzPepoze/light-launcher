<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import { isDirectLink } from "./addLink";
	import type { DownloadsState } from "./downloadsState.svelte";

	let { downloads }: { downloads: DownloadsState } = $props();

	let url = $state("");
	let input: HTMLInputElement | undefined = $state();

	const ready = $derived(isDirectLink(url));
	const canPauseAll = $derived(downloads.items.some((item) => item.status === "downloading"));
	const canResumeAll = $derived(
		downloads.items.some((item) => item.status === "paused" || item.status === "queued")
	);

	export function focusInput() {
		input?.focus();
	}

	/** Pasting a link fills the field with it, without spaces around it. */
	function onPaste(event: ClipboardEvent) {
		const text = event.clipboardData?.getData("text") ?? "";
		if (!isDirectLink(text)) return;
		event.preventDefault();
		url = text.trim();
	}

	function submitUrl(event: SubmitEvent) {
		event.preventDefault();
		if (!ready) return;
		downloads.enqueue(url.trim());
		url = "";
	}
</script>

<form class="bar" onsubmit={submitUrl}>
	<Icon name="add_link" size={22} />
	<input
		bind:this={input}
		bind:value={url}
		class="input"
		type="url"
		placeholder="Paste a direct link"
		aria-label="Direct link"
		onpaste={onPaste}
	/>
	<Button type="submit" variant="primary" size="sm" disabled={!ready}>Add</Button>
	<IconButton
		icon="pause"
		label="Pause all"
		disabled={!canPauseAll}
		onclick={() => downloads.pauseAll()}
	/>
	<IconButton
		icon="play_arrow"
		label="Resume all"
		disabled={!canResumeAll}
		onclick={() => downloads.resumeAll()}
	/>
</form>

{#if url.trim() && !ready}
	<p class="hint">Use a link that starts with http:// or https://.</p>
{/if}

<style lang="scss">
	.bar {
		display: flex;
		align-items: center;
		gap: 8px;
		margin: 0 0 8px;
		color: var(--text-muted);
	}

	.input {
		flex: 1;
		min-width: 0;
		font-size: 0.9rem;
	}

	.hint {
		margin: 0 0 16px;
		color: var(--text-muted);
		font-size: 0.8rem;
	}
</style>
