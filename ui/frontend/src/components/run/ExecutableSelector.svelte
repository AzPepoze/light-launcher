<script lang="ts">
	import { loadExeIcon } from "@lib/iconService";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import BrowseInput from "@components/shared/BrowseInput.svelte";

	export let launcherPath = "";
	export let gamePath = "";
	export let useGamePath = false;
	export let launcherIcon = "";
	export let gameIcon = "";
	export let onBrowseLauncher: () => Promise<void>;
	export let onBrowseGame: () => Promise<void>;

	let launcherIconFailed = false;
	let gameIconFailed = false;
	let prevLauncherPath = "";
	let prevGamePath = "";

	// Internal state for inputs - keep in sync with props
	let internalLauncherPath = launcherPath;
	let internalGamePath = gamePath;

	// When props change, update internal state
	$: if (launcherPath !== internalLauncherPath) {
		console.log(
			"ExecutableSelector: launcherPath prop changed to:",
			launcherPath,
		);
		internalLauncherPath = launcherPath;
	}

	$: if (gamePath !== internalGamePath) {
		console.log(
			"ExecutableSelector: gamePath prop changed to:",
			gamePath,
		);
		internalGamePath = gamePath;
	}

	// Load launcher icon when launcher path changes
	$: if (internalLauncherPath && internalLauncherPath !== prevLauncherPath) {
		prevLauncherPath = internalLauncherPath;
		launcherIconFailed = false;
		(async () => {
			const icon = await loadExeIcon(internalLauncherPath);
			if (icon) {
				launcherIcon = icon;
				launcherIconFailed = false;
			} else {
				launcherIconFailed = true;
			}
		})();
	}

	// Reload game icon when game path changes
	$: if (internalGamePath && internalGamePath !== prevGamePath) {
		prevGamePath = internalGamePath;
		gameIconFailed = false;
		if (useGamePath) {
			(async () => {
				const icon = await loadExeIcon(internalGamePath);
				if (icon) {
					gameIcon = icon;
					gameIconFailed = false;
				} else {
					gameIconFailed = true;
				}
			})();
		}
	}

	async function handleBrowseLauncherClick() {
		await onBrowseLauncher();
		// Give parent time to update binding, then force update
		await new Promise((r) => setTimeout(r, 0));
	}

	async function handleBrowseGameClick() {
		await onBrowseGame();
		// Give parent time to update binding, then force update
		await new Promise((r) => setTimeout(r, 0));
	}
</script>

<div class="exe-selector">
	<!-- Launcher Executable Section -->
	<div class="launcher-exe-section">
		<label for="launcherExe">
			Launcher Executable <span class="required-tag">Required</span>
		</label>

		<div class="launcher-exe-wrapper">
			<div class="exe-icon-display launcher-icon">
				{#if launcherIcon && !launcherIconFailed}
					<img
						src={launcherIcon}
						alt="Launcher Icon"
						class="exe-icon"
						on:load={() => {
							launcherIconFailed = false;
						}}
						on:error={() => {
							launcherIconFailed = true;
						}}
					/>
				{:else}
					<div class="exe-icon-placeholder">
						<span
							class="material-icons"
							style="font-size: 32px;">laptop_windows</span
						>
					</div>
				{/if}
			</div>

			<div class="launcher-input-group" style="flex: 1;">
				<BrowseInput
					id="launcherExe"
					bind:value={internalLauncherPath}
					placeholder="Path to launcher.exe (main executable to run)..."
					type="file"
					focusBorderColor="var(--accent-primary)"
					browseHandler={handleBrowseLauncherClick}
				/>
			</div>
		</div>

		<p class="exe-note launcher-note">
			Main executable to launch. Required for game execution.
		</p>
	</div>

	<!-- Game Executable Toggle -->
	<SlideButton
		bind:checked={useGamePath}
		label="Use Game Exe (for LSFG-VK)"
		subtitle="Configure different game exe for LSFG-VK profile"
	/>

	<!-- Game Executable Section (Conditional) -->
	{#if useGamePath}
		<div class="game-exe-section">
			<label for="gameExe">
				Game Executable <span class="optional-tag">For LSFG-VK</span
				>
			</label>

			<div class="game-exe-wrapper">
				<div class="exe-icon-display game-icon">
					{#if gameIcon && !gameIconFailed}
						<img
							src={gameIcon}
							alt="Game Icon"
							class="exe-icon"
							on:load={() => {
								gameIconFailed = false;
							}}
							on:error={() => {
								gameIconFailed = true;
							}}
						/>
					{:else}
						<div class="exe-icon-placeholder">
							<span
								class="material-icons"
								style="font-size: 32px;"
								>laptop_windows</span
							>
						</div>
					{/if}
				</div>

				<div class="game-input-group" style="flex: 1;">
					<BrowseInput
						id="gameExe"
						bind:value={internalGamePath}
						placeholder="Select game .exe file..."
						type="file"
						focusBorderColor="var(--accent-secondary)"
						browseHandler={handleBrowseGameClick}
					/>
				</div>
			</div>

			<p class="exe-note game-note">
				Used for LSFG-VK profile matching and configuration.
			</p>
		</div>
	{/if}
</div>

<style lang="scss">
	.exe-selector {
		display: flex;
		flex-direction: column;
		gap: 18px;
		margin-bottom: 24px;
	}

	.launcher-exe-section,
	.game-exe-section {
		padding: 20px;
		background: var(--bg-surface);
		border-radius: var(--radius-lg);
		border: 2px solid var(--glass-border);
	}

	label {
		display: flex;
		align-items: center;
		font-size: 0.9rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 1px;
		margin-bottom: 14px;
	}

	.launcher-exe-section label {
		color: var(--accent-primary);
	}

	.game-exe-section label {
		color: var(--accent-secondary);
	}

	.required-tag {
		font-size: 0.7rem;
		padding: 4px 8px;
		border-radius: var(--radius-sm);
		margin-left: 10px;
		text-transform: uppercase;
		font-weight: 900;
		background: var(--accent-primary);
		color: var(--bg-base);
	}

	.optional-tag {
		font-size: 0.7rem;
		padding: 4px 8px;
		border-radius: var(--radius-sm);
		margin-left: 10px;
		text-transform: uppercase;
		font-weight: 900;
		background: var(--accent-secondary);
		color: #000000;
	}

	.launcher-exe-wrapper,
	.game-exe-wrapper {
		display: flex;
		gap: 16px;
		align-items: center;
	}

	.exe-icon-display {
		flex-shrink: 0;
		width: 56px;
		height: 56px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-md);
		background: var(--bg-base);
		border: 2px solid var(--glass-border);
		overflow: hidden;
		transition: transform var(--transition-spring), border-color var(--transition-fast);

		&:hover {
			transform: scale(1.1);
			border-color: var(--glass-border-bright);
		}
	}

	.exe-icon {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.exe-icon-placeholder {
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--text-muted);
		width: 100%;
		height: 100%;
	}

	.launcher-input-group,
	.game-input-group {
		flex: 1;
		display: flex;
		gap: 12px;
	}



	.exe-note {
		font-size: 0.85rem;
		font-weight: 700;
		margin-top: 12px;
	}

	.launcher-note {
		color: var(--text-muted);
	}

	.game-note {
		color: var(--text-muted);
	}
</style>
