<script lang="ts">
	import { onMount } from "svelte";
	import AppMenu from "@components/shared/AppMenu.svelte";
	import MenuTrigger from "@components/shared/MenuTrigger.svelte";

	export let icon: string = "folder";
	export let name: string = "";
	export let subtitle: string = "";
	export let count: number = 0;
	export let hasMenu: boolean = false;
	export let isMenuOpen: boolean = false;
	export let onToggleMenu: (event: MouseEvent) => void = () => {};
	export let onCloseMenu: () => void = () => {};
	export let onRescan: () => void = () => {};
	export let onConfigureFolder: () => void = () => {};
	export let onRemoveFolder: () => void = () => {};
	/** Horizontal px to bleed out of a padded parent container when stuck, so the bar matches the Custom Profiles header width. */
	export let stuckBleed = 0;

	let headerEl: HTMLDivElement;
	let isStuck = false;

	function findScrollParent(el: HTMLElement | null): HTMLElement | null {
		let p: HTMLElement | null = el?.parentElement ?? null;
		while (p) {
			const style = getComputedStyle(p);
			if (style.overflowY === "auto" || style.overflowY === "scroll") return p;
			p = p.parentElement;
		}
		return null;
	}

	onMount(() => {
		const scrollParent = findScrollParent(headerEl);
		const target = scrollParent ?? window;
		const check = () => {
			if (!headerEl) return;
			const rect = headerEl.getBoundingClientRect();
			const parentRect = scrollParent ? scrollParent.getBoundingClientRect() : { top: 0 } as DOMRect;
			isStuck = rect.top <= parentRect.top + 1;
		};
		target.addEventListener("scroll", check, { passive: true } as any);
		window.addEventListener("scroll", check, { passive: true } as any);
		window.addEventListener("resize", check);
		check();
		return () => {
			target.removeEventListener("scroll", check as any);
			window.removeEventListener("scroll", check as any);
			window.removeEventListener("resize", check);
		};
	});
</script>

<div
	bind:this={headerEl}
	class="folder-group-header"
	class:is-stuck={isStuck}
	style:--stuck-bleed="{stuckBleed}px"
>
	<div class="folder-title" title={subtitle || name}>
		<span class="material-icons folder-icon-main">{icon}</span>
		<div class="folder-metadata">
			<span class="folder-name">{name}</span>
			{#if subtitle}
				<span class="folder-path">{subtitle}</span>
			{/if}
		</div>
		<span class="badge">{count}</span>
	</div>

	{#if hasMenu}
		<div class="folder-actions-wrapper">
			<MenuTrigger
				active={isMenuOpen}
				label="Folder Options"
				on:click={onToggleMenu}
			/>
			<AppMenu visible={isMenuOpen} mode="anchor" onClose={onCloseMenu}>
				<button class="menu-item" on:click={onRescan}>
					<span class="material-icons">refresh</span>
					<span>Rescan Folder</span>
				</button>
				<button class="menu-item" on:click={onConfigureFolder}>
					<span class="material-icons">settings</span>
					<span>Folder Settings</span>
				</button>
				<button class="menu-item danger" on:click={onRemoveFolder}>
					<span class="material-icons">delete_outline</span>
					<span>Remove Watch</span>
				</button>
			</AppMenu>
		</div>
	{/if}
</div>

<style lang="scss">
	.folder-group-header {
		position: sticky;
		top: 0;
		z-index: 7;
		display: flex;
		justify-content: space-between;
		align-items: center;
		box-sizing: border-box;
		height: 64px;
		margin: 12px;
		padding: 12px 16px;
		border: 1px solid transparent;
		border-bottom: 1px solid rgba(255, 255, 255, 0.06);
		background: transparent;
		border-radius: 0;
		transition: background var(--transition-fast), border-color var(--transition-fast), border-radius var(--transition-fast);

		&.is-stuck {
			background: var(--bg-base);
			border: 1px solid rgba(255, 255, 255, 0.06);
			border-radius: var(--radius-md);
			margin-left: calc(12px - var(--stuck-bleed, 0px));
			margin-right: calc(12px - var(--stuck-bleed, 0px));
		}
	}

	.folder-title {
		display: flex;
		align-items: center;
		gap: 14px;
		min-width: 0;
		flex: 1;
	}

	.folder-icon-main {
		font-size: 26px;
		color: var(--accent-secondary);
		filter: drop-shadow(0 0 8px rgba(255, 102, 171, 0.3));
	}

	.folder-metadata {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}

	.folder-name {
		font-size: 1.15rem;
		line-height: 22px;
		font-weight: 800;
		color: var(--text-main);
		letter-spacing: -0.3px;
	}

	.folder-path {
		font-size: 0.75rem;
		line-height: 14px;
		color: var(--text-muted);
		opacity: 0.6;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 400px;
	}

	.badge {
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-sm);
		padding: 2px 8px;
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--text-muted);
	}

	.folder-actions-wrapper {
		position: relative;
	}
</style>
