<script lang="ts">
	import { fly } from "svelte/transition";

	export let sessions: any[] = [];
	export let onKill: (pid: number, name: string) => void;
</script>

{#if sessions.length > 0}
	<div class="sessions-section">
		<h2 class="section-title">Running Sessions</h2>
		<div class="sessions-grid">
			{#each sessions as session}
				<div
					class="session-card"
					in:fly={{ y: -20, duration: 400 }}
				>
					<div class="session-info">
						<div class="session-title">
							{session.gameName}
						</div>
						<div class="session-pid">PID: {session.pid}</div>
					</div>
					<button
						class="kill-btn"
						on:click={() =>
							onKill(session.pid, session.gameName)}
					>
						Terminate
					</button>
				</div>
			{/each}
		</div>
	</div>
{/if}

<style lang="scss">
	.section-title {
		font-size: 1.4rem;
		font-weight: 800;
		color: var(--danger);
		text-transform: uppercase;
		letter-spacing: 2px;
		margin-bottom: 16px;
	}

	.sessions-section {
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		gap: 16px;
		background: var(--bg-surface);
		padding: 28px;
		border-radius: var(--radius-lg);
		border: 2px solid var(--danger);
		box-shadow: var(--glass-shadow);
		animation: slide-down 0.4s var(--ease-spring);
	}

	.sessions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 20px;
	}

	.session-card {
		background: var(--bg-elevated);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		padding: 18px 24px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		transition: transform var(--transition-spring), border-color var(--transition-fast);

		&:hover {
			border-color: var(--danger);
			transform: scale(1.03) translateX(4px);
		}

		.session-info {
			display: flex;
			flex-direction: column;
			gap: 4px;
		}

		.session-title {
			font-weight: 800;
			color: var(--text-main);
			font-size: 1.1rem;
			letter-spacing: -0.3px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			max-width: 180px;
		}

		.session-pid {
			font-size: 0.75rem;
			color: var(--text-muted);
			font-family: monospace;
			font-weight: 600;
		}

		.kill-btn {
			background: var(--danger);
			color: #ffffff;
			padding: 10px 22px;
			border: none;
			border-radius: var(--radius-pill);
			font-size: 0.8rem;
			font-weight: 800;
			cursor: pointer;
			transition: transform var(--transition-spring), filter var(--transition-fast), box-shadow var(--transition-fast);
			box-shadow: 0 4px 12px rgba(255, 68, 102, 0.2);

			&:hover {
				filter: brightness(1.15);
				transform: scale(1.08);
				box-shadow: 0 6px 16px rgba(255, 68, 102, 0.35);
			}

			&:active {
				transform: scale(0.95);
			}
		}
	}

	@keyframes slide-down {
		from {
			opacity: 0;
			transform: translateY(-20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style>
