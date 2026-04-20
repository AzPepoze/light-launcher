<script lang="ts">
	export let checked = false;
	export let label = "";
	export let subtitle = "";
	export let hasConfig = false;
	export let onChange: (value: boolean) => void = () => {};
	export let onConfig: () => void = () => {};

	function toggle() {
		checked = !checked;
		onChange(checked);
	}

	function openConfig(e: MouseEvent) {
		e.stopPropagation();
		onConfig();
	}
</script>

<div
	class="slide-button-card"
	class:active={checked}
	on:click={toggle}
	on:keydown={(e) => (e.key === " " || e.key === "Enter") && toggle()}
	role="button"
	tabindex="0"
>
	<div class="info">
		<div class="title">{label}</div>
		{#if subtitle}
			<div class="subtitle">{subtitle}</div>
		{/if}
	</div>
	<div class="actions">
		{#if hasConfig}
			<button
				class="config-btn"
				on:click={openConfig}
				title="Configure"
			>
				<span class="material-icons" style="font-size: 16px;"
					>settings</span
				>
			</button>
		{/if}
		<div class="switch-container">
			<input
				type="checkbox"
				{checked}
				on:change|stopPropagation={toggle}
			/>
			<span class="slider"></span>
		</div>
	</div>
</div>

<style lang="scss">
	.slide-button-card {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 16px 20px;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 12px;
		cursor: pointer;
		transition: all 0.2s;
		user-select: none;

		&:hover {
			background: var(--glass-border-bright);
			border-color: var(--accent-primary);

			.config-btn {
				opacity: 1;
			}
		}

		&.active {
			border-color: var(--accent-primary);
			background: var(--glass-surface);
		}
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.config-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		padding: 6px;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0.8;
		transition: all 0.2s;

		&:hover {
			background: var(--glass-hover);
			color: var(--text-main);
			opacity: 1;
		}
	}

	.switch-container {
		position: relative;
		width: 38px;
		height: 20px;
		flex-shrink: 0;

		input {
			opacity: 0;
			width: 0;
			height: 0;

			&:checked + .slider {
				background-color: var(--accent-primary);
			}

			&:checked + .slider:before {
				transform: translateX(18px);
				background-color: var(--glass-bg);
			}
		}
	}

	.slider {
		position: absolute;
		cursor: pointer;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-color: var(--glass-bg);
		border: 1px solid var(--glass-border);
		transition: 0.3s;
		border-radius: 34px;

		&:before {
			position: absolute;
			content: "";
			height: 14px;
			width: 14px;
			left: 3px;
			bottom: 3px;
			background-color: var(--text-dim);
			transition: 0.3s;
			border-radius: 50%;
		}
	}

	.info {
		.title {
			font-weight: 600;
			color: var(--text-main);
			font-size: 0.9rem;
		}
		.subtitle {
			font-size: 0.7rem;
			color: var(--text-dim);
			margin-top: 2px;
		}
	}
</style>
