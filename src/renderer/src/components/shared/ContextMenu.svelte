<script lang="ts">
	import AppMenu from "@components/shared/AppMenu.svelte";

	export let x: number;
	export let y: number;
	export let visible: boolean;
	export let isAutoScanned: boolean = false;
	export let isRunning: boolean = false;
	export let hasCustomIcon: boolean = false;

	export let onLaunch: () => void;
	export let onLaunchWithLogs: () => void = () => {};
	export let onKill: () => void = () => {};
	export let onConfigure: () => void;
	export let onSetCustomIcon: () => void = () => {};
	export let onClearCustomIcon: () => void = () => {};
	export let onAction: () => void; // Blacklist or Delete
	export let onClose: () => void;
</script>

<AppMenu bind:x bind:y bind:visible mode="cursor" {onClose}>
	{#if isRunning}
		<button class="menu-item danger-soft" on:click={() => { onKill(); onClose(); }}>
			<span class="material-icons">stop_circle</span>
			<span>Stop Game</span>
		</button>
	{:else}
		<button class="menu-item" on:click={() => { onLaunch(); onClose(); }}>
			<span class="material-icons">play_arrow</span>
			<span>Launch Game</span>
		</button>

		<button class="menu-item" on:click={() => { onLaunchWithLogs(); onClose(); }}>
			<span class="material-icons">terminal</span>
			<span>Run with Logs</span>
		</button>
	{/if}

	<button class="menu-item" on:click={() => { onConfigure(); onClose(); }}>
		<span class="material-icons">settings</span>
		<span>Configure Profile</span>
	</button>

	{#if !isAutoScanned}
		<div class="menu-divider"></div>
		<button class="menu-item" on:click={() => { onSetCustomIcon(); onClose(); }}>
			<span class="material-icons">image</span>
			<span>{hasCustomIcon ? 'Replace Custom Icon' : 'Set Custom Icon'}</span>
		</button>
		{#if hasCustomIcon}
			<button class="menu-item" on:click={() => { onClearCustomIcon(); onClose(); }}>
				<span class="material-icons">hide_image</span>
				<span>Use Executable Icon</span>
			</button>
		{/if}
	{/if}

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
</AppMenu>
