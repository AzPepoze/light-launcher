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
		font-size: 1.2rem;
		font-weight: 800;
		color: var(--text-dim);
		opacity: 0.4;
		text-transform: uppercase;
		letter-spacing: 2px;
		margin-bottom: 20px;
	}

	.sessions-section {
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		gap: 20px;
		background: linear-gradient(
			135deg,
			rgba(239, 68, 68, 0.1) 0%,
			rgba(239, 68, 68, 0.02) 100%
		);
		padding: 24px;
		border-radius: 24px;
		border: 1px solid rgba(239, 68, 68, 0.2);
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
		animation: slide-down 0.5s cubic-bezier(0.23, 1, 0.32, 1);

		.section-title {
			margin-bottom: 10px;
			color: #ef4444;
		}
	}

	.sessions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 16px;
	}

	.session-card {
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 16px;
		padding: 14px 20px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		transition: all 0.3s;

		&:hover {
			border-color: rgba(239, 68, 68, 0.4);
			background: var(--glass-surface);
			transform: translateX(4px);
		}

		.session-info {
			display: flex;
			flex-direction: column;
			gap: 2px;
		}

		.session-title {
			font-weight: 800;
			color: var(--text-main);
			font-size: 1rem;
			letter-spacing: -0.3px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			max-width: 200px;
		}

		.session-pid {
			font-size: 0.7rem;
			color: var(--text-dim);
			font-family: monospace;
			font-weight: 600;
		}

		.kill-btn {
			background: #ef4444;
			color: #fff;
			padding: 8px 16px;
			border: none;
			border-radius: 10px;
			font-size: 0.75rem;
			font-weight: 800;
			cursor: pointer;
			transition: all 0.2s;
			box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);

			&:hover {
				filter: brightness(1.2);
				transform: translateY(-2px);
				box-shadow: 0 6px 16px rgba(239, 68, 68, 0.4);
			}

			&:active {
				transform: translateY(0);
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
