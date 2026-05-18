<script lang="ts">
	import StatusUtilityButton from "@components/shared/StatusUtilityButton.svelte";

	export let isCleaning: boolean;
	export let showCleanupSuccess: boolean;
	export let handleCleanup: () => Promise<void>;

	export let isClearingCache: boolean;
	export let showCacheSuccess: boolean;
	export let handleClearCache: () => Promise<void>;
	export let shaderCacheSize: string;

	export let isDroppingCaches: boolean;
	export let showDropSuccess: boolean;
	export let handleDropCaches: () => Promise<void>;

	export let isClearingSwap: boolean;
	export let showSwapSuccess: boolean;
	export let handleClearSwap: () => Promise<void>;
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
