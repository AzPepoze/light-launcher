<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import Icon from "./Icon.svelte";

	export let active: boolean = false;
	export let label: string = "Open menu";
	export let icon: string = "more_vert";

	const dispatch = createEventDispatcher<{ click: MouseEvent }>();
</script>

<button
	class="menu-trigger"
	class:active
	on:click|stopPropagation={(event) => dispatch("click", event)}
	type="button"
	aria-label={label}
	aria-haspopup="menu"
	aria-expanded={active}
>
	<Icon name={icon} size={20} class="menu-icon" />
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

		:global(.menu-icon) {
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
			:global(.menu-icon) {
				transform: rotate(90deg);
			}
		}
	}
</style>
