<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import StatusUtilityButton from "@components/shared/StatusUtilityButton.svelte";
	import SystemResources from "@components/shared/SystemResources.svelte";
	import CleanupActions from "@components/shared/CleanupActions.svelte";
	import { StatusDrawerState } from "./StatusDrawerState.svelte";

	const state = new StatusDrawerState();

	onMount(() => {
		state.initialize();
	});

	onDestroy(() => {
		state.destroy();
	});
</script>

<div class="status-drawer-wrapper" class:expanded={state.isExpanded}>
	<button class="toggle-btn" on:click={() => (state.isExpanded = !state.isExpanded)}>
		<span class="material-icons">tune</span>
		<span class="trigger-text"
			>{state.isExpanded
				? "CLOSE DRAWER"
				: "SYSTEM STATUS & UTILITIES"}</span
		>
		<span class="material-icons">
			{state.isExpanded ? "keyboard_arrow_down" : "keyboard_arrow_up"}
		</span>
	</button>

	<div class="drawer-content">
		<SystemResources sysInfo={state.sysInfo} sysUsage={state.sysUsage} />

		<div class="divider"></div>

		<CleanupActions
			isCleaning={state.isCleaning}
			showCleanupSuccess={state.showCleanupSuccess}
			handleCleanup={() => state.handleCleanup()}
			isClearingCache={state.isClearingCache}
			showCacheSuccess={state.showCacheSuccess}
			handleClearCache={() => state.handleClearCache()}
			shaderCacheSize={state.shaderCacheSize}
			isDroppingCaches={state.isDroppingCaches}
			showDropSuccess={state.showDropSuccess}
			handleDropCaches={() => state.handleDropCaches()}
			isClearingSwap={state.isClearingSwap}
			showSwapSuccess={state.showSwapSuccess}
			handleClearSwap={() => state.handleClearSwap()}
		/>
	</div>
</div>

<style lang="scss">
	.status-drawer-wrapper {
		position: fixed;
		bottom: 20px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		transform: translateY(calc(100% - 78px));
		transition: transform 0.6s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.3s ease;
		z-index: 100;
		padding: 0 24px 24px 24px;
		box-shadow: 0 10px 50px rgba(0, 0, 0, 0.4);
		overflow: hidden;
		margin-right: 24px;
		width: -webkit-fill-available;
		box-sizing: border-box;

		&.expanded {
			transform: translateY(0);
			box-shadow: 0 10px 60px rgba(255, 102, 171, 0.1);
		}
	}

	.toggle-btn {
		width: 100%;
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		cursor: pointer;
		background: var(--bg-elevated);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-pill);
		margin: 12px 0;
		color: var(--accent-primary);
		transition: transform var(--transition-spring), border-color var(--transition-fast), background var(--transition-fast), box-shadow var(--transition-fast);

		.trigger-text {
			font-size: 0.85rem;
			font-weight: 800;
			color: inherit;
			letter-spacing: 1.5px;
			text-transform: uppercase;
		}

		.material-icons {
			font-size: 20px;
			color: inherit;
		}

		&:hover {
			background: var(--accent-primary);
			border-color: var(--accent-primary);
			color: var(--bg-base);
			box-shadow: 0 4px 15px var(--accent-glow);
		}

		&:active {
			transform: scale(0.97);
		}
	}

	.drawer-content {
		padding-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}
</style>
