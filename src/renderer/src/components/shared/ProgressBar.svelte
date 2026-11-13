<script lang="ts">
	/* value is 0 to 1. null means unknown: a segment slides instead. */
	let {
		value,
		tone = "accent",
		active = false,
		label,
		height = 6
	}: {
		value: number | null;
		tone?: "accent" | "success" | "danger" | "muted";
		/** Adds a slow sheen to a running transfer. */
		active?: boolean;
		label: string;
		height?: number;
	} = $props();

	const clamped = $derived(value === null ? null : Math.min(1, Math.max(0, value)));
	const ariaNow = $derived(clamped === null ? undefined : Math.round(clamped * 100));
</script>

<div
	class="track {tone}"
	role="progressbar"
	aria-label={label}
	aria-valuemin={0}
	aria-valuemax={100}
	aria-valuenow={ariaNow}
	style:height="{height}px"
>
	{#if clamped === null}
		<span class="slide"></span>
	{:else}
		<span class="fill" style:width="{clamped * 100}%">
			{#if active}<span class="sheen"></span>{/if}
		</span>
	{/if}
</div>

<style lang="scss">
	@use "./tones" as *;

	.track {
		--tone: var(--accent-primary);
		position: relative;
		width: 100%;
		overflow: hidden;
		border-radius: var(--radius-pill);
		background: var(--bg-elevated);

		@include tone-classes;
	}

	.fill {
		position: relative;
		display: block;
		height: 100%;
		overflow: hidden;
		border-radius: var(--radius-pill);
		background: var(--tone);
		transition: width var(--transition-normal);
	}

	.slide {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 35%;
		border-radius: var(--radius-pill);
		background: var(--tone);
		animation: ll-progress-slide 1.6s var(--ease-out) infinite;
	}

	.sheen {
		position: absolute;
		inset: 0;
		background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.35), transparent);
		transform: translateX(-100%);
		animation: ll-progress-sheen 2s linear infinite;
	}

	@keyframes ll-progress-slide {
		from {
			left: -35%;
		}
		to {
			left: 100%;
		}
	}

	@keyframes ll-progress-sheen {
		to {
			transform: translateX(100%);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.slide {
			left: 0;
			width: 100%;
			opacity: 0.35;
			animation: none;
		}

		.sheen {
			display: none;
		}
	}
</style>
