<script lang="ts">
	import AppMenu from "@components/shared/AppMenu.svelte";
	import Icon from "./Icon.svelte";

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
	export let onOpenLocation: () => void = () => {};
	export let onSetCustomIcon: () => void = () => {};
	export let onClearCustomIcon: () => void = () => {};
	export let onAction: () => void; // Blacklist or Delete
	export let onClose: () => void;
</script>

<AppMenu bind:x bind:y bind:visible mode="cursor" {onClose}>
	{#if isRunning}
		<button class="menu-item danger-soft" on:click={() => { onKill(); onClose(); }}>
			<Icon name="stop_circle" size={18} />
			<span>Stop Game</span>
		</button>
	{:else}
		<button class="menu-item" on:click={() => { onLaunch(); onClose(); }}>
			<Icon name="play_arrow" size={18} />
			<span>Launch Game</span>
		</button>

		<button class="menu-item" on:click={() => { onLaunchWithLogs(); onClose(); }}>
			<Icon name="terminal" size={18} />
			<span>Run with Logs</span>
		</button>
	{/if}

	<button class="menu-item" on:click={() => { onConfigure(); onClose(); }}>
		<Icon name="settings" size={18} />
		<span>Configure Profile</span>
	</button>

	<button class="menu-item" on:click={() => { onOpenLocation(); onClose(); }}>
		<Icon name="folder_open" size={18} />
		<span>Open File Location</span>
	</button>

	{#if !isAutoScanned}
		<div class="menu-divider"></div>
		<button class="menu-item" on:click={() => { onSetCustomIcon(); onClose(); }}>
			<Icon name="image" size={18} />
			<span>{hasCustomIcon ? 'Replace Custom Icon' : 'Set Custom Icon'}</span>
		</button>
		{#if hasCustomIcon}
			<button class="menu-item" on:click={() => { onClearCustomIcon(); onClose(); }}>
				<Icon name="hide_image" size={18} />
				<span>Use Executable Icon</span>
			</button>
		{/if}
	{/if}

	<div class="menu-divider"></div>

	{#if isAutoScanned}
		<button class="menu-item danger" on:click={() => { onAction(); onClose(); }}>
			<Icon name="block" size={18} />
			<span>Hide / Blacklist</span>
		</button>
	{:else}
		<button class="menu-item danger" on:click={() => { onAction(); onClose(); }}>
			<Icon name="delete" size={18} />
			<span>Remove Profile</span>
		</button>
	{/if}
</AppMenu>
