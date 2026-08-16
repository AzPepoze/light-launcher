<script lang="ts">
	import { fade } from "svelte/transition";

	export let onAddFile: () => void;
	export let onAddFolder: () => void;
</script>

<div class="selection-view" transition:fade={{ duration: 200 }}>
	<div class="mode-buttons">
		<button class="mode-btn" on:click={onAddFile}>
			<div class="icon-box">
				<span class="material-icons" style="font-size: 28px;"
					>add_box</span
				>
			</div>
			<div class="btn-text">
				<span class="title">Single File</span>
				<span class="desc">Add a specific .exe game file</span>
			</div>
		</button>

		<button class="mode-btn" on:click={onAddFolder}>
			<div class="icon-box">
				<span class="material-icons" style="font-size: 28px;"
					>create_new_folder</span
				>
			</div>
			<div class="btn-text">
				<span class="title">Scan Folder</span>
				<span class="desc">Batch add games from directory</span>
			</div>
		</button>
	</div>
</div>

<style lang="scss">
	.selection-view {
		display: flex;
		flex-direction: column;
	}

	.mode-buttons {
		display: grid;
		grid-template-columns: 1fr;
		gap: 16px;
	}

	.mode-btn {
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 20px;
		padding: 24px;
		display: flex;
		align-items: center;
		gap: 24px;
		cursor: pointer;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		text-align: left;
		position: relative;
		overflow: hidden;

		&::after {
			content: "";
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			bottom: 0;
			background: radial-gradient(
				circle at top left,
				var(--glass-high-alpha),
				transparent 70%
			);
			opacity: 0;
			transition: opacity 0.3s;
		}

		&:hover {
			background: var(--glass-border-bright);
			border-color: var(--accent-primary);
			transform: translateY(-4px);
			box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);

			&::after {
				opacity: 1;
			}

			.icon-box {
				background: var(--accent-primary);
				color: var(--glass-bg);
				transform: rotate(-10deg) scale(1.1);
			}
		}

		.icon-box {
			width: 64px;
			height: 64px;
			background: var(--glass-bg);
			border-radius: 18px;
			display: flex;
			align-items: center;
			justify-content: center;
			color: var(--text-main);
			transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
			border: 1px solid var(--glass-border);
		}

		.btn-text {
			.title {
				display: block;
				font-size: 1.25rem;
				font-weight: 800;
				color: var(--text-main);
				margin-bottom: 4px;
			}
			.desc {
				font-size: 0.9rem;
				color: var(--text-dim);
			}
		}
	}
</style>
