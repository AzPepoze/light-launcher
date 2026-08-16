import { LoadPrefixConfig } from "@lib/api";
import * as core from "@shared";

export class ProtonState {
	protonVersions = $state<core.ProtonTool[]>([]);
	protonOptions = $state<string[]>([]);
	selectedProton = $state("");
	isLoadingProton = $state(true);
	prefixDefaultProton = $state("");

	handleProtonChange(value: string) {
		this.selectedProton = value;
	}

	async updatePrefixDefaultProton(prefixName: string) {
		if (!prefixName || prefixName === "Custom...") return;
		try {
			const cfg = await LoadPrefixConfig(prefixName);
			if (cfg && cfg.ProtonPath) {
				const match = this.protonVersions.find((p) => p.Path === cfg.ProtonPath);
				this.prefixDefaultProton = match ? match.DisplayName : cfg.ProtonPath;
			} else {
				if (this.protonVersions.length > 0) {
					this.prefixDefaultProton = this.protonVersions[0].DisplayName;
				}
			}
		} catch (e) {
			console.error("Failed to load prefix config for default proton:", e);
			if (this.protonVersions.length > 0) {
				this.prefixDefaultProton = this.protonVersions[0].DisplayName;
			}
		}
	}
}
