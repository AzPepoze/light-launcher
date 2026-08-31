<script lang="ts">
	import CardRunningIndicator from "./shared/CardRunningIndicator.svelte";
	import CardSelectionCheckbox from "./shared/CardSelectionCheckbox.svelte";
	import AnimatedPlayIcon from "./shared/AnimatedPlayIcon.svelte";
	import { getDominantColor } from "@lib/dominantColor";

	export let game: any;
	export let icon: string = "";
	export let isRunning: boolean = false;
	export let isSelectionMode: boolean = false;
	export let isSelected: boolean = false;
	export let onLaunch: (game: any) => void = () => {};
	export let onConfigure: (game: any) => void = () => {};
	export let onSelect: (game: any, shiftKey: boolean) => void = () => {};

	let cardElement: HTMLElement;
	let spotlightX = 50;
	let spotlightY = 50;
	let tiltX = 0;
	let tiltY = 0;
	let isHovering = false;
	let rafId: number | null = null;
	let pendingX = 50;
	let pendingY = 50;
	let pendingTiltX = 0;
	let pendingTiltY = 0;

	let dominantRGB = "255,255,255";
	let iconNonce = 0;

	$: if (icon) {
		const current = ++iconNonce;
		getDominantColor(icon).then((c) => {
			if (current !== iconNonce) return;
			if (c) dominantRGB = `${c[0]},${c[1]},${c[2]}`;
		});
	} else {
		dominantRGB = "255,255,255";
	}

	function handleLaunch(event?: MouseEvent) {
		if (isSelectionMode) {
			onSelect(game, event ? event.shiftKey : false);
			return;
		}
		onLaunch(game);
	}

	function handleConfigure() {
		onConfigure(game);
	}

	function scheduleFrame() {
		if (rafId !== null) return;
		rafId = requestAnimationFrame(() => {
			rafId = null;
			spotlightX = pendingX;
			spotlightY = pendingY;
			tiltX = pendingTiltX;
			tiltY = pendingTiltY;
		});
	}

	function handleMouseMove(event: MouseEvent) {
		if (isSelectionMode) return;
		const rect = cardElement.getBoundingClientRect();
		const x = ((event.clientX - rect.left) / rect.width) * 100;
		const y = ((event.clientY - rect.top) / rect.height) * 100;
		pendingX = Math.max(0, Math.min(100, x));
		pendingY = Math.max(0, Math.min(100, y));

		// subtle 3D tilt ±4deg
		const centerX = rect.left + rect.width / 2;
		const centerY = rect.top + rect.height / 2;
		pendingTiltY = ((event.clientX - centerX) / rect.width) * 8;
		pendingTiltX = ((centerY - event.clientY) / rect.height) * 8;
		pendingTiltX = Math.max(-4, Math.min(4, pendingTiltX));
		pendingTiltY = Math.max(-4, Math.min(4, pendingTiltY));
		scheduleFrame();
	}

	function handleMouseEnter() {
		isHovering = true;
	}

	function handleMouseLeave() {
		isHovering = false;
		pendingTiltX = 0;
		pendingTiltY = 0;
		if (rafId !== null) {
			cancelAnimationFrame(rafId);
			rafId = null;
		}
		tiltX = 0;
		tiltY = 0;
	}
</script>

<div
	bind:this={cardElement}
	class="game-card"
	class:running={isRunning}
	class:selection-mode={isSelectionMode}
	class:selected={isSelected}
	class:hovering={isHovering}
	style="--mx: {spotlightX}%; --my: {spotlightY}%; --tilt-x: {tiltX}deg; --tilt-y: {tiltY}deg; --spotlight-rgb: {dominantRGB};"
	on:mousemove={handleMouseMove}
	on:mouseenter={handleMouseEnter}
	on:mouseleave={handleMouseLeave}
	role="group"
>
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="game-icon-container"
		on:click={handleLaunch}
		role="button"
		tabindex="0"
	>
		{#if isSelectionMode}
			<CardSelectionCheckbox {isSelected} />
		{/if}

		{#if isRunning}
			<CardRunningIndicator />
		{/if}

		<div class="icon-wrapper">
			{#if icon}
				<img src={icon} alt={game.name} class="game-icon" loading="lazy" />
			{:else}
				<span class="material-icons system-icon">rocket_launch</span>
			{/if}
		</div>

		<div class="play-overlay">
			<span class="play-ripple" aria-hidden="true"></span>
			<span class="launch-icon-large">
				<AnimatedPlayIcon size={64} />
			</span>
		</div>
	</div>

	<div class="game-footer">
		<span class="game-name" title={game.name}>{game.name}</span>
		<button
			class="config-btn"
			title="Configure"
			on:click|stopPropagation={handleConfigure}
		>
			<span class="material-icons">settings</span>
		</button>
	</div>
</div>

<style lang="scss">
	.game-card {
		display: flex;
		flex-direction: column;
		gap: 12px;
		transition: transform 240ms var(--ease-spring), filter var(--transition-fast);
		position: relative;
		width: 100%;
		max-width: 200px;
		margin: 6px;
		perspective: 900px;
		transform-style: preserve-3d;
		--mx: 50%;
		--my: 50%;
		--spotlight-rgb: 255, 255, 255;

		&.selected {
			.game-icon-container {
				border-color: var(--accent-primary);
				box-shadow: 0 0 24px var(--accent-glow);
			}
		}

		&:hover,
		&:focus-within {
			transform: perspective(900px) rotateX(var(--tilt-x)) rotateY(var(--tilt-y));

			.game-icon-container {
				border-color: rgba(var(--spotlight-rgb), 0.25);
				box-shadow:
					0 2px 8px rgba(0, 0, 0, 0.25),
					0 12px 32px rgba(var(--spotlight-rgb), 0.14);

				&::before {
					opacity: 1;
				}

				.play-overlay {
					opacity: 1;
				}

				img.game-icon {
					transform: scale(1.08);
					filter: brightness(1.05);
				}
			}

			.game-footer .game-name {
				color: var(--text-main);
				white-space: normal;
				overflow: visible;
				word-break: break-all;
			}

			.config-btn {
				opacity: 1;
				visibility: visible;
				transform: scale(1);
			}
		}

		&:active {
			transform: perspective(900px) scale(0.98) rotateX(var(--tilt-x)) rotateY(var(--tilt-y));
		}

		&.selection-mode:hover {
			transform: none;

			.game-icon-container::before {
				opacity: 0 !important;
			}

			.play-overlay {
				opacity: 0 !important;
			}
		}

		&.running {
			.game-icon-container {
				border-color: var(--success, #44ffaa);
				box-shadow:
					0 0 24px rgba(68, 255, 170, 0.35),
					0 8px 24px rgba(0, 0, 0, 0.3);

				&::before {
					opacity: 0.06;
				}
			}
		}

		&.running:hover .game-icon-container::before {
			opacity: 0.06;
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
		transition: transform 240ms var(--ease-spring), border-color var(--transition-fast), box-shadow var(--transition-fast);
		z-index: 1;
		overflow: hidden;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: radial-gradient(
				420px circle at var(--mx) var(--my),
				rgba(var(--spotlight-rgb), 0.15),
				transparent 42%
			);
			opacity: 0;
			transition: opacity var(--transition-fast);
			z-index: 4;
			pointer-events: none;
			border-radius: inherit;
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
		transition: transform 0.5s var(--ease-spring), filter var(--transition-fast);
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

	.game-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0 4px;

		.game-name {
			font-size: 0.95rem;
			font-weight: 800;
			color: var(--text-muted);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			flex: 1;
			transition: color var(--transition-fast);
			letter-spacing: -0.2px;
			line-height: 1.2;
		}

		.config-btn {
			background: var(--bg-surface);
			border: 2px solid rgba(255, 255, 255, 0.05);
			padding: 8px;
			border-radius: var(--radius-pill);
			cursor: pointer;
			display: flex;
			align-items: center;
			justify-content: center;
			transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast);
			color: var(--text-main);
			margin-left: 6px;

			&:hover {
				background: var(--bg-elevated);
				transform: rotate(90deg) scale(1.1);
				border-color: var(--accent-secondary);
			}

			.material-icons {
				font-size: 16px;
			}
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.game-card {
			&:hover,
			&:focus-within {
				transform: none;
			}
		}

		.game-icon-container::before {
			display: none;
		}

		.game-icon {
			transition: none;
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
