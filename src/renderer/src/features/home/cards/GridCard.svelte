<script lang="ts">
	import CardRunningIndicator from "./shared/CardRunningIndicator.svelte";
	import CardSelectionCheckbox from "./shared/CardSelectionCheckbox.svelte";
	import AnimatedPlayIcon from "./shared/AnimatedPlayIcon.svelte";
	import Marquee from "@components/shared/Marquee.svelte";
	import LauncherMenu from "../shared/LauncherMenu.svelte";
	import { getDominantColor } from "@lib/dominantColor";
	import { iconFade } from "@lib/iconReveal";

	export let game: any;
	export let icon: string = "";
	export let isRunning: boolean = false;
	export let isSelectionMode: boolean = false;
	export let isSelected: boolean = false;
	export let onLaunch: (game: any) => void = () => {};
	export let onConfigure: (game: any) => void = () => {};
	export let onChooseLauncher: (launcherPath: string | null) => void = () => {};
	export let onSelect: (game: any, shiftKey: boolean, ctrlKey?: boolean) => void = () => {};

	let cardElement: HTMLElement;
	let spotlightX = 50;
	let spotlightY = 50;
	let isHovering = false;
	let rafId: number | null = null;
	let pendingX = 50;
	let pendingY = 50;

	let dominantRGB = "255,255,255";
	let colorFetched = false;
	let iconImgEl: HTMLImageElement | undefined;
	let colorNonce = 0;
	let lastIcon = "";

	$: if (icon !== lastIcon) {
		lastIcon = icon;
		colorFetched = false;
		dominantRGB = "255,255,255";
		iconImgEl = undefined;
	}

	function handleLaunch(event?: MouseEvent) {
		if (isSelectionMode) {
			onSelect(game, !!event?.shiftKey, !!(event?.ctrlKey || event?.metaKey));
			return;
		}
		onLaunch(game);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === "Enter") handleLaunch();
	}

	function handleConfigure() {
		onConfigure(game);
	}

	// Keeps clicks and keys on the card actions from reaching the icon's launch handler.
	function stopPropagation(event: Event) {
		event.stopPropagation();
	}

	function scheduleFrame() {
		if (rafId !== null) return;
		rafId = requestAnimationFrame(() => {
			rafId = null;
			spotlightX = pendingX;
			spotlightY = pendingY;
		});
	}

	function handleMouseMove(event: MouseEvent) {
		if (isSelectionMode) return;
		const rect = cardElement.getBoundingClientRect();
		const x = ((event.clientX - rect.left) / rect.width) * 100;
		const y = ((event.clientY - rect.top) / rect.height) * 100;
		pendingX = Math.max(0, Math.min(100, x));
		pendingY = Math.max(0, Math.min(100, y));
		scheduleFrame();
	}

	function handleMouseEnter() {
		isHovering = true;
		if (icon && !colorFetched) {
			colorFetched = true;
			const nonce = ++colorNonce;
			getDominantColor(icon, iconImgEl).then((c) => {
				if (nonce !== colorNonce) return;
				if (c) dominantRGB = `${c[0]},${c[1]},${c[2]}`;
			});
		}
	}

	function handleMouseLeave() {
		isHovering = false;
		if (rafId !== null) {
			cancelAnimationFrame(rafId);
			rafId = null;
		}
	}
</script>

<div
	bind:this={cardElement}
	class="game-card"
	class:running={isRunning}
	class:selection-mode={isSelectionMode}
	class:selected={isSelected}
	class:hovering={isHovering}
	style="--mx: {spotlightX}%; --my: {spotlightY}%; --spotlight-rgb: {dominantRGB};"
	on:mousemove={handleMouseMove}
	on:mouseenter={handleMouseEnter}
	on:mouseleave={handleMouseLeave}
	role="group"
>
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="game-icon-container"
		on:click={handleLaunch}
		on:keydown={handleKeydown}
		role="button"
		tabindex="0"
		aria-label={game.name}
	>
		{#if isSelectionMode}
			<CardSelectionCheckbox {isSelected} />
		{/if}

		{#if isRunning}
			<CardRunningIndicator />
		{/if}

		<div class="icon-wrapper">
			{#if icon}
				<img
					bind:this={iconImgEl}
					use:iconFade={icon}
					src={icon}
					alt=""
					class="game-icon"
					loading="lazy"
					decoding="async"
					draggable="false"
				/>
			{:else}
				<span class="material-icons system-icon">rocket_launch</span>
			{/if}
		</div>

		<div class="play-overlay">
			{#if isHovering}
				<span class="play-ripple" aria-hidden="true"></span>
				<span class="launch-icon-large">
					<AnimatedPlayIcon size={64} />
				</span>
			{/if}
		</div>

		<div class="card-overlay">
			<Marquee
				text={game.name}
				title={game.name}
				class="game-name"
				viewportClass="name-viewport"
			/>
			<div class="card-actions">
				<button
					class="config-btn"
					type="button"
					title="Configure"
					aria-label="Configure {game.name}"
					on:click|stopPropagation={handleConfigure}
				>
					<span class="material-icons">settings</span>
				</button>
			</div>
		</div>

		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="launcher-slot" on:click={stopPropagation} on:keydown={stopPropagation}>
			<LauncherMenu {game} onChoose={onChooseLauncher} />
		</div>
	</div>
</div>

<style lang="scss">
	.game-card {
		--card-pad: 8px;
		position: relative;
		width: 100%;
		max-width: 190px;
		padding: var(--card-pad);
		--mx: 50%;
		--my: 50%;
		--spotlight-rgb: 255, 255, 255;
		transition: transform var(--motion-base) var(--ease-out);

		&.selected {
			.game-icon-container {
				border-color: var(--accent-primary);
				box-shadow: 0 0 10px var(--accent-glow);
			}
		}

		&:hover,
		&:focus-within {
			transform: translateY(-4px) scale(1.015);

			.game-icon-container {
				border-color: rgba(var(--spotlight-rgb), 0.32);
				box-shadow: 0 0 16px rgba(var(--spotlight-rgb), 0.35);

				.play-overlay {
					opacity: 1;
				}

				img.game-icon {
					transform: scale(1.06);
					filter: brightness(1.05);
				}
			}

			.card-overlay,
			.launcher-slot {
				opacity: 1;
				transform: none;
				pointer-events: auto;
			}
		}

		&:active {
			transform: scale(0.98);
		}

		// :focus-within too, or the clicked card keeps the play overlay stuck.
		&.selection-mode:hover,
		&.selection-mode:focus-within {
			transform: none;

			.game-icon-container {
				border-color: var(--accent-primary);

				.play-overlay {
					opacity: 0 !important;
				}

				img.game-icon {
					transform: none;
					filter: none;
				}
			}

			&.selected .game-icon-container {
				box-shadow:
					0 2px 8px rgba(0, 0, 0, 0.35),
					0 0 10px var(--accent-glow);
			}

			:global(.selection-overlay .checkbox) {
				border-color: var(--accent-primary);
				transform: scale(1.12);
			}

			.launch-icon-large,
			.play-ripple {
				animation: none;
			}
		}

		&.running {
			.game-icon-container {
				border-color: var(--success, #44ffaa);
				animation: running-breathe 2.4s ease-in-out infinite;
			}
		}
	}

	.game-icon-container {
		aspect-ratio: 1;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.06);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		position: relative;
		cursor: pointer;
		transition: border-color var(--motion-base) var(--ease-out);
		z-index: 1;
		overflow: hidden;

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 3px;
		}

	}

	.icon-wrapper {
		position: absolute;
		inset: 2px;
		background: var(--bg-base);
		border-radius: calc(var(--radius-lg) - 2px);
		z-index: 2;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.game-icon {
		width: 100%;
		height: 100%;
		object-fit: cover;
		opacity: 0;
		transform: scale(0.96);
		transition: opacity 240ms var(--ease-out), transform 320ms var(--ease-spring);

		&:global(.loaded) {
			opacity: 1;
			transform: none;
		}
	}

	.system-icon {
		font-size: 64px;
		color: var(--text-muted);
		opacity: 0.3;
	}

	@keyframes play-bounce {
		0% {
			transform: scale(0.8);
		}
		55% {
			transform: scale(1.15);
		}
		75% {
			transform: scale(0.96);
		}
		100% {
			transform: scale(1);
		}
	}

	@keyframes ripple-pulse {
		0% {
			transform: scale(0.85);
			opacity: 0.28;
		}
		100% {
			transform: scale(1.45);
			opacity: 0;
		}
	}

	@keyframes play-pulse {
		0%,
		100% {
			transform: scale(1);
		}
		50% {
			transform: scale(1.08);
		}
	}

	@keyframes running-breathe {
		0%,
		100% {
			box-shadow: 0 0 6px rgba(68, 255, 170, 0.15);
		}
		50% {
			box-shadow: 0 0 16px rgba(68, 255, 170, 0.4);
		}
	}

	.play-overlay {
		position: absolute;
		inset: 0;
		background: rgba(var(--bg-base-rgb), 0.6);
		border: 1px solid rgba(255, 255, 255, 0.08);
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity var(--transition-fast);
		z-index: 3;
		border-radius: calc(var(--radius-lg) - 2px);

		.play-ripple {
			position: absolute;
			width: 84px;
			height: 84px;
			border-radius: 50%;
			background: rgba(var(--spotlight-rgb), 0.18);
			opacity: 0;
			pointer-events: none;
			z-index: 1;
		}

		.launch-icon-large {
			display: flex;
			align-items: center;
			justify-content: center;
			transform: scale(0.8);
			transition: transform 240ms var(--ease-spring);
			position: relative;
			z-index: 2;
		}
	}

	.game-card:hover .launch-icon-large,
	.game-card:focus-within .launch-icon-large {
		animation:
			play-bounce 480ms var(--ease-spring) forwards,
			play-pulse 1.8s ease-in-out 0.48s infinite;
	}

	.game-card:hover .play-ripple,
	.game-card:focus-within .play-ripple {
		animation: ripple-pulse 1.8s ease-out infinite;
	}

	.game-card:active .launch-icon-large {
		animation: none;
		transform: scale(0.92);
	}

	// Name and settings bar over the icon's bottom edge. Shown on hover.
	.card-overlay {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		z-index: 4;
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 18px 8px 8px 10px;
		background: linear-gradient(to top, rgba(0, 0, 0, 0.75), transparent);
		opacity: 0;
		transform: translateY(8px);
		transition:
			opacity var(--motion-quick) var(--ease-out),
			transform var(--motion-quick) var(--ease-out);
		pointer-events: none;

		:global(.name-viewport) {
			flex: 1;
		}

		:global(.game-name) {
			display: inline-block;
			font-size: 0.9rem;
			font-weight: 800;
			letter-spacing: -0.2px;
			line-height: 1.2;
			color: var(--text-main);
		}
	}

	.card-actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		min-width: 0;
	}

	.config-btn {
		flex-shrink: 0;
		width: 28px;
		height: 28px;
		padding: 0;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		background: rgba(var(--bg-surface-rgb), 0.7);
		border: 1px solid var(--glass-border);
		border-radius: var(--radius-pill);
		color: var(--text-main);
		cursor: pointer;
		transition:
			transform var(--motion-quick) var(--ease-spring),
			background var(--motion-quick) var(--ease-out),
			border-color var(--motion-quick) var(--ease-out);

		.material-icons {
			font-size: 16px;
		}

		&:hover {
			background: var(--bg-elevated);
			border-color: var(--accent-secondary);
			transform: rotate(90deg);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}

	// Launcher pill at the icon's top right. Shown on hover or touch.
	.launcher-slot {
		position: absolute;
		top: 6px;
		right: 6px;
		z-index: 5;
		opacity: 0;
		transition: opacity var(--motion-base) var(--ease-out);
		pointer-events: none;
	}

	// No hover: show the settings button without hover.
	@media (hover: none) {
		.card-overlay,
		.launcher-slot {
			opacity: 1;
			transform: none;
			pointer-events: auto;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.game-card {
			&:hover,
			&:focus-within,
			&:active {
				transform: none;
			}

		}

		.game-card.running .game-icon-container {
			animation: none;
			box-shadow: 0 0 10px rgba(68, 255, 170, 0.3);
		}

		.card-overlay,
		.launcher-slot {
			transform: none;
			transition: none;
		}

		.game-icon-container::before {
			display: none;
		}

		.game-icon {
			transition: none;
			transform: none;
			opacity: 1;
		}

		.play-ripple {
			display: none !important;
		}

		.game-card:hover .launch-icon-large,
		.game-card:focus-within .launch-icon-large {
			animation: none;
			transform: scale(1);
		}
	}
</style>
