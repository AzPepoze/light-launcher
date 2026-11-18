<script lang="ts">
	import Icon from "@components/shared/Icon.svelte";

	export let runningToolName: string = "";
	/** Why the tools cannot run yet. Empty when they are ready. */
	export let blockReason: string = "";
	export let onRunTool: (tool: string) => void;

	const tools = [
		{ id: "winecfg", icon: "settings", title: "Winecfg", description: "Wine settings" },
		{ id: "regedit", icon: "edit", title: "Regedit", description: "Edit the registry" },
		{ id: "cmd", icon: "terminal", title: "CMD", description: "Command prompt" },
		{ id: "winetricks", icon: "auto_fix_high", title: "Winetricks", description: "Install Windows components" },
		{ id: "taskmgr", icon: "bar_chart", title: "Task Manager", description: "See running processes" },
		{ id: "explorer", icon: "folder", title: "Explorer", description: "Browse the prefix files" }
	];
</script>

<div class="tools-grid">
	{#each tools as tool (tool.id)}
		{@const isRunning = runningToolName === tool.id}
		<button
			type="button"
			class="tool-tile"
			disabled={!!blockReason || isRunning}
			onclick={() => onRunTool(tool.id)}
		>
			<span class="tool-icon">
				<Icon name={isRunning ? "progress_activity" : tool.icon} size={36} class={isRunning ? "spin-icon" : ""} />
			</span>
			<span class="tool-text">
				<span class="tool-title">{tool.title}</span>
				{#if blockReason}
					<span class="tool-line reason">{blockReason}</span>
				{:else}
					<span class="tool-line">{isRunning ? "Opening…" : tool.description}</span>
				{/if}
			</span>
		</button>
	{/each}
</div>

<style lang="scss">
	.tools-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 12px;
	}

	.tool-tile {
		display: flex;
		align-items: center;
		gap: 14px;
		min-width: 0;
		padding: 16px;
		background: var(--bg-elevated);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
		color: var(--text-main);
		text-align: left;
		cursor: pointer;
		transition: border-color var(--transition-fast), background var(--transition-fast);

		&:hover:not(:disabled) {
			border-color: var(--accent-primary);
			background: var(--bg-surface);
		}

		&:disabled {
			cursor: not-allowed;

			.tool-title {
				color: var(--text-muted);
			}
		}

		:global(.spin-icon) {
			animation: ll-spin 1.2s linear infinite;
		}
	}

	.tool-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		flex-shrink: 0;
		color: var(--accent-primary);
	}

	.tool-text {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}

	.tool-title {
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--text-main);
	}

	.tool-line {
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--text-muted);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;

		&.reason {
			color: var(--warning);
		}
	}

	@media (max-width: 900px) {
		.tools-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 600px) {
		.tools-grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
