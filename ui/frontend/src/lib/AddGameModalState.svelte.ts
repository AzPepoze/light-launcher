import {
	PickFile,
	PickFolder,
	ListPrefixes,
	GetPrefixBaseDir,
} from "@bindings/light-launcher/internal/app/app";
import { notifications } from "@stores/notificationStore";
import { loadExeIcon } from "@lib/iconService";
import { getDefaultPrefixName } from "./prefixService";
import * as service from "@lib/gameService";

export class AddGameModalState {
	addMode = $state<"select" | "folder-config" | "folder-review">("select");
	searchDepth = $state("2");
	excludeNames = $state("UnityCrashHandler64, uninstall, redist");
	selectedFolder = $state("");
	foundExecutables = $state<service.ScannedExecutable[]>([]);
	discardedExecutables = $state(new Set<string>());
	isSearching = $state(false);

	prefixes = $state<string[]>([]);
	selectedPrefix = $state("default");
	prefixBaseDir = $state("");

	async loadPrefixes() {
		try {
			const list = await ListPrefixes();
			this.prefixes = list || [];
			this.prefixBaseDir = await GetPrefixBaseDir();
			this.selectedPrefix = getDefaultPrefixName(this.prefixes);
		} catch (err) {
			console.error("Failed to load prefixes:", err);
		}
	}

	async initialize() {
		await this.loadPrefixes();
	}

	resetState() {
		this.addMode = "select";
		this.selectedFolder = "";
		this.foundExecutables = [];
		this.discardedExecutables = new Set<string>();
		this.searchDepth = "2";
		this.excludeNames = "UnityCrashHandler64, uninstall, redist";
		this.isSearching = false;
		this.loadPrefixes();
	}

	async handleAddFile(onRefresh: () => void, onClose: () => void) {
		try {
			const path = await PickFile();
			if (path) {
				const name = path.split("/").pop()?.replace(".exe", "") || "Game";
				await service.registerGame(path, `${this.prefixBaseDir}/${this.selectedPrefix}`);
				notifications.add(`Added ${name}`, "success");
				onRefresh();
				onClose();
			}
		} catch (err) {
			notifications.add(`Failed to add game: ${err}`, "error");
		}
	}

	async handleAddFolder() {
		try {
			const folder = await PickFolder();
			if (folder) {
				this.selectedFolder = folder;
				this.addMode = "folder-config";
			}
		} catch (err) {
			notifications.add(`Failed to select folder: ${err}`, "error");
		}
	}

	async startFolderScan() {
		if (!this.selectedFolder) return;
		this.isSearching = true;
		try {
			const depth = parseInt(this.searchDepth) || 2;
			const excludes = this.excludeNames.split(",").map((e) => e.trim()).filter(Boolean);

			const results = await service.scanFolderForExecutables(this.selectedFolder, depth, excludes);
			
			if (results.length > 0) {
				this.foundExecutables = results;
				this.addMode = "folder-review";
				this.discardedExecutables = new Set<string>();
				
				this.foundExecutables.forEach((item, index) => {
					loadExeIcon(item.path).then((icon) => {
						if (icon) {
							this.foundExecutables[index].icon = icon;
							this.foundExecutables = [...this.foundExecutables];
						}
					});
				});
			} else {
				notifications.add("No executables found in folder", "info");
				this.addMode = "select";
			}
		} finally {
			this.isSearching = false;
		}
	}

	toggleDiscard(path: string) {
		if (this.discardedExecutables.has(path)) {
			this.discardedExecutables.delete(path);
		} else {
			this.discardedExecutables.add(path);
		}
		this.discardedExecutables = new Set(this.discardedExecutables);
	}

	async confirmAddFolder(onRefresh: () => void, onClose: () => void) {
		const targetPrefixPath = `${this.prefixBaseDir}/${this.selectedPrefix}`;
		const addedCount = await service.batchRegisterGames(this.foundExecutables, this.discardedExecutables, targetPrefixPath);
		
		if (addedCount > 0) {
			onRefresh();
			onClose();
		} else if (this.foundExecutables.length > 0) {
			onClose();
		}
	}
}
