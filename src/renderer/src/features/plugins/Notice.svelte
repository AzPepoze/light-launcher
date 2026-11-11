<script lang="ts">
	import type { Snippet } from "svelte";

	let { icon, children }: { icon: string; children: Snippet } = $props();
</script>

<div class="notice" role="note">
	<span class="material-icons" aria-hidden="true">{icon}</span>
	<div class="text">{@render children()}</div>
</div>

<style lang="scss">
	.notice {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 12px 14px;
		border-radius: var(--radius-md);
		border: 1px solid var(--danger);
		background: color-mix(in srgb, var(--danger) 12%, transparent);
		font-size: 1rem;
		line-height: 1.5;

		.material-icons {
			color: var(--danger);
			font-size: 1.3rem;
		}

		.text {
			min-width: 0;
		}

		:global(p) {
			margin: 0;
		}
	}
</style>
