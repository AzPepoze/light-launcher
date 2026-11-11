<script lang="ts">
	import { onMount } from "svelte";
	import CoverImage from "@components/shared/CoverImage.svelte";
	import ScreenshotImage from "@components/shared/ScreenshotImage.svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { libraryApi } from "../../lib/features/libraryApi";
	import type { Game, GameChange, GameFile } from "../../lib/features/types";
	import { ago, playtime } from "../../lib/features/format";
	import GameFiles from "./GameFiles.svelte";
	import GameInstall from "./GameInstall.svelte";
	import { groupGames } from "./libraryGroups";
	import type { LibraryState } from "./libraryState.svelte";

	let {
		game,
		library,
		onBack,
		onRemoved,
	}: {
		game: Game;
		library: LibraryState;
		onBack: () => void;
		onRemoved: (id: number) => void;
	} = $props();

	let files = $state<GameFile[]>([]);
	let changes = $state<GameChange[]>([]);
	let notes = $state("");
	let confirmingRemove = $state(false);
	let message = $state("");
	let busy = $state(false);
	let closing = $state(false);
	let titleEl = $state<HTMLElement>();

	const sourceName = $derived(groupGames([game], library.sources)[0]?.label ?? game.source);
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
			message = errorMessage(err);
		}
	});

	async function run(action: () => Promise<void>) {
		busy = true;
		message = "";
		try {
			await action();
		} catch (err) {
			message = errorMessage(err);
		} finally {
			busy = false;
		}
	}

	function toggleFavorite() {
		run(async () => {
			library.replaceGame(await libraryApi.setFavorite(game.id, !game.favorite));
		});
	}

	function launch() {
		run(() => libraryApi.launch(game.id));
	}

	function saveNotes() {
		if (notes === (game.notes ?? "")) return;
		run(async () => {
			library.replaceGame(await libraryApi.update(game.id, { notes }));
		});
	}

	function remove() {
		run(async () => {
			await libraryApi.remove(game.id);
			onRemoved(game.id);
		});
	}

	/** Starts the slide-out. The panel is removed when the slide ends. */
	function close() {
		closing = true;
	}

	function onAnimationEnd(event: AnimationEvent) {
		if (closing && event.animationName === "detail-out") onBack();
	}
</script>

<div class="detail" class:closing onanimationend={onAnimationEnd}>
	<button class="btn sm back" onclick={close}>
		<span class="material-icons" aria-hidden="true">arrow_back</span> Library
	</button>

	<header class="hero">
		<div class="cover">
			{#if game.cover || game.coverLocal}
				<CoverImage {game} />
			{:else}
				<span class="material-icons" aria-hidden="true">videogame_asset</span>
			{/if}
		</div>
		<div class="summary">
			<h2 bind:this={titleEl} tabindex="-1">{game.title}</h2>
			<p class="meta">{versionLine}</p>

			<div class="actions">
				{#if game.installed && game.exePath}
					<button class="btn sm primary" disabled={busy} onclick={launch}>Play</button>
				{/if}
				<button class="btn sm" aria-pressed={game.favorite} disabled={busy} onclick={toggleFavorite}>
					{game.favorite ? "Unfavorite" : "Favorite"}
				</button>
				<a class="btn sm" href={game.url} target="_blank" rel="noreferrer">Open page</a>
				{#if confirmingRemove}
					<button class="btn sm danger" disabled={busy} onclick={remove}>Confirm remove</button>
					<button class="btn sm" onclick={() => (confirmingRemove = false)}>Cancel</button>
				{:else}
					<button class="btn sm" onclick={() => (confirmingRemove = true)}>Remove</button>
				{/if}
			</div>
			{#if message}
				<p class="error" role="alert">{message}</p>
			{/if}
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

	<GameInstall {game} {library} />

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
	@keyframes detail-in {
		from {
			opacity: 0;
			transform: translateX(24px);
		}
	}

	@keyframes detail-out {
		to {
			opacity: 0;
			transform: translateX(24px);
		}
	}

	.detail {
		display: flex;
		flex-direction: column;
		gap: 24px;
		padding-bottom: 24px;
		animation: detail-in var(--motion-base) var(--ease-snap) both;

		&.closing {
			animation: detail-out var(--motion-base) var(--ease-snap) both;
		}
	}

	.back {
		align-self: flex-start;
	}

	.hero {
		display: flex;
		gap: 24px;
		flex-wrap: wrap;
	}

	.cover {
		width: 180px;
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
		flex: 1 1 260px;

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

	.error {
		margin: 0;
		color: var(--danger);
		font-size: 0.875rem;
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
