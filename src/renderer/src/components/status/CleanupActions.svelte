<script lang="ts">
	import StatusUtilityButton from "./StatusUtilityButton.svelte";

	interface Props {
		isCleaning: boolean;
		showCleanupSuccess: boolean;
		handleCleanup: () => Promise<void>;
		isClearingCache: boolean;
		showCacheSuccess: boolean;
		handleClearCache: () => Promise<void>;
		shaderCacheSize: string;
		isDroppingCaches: boolean;
		showDropSuccess: boolean;
		handleDropCaches: () => Promise<void>;
		isClearingSwap: boolean;
		showSwapSuccess: boolean;
		handleClearSwap: () => Promise<void>;
	}

	let {
		isCleaning,
		showCleanupSuccess,
		handleCleanup,
		isClearingCache,
		showCacheSuccess,
		handleClearCache,
		shaderCacheSize,
		isDroppingCaches,
		showDropSuccess,
		handleDropCaches,
		isClearingSwap,
		showSwapSuccess,
		handleClearSwap
	}: Props = $props();
</script>

<div class="utilities-row">
	<StatusUtilityButton
		icon="delete"
		title="Cleanup System"
		subtitle="Terminate all running processes"
		isPulsing={isCleaning}
		showSuccess={showCleanupSuccess}
		btnClass="cleanup"
		onclick={handleCleanup}
	/>

	<StatusUtilityButton
		icon="rocket_launch"
		title="Clear Shader Cache"
		subtitle={shaderCacheSize}
		isPulsing={isClearingCache}
		showSuccess={showCacheSuccess}
		btnClass="cache"
		onclick={handleClearCache}
	/>

	<StatusUtilityButton
		icon="auto_fix_high"
		title="Drop Caches"
		subtitle="Free up pagecache & dentries"
		isPulsing={isDroppingCaches}
		showSuccess={showDropSuccess}
		btnClass="drop-caches"
		onclick={handleDropCaches}
	/>

	<StatusUtilityButton
		icon="swap_horiz"
		title="Clear Swap"
		subtitle="Reset swap usage"
		isPulsing={isClearingSwap}
		showSuccess={showSwapSuccess}
		btnClass="clear-swap"
		onclick={handleClearSwap}
	/>
</div>

<style lang="scss">
	.utilities-row {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 20px;
		align-items: stretch;
	}
</style>
