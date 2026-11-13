<script lang="ts">
	import type { Snippet } from "svelte";
	import Icon from "./Icon.svelte";

	let {
		tone = "info",
		icon,
		role,
		children
	}: {
		tone?: "info" | "warning" | "error" | "success";
		icon?: string;
		/** Defaults to alert for errors and status for the rest. */
		role?: "alert" | "status";
		children?: Snippet;
	} = $props();

	const resolvedRole = $derived(role ?? (tone === "error" ? "alert" : "status"));
</script>

<div class="notice {tone}" role={resolvedRole}>
	{#if icon}
		<span class="notice-icon"><Icon name={icon} size={20} /></span>
	{/if}
	<div class="notice-body">
		{#if children}{@render children()}{/if}
	</div>
</div>

<style lang="scss">
	@use "./tones" as *;

	.notice {
		--tone: var(--accent-primary);
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 12px 16px;
		border: 2px solid color-mix(in srgb, var(--tone) 40%, transparent);
		border-radius: var(--radius-md);
		background: color-mix(in srgb, var(--tone) 8%, transparent);
		color: var(--text-main);
		font-size: 0.9rem;
		line-height: 1.45;

		@include tone-classes;
	}

	.notice-icon {
		color: var(--tone);
		padding-top: 1px;
	}

	.notice-body {
		min-width: 0;
		flex: 1;
	}
</style>
