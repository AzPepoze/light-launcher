<script lang="ts">
	import { onMount } from "svelte";
	import AppMenu from "@components/shared/AppMenu.svelte";
	import Button from "@components/shared/Button.svelte";
	import CoverImage from "@components/shared/CoverImage.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import MenuTrigger from "@components/shared/MenuTrigger.svelte";
	import ScreenshotImage from "@components/shared/ScreenshotImage.svelte";
	import { notifications } from "@stores/notificationStore";
	import { libraryApi } from "../../lib/features/libraryApi";
	import type {
		Game,
		GameChange,
		GameFile,
		LibraryGetResponse,
		SourceAction,
	} from "../../lib/features/types";
	import { ago, playtime } from "../../lib/features/format";
	import { loadGameIcon } from "../../lib/iconService";
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
	let notesSaved = $state(false);
	let menuOpen = $state(false);
	/** The overflow menu's remove needs a second click, so one slip cannot delete a game. */
	let confirmingRemove = $state(false);
	let runningAction = $state("");
	let refreshingFiles = $state(false);
	let busy = $state(false);
	let titleEl = $state<HTMLElement>();
	/** The real game icon, shown for local games that have no cover art. */
	let gameIcon = $state("");

	const versionLine = $derived(
		[sourceName, game.version ? `Version ${game.version}` : null].filter(Boolean).join(" · "),
	);
	const facts = $derived([game.developer, game.engine].filter(Boolean).join(" · "));

	// A local game with no cover shows its custom or executable icon instead of a generic glyph.
	$effect(() => {
		gameIcon = "";
		if (game.cover || game.coverLocal) return;
		const path = game.exePath || game.localConfig?.GamePath || game.localConfig?.LauncherPath || "";
		if (!path) return;
		let current = true;
		void loadGameIcon(
			path,
			game.localConfig?.CustomIconPath ?? null,
			"low",
			false
		).then((result) => {
			if (current) gameIcon = result.icon;
		});
		return () => {
			current = false;
		};
	});

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

	/** Shows the page a detail call returned, and keeps the list row in step with it. */
	function applyDetail(detail: LibraryGetResponse) {
		files = detail.files;
		changes = detail.changes;
		onChanged(detail.game);
	}

	/** Runs one game-scope action. A refresh reloads the game and its files. */
	async function runGameAction(action: SourceAction) {
		runningAction = action.id;
		try {
			const result = await libraryApi.runAction(game.source, action.id, game.id);
			if (result.message) notifications.success(result.message);
			if (result.refresh) applyDetail(await libraryApi.get(game.id));
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		} finally {
			runningAction = "";
		}
	}

	/** Re-reads the source metadata and plugin file ids, then shows the returned page. */
	async function refreshFiles() {
		if (refreshingFiles) return;
		refreshingFiles = true;
		try {
			applyDetail(await libraryApi.refreshFiles(game.id));
			notifications.success("Files refreshed.");
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		} finally {
			refreshingFiles = false;
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
			notesSaved = true;
		});
	}

	function editNotes() {
		notesSaved = false;
	}

	function remove() {
		run(async () => {
			await libraryApi.remove(game.id);
			onRemoved(game.id);
		});
	}

	function toggleMenu() {
		if (menuOpen) closeMenu();
		else menuOpen = true;
	}

	/** Closing the menu also drops a pending remove, so the next open starts clean. */
	function closeMenu() {
		menuOpen = false;
		confirmingRemove = false;
	}

	function askRemove() {
		confirmingRemove = true;
	}

	function cancelRemove() {
		confirmingRemove = false;
	}
</script>

<div class="detail">
	<header class="hero">
		<div class="cover">
			{#if game.cover || game.coverLocal}
				<CoverImage {game} />
			{:else if gameIcon}
				<img class="cover-icon" src={gameIcon} alt="" />
			{:else}
				<Icon name="videogame_asset" size={40} />
			{/if}
		</div>
		<div class="summary">
			<h2 bind:this={titleEl} tabindex="-1">{game.title}</h2>
			<p class="meta">{versionLine}</p>
		</div>
	</header>

	{#if game.installed && game.exePath}
		<div class="play-row">
			<Button class="play" variant="primary" icon="play_arrow" disabled={busy} onclick={launch}>Play</Button>
		</div>
	{/if}

	<div class="actions">
		<IconButton
			icon="favorite"
			label={game.favorite ? "Unfavorite" : "Favorite"}
			filled={game.favorite}
			disabled={busy}
			onclick={toggleFavorite}
		/>
		<a class="icon-link" href={game.url} target="_blank" rel="noreferrer" title="Open page" aria-label="Open page">
			<Icon name="open_in_new" size={20} />
		</a>
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
		{#if game.sourceKind === "plugin"}
			<Button
				size="sm"
				icon="refresh"
				loading={refreshingFiles}
				disabled={runningAction !== "" || busy}
				onclick={refreshFiles}
			>
				Refresh files
			</Button>
		{/if}

		<div class="overflow">
			<MenuTrigger active={menuOpen} label="More actions" on:click={toggleMenu} />
			<AppMenu visible={menuOpen} mode="anchor" onClose={closeMenu}>
				{#if confirmingRemove}
					<button class="menu-item danger" disabled={busy} onclick={remove}>
						<Icon name="delete_forever" size={18} />
						<span>Confirm remove</span>
					</button>
					<button class="menu-item" onclick={cancelRemove}>
						<Icon name="close" size={18} />
						<span>Cancel</span>
					</button>
				{:else}
					<button class="menu-item danger" onclick={askRemove}>
						<Icon name="delete_outline" size={18} />
						<span>Remove game</span>
					</button>
				{/if}
			</AppMenu>
		</div>
	</div>

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
		<section class="card">
			<h3>About</h3>
			<p class="description">{game.description}</p>
		</section>
	{/if}

	<section class="card">
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

	<div class="card">
		<GameInstall {game} {onChanged} />
	</div>

	<div class="card">
		<GameFiles gameId={game.id} {files} />
	</div>

	<section class="card">
		<div class="card-head">
			<h3 id="notes-title">Notes</h3>
			{#if notesSaved}<span class="saved" role="status">Saved</span>{/if}
		</div>
		<textarea
			aria-labelledby="notes-title"
			class="input"
			rows="4"
			bind:value={notes}
			oninput={editNotes}
			onblur={saveNotes}
			placeholder="Write a note about this game"
		></textarea>
	</section>

	{#if changes.length}
		<section class="card">
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
	/* The game panel is narrow: the cover sits beside the title, and the actions wrap below. */
	.detail {
		display: flex;
		flex-direction: column;
		gap: 16px;
		min-width: 0;
	}

	.hero {
		display: flex;
		gap: 16px;
		align-items: flex-end;
	}

	.cover {
		width: 112px;
		aspect-ratio: 3 / 4;
		flex-shrink: 0;
		border-radius: var(--radius-lg);
		overflow: hidden;
		background: var(--bg-elevated);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-muted);
	}

	.cover-icon {
		width: 74%;
		height: 74%;
		object-fit: contain;
	}

	.summary {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
		flex: 1 1 180px;

		h2 {
			margin: 0;
			font-size: 1.5rem;
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
		font-size: 0.9375rem;
		color: var(--text-muted);
	}

	/* One large Play button across the panel. */
	.play-row :global(.play) {
		width: 100%;
		height: 48px;
		font-size: 1rem;
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
	}

	.icon-link {
		display: inline-grid;
		place-items: center;
		width: 36px;
		height: 36px;
		border-radius: var(--radius-pill);
		color: var(--text-muted);
		transition:
			background var(--transition-fast),
			color var(--transition-fast);

		&:hover {
			background: rgb(255 255 255 / 0.08);
			color: var(--text-main);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}

	/* The overflow panel opens under this wrapper, aligned to its right edge. */
	.overflow {
		position: relative;
		margin-left: auto;
	}

	.muted {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.875rem;
	}

	/* A rounded group: each block of the panel sits on its own surface. */
	.card {
		display: flex;
		flex-direction: column;
		gap: 10px;
		min-width: 0;
		padding: 14px 16px;
		border-radius: var(--radius-lg);
		background: var(--bg-elevated);

		h3 {
			margin: 0;
			font-size: 1.0625rem;
			font-weight: 600;
		}
	}

	.card-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.saved {
		color: var(--text-muted);
		font-size: 0.8125rem;
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
			background: var(--bg-input);
			color: var(--text-muted);
		}
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

	.description {
		max-width: 68ch;
		white-space: pre-line;
		font-size: 0.9375rem;
		line-height: 1.5;
		margin: 0;
	}

	textarea {
		resize: vertical;
		font-size: 0.9375rem;
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
