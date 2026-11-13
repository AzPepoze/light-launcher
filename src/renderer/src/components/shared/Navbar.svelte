<script lang="ts">
	import appIcon from "@icons/appicon.png";
	import { isPageAvailable, platform } from "@stores/platformStore";
	import { notifications } from "@stores/notificationStore";
	import { routes, type AppRoute } from "../../app/routes";

	export let activePage: string = "home";
	export let onNavigate: (page: string) => void = () => {};

	const navItems = routes.filter((route) => !route.hidden);
	const unreadCount = notifications.unreadCount;

	$: visibleItems = navItems.filter((item) => isPageAvailable(item.id, $platform.features));
	$: mainItems = visibleItems.filter((item) => !item.bottom);
	$: bottomItems = visibleItems.filter((item) => item.bottom);

	let navbarRef: HTMLElement;
	let indicatorStyle = "";

	function navIcon(item: AppRoute, unread: number): string {
		if (item.id !== "notifications") return item.icon;
		return unread > 0 ? "notifications_active" : "notifications";
	}

	function badgeLabel(unread: number): string {
		return unread > 99 ? "99+" : String(unread);
	}

	function updateIndicator(id: string) {
		if (!navbarRef) return;
		const activeEl = navbarRef.querySelector(
			`button[data-id="${id}"]`,
		) as HTMLElement;
		if (activeEl) {
			const navRect = navbarRef.getBoundingClientRect();
			const btnRect = activeEl.getBoundingClientRect();
			const top = btnRect.top - navRect.top;
			indicatorStyle = `top: ${top}px; height: ${btnRect.height}px; opacity: 1;`;
		}
	}

	$: {
		if (activePage && navbarRef) {
			setTimeout(() => updateIndicator(activePage), 0);
		}
	}

	function setActive(id: string) {
		onNavigate(id);
	}
</script>

{#snippet navItem(item: AppRoute)}
	<button
		class="nav-item"
		data-id={item.id}
		class:active={activePage === item.id}
		on:click={() => setActive(item.id)}
		title={item.label}
		aria-label={item.id === "notifications" && $unreadCount > 0
			? `${item.label}, ${$unreadCount} unread`
			: item.label}
		aria-current={activePage === item.id ? "page" : undefined}
	>
		{#if item.isCustomIcon}
			<img src={item.icon} class="custom-icon" alt="" />
		{:else}
			<span class="material-icons icon">{navIcon(item, $unreadCount)}</span>
		{/if}
		{#if item.id === "notifications" && $unreadCount > 0}
			<span class="nav-badge" aria-hidden="true">{badgeLabel($unreadCount)}</span>
		{/if}
	</button>
{/snippet}

<div class="navbar-wrapper">
	<div class="brand-logo">
		<img src={appIcon} alt="App Logo" />
	</div>

	<nav class="navbar" bind:this={navbarRef}>
		<div class="indicator" style={indicatorStyle}></div>
		<div class="nav-main">
			{#each mainItems as item (item.id)}
				{@render navItem(item)}
			{/each}
		</div>
		<div class="nav-bottom">
			{#each bottomItems as item (item.id)}
				{@render navItem(item)}
			{/each}
		</div>
	</nav>
</div>

<style lang="scss">
	.navbar-wrapper {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: flex-start;
		padding: 16px 0;
		gap: 24px;
		box-sizing: border-box;
		position: relative;
		overflow-x: hidden;
		overflow-y: auto;
		scrollbar-width: thin;
		scrollbar-color: var(--glass-border) transparent;
	}

	.brand-logo {
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: transform var(--transition-spring);
		/* In the flow above the icons, so a short window scrolls the icons and never slides them under the logo. */
		flex-shrink: 0;

		&:hover {
			transform: scale(1.2) rotate(10deg);
		}

		img {
			width: 100%;
			height: 100%;
			object-fit: contain;
		}
	}

	.navbar {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		position: relative;
		flex: 1;
		min-height: 0;
		margin-block: 0;
	}

	/* Main pages stay centred; the bottom page is pinned to the rail's bottom edge. */
	.nav-main,
	.nav-bottom {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
	}

	.nav-main {
		margin-block: auto;
	}

	.nav-bottom {
		flex-shrink: 0;
	}

	.indicator {
		position: absolute;
		left: 0;
		width: 3px;
		border-radius: 0 var(--radius-pill) var(--radius-pill) 0;
		background: var(--accent-primary);
		box-shadow: 0 0 12px var(--accent-glow);
		opacity: 0;
		transition:
			top 400ms var(--ease-spring),
			height 300ms var(--ease-out),
			opacity 200ms ease;
		z-index: 2;
	}

	.nav-item {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		flex-shrink: 0;
		border: none;
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		border-radius: var(--radius-md);
		transition:
			transform var(--transition-spring),
			color var(--transition-fast);
		outline: none;
		position: relative;
		z-index: 3;

		.icon {
			font-size: 22px;
			color: inherit;
			transition:
				color var(--transition-fast),
				transform var(--transition-spring);
		}

		.custom-icon {
			width: 22px;
			height: 22px;
			object-fit: contain;
			filter: grayscale(100%) opacity(0.6);
			transition:
				filter var(--transition-fast),
				transform var(--transition-spring);
		}

		&:hover {
			color: var(--text-main);
			transform: scale(1.15);
			.custom-icon {
				filter: grayscale(0%) opacity(0.9);
			}
		}

		&:active {
			transform: scale(0.9);
		}

		&.active {
			color: var(--accent-primary);

			.icon {
				color: var(--accent-primary);
			}

			.custom-icon {
				filter: grayscale(0%) opacity(1) drop-shadow(0 0 4px var(--accent-glow));
			}
		}
	}

	.nav-badge {
		position: absolute;
		top: 2px;
		right: 0;
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		box-sizing: border-box;
		border-radius: var(--radius-pill);
		background: var(--accent-primary);
		color: var(--bg-base);
		font-size: 0.7rem;
		font-weight: 700;
		line-height: 18px;
		text-align: center;
		pointer-events: none;
	}
</style>
