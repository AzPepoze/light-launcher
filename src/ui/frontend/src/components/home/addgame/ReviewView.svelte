<script lang="ts">
	import { fade } from "svelte/transition";

	export let foundExecutables: {
		path: string;
		name: string;
		icon: string | null;
	}[];
	export let discardedExecutables: Set<string>;
	export let onToggleDiscard: (path: string) => void;
</script>

<div class="review-view" transition:fade={{ duration: 200 }}>
	<div class="found-grid">
		{#each foundExecutables as exe}
			<div
				class="exe-card"
				class:discarded={discardedExecutables.has(exe.path)}
				on:click={() => onToggleDiscard(exe.path)}
				role="button"
				tabindex="0"
				on:keydown={(e) =>
					e.key === "Enter" && onToggleDiscard(exe.path)}
			>
				<div class="card-content">
					<div class="icon-container">
						{#if exe.icon}
							<img
								src={exe.icon}
								alt={exe.name}
								class="game-icon"
							/>
						{:else}
							<div class="icon-placeholder">
								<span
									class="material-icons"
									style="font-size: 24px;"
									>rocket_launch</span
								>
							</div>
						{/if}
					</div>
					<div class="name" title={exe.path}>{exe.name}</div>
				</div>
				<div class="discard-overlay">
					<span class="material-icons">delete</span>
				</div>
			</div>
		{/each}
	</div>
</div>

<style lang="scss">
	.review-view {
		display: flex;
		flex-direction: column;
	}

	.found-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
		gap: 16px;
	}

	.exe-card {
		position: relative;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 20px;
		padding: 20px 16px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		cursor: pointer;
		overflow: hidden;

		&::after {
			content: "";
			position: absolute;
			inset: 0;
			background: radial-gradient(
				circle at top left,
				var(--glass-low-alpha),
				transparent 70%
			);
			opacity: 0;
			transition: opacity 0.3s;
		}

		.discard-overlay {
			position: absolute;
			top: 0;
			left: 0;
			width: 100%;
			height: 100%;
			background: rgba(239, 68, 68, 0.2);
			display: flex;
			align-items: center;
			justify-content: center;
			opacity: 0;
			transition: all 0.2s;
			color: #fff;
		}

		&.discarded {
			filter: grayscale(1);
			opacity: 0.4;
			border-color: rgba(239, 68, 68, 0.3);

			.discard-overlay {
				opacity: 1;
			}
		}

		&:hover:not(.discarded) {
			background: var(--glass-border-bright);
			border-color: var(--accent-primary);
			transform: translateY(-4px);
		}

		.card-content {
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 12px;
			width: 100%;
		}

		.icon-container {
			width: 56px;
			height: 56px;
			display: flex;
			align-items: center;
			justify-content: center;

			.game-icon {
				width: 100%;
				height: 100%;
				object-fit: contain;
			}

			.icon-placeholder {
				color: var(--text-dim);
				opacity: 0.3;
			}
		}

		.name {
			font-size: 0.85rem;
			font-weight: 700;
			color: var(--text-main);
			text-align: center;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			width: 100%;
		}
	}
</style>
