<script lang="ts">
	import CoverImage from "@components/shared/CoverImage.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { onMount } from "svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { libraryApi } from "../../lib/features/libraryApi";
	import { ago } from "../../lib/features/format";
	import type { UpdatesResponse } from "../../lib/features/types";

	let updates = $state<UpdatesResponse | null>(null);
	let checking = $state(false);
	let error = $state("");

	const subtitle = $derived(
		updates
			? `${updates.count} ${updates.count === 1 ? "game has" : "games have"} a newer release`
			: "Checking for new releases",
	);

	onMount(() => {
		check();
	});

	async function check() {
		checking = true;
		error = "";
		try {
			updates = await libraryApi.checkUpdates();
		} catch (err) {
			error = errorMessage(err);
		} finally {
			checking = false;
		}
	}
</script>

<PageHeader title="Updates" {subtitle} icon="system_update_alt">
	<button class="btn sm primary" disabled={checking} onclick={() => check()}>
		{checking ? "Checking..." : "Check now"}
	</button>
</PageHeader>

{#if error}
	<p class="notice" role="alert">{error}</p>
{/if}

{#each updates?.errors ?? [] as failure (failure)}
	<p class="notice" role="alert">{failure}</p>
{/each}

{#if updates && updates.games.length === 0}
	<div class="empty">
		<p>Everything is up to date.</p>
	</div>
{:else if updates}
	<ul class="list" aria-label="Games with updates">
		{#each updates.games as game (game.id)}
			<li class="row">
				<div class="cover">
					{#if game.cover || game.coverLocal}
						<CoverImage {game} />
					{/if}
				</div>
				<div class="info">
					<p class="title">{game.title}</p>
					<p class="versions">
						<span class="old">{game.installedVersion ?? "unknown"}</span>
						<span aria-hidden="true">→</span>
						<span class="new">{game.version ?? "new"}</span>
					</p>
				</div>
			</li>
		{/each}
	</ul>
{/if}

{#if updates && updates.history.length}
	<section aria-labelledby="history-title">
		<h2 id="history-title">Recent checks</h2>
		<ul class="history">
			{#each updates.history.slice(0, 25) as check (check.id)}
				<li>
					<span class="title">{check.title}</span>
					<span class="versions">{check.from ?? "?"} → {check.to ?? "?"}</span>
					{#if check.applied}<span class="tag">Installed</span>{/if}
					<span class="when">{ago(check.checkedAt)}</span>
				</li>
			{/each}
		</ul>
	</section>
{/if}

<style lang="scss">
	.list,
	.history {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 12px;
		border-radius: var(--radius-md);
		background: var(--bg-surface);
	}

	.cover {
		width: 48px;
		height: 64px;
		flex-shrink: 0;
		border-radius: var(--radius-sm);
		overflow: hidden;
		background: var(--bg-elevated);
	}

	.info {
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.title {
		margin: 0;
		font-weight: 700;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.versions {
		margin: 0;
		font-size: 0.85rem;
		color: var(--text-muted);

		.old {
			text-decoration: line-through;
		}

		.new {
			color: var(--accent-primary);
			font-weight: 700;
		}
	}

	section {
		margin-top: 28px;
	}

	h2 {
		margin: 0 0 10px;
		font-size: 0.8rem;
		font-weight: 800;
		letter-spacing: 1px;
		text-transform: uppercase;
		color: var(--text-muted);
	}

	.history li {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 8px 12px;
		border-radius: var(--radius-sm);
		background: var(--bg-surface);
		font-size: 0.85rem;

		.title {
			flex: 1;
			font-size: 0.85rem;
		}

		.versions,
		.when {
			color: var(--text-muted);
			white-space: nowrap;
		}

		.tag {
			font-size: 0.7rem;
			padding: 2px 8px;
			border-radius: var(--radius-pill);
			border: 1px solid var(--glass-border-bright);
		}
	}

	.notice {
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
			margin: 0;
		}
	}
</style>
