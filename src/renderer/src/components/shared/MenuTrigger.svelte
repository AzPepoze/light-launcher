<script lang="ts">
	export let active: boolean = false;
	export let label: string = "Open menu";
	export let icon: string = "more_vert";
</script>

<button
	class="menu-trigger"
	class:active
	on:click|stopPropagation
	type="button"
	aria-label={label}
	aria-haspopup="menu"
	aria-expanded={active}
>
	<span class="material-icons">{icon}</span>
</button>

<style lang="scss">
	.menu-trigger {
		background: transparent;
		border: 1px solid transparent;
		color: var(--text-muted);
		width: 32px;
		height: 32px;
		border-radius: var(--radius-md);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition:
			background var(--transition-fast),
			color var(--transition-fast),
			border-color var(--transition-fast),
			transform var(--transition-spring);

		.material-icons {
			font-size: 20px;
			transition: transform var(--transition-spring);
		}

		&:hover {
			background: rgba(255, 255, 255, 0.08);
			color: var(--text-main);
			border-color: rgba(255, 255, 255, 0.06);
			transform: scale(1.05);
		}

		&:active {
			transform: scale(0.95);
		}

		&.active {
			background: var(--accent-glow);
			color: var(--accent-primary);
			border-color: var(--glass-border);
			.material-icons {
				transform: rotate(90deg);
			}
		}
	}
</style>
