<script lang="ts">
	import { onMount } from "svelte";
	import Button from "@components/shared/Button.svelte";
	import CoverImage from "@components/shared/CoverImage.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import ScreenshotImage from "@components/shared/ScreenshotImage.svelte";
	import { notifications } from "@stores/notificationStore";
	import { libraryApi } from "../../lib/features/libraryApi";
	import type { Game, GameChange, GameFile, SourceAction } from "../../lib/features/types";
	import { ago, playtime } from "../../lib/features/format";
	import { libraryErrorMessage } from "./libraryError";
	import GameFiles from "./GameFiles.svelte";
	import GameInstall from "./GameInstall.svelte";

	let {
		game,
		sourceName,
		actions,
		onRemoved,
		onChanged,
	}: {
		game: Game;
		sourceName: string;
		/** The source's game-scope actions. */
		actions: SourceAction[];
		onRemoved: (id: number) => void;
		onChanged: (game: Game) => void;
	} = $props();

	let files = $state<GameFile[]>([]);
	let changes = $state<GameChange[]>([]);
	let notes = $state("");
	let confirmingRemove = $state(false);
	let runningAction = $state("");
	let busy = $state(false);
	let titleEl = $state<HTMLElement>();

	const versionLine = $derived(
		[sourceName, game.version ? `Version ${game.version}` : null].filter(Boolean).join(" · "),
	);
	const facts = $derived([game.developer, game.engine].filter(Boolean).join(" · "));

	onMount(() => {
		titleEl?.focus();
	});

	onMount(async () => {
		notes = game.notes ?? "";
		try {
			const detail = await libraryApi.get(game.id);
			files = detail.files;
			changes = detail.changes;
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		}
	});

	async function run(action: () => Promise<void>) {
		busy = true;
		try {
			await action();
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		} finally {
			busy = false;
		}
	}

	/** Runs one game-scope action. A refresh reloads the game and its files. */
	async function runGameAction(action: SourceAction) {
		runningAction = action.id;
		try {
			const result = await libraryApi.runAction(game.source, action.id, game.id);
			if (result.message) notifications.success(result.message);
			if (result.refresh) {
				const detail = await libraryApi.get(game.id);
				files = detail.files;
				changes = detail.changes;
				onChanged(detail.game);
			}
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		} finally {
			runningAction = "";
		}
	}

	function toggleFavorite() {
		run(async () => {
			onChanged(await libraryApi.setFavorite(game.id, !game.favorite));
		});
	}

	function launch() {
		run(() => libraryApi.launch(game.id));
	}

	function saveNotes() {
		if (notes === (game.notes ?? "")) return;
		run(async () => {
			onChanged(await libraryApi.update(game.id, { notes }));
		});
	}

	function remove() {
		run(async () => {
			await libraryApi.remove(game.id);
			onRemoved(game.id);
		});
	}

</script>

<div class="detail">
	<header class="hero">
		<div class="cover">
			{#if game.cover || game.coverLocal}
				<CoverImage {game} />
			{:else}
				<Icon name="videogame_asset" size={40} />
			{/if}
		</div>
		<div class="summary">
			<h2 bind:this={titleEl} tabindex="-1">{game.title}</h2>
			<p class="meta">{versionLine}</p>

			<div class="actions">
				{#if game.installed && game.exePath}
					<Button size="sm" variant="primary" disabled={busy} onclick={launch}>Play</Button>
				{/if}
				{#each actions as action (action.id)}
					<Button
						size="sm"
						icon={action.icon}
						loading={runningAction === action.id}
						disabled={runningAction !== "" || busy}
						onclick={() => runGameAction(action)}
					>
						{action.label}
					</Button>
				{/each}
				<Button size="sm" aria-pressed={game.favorite} disabled={busy} onclick={toggleFavorite}>
					{game.favorite ? "Unfavorite" : "Favorite"}
				</Button>
				<a class="btn sm" href={game.url} target="_blank" rel="noreferrer">Open page</a>
				{#if confirmingRemove}
					<Button size="sm" variant="danger" disabled={busy} onclick={remove}>Confirm remove</Button>
					<Button size="sm" onclick={() => (confirmingRemove = false)}>Cancel</Button>
				{:else}
					<Button size="sm" onclick={() => (confirmingRemove = true)}>Remove</Button>
				{/if}
			</div>
		</div>
	</header>

	{#if game.screenshots.length}
		<ul class="shots" aria-label="Screenshots">
			{#each game.screenshots as shot, index (shot)}
				<li>
					<ScreenshotImage
						gameId={game.id}
						{index}
						local={game.screenshotLocals[index] ?? null}
						remote={shot}
					/>
				</li>
			{/each}
		</ul>
	{/if}

	{#if game.description}
		<section>
			<h3>About</h3>
			<p class="description">{game.description}</p>
		</section>
	{/if}

	<section class="facts-block">
		{#if facts}
			<p class="muted">{facts}</p>
		{/if}
		{#if game.tags.length}
			<ul class="tags" aria-label="Tags">
				{#each game.tags as tag (tag)}
					<li>{tag}</li>
				{/each}
			</ul>
		{/if}
		<p class="muted">Played {playtime(game.playtimeS)} · Last played {ago(game.lastPlayedAt)}</p>
	</section>

	<GameInstall {game} {onChanged} />

	<GameFiles gameId={game.id} {files} />

	<section>
		<h3 id="notes-title">Notes</h3>
		<textarea
			aria-labelledby="notes-title"
			class="input"
			rows="4"
			bind:value={notes}
			onblur={saveNotes}
			placeholder="Write a note about this game"
		></textarea>
	</section>

	{#if changes.length}
		<section>
			<h3>Changes</h3>
			<ul class="changes">
				{#each changes as change (change.id)}
					<li>
						<span>{change.from ?? "?"} → {change.to ?? "?"}</span>
						<span class="muted">{ago(change.checkedAt)}</span>
					</li>
				{/each}
			</ul>
		</section>
	{/if}
</div>

<style lang="scss">
	/* The game panel is narrow: the cover sits beside the title and the actions wrap below. */
	.detail {
		display: flex;
		flex-direction: column;
		gap: 24px;
		min-width: 0;
	}

	.hero {
		display: flex;
		gap: 20px;
		flex-wrap: wrap;
	}

	.cover {
		width: 128px;
		aspect-ratio: 3 / 4;
		flex-shrink: 0;
		border-radius: var(--radius-md);
		overflow: hidden;
		background: var(--bg-elevated);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-muted);
	}

	.summary {
		display: flex;
		flex-direction: column;
		gap: 10px;
		min-width: 0;
		flex: 1 1 180px;

		h2 {
			margin: 0;
			font-size: 1.75rem;
			font-weight: 700;
			line-height: 1.2;
			overflow-wrap: anywhere;

			&:focus {
				outline: none;
			}
		}
	}

	.meta {
		margin: 0;
		font-size: 1rem;
		color: var(--text-muted);
	}

	.muted {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.875rem;
	}

	.facts-block {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		list-style: none;
		margin: 0;
		padding: 0;

		li {
			font-size: 0.875rem;
			padding: 3px 10px;
			border-radius: var(--radius-pill);
			background: var(--bg-elevated);
			color: var(--text-muted);
		}
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.shots {
		display: flex;
		gap: 10px;
		overflow-x: auto;
		list-style: none;
		margin: 0;
		padding: 0 0 4px;

		li {
			flex: 0 0 240px;
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
	}

	.description {
		max-width: 68ch;
		white-space: pre-line;
		font-size: 1rem;
		line-height: 1.5;
		margin: 0;
	}

	textarea {
		resize: vertical;
		font-size: 1rem;
	}

	.changes {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 0.875rem;

		li {
			display: flex;
			justify-content: space-between;
			gap: 12px;
		}
	}
</style>
