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
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		cursor: pointer;
		transition: transform var(--transition-spring), border-color var(--transition-fast), background var(--transition-fast);
		user-select: none;

		&:hover {
			border-color: var(--accent-secondary);
			transform: scale(1.02);

			.config-btn {
				opacity: 1;
			}
		}

		&.active {
			border-color: var(--accent-primary);
			background: var(--bg-elevated);
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
		border-radius: var(--radius-sm);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0.8;
		transition: transform var(--transition-spring), color var(--transition-fast), opacity var(--transition-fast);

		&:hover {
			color: var(--accent-primary);
			opacity: 1;
			transform: rotate(45deg);
		}
	}

	.switch-container {
		position: relative;
		width: 44px;
		height: 24px;
		flex-shrink: 0;

		input {
			opacity: 0;
			width: 0;
			height: 0;

			&:checked + .slider {
				background-color: var(--accent-primary);
			}

			&:checked + .slider:before {
				transform: translateX(20px);
				background-color: #ffffff;
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
		background-color: var(--bg-input);
		border: 2px solid rgba(255, 255, 255, 0.03);
		transition: background-color var(--transition-fast);
		border-radius: var(--radius-pill);

		&:before {
			position: absolute;
			content: "";
			height: 14px;
			width: 14px;
			left: 3px;
			bottom: 3px;
			background-color: var(--text-muted);
			transition: transform var(--transition-spring), background-color var(--transition-fast);
			border-radius: 50%;
		}
	}

	.info {
		.title {
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			font-size: 0.85rem;
			letter-spacing: 0.5px;
		}
		.subtitle {
			font-size: 0.75rem;
			font-weight: 600;
			color: var(--text-muted);
			margin-top: 4px;
		}
	}
</style>
