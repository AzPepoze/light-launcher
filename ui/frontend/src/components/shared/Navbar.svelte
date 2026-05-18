<script lang="ts">
	import appIcon from "@icons/appicon.png";

	export let activePage: string = "home";
	export let onNavigate: (page: string) => void = () => {};
	export let toggleTheme: () => void = () => {};

	const navItems = [
		{ id: "home", label: "Home", icon: "home" },
		{ id: "run", label: "Run", icon: "rocket_launch" },
		{ id: "versions", label: "Versions", icon: "history" },
		{ id: "prefix", label: "Prefix", icon: "folder_shared" },
		{ id: "utils", label: "Utils", icon: "handyman" },
		{ id: "settings", label: "Settings", icon: "settings" },
	];

	let navbarRef: HTMLElement;
	let indicatorStyle = "";

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

<div class="navbar-wrapper">
	<div class="brand-logo">
		<img src={appIcon} alt="App Logo" />
	</div>

	<nav class="navbar" bind:this={navbarRef}>
		<div class="indicator" style={indicatorStyle}></div>
		{#each navItems as item}
			<button
				class="nav-item"
				data-id={item.id}
				class:active={activePage === item.id}
				on:click={() => setActive(item.id)}
				title={item.label}
			>
				<span class="material-icons icon">{item.icon}</span>
			</button>
		{/each}
	</nav>
</div>

<style lang="scss">
	.navbar-wrapper {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 16px 0;
		gap: 24px;
		box-sizing: border-box;
		position: relative;
	}

	.brand-logo {
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: transform var(--transition-spring);
		position: absolute;
		top: 20px;

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
	}

	.indicator {
		position: absolute;
		left: 0;
		width: 3px;
		border-radius: 0 var(--radius-pill) var(--radius-pill) 0;
		background: var(--accent-primary);
		box-shadow: 0 0 12px var(--accent-glow);
		opacity: 0;
		transition: top 400ms var(--ease-spring), height 300ms var(--ease-out), opacity 200ms ease;
		z-index: 2;
	}

	.nav-item {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		border: none;
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		border-radius: var(--radius-md);
		transition: transform var(--transition-spring), color var(--transition-fast);
		outline: none;
		position: relative;
		z-index: 3;

		.icon {
			font-size: 22px;
			color: inherit;
			transition: color var(--transition-fast), transform var(--transition-spring);
		}

		&:hover {
			color: var(--text-main);
			transform: scale(1.15);
		}

		&:active {
			transform: scale(0.9);
		}

		&.active {
			color: var(--accent-primary);

			.icon {
				color: var(--accent-primary);
			}
		}
	}
</style>
