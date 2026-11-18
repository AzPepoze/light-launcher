<script lang="ts">
	import { onMount } from "svelte";
	import Button from "@components/shared/Button.svelte";
	import CoverImage from "@components/shared/CoverImage.svelte";
	import ExpandableText from "@components/shared/ExpandableText.svelte";
	import Icon from "@components/shared/Icon.svelte";
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
	let confirmingRemove = $state(false);
	let runningAction = $state("");
	let refreshingFiles = $state(false);
	let busy = $state(false);
	let titleEl = $state<HTMLElement>();
	/** The real game icon, shown for local games that have no cover art. */
	let gameIcon = $state("");

	const developerLine = $derived([game.developer, game.engine].filter(Boolean).join(" · "));
	const versionText = $derived(game.version ? `Version ${game.version}` : "Version unknown");
	const hasArt = $derived(Boolean(game.cover || game.coverLocal));
	/** Tags show a short row until the person asks for the rest. */
	const TAG_PREVIEW_COUNT = 8;
	let showAllTags = $state(false);
	const hiddenTagCount = $derived(Math.max(0, game.tags.length - TAG_PREVIEW_COUNT));
	const visibleTags = $derived(showAllTags ? game.tags : game.tags.slice(0, TAG_PREVIEW_COUNT));

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
		<div class="art" class:bare={!hasArt}>
			{#if hasArt}
				<CoverImage {game} />
			{:else if gameIcon}
				<img class="art-icon" src={gameIcon} alt="" />
			{:else}
				<Icon name="videogame_asset" size={48} />
			{/if}
		</div>
		<div class="title-block">
			<h2 bind:this={titleEl} tabindex="-1">{game.title}</h2>
			{#if developerLine}<p class="meta">{developerLine}</p>{/if}
		</div>
	</header>

	<div class="actions">
		{#if game.installed && game.exePath}
			<Button variant="primary" disabled={busy} onclick={launch}>Play</Button>
		{/if}
		<Button aria-pressed={game.favorite} disabled={busy} onclick={toggleFavorite}>
			<Icon name={game.favorite ? "star" : "star_border"} size={18} />
			{game.favorite ? "Favorited" : "Favorite"}
		</Button>
		<a class="btn sm" href={game.url} target="_blank" rel="noreferrer">Open page</a>
		{#each actions as action (action.id)}
			<Button
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
				icon="refresh"
				loading={refreshingFiles}
				disabled={runningAction !== "" || busy}
				onclick={refreshFiles}
			>
				Refresh files
			</Button>
		{/if}
	</div>

	<dl class="facts">
		<div>
			<dt>Source</dt>
			<dd>{sourceName}</dd>
		</div>
		<div>
			<dt>Version</dt>
			<dd>{game.version ?? "Unknown"}</dd>
		</div>
		<div>
			<dt>Played</dt>
			<dd>{playtime(game.playtimeS)}</dd>
		</div>
		<div>
			<dt>Last played</dt>
			<dd>{ago(game.lastPlayedAt)}</dd>
		</div>
	</dl>

	{#if game.tags.length}
		<div class="tags-block">
			<ul class="tags" aria-label="Tags">
				{#each visibleTags as tag (tag)}
					<li>{tag}</li>
				{/each}
			</ul>
			{#if hiddenTagCount > 0 || showAllTags}
				<Button
					variant="text"
					size="sm"
					icon={showAllTags ? "expand_less" : "expand_more"}
					aria-expanded={showAllTags}
					onclick={() => (showAllTags = !showAllTags)}
				>
					{showAllTags ? "Show fewer tags" : `Show all ${game.tags.length} tags`}
				</Button>
			{/if}
		</div>
	{/if}

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
			<ExpandableText text={game.description} />
		</section>
	{/if}

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

	<footer class="danger">
		{#if confirmingRemove}
			<p class="muted">Remove this game from the library? Files on disk stay.</p>
			<div class="danger-actions">
				<Button variant="danger" disabled={busy} onclick={remove}>Confirm remove</Button>
				<Button onclick={() => (confirmingRemove = false)}>Cancel</Button>
			</div>
		{:else}
			<Button variant="text" icon="delete" onclick={() => (confirmingRemove = true)}>
				Remove from library
			</Button>
		{/if}
	</footer>
</div>

<style lang="scss">
	/* The hero bleeds to the drawer edges; everything below it keeps the drawer's side padding. */
	.detail {
		display: flex;
		flex-direction: column;
		gap: 24px;
		min-width: 0;
		padding: 0 20px;
	}

	.hero {
		position: relative;
		margin: 0 -20px;
		min-width: 0;
	}

	/* The art runs to the top edge and fades into the drawer surface below. */
	.art {
		position: relative;
		height: 240px;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-elevated);
		color: var(--text-muted);

		&::after {
			content: "";
			position: absolute;
			inset: 0;
			background: linear-gradient(to bottom, transparent 40%, var(--bg-surface) 100%);
			pointer-events: none;
		}

		&.bare {
			height: 160px;
		}
	}

	.art-icon {
		width: 40%;
		height: 40%;
		object-fit: contain;
	}

	.title-block {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		z-index: 1;
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
		padding: 0 20px;
	}

	.hero:has(.art.bare) .title-block {
		position: static;
		padding-top: 12px;
	}

	h2 {
		margin: 0;
		font-size: 2rem;
		font-weight: 700;
		line-height: 1.15;
		overflow-wrap: anywhere;
		text-shadow: 0 2px 12px rgb(0 0 0 / 0.5);

		&:focus {
			outline: none;
		}
	}

	.meta {
		margin: 0;
		font-size: 0.95rem;
		color: var(--text-muted);
	}

	.muted {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.875rem;
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;

		a {
			text-decoration: none;
		}
	}

	.facts {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 14px 20px;
		margin: 0;

		div {
			min-width: 0;
		}

		dt {
			font-size: 0.8rem;
			color: var(--text-muted);
		}

		dd {
			margin: 2px 0 0;
			font-size: 0.95rem;
			font-weight: 600;
			overflow-wrap: anywhere;
		}
	}

	.tags-block {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 6px;
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

	.shots {
		display: flex;
		gap: 10px;
		overflow-x: auto;
		scroll-snap-type: x proximity;
		list-style: none;
		margin: 0;
		padding: 0 0 4px;

		li {
			flex: 0 0 240px;
			scroll-snap-align: start;
		}
	}

	section {
		display: flex;
		flex-direction: column;
		gap: 10px;

		h3 {
			margin: 0;
			font-size: 1.1rem;
			font-weight: 600;
		}
	}

	/* Notes keep a fixed box: long text scrolls inside it instead of growing the panel. */
	textarea {
		resize: vertical;
		max-height: 240px;
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

	.danger {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 10px;
		padding-top: 16px;
		border-top: 1px solid var(--glass-border);
	}

	.danger-actions {
		display: flex;
		gap: 8px;
	}
</style>
