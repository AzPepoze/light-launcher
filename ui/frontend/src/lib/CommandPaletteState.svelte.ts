import { GetAllGames, RunGame } from "@bindings/light-launcher/internal/app/app";
import { navigationCommand } from "@stores/navigationStore";
import { notifications } from "@stores/notificationStore";
import { loadExeIcon } from "@lib/iconService";
import protonIcon from "@icons/protron_forked.png";

export class CommandPaletteState {
	show = $state(false);
	searchQuery = $state("");
	games = $state<any[]>([]);
	gameIcons = $state<Record<string, string>>({});
	filteredItems = $state<any[]>([]);
	selectedIndex = $state(0);

	inputElement = $state<HTMLInputElement | null>(null);
	resultsContainer = $state<HTMLDivElement | null>(null);

	readonly PAGES = [
		{ name: "Go to Home", icon: "home", action: () => this.navigateTo("home") },
		{ name: "Go to Launch Configuration", icon: "play_arrow", action: () => this.navigateTo("run") },
		{ name: "Go to Utilities", icon: "handyman", action: () => this.navigateTo("utils") },
		{ name: "Go to Appearance & Settings", icon: "settings", action: () => this.navigateTo("settings") },
		{ name: "Go to Proton Versions", icon: protonIcon, isCustomIcon: true, action: () => this.navigateTo("versions") },
		{ name: "Go to WINE Prefixes", icon: "folder", action: () => this.navigateTo("prefix") }
	];

	async loadGames() {
		try {
			const fetched = await GetAllGames();
			this.games = fetched || [];
			
			for (const game of this.games) {
				const path = game.path || game.config?.LauncherPath;
				if (path && !this.gameIcons[path]) {
					loadExeIcon(path).then((icon) => {
						if (icon) {
							this.gameIcons[path] = icon;
						}
					});
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
				...this.PAGES.map(p => ({ ...p, type: "page" })),
				...this.games.slice(0, 5).map(g => ({ name: `Launch ${g.name}`, icon: "sports_esports", type: "game", game: g }))
			];
			this.selectedIndex = 0;
			return;
		}

		const matchedPages = this.PAGES.filter(p => p.name.toLowerCase().includes(query)).map(p => ({ ...p, type: "page" }));
		const matchedGames = this.games.filter(g => g.name.toLowerCase().includes(query)).map(g => ({ name: `Launch ${g.name}`, icon: "sports_esports", type: "game", game: g }));
		
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
				notifications.add(`Launching ${item.game.name}...`, "info");
				this.close();
				await RunGame(item.game.config, false);
			} catch (err) {
				notifications.add(`Launch failed: ${err}`, "error");
			}
		}
	}

	close() {
		this.show = false;
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
			this.selectedIndex = (this.selectedIndex - 1 + this.filteredItems.length) % Math.max(1, this.filteredItems.length);
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
		this.show = newShow;
		if (newShow) {
			this.searchQuery = "";
			this.selectedIndex = 0;
			this.loadGames();
			setTimeout(() => {
				if (this.inputElement) this.inputElement.focus();
			}, 50);
		}
	}
}

export const commandPaletteState = new CommandPaletteState();
