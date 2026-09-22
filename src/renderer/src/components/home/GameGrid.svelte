<script lang="ts">
	import { afterUpdate, onDestroy, onMount } from "svelte";
	import ContextMenu from "@components/shared/ContextMenu.svelte";
	import SidebarPanel from "@components/home/SidebarPanel.svelte";
	import FolderGroup from "@components/home/FolderGroup.svelte";
	import FolderSettingsModal from "@components/home/FolderSettingsModal.svelte";
	import SidebarProfilesSection from "@components/home/SidebarProfilesSection.svelte";
	import GameCardGrid from "@components/home/shared/GameCardGrid.svelte";
	import {
		BlacklistGame,
		RemoveGame,
		RemoveScanFolder,
		KillSession,
		PickFileCustom,
		SaveGameConfig,
		OpenFileLocation
	} from "@lib/api";
	import { notifications } from "@stores/notificationStore";
	import { saveHomeScroll, getHomeScroll } from "@stores/homeScrollStore";
	import Marquee from "@components/home/shared/Marquee.svelte";
	import { setIconQueuePaused } from "@lib/iconService";

	export let currentView: "grid" | "list-grid" | "sidebar-grid" = "grid";
	export let games: any[] = [];
	export let filteredGames: any[] = [];
	export let scannedFolderGroups: any[] = [];
	export let gameIcons: Record<string, string> = {};
	export let searchQuery = "";
	export let selectedPrefixFilter = "All Prefixes";
	export let isSelectionMode = false;
	export let selectedPaths = new Set<string>();

	export let isGameRunning: (game: any, sessionsList: any[]) => boolean;
	export let sessions: any[] = [];
	export let handleQuickLaunch: (game: any, showLogs?: boolean) => Promise<void>;
	export let handleConfigure: (game: any) => void;
	export let toggleGameSelection: (game: any, shiftKey: boolean, ctrlKey?: boolean) => void;
	export let onRefresh: () => void = () => {};
	export let loadIcon: (path: string, customIconPath?: string | null) => void = () => {};
	export let onMarqueeSelect: (paths: string[], additive: boolean) => void = () => {};
	export let onSelectAll: () => void = () => {};
	export let onCancelSelection: () => void = () => {};
	export let modalOpen: boolean = false;

	let menuX = 0;
	let menuY = 0;
	let menuVisible = false;
	let activeMenuGame: any = null;

	let activeFolderMenu: string | null = null;

	function toggleFolderMenu(event: MouseEvent, folderPath: string) {
		if (activeFolderMenu === folderPath) {
			activeFolderMenu = null;
		} else {
			activeFolderMenu = folderPath;
		}
	}

	async function handleRemoveFolder(folderPath: string) {
		try {
			await RemoveScanFolder(folderPath);
			notifications.add("Removed watched folder", "success");
			activeFolderMenu = null;
			onRefresh();
		} catch (err) {
			notifications.add(`Failed to remove watched folder: ${err}`, "error");
		}
	}

	function handleRescan() {
		notifications.add("Rescanning watched folders...", "info");
		activeFolderMenu = null;
		onRefresh();
	}

	let showFolderSettings = false;
	let folderSettingsPath = "";

	function handleConfigureFolder(folderPath: string) {
		folderSettingsPath = folderPath;
		showFolderSettings = true;
		activeFolderMenu = null;
	}

	function handleRightClick(event: MouseEvent, game: any) {
		if (isSelectionMode) return;
		menuX = event.clientX;
		menuY = event.clientY;
		activeMenuGame = game;
		menuVisible = true;
	}

	function getGamePath(game: any): string {
		return game?.path || game?.config?.GamePath || game?.config?.LauncherPath || "";
	}

	async function handleOpenFileLocation() {
		if (!activeMenuGame) return;
		const targetPath = getGamePath(activeMenuGame);
		if (!targetPath) {
			notifications.add("No file location found for this game.", "error");
			return;
		}

		try {
			await OpenFileLocation(targetPath);
		} catch (err) {
			notifications.add(`Failed to open file location: ${err}`, "error");
		}
	}

	async function handleKillActiveGame() {
		if (!activeMenuGame) return;
		const activePath = getGamePath(activeMenuGame);
		const session = sessions.find((item) => item.gamePath === activePath);
		if (!session) {
			notifications.add("The running game session was not found.", "error");
			return;
		}

		try {
			await KillSession(session.pid);
			notifications.add(`Stopping ${activeMenuGame.name}...`, "info");
			onRefresh();
		} catch (err) {
			notifications.add(`Failed to stop game: ${err}`, "error");
		}
	}

	async function handleSetCustomIcon() {
		if (!activeMenuGame || activeMenuGame.isAutoScanned) return;
		try {
			const iconPath = await PickFileCustom("Select Game Icon", [
				{
					displayName: "Images",
					pattern: "*.png;*.jpg;*.jpeg;*.webp;*.svg;*.ico"
				}
			]);
			if (!iconPath) return;

			const config = structuredClone(activeMenuGame.config);
			config.CustomIconPath = iconPath;
			await SaveGameConfig(config);
			notifications.add("Custom game icon saved", "success");
			onRefresh();
		} catch (err) {
			notifications.add(`Failed to set custom icon: ${err}`, "error");
		}
	}

	async function handleClearCustomIcon() {
		if (!activeMenuGame || activeMenuGame.isAutoScanned) return;
		try {
			const config = structuredClone(activeMenuGame.config);
			config.CustomIconPath = "";
			await SaveGameConfig(config);
			notifications.add("Using executable icon again", "success");
			onRefresh();
		} catch (err) {
			notifications.add(`Failed to clear custom icon: ${err}`, "error");
		}
	}

	async function handleAction() {
		if (!activeMenuGame) return;
		const path = activeMenuGame.path || activeMenuGame.config.LauncherPath;
		if (activeMenuGame.isAutoScanned) {
			try {
				await BlacklistGame(path);
				notifications.add("Game hidden from library", "success");
				onRefresh();
			} catch (err) {
				notifications.add(`Failed to hide game: ${err}`, "error");
			}
		} else {
			try {
				await RemoveGame(path);
				notifications.add("Game profile removed", "success");
				onRefresh();
			} catch (err) {
				notifications.add(`Failed to remove game: ${err}`, "error");
			}
		}
	}

	let selectedGroupKey = "no-folder";

	let gamesScrollerEl: HTMLElement | null = null;
	let sidebarScrollerEl: HTMLElement | null = null;
	let restoredForView: string | null = null;
	let scanHeaderEl: HTMLElement | null = null;
	let isScanStuck = false;
	let stuckByFolder: Record<string, boolean> = {};
	let scrollSaveRaf: number | null = null;
	let pendingScrollTop = 0;
	let stuckRaf: number | null = null;
	let iconResumeTimer: ReturnType<typeof setTimeout> | null = null;
	let userScrolling = false;

	function markUserScroll() {
		userScrolling = true;
	}

	function getScroller(): HTMLElement | null {
		return currentView === "sidebar-grid" ? sidebarScrollerEl : gamesScrollerEl;
	}

	function updateStuck() {
		const scroller = getScroller();
		if (!scroller) return;
		const scrollerTop = scroller.getBoundingClientRect().top;
		if (scanHeaderEl) {
			const stuck = scanHeaderEl.getBoundingClientRect().top <= scrollerTop + 1;
			if (stuck !== isScanStuck) isScanStuck = stuck;
		}
		const next: Record<string, boolean> = {};
		for (const header of scroller.querySelectorAll<HTMLElement>("[data-sticky-header][data-stuck-key]")) {
			const key = header.dataset.stuckKey;
			if (!key) continue;
			next[key] = header.getBoundingClientRect().top <= scrollerTop + 1;
		}
		let changed = false;
		for (const key of Object.keys(next)) {
			if (next[key] !== stuckByFolder[key]) {
				changed = true;
				break;
			}
		}
		if (!changed) {
			for (const key of Object.keys(stuckByFolder)) {
				if (!(key in next)) {
					changed = true;
					break;
				}
			}
		}
		if (changed) stuckByFolder = next;
	}

	function scheduleStuckCheck() {
		if (stuckRaf !== null) return;
		stuckRaf = requestAnimationFrame(() => {
			stuckRaf = null;
			updateStuck();
		});
	}

	function handleScrollerScroll(e: Event) {
		const scroller = e.currentTarget as HTMLElement;
		if (scroller !== getScroller()) return;
		pendingScrollTop = scroller.scrollTop;
		if (scrollSaveRaf === null) {
			scrollSaveRaf = requestAnimationFrame(() => {
				scrollSaveRaf = null;
				saveHomeScroll(pendingScrollTop);
			});
		}
		if (userScrolling) {
			// Hold icon extraction during user scroll; resume shortly after it stops.
			setIconQueuePaused(true);
			if (iconResumeTimer !== null) clearTimeout(iconResumeTimer);
			iconResumeTimer = setTimeout(() => {
				iconResumeTimer = null;
				setIconQueuePaused(false);
			}, 160);
		}
		scheduleStuckCheck();
	}

	function handleResize() {
		scheduleStuckCheck();
	}

	onMount(() => {
		window.addEventListener("resize", handleResize);
		window.addEventListener("wheel", markUserScroll, { passive: true });
		window.addEventListener("touchstart", markUserScroll, { passive: true });
		window.addEventListener("pointerdown", markUserScroll, { passive: true });
		scheduleStuckCheck();
	});

	onDestroy(() => {
		window.removeEventListener("resize", handleResize);
		window.removeEventListener("wheel", markUserScroll);
		window.removeEventListener("touchstart", markUserScroll);
		window.removeEventListener("pointerdown", markUserScroll);
		if (scrollSaveRaf !== null) {
			cancelAnimationFrame(scrollSaveRaf);
			scrollSaveRaf = null;
		}
		if (stuckRaf !== null) {
			cancelAnimationFrame(stuckRaf);
			stuckRaf = null;
		}
		if (iconResumeTimer !== null) {
			clearTimeout(iconResumeTimer);
			iconResumeTimer = null;
		}
		setIconQueuePaused(false);
	});

	afterUpdate(() => {
		if (restoredForView !== currentView) {
			const el = getScroller();
			if (el) {
				restoredForView = currentView;
				const saved = getHomeScroll();
				if (saved > 0 && el.scrollTop !== saved) {
					el.scrollTop = saved;
				}
			}
		}
		scheduleStuckCheck();
	});

	// Rubber-band marquee: pointerdown arms it, >5px movement starts it.
	type MarqueeRect = { x: number; y: number; w: number; h: number };
	const DRAG_THRESHOLD = 5;

	let marqueeRect: MarqueeRect | null = null;
	let dragPending = false;
	let dragActive = false;
	let dragScroller: HTMLElement | null = null;
	let dragStartX = 0; // container content coords
	let dragStartY = 0;
	let pointerX = 0; // viewport coords
	let pointerY = 0;
	let moveRaf: number | null = null;
	let scrollRaf: number | null = null;
	let suppressClicksUntil = 0;

	function handleScrollerPointerDown(e: PointerEvent) {
		if (!isSelectionMode || e.button !== 0) return;
		const scroller = e.currentTarget as HTMLElement;
		if (scroller !== getScroller()) return;
		// Leave scrollbar drags alone.
		const rect = scroller.getBoundingClientRect();
		if (e.clientX - rect.left >= scroller.clientWidth || e.clientY - rect.top >= scroller.clientHeight) return;

		dragPending = true;
		dragActive = false;
		dragScroller = scroller;
		dragStartX = e.clientX - rect.left + scroller.scrollLeft;
		dragStartY = e.clientY - rect.top + scroller.scrollTop;
		pointerX = e.clientX;
		pointerY = e.clientY;

		window.addEventListener("pointermove", handleDragMove);
		window.addEventListener("pointerup", handleDragUp);
		window.addEventListener("pointercancel", cancelDrag);
		window.addEventListener("blur", cancelDrag);
	}

	function handleDragMove(e: PointerEvent) {
		if (!dragPending || !dragScroller) return;
		pointerX = e.clientX;
		pointerY = e.clientY;

		if (!dragActive) {
			const rect = dragScroller.getBoundingClientRect();
			const startX = rect.left + dragStartX - dragScroller.scrollLeft;
			const startY = rect.top + dragStartY - dragScroller.scrollTop;
			if (Math.hypot(e.clientX - startX, e.clientY - startY) < DRAG_THRESHOLD) return;
			dragActive = true;
			marqueeRect = { x: 0, y: 0, w: 0, h: 0 };
			startAutoscroll();
		}

		if (moveRaf === null) {
			moveRaf = requestAnimationFrame(() => {
				moveRaf = null;
				updateMarquee();
			});
		}
	}

	function updateMarquee() {
		if (!dragActive || !dragScroller) return;
		const scroller = dragScroller;
		const rect = scroller.getBoundingClientRect();
		// Origin is content-anchored so it follows the list while autoscrolling;
		// the live corner sticks to the pointer.
		const originX = rect.left + dragStartX - scroller.scrollLeft;
		const originY = rect.top + dragStartY - scroller.scrollTop;
		let x1 = Math.min(originX, pointerX);
		let x2 = Math.max(originX, pointerX);
		let y1 = Math.min(originY, pointerY);
		let y2 = Math.max(originY, pointerY);
		x1 = Math.max(x1, rect.left);
		x2 = Math.min(x2, rect.right);
		y1 = Math.max(y1, rect.top);
		y2 = Math.min(y2, rect.bottom);
		marqueeRect = { x: x1, y: y1, w: Math.max(0, x2 - x1), h: Math.max(0, y2 - y1) };
	}

	function startAutoscroll() {
		if (scrollRaf !== null) return;
		const EDGE = 70;
		const MAX = 18;
		const tick = () => {
			scrollRaf = null;
			if (!dragActive || !dragScroller) return;
			const rect = dragScroller.getBoundingClientRect();
			let delta = 0;
			if (pointerY < rect.top + EDGE) {
				delta = -MAX * Math.min(1, (rect.top + EDGE - pointerY) / EDGE);
			} else if (pointerY > rect.bottom - EDGE) {
				delta = MAX * Math.min(1, (pointerY - (rect.bottom - EDGE)) / EDGE);
			}
			if (delta !== 0) {
				const before = dragScroller.scrollTop;
				dragScroller.scrollTop += delta;
				if (dragScroller.scrollTop !== before) updateMarquee();
			}
			scrollRaf = requestAnimationFrame(tick);
		};
		scrollRaf = requestAnimationFrame(tick);
	}

	function collectIntersecting(scroller: HTMLElement, r: MarqueeRect): string[] {
		const paths: string[] = [];
		for (const el of scroller.querySelectorAll<HTMLElement>("[data-game-path]")) {
			const b = el.getBoundingClientRect();
			if (b.width === 0 || b.height === 0) continue;
			if (b.left < r.x + r.w && b.right > r.x && b.top < r.y + r.h && b.bottom > r.y) {
				paths.push(el.dataset.gamePath!);
			}
		}
		return paths;
	}

	function teardownDrag() {
		window.removeEventListener("pointermove", handleDragMove);
		window.removeEventListener("pointerup", handleDragUp);
		window.removeEventListener("pointercancel", cancelDrag);
		window.removeEventListener("blur", cancelDrag);
		if (moveRaf !== null) cancelAnimationFrame(moveRaf);
		if (scrollRaf !== null) cancelAnimationFrame(scrollRaf);
		moveRaf = null;
		scrollRaf = null;
		dragPending = false;
		dragActive = false;
		dragScroller = null;
	}

	function cancelDrag() {
		if (!dragPending) return;
		teardownDrag();
		marqueeRect = null;
	}

	function handleDragUp(e: PointerEvent) {
		if (!dragPending) return;
		const wasActive = dragActive;
		const scroller = dragScroller;
		const rect = marqueeRect;
		const additive = e.ctrlKey || e.metaKey;
		teardownDrag();
		marqueeRect = null;
		if (!wasActive || !scroller || !rect) return;
		onMarqueeSelect(collectIntersecting(scroller, rect), additive);
		// The click that follows this pointerup would toggle a card — swallow it.
		suppressClicksUntil = Date.now() + 500;
	}

	function handleWindowClickCapture(e: MouseEvent) {
		if (Date.now() > suppressClicksUntil) return;
		suppressClicksUntil = 0;
		e.preventDefault();
		e.stopImmediatePropagation();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!isSelectionMode) return;
		const target = e.target as HTMLElement | null;
		if (target && (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable)) return;
		if (e.key === "Escape") {
			if (dragPending) {
				cancelDrag();
				e.preventDefault();
			} else if (!modalOpen) {
				onCancelSelection();
			}
			return;
		}
		if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "a") {
			if (modalOpen) return;
			e.preventDefault();
			onSelectAll();
		}
	}

	$: {
		if (selectedGroupKey !== "no-folder" && !scannedFolderGroups.some(g => g.folderPath === selectedGroupKey)) {
			selectedGroupKey = "no-folder";
		}
	}

	$: showCustomProfiles = currentView !== "sidebar-grid" ? filteredGames.length > 0 : selectedGroupKey === "no-folder";
	$: foldersToRender = currentView !== "sidebar-grid" ? scannedFolderGroups : scannedFolderGroups.filter(g => g.folderPath === selectedGroupKey);
</script>

<svelte:window on:keydown={handleKeydown} on:click|capture={handleWindowClickCapture} />

<div
	class="games-container"
	class:grid-view={currentView === "grid"}
	class:list-view={currentView === "list-grid"}
	class:sidebar-layout-view={currentView === "sidebar-grid"}
	class:selecting={isSelectionMode}
	bind:this={gamesScrollerEl}
	on:scroll={handleScrollerScroll}
	on:pointerdown={handleScrollerPointerDown}
>
	{#if currentView === "sidebar-grid"}
		<SidebarPanel
			{filteredGames}
			{scannedFolderGroups}
			bind:selectedGroupKey
		/>
	{/if}

	<div
		class="main-content-panel"
		bind:this={sidebarScrollerEl}
		on:scroll={handleScrollerScroll}
		on:pointerdown={handleScrollerPointerDown}
	>
		{#if showCustomProfiles}
			{#if currentView !== "sidebar-grid" && scannedFolderGroups.length > 0}
				<h2 bind:this={scanHeaderEl} class="scan-section-title" class:is-stuck={isScanStuck}>
					<span class="material-icons">library_books</span>
					Custom Profiles <span class="badge">{filteredGames.length}</span>
				</h2>
			{/if}

			{#if currentView === "sidebar-grid"}
				<SidebarProfilesSection
					{filteredGames}
					{gameIcons}
					{isGameRunning}
					{sessions}
					{isSelectionMode}
					{selectedPaths}
					{handleRightClick}
					{handleQuickLaunch}
					{handleConfigure}
					{toggleGameSelection}
					{loadIcon}
				/>
			{:else if filteredGames.length > 0}
				<GameCardGrid
					games={filteredGames}
					view={currentView}
					{gameIcons}
					{isGameRunning}
					{sessions}
					{isSelectionMode}
					{selectedPaths}
					{handleRightClick}
					{handleQuickLaunch}
					{handleConfigure}
					{toggleGameSelection}
					{loadIcon}
				/>
			{/if}
		{/if}

		{#each foldersToRender as group (group.folderPath)}
			<FolderGroup
				{group}
				{currentView}
				{gameIcons}
				{isGameRunning}
				{sessions}
				{isSelectionMode}
				{selectedPaths}
				isStuck={!!stuckByFolder[group.folderPath]}
				bind:activeFolderMenu
				{toggleFolderMenu}
				{handleRescan}
				{handleRemoveFolder}
				{handleConfigureFolder}
				{handleRightClick}
				{handleQuickLaunch}
				{handleConfigure}
				{toggleGameSelection}
				{loadIcon}
			/>
		{/each}

		{#if filteredGames.length === 0 && scannedFolderGroups.every(g => g.games.length === 0) && (games.length > 0 || scannedFolderGroups.length > 0)}
			<div class="no-results">
				<p>
					No games matching
					{#if searchQuery}"{searchQuery}"{/if}
					{#if selectedPrefixFilter !== "All Prefixes"}
						in prefix <b>{selectedPrefixFilter}</b>
					{/if}
				</p>
				<button
					class="link-btn"
					on:click={() => {
						searchQuery = "";
						selectedPrefixFilter = "All Prefixes";
					}}>Clear all filters</button
				>
			</div>
		{/if}
	</div>
</div>

<Marquee rect={marqueeRect} />

{#if activeMenuGame}
	<ContextMenu
		bind:x={menuX}
		bind:y={menuY}
		bind:visible={menuVisible}
		isAutoScanned={activeMenuGame.isAutoScanned}
		isRunning={isGameRunning(activeMenuGame, sessions)}
		hasCustomIcon={Boolean(activeMenuGame.config?.CustomIconPath)}
		onLaunch={() => handleQuickLaunch(activeMenuGame, false)}
		onLaunchWithLogs={() => handleQuickLaunch(activeMenuGame, true)}
		onKill={handleKillActiveGame}
		onConfigure={() => handleConfigure(activeMenuGame)}
		onOpenLocation={handleOpenFileLocation}
		onSetCustomIcon={handleSetCustomIcon}
		onClearCustomIcon={handleClearCustomIcon}
		onAction={handleAction}
		onClose={() => { menuVisible = false; activeMenuGame = null; }}
	/>
{/if}

<FolderSettingsModal
	show={showFolderSettings}
	folderPath={folderSettingsPath}
	onClose={() => { showFolderSettings = false; }}
	onSave={onRefresh}
/>

<style lang="scss">
	@keyframes games-mount {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	.games-container {
		flex: 1;
		min-height: 0;
		min-width: 0;
		overflow-y: auto;
		overflow-x: hidden;
		padding-right: 8px;
		box-sizing: border-box;
		max-width: 100%;
		animation: games-mount 260ms cubic-bezier(0.215, 0.61, 0.355, 1);

		.main-content-panel {
			min-width: 0;
			max-width: 100%;
			overflow: visible;
		}

		&.selecting {
			cursor: crosshair;
		}
	}

	.scan-section-title {
		position: sticky;
		top: 0;
		z-index: 5;
		display: flex;
		align-items: center;
		gap: 10px;
		box-sizing: border-box;
		height: 64px;
		margin: 12px;
		padding: 12px 16px;
		font-size: 1.1rem;
		font-weight: 800;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		border: 1px solid transparent;
		border-bottom: 1px solid rgba(255, 255, 255, 0.06);
		background: transparent;
		border-radius: 0;
		transition: background var(--transition-fast), border-color var(--transition-fast), border-radius var(--transition-fast);

		&.is-stuck {
			background: var(--bg-base);
			border: 1px solid rgba(255, 255, 255, 0.06);
			border-radius: var(--radius-md);
		}

		.material-icons {
			font-size: 20px;
			color: var(--accent-secondary);
		}

		.badge {
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.08);
			border-radius: var(--radius-sm);
			padding: 2px 8px;
			font-size: 0.75rem;
			font-weight: 700;
			color: var(--text-muted);
		}
	}

	.no-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 80px 20px;
		color: var(--text-muted);
		text-align: center;
		background: var(--bg-surface);
		border-radius: var(--radius-xl);
		border: 2px dashed rgba(255, 255, 255, 0.08);
		margin-top: 20px;

		p {
			font-size: 1.15rem;
			font-weight: 700;
			margin-bottom: 20px;
		}
	}

	.link-btn {
		background: none;
		border: none;
		color: var(--accent-primary);
		font-weight: 800;
		text-decoration: underline;
		cursor: pointer;
		padding: 0;
		font-size: inherit;

		&:hover {
			filter: brightness(1.2);
		}
	}

	.games-container.sidebar-layout-view {
		overflow: hidden;
		padding-right: 0;
		display: flex;
		flex-direction: row;
		gap: 24px;
		flex: 1;
		min-height: 0;
		height: 100%;
		max-height: 100%;

		.main-content-panel {
			flex: 1;
			min-height: 0;
			height: 100%;
			max-height: 100%;
			overflow-y: auto;
			overflow-x: hidden;
			padding-right: 8px;
			display: flex;
			flex-direction: column;
			box-sizing: border-box;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.games-container {
			animation: none;
		}
	}
</style>
