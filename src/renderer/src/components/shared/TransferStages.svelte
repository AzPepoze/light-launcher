<script lang="ts">
	import Icon from "./Icon.svelte";

	type Stage = "download" | "extract" | "install" | "done";

	let { stage, failed = false }: { stage: Stage; failed?: boolean } = $props();

	const STEPS = [
		{ key: "download", label: "Download", icon: "download" },
		{ key: "extract", label: "Extract", icon: "unarchive" },
		{ key: "install", label: "Install", icon: "install_desktop" }
	] as const;

	/* Index of the current step. A done stage is past the last step. */
	const current = $derived(
		stage === "done" ? STEPS.length : STEPS.findIndex((step) => step.key === stage)
	);

	function stepState(index: number): "done" | "current" | "failed" | "later" {
		if (index < current) return "done";
		if (index > current) return "later";
		return failed ? "failed" : "current";
	}
</script>

<ol class="stages">
	{#each STEPS as step, index (step.key)}
		{@const state = stepState(index)}
		<li class="step {state}" aria-current={state === "current" ? "step" : undefined}>
			<span class="step-icon">
				{#if state === "done"}
					<Icon name="check" size={16} />
				{:else if state === "failed"}
					<Icon name="error" size={16} />
				{:else}
					<Icon name={step.icon} size={16} filled={state === "current"} />
				{/if}
				{#if state === "current"}<span class="pulse" aria-hidden="true"></span>{/if}
			</span>
			<span class="step-label">{step.label}</span>
		</li>
	{/each}
</ol>

<style lang="scss">
	.stages {
		display: flex;
		align-items: center;
		gap: 8px;
		margin: 0;
		padding: 0;
		list-style: none;
		font-size: 0.8rem;
	}

	.step {
		display: flex;
		align-items: center;
		gap: 6px;
		color: var(--text-dim);

		&.done {
			color: var(--text-muted);

			.step-icon {
				color: var(--success);
			}
		}

		&.current {
			color: var(--text-main);
			font-weight: 700;

			.step-icon {
				color: var(--accent-primary);
			}
		}

		&.failed {
			color: var(--danger);
			font-weight: 700;

			.step-icon {
				color: var(--danger);
			}
		}

		&:not(:last-child)::after {
			content: "";
			width: 16px;
			height: 2px;
			margin-left: 2px;
			border-radius: var(--radius-pill);
			background: var(--glass-border);
		}

		&.done:not(:last-child)::after {
			background: var(--success);
		}
	}

	.step-icon {
		position: relative;
		display: inline-flex;
	}

	.pulse {
		position: absolute;
		top: -2px;
		right: -3px;
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--accent-primary);
		animation: ll-stage-pulse 1.4s var(--ease-out) infinite;
	}

	@keyframes ll-stage-pulse {
		50% {
			opacity: 0.3;
			transform: scale(0.7);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.pulse {
			animation: none;
		}
	}
</style>
