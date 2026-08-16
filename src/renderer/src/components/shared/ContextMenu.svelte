<script lang="ts">
	import { fade, scale } from "svelte/transition";
	import { onMount } from "svelte";

	export let x: number;
	export let y: number;
	export let visible: boolean;
	export let isAutoScanned: boolean = false;
	export let isRunning: boolean = false;

	export let onLaunch: () => void;
	export let onConfigure: () => void;
	export let onAction: () => void; // Blacklist or Delete
	export let onClose: () => void;

	let menuEl: HTMLDivElement;

	onMount(() => {
		// Prevent context menu from going off screen
		if (menuEl) {
			const menuRect = menuEl.getBoundingClientRect();
			const windowWidth = window.innerWidth;
			const windowHeight = window.innerHeight;

			if (x + menuRect.width > windowWidth) {
				x = windowWidth - menuRect.width - 10;
			}
			if (y + menuRect.height > windowHeight) {
				y = windowHeight - menuRect.height - 10;
			}
		}
	});
</script>

{#if visible}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="context-menu-backdrop" on:click={onClose} on:contextmenu|preventDefault={onClose} transition:fade={{ duration: 100 }}>
		<div
			bind:this={menuEl}
			class="context-menu glass"
			style="left: {x}px; top: {y}px;"
			on:click|stopPropagation
			transition:scale={{ duration: 120, start: 0.95 }}
		>
			<button class="menu-item" on:click={() => { onLaunch(); onClose(); }}>
				<span class="material-icons">{isRunning ? 'stop' : 'play_arrow'}</span>
				<span>{isRunning ? 'Kill Process' : 'Launch Game'}</span>
			</button>

			<button class="menu-item" on:click={() => { onConfigure(); onClose(); }}>
				<span class="material-icons">settings</span>
				<span>Configure Config</span>
			</button>

			<div class="menu-divider"></div>

			{#if isAutoScanned}
				<button class="menu-item danger" on:click={() => { onAction(); onClose(); }}>
					<span class="material-icons">block</span>
					<span>Hide / Blacklist</span>
				</button>
			{:else}
				<button class="menu-item danger" on:click={() => { onAction(); onClose(); }}>
					<span class="material-icons">delete</span>
					<span>Remove Profile</span>
				</button>
			{/if}
		</div>
	</div>
{/if}

<style lang="scss">
	.context-menu-backdrop {
		position: fixed;
		inset: 0;
		z-index: 9999;
		background: transparent;
	}

	.context-menu {
		position: absolute;
		min-width: 180px;
		background: rgba(var(--bg-surface-rgb, 12, 12, 12), 0.75);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid var(--glass-border);
		border-radius: var(--radius-md);
		padding: 6px;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
		display: flex;
		flex-direction: column;
		gap: 2px;
		z-index: 10000;
	}

	.menu-item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 14px;
		background: none;
		border: none;
		border-radius: var(--radius-sm);
		color: var(--text-main);
		font-size: 0.88rem;
		font-weight: 700;
		cursor: pointer;
		text-align: left;
		transition: all var(--transition-fast);

		&:hover {
			background: var(--accent-glow);
			color: var(--accent-primary);
		}

		&.danger:hover {
			background: rgba(255, 68, 68, 0.1);
			color: var(--danger, #ff4444);
		}

		.material-icons {
			font-size: 18px;
		}
	}

	.menu-divider {
		height: 1px;
		background: rgba(255, 255, 255, 0.06);
		margin: 4px 6px;
	}
</style>
