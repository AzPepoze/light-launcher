<script lang="ts">
	import type { Snippet } from "svelte";
	import Icon from "./Icon.svelte";

	let {
		tone = "neutral",
		icon,
		children
	}: {
		tone?: "neutral" | "accent" | "success" | "warning" | "danger";
		icon?: string;
		children?: Snippet;
	} = $props();
</script>

<span class="badge {tone}">
	{#if icon}<Icon name={icon} size={14} />{/if}
	{#if children}{@render children()}{/if}
</span>

<style lang="scss">
	@use "./tones" as *;

	.badge {
		--tone: var(--text-muted);
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 2px 10px;
		border: 1px solid color-mix(in srgb, var(--tone) 45%, transparent);
		border-radius: var(--radius-pill);
		background: color-mix(in srgb, var(--tone) 10%, transparent);
		color: var(--tone);
		font-size: 0.75rem;
		font-weight: 700;
		line-height: 1.6;
		white-space: nowrap;

		@include tone-classes;
	}
</style>
