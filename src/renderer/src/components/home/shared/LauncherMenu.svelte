<script lang="ts">
	import AppMenu from "@components/shared/AppMenu.svelte";
	import { launcherCountLabel, launcherKindLabel } from "@lib/features/launcher";
	import type { GameInfo } from "@shared";

	export let game: GameInfo;
	/** Receives the picked launcher path, or null for automatic. */
	export let onChoose: (launcherPath: string | null) => void = () => {};
	/** Cards paint-clip their box, so the panel has to open back over the cover. */
	export let placement: "bottom" | "top" = "bottom";

	let isOpen = false;

	function choose(launcherPath: string | null) {
		isOpen = false;
		onChoose(launcherPath);
	}
</script>

{#if game.launcherCount > 1}
	<div class="launcher-menu" class:up={placement === "top"}>
		<!-- Stops propagation so the window click that closes AppMenu does not run for this button. -->
		<button
			type="button"
			class="launcher-note"
			class:open={isOpen}
			title="Choose launcher"
			aria-haspopup="menu"
			aria-expanded={isOpen}
			on:click|stopPropagation={() => (isOpen = !isOpen)}
		>
			<span class="material-icons">tune</span>
			<span>{launcherCountLabel(game.launcherCount)}</span>
		</button>
		<AppMenu visible={isOpen} mode="anchor" {placement} onClose={() => (isOpen = false)}>
			<button type="button" class="menu-item" on:click|stopPropagation={() => choose(null)}>
				<span class="material-icons">autorenew</span>
				<span>Use automatic</span>
			</button>
			{#each game.launchers as launcher (launcher.path)}
				<button type="button" class="menu-item" on:click|stopPropagation={() => choose(launcher.path)}>
					<span class="material-icons">{launcher.primary ? "check" : "play_arrow"}</span>
					<span class="launcher-text">
						<span class="launcher-name">{launcher.name}</span>
						<span class="launcher-kind">{launcherKindLabel(launcher.kind)}</span>
					</span>
				</button>
			{/each}
		</AppMenu>
	</div>
{/if}

<style lang="scss">
	.launcher-menu {
		position: relative;
		display: inline-flex;

		// Keep the panel inside the card, whose box is clipped by content-visibility.
		&.up {
			--app-menu-max-height: 132px;
			--app-menu-anchor-width: 158px;
		}
	}

	.launcher-note {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 2px 8px;
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-sm);
		background: rgba(255, 255, 255, 0.04);
		color: var(--text-muted);
		font-size: 0.7rem;
		font-weight: 700;
		white-space: nowrap;
		cursor: pointer;
		transition:
			color var(--motion-quick),
			border-color var(--motion-quick),
			background var(--motion-quick);

		.material-icons {
			font-size: 14px;
		}

		&:hover,
		&.open {
			color: var(--text-main);
			border-color: var(--accent-primary);
		}
	}

	.launcher-text {
		display: flex;
		flex-direction: column;
		min-width: 0;
		text-align: left;
	}

	.launcher-name {
		font-weight: 700;
		color: var(--text-main);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.launcher-kind {
		font-size: 0.7rem;
		color: var(--text-muted);
	}

	@media (prefers-reduced-motion: reduce) {
		.launcher-note {
			transition: none;
		}
	}
</style>
