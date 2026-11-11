<script lang="ts">
	import { platform } from "@stores/platformStore";
	import { downloadsApi } from "../../lib/features/downloadsApi";
	import { bytes } from "../../lib/features/format";
	import { isFileForOs } from "../../lib/features/platform";
	import type { GameFile } from "../../lib/features/types";
	import { PREPARING_TEXT, queueGameFile } from "./gameFileQueue";

	let { gameId, files }: { gameId: number; files: GameFile[] } = $props();

	let message = $state("");
	let preparingId = $state<number | null>(null);
	let flashedId = $state<number | null>(null);

	async function queue(file: GameFile) {
		message = "";
		message = await queueGameFile(gameId, file, enqueueAndFlash, (id) => {
			preparingId = id;
		});
	}

	async function enqueueAndFlash(targetGameId: number, fileId: number): Promise<number> {
		const queueId = await downloadsApi.enqueueFile(targetGameId, fileId);
		flashedId = fileId;
		return queueId;
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
							{#if !isFileForOs(file.platform, $platform.os)}
								· Other OS
							{/if}
							{#if file.requiresAuth}
								· Needs sign-in
							{/if}
						</p>
					</div>
					<button
						class="btn sm download"
						class:flash={flashedId === file.id}
						disabled={preparingId !== null}
						aria-busy={preparingId === file.id}
						aria-label="Download {file.label}"
						onclick={() => queue(file)}
						onanimationend={() => (flashedId = null)}
					>
						{#if preparingId === file.id}
							<span class="spinner" aria-hidden="true"></span>
						{/if}
						{preparingId === file.id ? PREPARING_TEXT : "Download"}
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
	@keyframes done-flash {
		from {
			background-color: var(--success);
			border-color: var(--success);
		}
	}

	section {
		display: flex;
		flex-direction: column;
		gap: 10px;

		h3 {
			margin: 0;
			font-size: 1.25rem;
			font-weight: 600;
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
			padding: 12px 14px;
			border-radius: var(--radius-md);
			background: var(--bg-surface);
		}
	}

	.info {
		flex: 1;
		min-width: 0;
	}

	.label {
		font-weight: 600;
		font-size: 1rem;
		overflow-wrap: anywhere;
	}

	.download {
		gap: 8px;

		&.flash {
			animation: done-flash var(--motion-slow) var(--ease-out);
		}
	}

	.spinner {
		width: 12px;
		height: 12px;
		border: 2px solid currentColor;
		border-right-color: transparent;
		border-radius: 50%;
		animation: ll-spin 0.8s linear infinite;
	}

	.muted {
		color: var(--text-muted);
		font-size: 0.875rem;
	}
</style>
