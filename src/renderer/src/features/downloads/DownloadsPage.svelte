<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { onMount } from "svelte";
	import { speed } from "../../lib/features/format";
	import DownloadControls from "./DownloadControls.svelte";
	import DownloadRow from "./DownloadRow.svelte";
	import { DownloadsState } from "./downloadsState.svelte";
	import { groupDownloads, installLine, totalSpeed } from "./downloadGroups";

	const downloads = new DownloadsState();
	const groups = $derived(groupDownloads(downloads.items));
	const installing = $derived(downloads.installs.filter((install) => install.state !== "done"));
	const running = $derived(totalSpeed(downloads.items));
	const subtitle = $derived(
		`${groups.active.length} active` +
			(running ? ` · ${speed(running)}` : "") +
			(groups.failed.length ? ` · ${groups.failed.length} failed` : ""),
	);

	onMount(() => {
		downloads.load();
		return downloads.subscribe();
	});
</script>

<PageHeader title="Downloads" {subtitle} icon="downloading" />

{#if downloads.toast}
	<p class="notice" role="alert">{downloads.toast.message}</p>
{/if}

{#if downloads.error}
	<p class="notice" role="alert">{downloads.error}</p>
{/if}

<DownloadControls {downloads} />

{#if downloads.loading}
	<p class="empty">Loading downloads...</p>
{:else if downloads.items.length === 0 && installing.length === 0}
	<div class="empty">
		<p>No downloads yet.</p>
		<p class="hint">Pick a file from a game, or paste a direct link above.</p>
	</div>
{:else}
	{#if groups.active.length}
		<section aria-labelledby="active-title">
			<h2 id="active-title">In progress</h2>
			<ul class="list">
				{#each groups.active as item (item.id)}
					<DownloadRow {item} onAct={(action, id) => downloads.act(action, id)} />
				{/each}
			</ul>
		</section>
	{/if}

	{#if installing.length}
		<section aria-labelledby="install-title">
			<h2 id="install-title">Installing</h2>
			<ul class="list">
				{#each installing as install (install.downloadId)}
					<li class="install" class:error={install.state === "error"}>
						<p class="name">{install.title}</p>
						<p class="meta">{install.error ?? installLine(install)}</p>
					</li>
				{/each}
			</ul>
		</section>
	{/if}

	{#if groups.finished.length}
		<section aria-labelledby="finished-title">
			<h2 id="finished-title">Finished</h2>
			<ul class="list">
				{#each groups.finished as item (item.id)}
					<DownloadRow {item} onAct={(action, id) => downloads.act(action, id)} />
				{/each}
			</ul>
		</section>
	{/if}

	{#if groups.failed.length}
		<section aria-labelledby="failed-title">
			<h2 id="failed-title">Failed</h2>
			<ul class="list">
				{#each groups.failed as item (item.id)}
					<DownloadRow {item} onAct={(action, id) => downloads.act(action, id)} />
				{/each}
			</ul>
		</section>
	{/if}
{/if}

<style lang="scss">
	section {
		margin-bottom: 24px;
	}

	h2 {
		margin: 0 0 10px;
		font-size: 0.8rem;
		font-weight: 800;
		letter-spacing: 1px;
		text-transform: uppercase;
		color: var(--text-muted);
	}

	.list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.install {
		padding: 12px 14px;
		border-radius: var(--radius-md);
		background: var(--bg-surface);

		p {
			margin: 0;
		}

		&.error .meta {
			color: var(--danger);
		}
	}

	.name {
		font-weight: 700;
		font-size: 0.9rem;
	}

	.meta,
	.hint {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.notice {
		margin: 0 0 14px;
		padding: 10px 14px;
		border-radius: var(--radius-md);
		border: 1px solid var(--danger);
		font-size: 0.85rem;
	}

	.empty {
		padding: 40px 16px;
		text-align: center;
		color: var(--text-muted);

		p {
			margin: 0 0 6px;
		}
	}
</style>
