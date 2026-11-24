<script lang="ts">
	import Icon from "./Icon.svelte";

	let {
		title,
		hint = "",
		icon = "file_download",
	}: { title: string; hint?: string; icon?: string } = $props();
</script>

<div class="drop-overlay" aria-hidden="true">
	<div class="panel">
		<span class="badge">
			<Icon name={icon} size={26} filled />
		</span>
		<div class="copy">
			<p class="title">{title}</p>
			{#if hint}
				<p class="hint">{hint}</p>
			{/if}
		</div>
	</div>
</div>

<style lang="scss">
	.drop-overlay {
		position: fixed;
		inset: 0;
		z-index: 50;
		display: flex;
		align-items: center;
		justify-content: center;
		background: color-mix(in srgb, var(--bg-base) 55%, transparent);
		backdrop-filter: blur(6px);
		pointer-events: none;
		animation: fade-in var(--motion-quick) var(--ease-out) both;

		/* Dashed frame inset from the window edges, marking the whole page as a drop target. */
		&::before {
			content: "";
			position: absolute;
			inset: 14px;
			border: 1.5px dashed var(--glass-border-bright);
			border-radius: var(--radius-lg);
		}
	}

	.panel {
		display: flex;
		align-items: center;
		gap: 18px;
		max-width: min(480px, calc(100% - 64px));
		padding: 18px 26px 18px 18px;
		background: var(--bg-elevated);
		border: 1px solid var(--glass-border);
		border-radius: var(--radius-xl);
		box-shadow: var(--glass-shadow);
		animation: panel-in var(--motion-base) var(--ease-out) both;
	}

	.badge {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 52px;
		height: 52px;
		border-radius: var(--radius-pill);
		background: var(--text-main);
		color: var(--bg-base);
	}

	.copy {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}

	.title {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 600;
		color: var(--text-main);
	}

	.hint {
		margin: 0;
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	@keyframes fade-in {
		from {
			opacity: 0;
		}
	}

	@keyframes panel-in {
		from {
			opacity: 0;
			transform: translateY(10px) scale(0.97);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.drop-overlay,
		.panel {
			animation: none;
		}
	}
</style>
