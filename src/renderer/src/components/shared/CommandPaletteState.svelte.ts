import { GetAllGames, GetImageBase64 } from "@lib/api";
import { navigationCommand } from "@stores/navigationStore";
import { notifications } from "@stores/notificationStore";
import { loadExeIcon } from "@lib/iconService";
import { launchGame } from "@lib/gameLaunchService";
import protonIcon from "@icons/protron_forked.png";

export class CommandPaletteState {
	show = $state(false);
	searchQuery = $state("");
	games = $state<any[]>([]);
	gameIcons = $state<Record<string, string>>({});
	filteredItems = $state<any[]>([]);
	selectedIndex = $state(0);
	onCloseCallback: () => void = () => {};

	inputElement = $state<HTMLInputElement | null>(null);
	resultsContainer = $state<HTMLDivElement | null>(null);

	readonly PAGES = [
		{ name: "Go to Home", icon: "home", action: () => this.navigateTo("home") },
		{
			name: "Go to Game Manager",
			icon: "monitor_heart",
			action: () => this.navigateTo("manager")
		},
		{
			name: "Go to Launch Configuration",
			icon: "play_arrow",
			action: () => this.navigateTo("run")
		},
		{ name: "Go to Utilities", icon: "handyman", action: () => this.navigateTo("utils") },
		{
			name: "Go to Appearance & Settings",
			icon: "settings",
			action: () => this.navigateTo("settings")
		},
		{
			name: "Go to Proton Versions",
			icon: protonIcon,
			isCustomIcon: true,
			action: () => this.navigateTo("versions")
		},
		{ name: "Go to WINE Prefixes", icon: "folder", action: () => this.navigateTo("prefix") }
	];

	async loadGames() {
		try {
			const fetched = await GetAllGames();
			this.games = fetched || [];

			for (const game of this.games) {
				const gamePath = game.path || game.config?.LauncherPath;
				if (!gamePath) continue;
				const customIconPath = game.config?.CustomIconPath;
				try {
					let icon = "";
					if (customIconPath) icon = (await GetImageBase64(customIconPath)) || "";
					if (!icon) icon = (await loadExeIcon(gamePath)) || "";
					if (icon) this.gameIcons[gamePath] = icon;
				} catch {
					const fallback = await loadExeIcon(gamePath).catch(() => "");
					if (fallback) this.gameIcons[gamePath] = fallback;
				}
			}

			this.filterItems();
		} catch (e) {
			console.error("Failed to load games for command palette", e);
		}
	}

	filterItems() {
		const query = this.searchQuery.trim().toLowerCase();

		if (!query) {
			this.filteredItems = [
				...this.PAGES.map((p) => ({ ...p, type: "page" })),
				...this.games
					.slice(0, 5)
					.map((g) => ({ name: `Launch ${g.name}`, icon: "sports_esports", type: "game", game: g }))
			];
			this.selectedIndex = 0;
			return;
		}

		const matchedPages = this.PAGES.filter((p) => p.name.toLowerCase().includes(query)).map(
			(p) => ({ ...p, type: "page" })
		);
		const matchedGames = this.games
			.filter((g) => g.name.toLowerCase().includes(query))
			.map((g) => ({ name: `Launch ${g.name}`, icon: "sports_esports", type: "game", game: g }));

		this.filteredItems = [...matchedPages, ...matchedGames];
		this.selectedIndex = 0;
	}

	navigateTo(page: string) {
		navigationCommand.set({ page });
		this.close();
	}

	async executeItem(item: any) {
		if (item.type === "page") {
			item.action();
		} else if (item.type === "game") {
			try {
				this.close();
				await launchGame(item.game.config, { showLogs: false });
			} catch {
				// Shared launch service already reports the error.
			}
		}
	}

	close() {
		this.show = false;
		this.searchQuery = "";
		this.selectedIndex = 0;
		this.onCloseCallback();
	}

	handleKeyDown(e: KeyboardEvent) {
		if (e.key === "Escape") {
			this.close();
			e.preventDefault();
		} else if (e.key === "ArrowDown") {
			this.selectedIndex = (this.selectedIndex + 1) % Math.max(1, this.filteredItems.length);
			e.preventDefault();
			this.scrollToSelected();
		} else if (e.key === "ArrowUp") {
			this.selectedIndex =
				(this.selectedIndex - 1 + this.filteredItems.length) %
				Math.max(1, this.filteredItems.length);
			e.preventDefault();
			this.scrollToSelected();
		} else if (e.key === "Enter") {
			if (this.filteredItems[this.selectedIndex]) {
				this.executeItem(this.filteredItems[this.selectedIndex]);
			}
			e.preventDefault();
		}
	}

	scrollToSelected() {
		if (!this.resultsContainer) return;
		const selectedEl = this.resultsContainer.children[this.selectedIndex] as HTMLElement;
		if (!selectedEl) return;

		const containerHeight = this.resultsContainer.clientHeight;
		const elTop = selectedEl.offsetTop;
		const elHeight = selectedEl.clientHeight;

		if (elTop < this.resultsContainer.scrollTop) {
			this.resultsContainer.scrollTop = elTop;
		} else if (elTop + elHeight > this.resultsContainer.scrollTop + containerHeight) {
			this.resultsContainer.scrollTop = elTop + elHeight - containerHeight;
		}
	}

	onShowChange(newShow: boolean) {
		if (newShow === this.show) return;
		this.show = newShow;
		if (newShow) {
			this.selectedIndex = 0;
			this.loadGames();
			setTimeout(() => this.focusInput(), 0);
		}
	}

	focusInput() {
		this.inputElement?.focus();
	}
}

export const commandPaletteState = new CommandPaletteState();
