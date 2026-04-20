<script lang="ts">
	export let activePage: string = "home";
	export let onNavigate: (page: string) => void = () => {};
	export let toggleTheme: () => void = () => {};

	const navItems = [
		{ id: "home", label: "Home", icon: "home" },
		{ id: "run", label: "Run", icon: "rocket_launch" },
		{ id: "versions", label: "Versions", icon: "history" },
		{ id: "prefix", label: "Prefix", icon: "folder_shared" },
		{ id: "utils", label: "Utils", icon: "handyman" },
	];

	let navbarRef: HTMLElement;
	let indicatorStyle = "";

	function updateIndicator(id: string) {
		if (!navbarRef) return;
		const activeEl = navbarRef.querySelector(
			`button[data-id="${id}"]`,
		) as HTMLElement;
		if (activeEl) {
			const top = activeEl.offsetTop;
			const height = activeEl.offsetHeight;
			indicatorStyle = `top: ${top}px; height: ${height}px; opacity: 1;`;
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
	<div class="brand-v-text">
		{#each "LIGHTLAUNCHER".split("") as char, i}
			<span style="animation-delay: {i * 0.1}s">{char}</span>
		{/each}
	</div>
	<nav class="navbar" bind:this={navbarRef}>
		<div class="moving-indicator" style={indicatorStyle}></div>
		{#each navItems as item}
			<button
				class="nav-item"
				data-id={item.id}
				class:active={activePage === item.id}
				on:click={() => setActive(item.id)}
			>
				<div class="icon-container">
					<span class="material-icons icon">{item.icon}</span>
					{#if activePage === item.id}
						<div class="glow-ring"></div>
					{/if}
				</div>
				<span class="label">{item.label}</span>
			</button>
		{/each}
	</nav>

	<button class="theme-toggle" on:click={toggleTheme} title="Toggle Theme">
		<span class="material-icons">contrast</span>
	</button>
</div>

<style lang="scss">
	.navbar-wrapper {
		z-index: 1000;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding-top: 20px;
		gap: 40px;
	}

	.brand-v-text {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		user-select: none;
		position: absolute;
		top: 50px;

		span {
			font-size: 0.75rem;
			font-weight: 950;
			color: var(--text-muted);
			opacity: 0.7;
			animation: char-wave 4s infinite ease-in-out;
			line-height: 1;
		}
	}

	@keyframes char-wave {
		0%,
		100% {
			transform: scale(0.8);
			opacity: 0.4;
		}
		50% {
			transform: scale(1);
			opacity: 1;
			text-shadow: 0 0 10px var(--text-main);
		}
	}

	.navbar {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 16px 0;
		background: transparent;
		border: none;
		box-shadow: none;
	}

	.moving-indicator {
		position: absolute;
		left: 0;
		width: 3px;
		background: var(--accent-primary);
		border-radius: 0 4px 4px 0;
		transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
		pointer-events: none;
		opacity: 0;
		z-index: 0;
	}

	.nav-item {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 2px;
		padding: 12px 0;
		border: none;
		background: transparent;
		color: var(--text-dim);
		cursor: pointer;
		border-radius: 20px;
		transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
		min-width: 68px; /* Narrower items */
		z-index: 1;

		&:hover {
			color: var(--text-main);

			.icon {
				transform: translateY(-2px);
				opacity: 1;
			}
		}

		&.active {
			color: var(--accent-primary);

			.icon {
				transform: scale(1.1);
				opacity: 1;
				color: var(--accent-primary);
			}

			.label {
				opacity: 1;
			}
		}
	}

	.icon-container {
		position: relative;
		width: 22px;
		height: 22px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.icon {
		font-size: 22px;
		color: var(--text-main);
		opacity: 0.6;
		transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
	}

	.label {
		font-size: 0.6rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.8px;
		opacity: 0.6;
		transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
	}

	.glow-ring {
		position: absolute;
		width: 140%;
		height: 140%;
		border-radius: 50%;
		background: radial-gradient(
			circle,
			var(--accent-primary) 0%,
			transparent 70%
		);
		opacity: 0.15;
		animation: pulse 3s infinite;
	}

	.theme-toggle {
		position: absolute;
		bottom: 24px;
		background: none;
		border: 1px solid var(--glass-border);
		color: var(--text-dim);
		cursor: pointer;
		padding: 8px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.3s;

		&:hover {
			background: var(--glass-border);
			color: var(--text-main);
			border-color: var(--glass-border-bright);
			transform: scale(1.1);
		}

		.material-icons {
			font-size: 20px;
		}
	}

	@keyframes pulse {
		0% {
			transform: scale(0.8);
			opacity: 0;
		}
		50% {
			transform: scale(1.2);
			opacity: 0.4;
		}
		100% {
			transform: scale(1.4);
			opacity: 0;
		}
	}
</style>
