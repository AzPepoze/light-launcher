<script lang="ts">
	import { onMount } from "svelte";
	import Button from "./Button.svelte";

	let {
		text,
		collapsedHeight = 160
	}: {
		text: string;
		/** Height in pixels shown before the person expands the text. */
		collapsedHeight?: number;
	} = $props();

	let body = $state<HTMLElement>();
	let expanded = $state(false);
	let overflowing = $state(false);

	function measure() {
		if (!body || expanded) return;
		overflowing = body.scrollHeight > body.clientHeight + 1;
	}

	// Re-measure when the text changes, and when the panel width changes the line count.
	$effect(() => {
		text;
		measure();
	});

	onMount(() => {
		const observer = new ResizeObserver(measure);
		if (body) observer.observe(body);
		return () => observer.disconnect();
	});
</script>

<div class="expandable">
	<div
		class="body"
		bind:this={body}
		style:max-height={expanded ? "none" : `${collapsedHeight}px`}
	>
		{text}
	</div>
	{#if overflowing || expanded}
		<Button
			variant="text"
			size="sm"
			icon={expanded ? "expand_less" : "expand_more"}
			aria-expanded={expanded}
			onclick={() => (expanded = !expanded)}
		>
			{expanded ? "Show less" : "Show more"}
		</Button>
	{/if}
</div>

<style lang="scss">
	.expandable {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 6px;
		min-width: 0;
	}

	.body {
		max-width: 68ch;
		width: 100%;
		overflow: hidden;
		white-space: pre-line;
		overflow-wrap: anywhere;
		font-size: 1rem;
		line-height: 1.5;
	}
</style>
