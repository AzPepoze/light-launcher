<script lang="ts">
	import { errorMessage } from "../../lib/features/rpc";
	import { downloadsApi } from "../../lib/features/downloadsApi";
	import { bytes } from "../../lib/features/format";
	import type { GameFile } from "../../lib/features/types";

	let { gameId, files }: { gameId: number; files: GameFile[] } = $props();

	let message = $state("");
	let queuingUrl = $state("");

	async function queue(file: GameFile) {
		queuingUrl = file.url;
		message = "";
		try {
			await downloadsApi.enqueue(file.url, gameId);
			message = `Added "${file.label}" to downloads.`;
		} catch (err) {
			message = errorMessage(err);
		} finally {
			queuingUrl = "";
		}
	}
</script>

<section>
	<h3 id="files-title">Files</h3>

	{#if files.length === 0}
		<p class="muted">No files found for this game.</p>
	{:else}
		<ul class="files" aria-labelledby="files-title">
			{#each files as file (file.id)}
				<li>
					<div class="info">
						<p class="label">{file.label}</p>
						<p class="muted">
							{[file.platform, file.host, file.size ? bytes(file.size) : null]
								.filter(Boolean)
								.join(" · ")}
							{#if file.requiresAuth}
								· Needs sign-in
							{/if}
						</p>
					</div>
					<button
						class="btn sm"
						disabled={queuingUrl === file.url}
						aria-label="Download {file.label}"
						onclick={() => queue(file)}
					>
						Download
					</button>
				</li>
			{/each}
		</ul>
	{/if}

	{#if message}
		<p class="muted" role="status">{message}</p>
	{/if}
</section>

<style lang="scss">
	section {
		display: flex;
		flex-direction: column;
		gap: 8px;

		h3 {
			margin: 0;
			font-size: 0.8rem;
			font-weight: 800;
			letter-spacing: 1px;
			text-transform: uppercase;
			color: var(--text-muted);
		}

		p {
			margin: 0;
		}
	}

	.files {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;

		li {
			display: flex;
			align-items: center;
			gap: 12px;
			padding: 10px 12px;
			border-radius: var(--radius-md);
			background: var(--bg-surface);
		}
	}

	.info {
		flex: 1;
		min-width: 0;
	}

	.label {
		font-weight: 700;
		font-size: 0.9rem;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.muted {
		color: var(--text-muted);
		font-size: 0.8rem;
	}
</style>
