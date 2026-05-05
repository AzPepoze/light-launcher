import * as core from "@bindings/light-launcher/internal/types/models";
import { GetConfig, LoadPrefixConfig, DetectLosslessDll } from "@bindings/light-launcher/internal/app/app";

export async function loadConfigForGame(
	path: string,
	options: core.LaunchOptions,
	prefixPath: string,
	baseDir: string,
	selectedPrefixName: string,
	protonVersions: core.ProtonTool[],
	updateOptions: (newOpts: core.LaunchOptions, pPath: string, pName: string, proton: string) => void
) {
	try {
		const config = await GetConfig(path);
		if (config) {
			const newPrefixPath = config.PrefixPath;
			let newPrefixName = selectedPrefixName;
			if (newPrefixPath.startsWith(baseDir)) {
				newPrefixName = newPrefixPath.replace(baseDir + "/", "");
			} else {
				newPrefixName = newPrefixPath.split('/').filter(Boolean).pop() || "Custom";
			}
			const updatedProton = applyConfigToOptions(config, options, protonVersions);
			updateOptions(options, newPrefixPath, newPrefixName, updatedProton);
		} else {
			await loadConfigForPrefix(selectedPrefixName, options, prefixPath, baseDir, protonVersions, updateOptions);
		}

		// Auto-detect Lossless.dll if not already set
		if (!options.Extras.Lsfg.DllPath) {
			try {
				const dll = await DetectLosslessDll();
				if (dll) {
					options.Extras.Lsfg.DllPath = dll;
					updateOptions(options, prefixPath, selectedPrefixName, ""); // triggers reactivity
				}
			} catch (err) {
				console.error("Failed to detect Lossless.dll:", err);
			}
		}
	} catch (err) {}
}

export async function loadConfigForPrefix(
	name: string,
	options: core.LaunchOptions,
	prefixPath: string,
	baseDir: string,
	protonVersions: core.ProtonTool[],
	updateOptions: (newOpts: core.LaunchOptions, pPath: string, pName: string, proton: string) => void
) {
	if (name === "Custom..." || !prefixPath.startsWith(baseDir)) return;
	try {
		const config = await LoadPrefixConfig(name);
		if (config) {
			const savedGamePath = options.GamePath;
			const savedRunnerPath = options.RunnerPath;
			const savedUseGamePath = options.UseGamePath;
			const savedPrefixPath = options.PrefixPath;

			const updatedProton = applyConfigToOptions(config, options, protonVersions);

			// Restore Target
			options.GamePath = savedGamePath;
			options.RunnerPath = savedRunnerPath;
			options.UseGamePath = savedUseGamePath;

			let newPrefixPath = prefixPath;
			if (savedPrefixPath) {
				options.PrefixPath = savedPrefixPath;
				newPrefixPath = savedPrefixPath;
			}
			updateOptions(options, newPrefixPath, name, updatedProton);
		}
	} catch (err) {}
}

export function applyConfigToOptions(
	config: core.LaunchOptions,
	options: core.LaunchOptions,
	protonVersions: core.ProtonTool[]
): string {
	let selectedProton = "";
	
	// Try matching by absolute path
	const matchByPath = protonVersions.find((p) => p.Path === config.ProtonPath);
	if (matchByPath) {
		selectedProton = matchByPath.DisplayName;
	} else if (config.ProtonPath) {
		// Fallback to full path if not found in scanned tools
		selectedProton = config.ProtonPath;
	}

	options.Name = config.Name || options.Name;
	options.CustomArgs = config.CustomArgs || "";
	
	// Copy Extras
	if (config.Extras) {
		const currentDll = options.Extras?.Lsfg?.DllPath;
		
		// Use structuredClone to ensure we have a deep copy of the config's extras
		const newExtras = structuredClone(config.Extras);
		
		// Preserve current DLL path if not set in config
		if (!newExtras.Lsfg.DllPath && currentDll) {
			newExtras.Lsfg.DllPath = currentDll;
		}
		
		options.Extras = newExtras;
	}
	
	if (!options.RunnerPath && config.RunnerPath) {
		options.RunnerPath = config.RunnerPath;
	}
	options.UseGamePath = config.UseGamePath === true; 

	if (!options.UseGamePath && options.RunnerPath) {
		options.GamePath = options.RunnerPath;
	} else if (options.UseGamePath && config.GamePath) {
		options.GamePath = config.GamePath;
	}

	return selectedProton;
}
