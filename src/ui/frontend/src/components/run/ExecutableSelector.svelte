<script lang="ts">
	import { loadExeIcon } from "@lib/iconService";
	import SlideButton from "@components/shared/SlideButton.svelte";

	export let launcherPath = "";
	export let mainExePath = "";
	export let haveGameExe = false;
	export let launcherIcon = "";
	export let gameIcon = "";
	export let onBrowseLauncher: () => Promise<void>;
	export let onBrowseGame: () => Promise<void>;

	let launcherIconFailed = false;
	let gameIconFailed = false;
	let prevLauncherPath = "";
	let prevMainExePath = "";

	// Internal state for inputs - keep in sync with props
	let internalLauncherPath = launcherPath;
	let internalMainExePath = mainExePath;

	// When props change, update internal state
	$: if (launcherPath !== internalLauncherPath) {
		console.log(
			"ExecutableSelector: launcherPath prop changed to:",
			launcherPath,
		);
		internalLauncherPath = launcherPath;
	}

	$: if (mainExePath !== internalMainExePath) {
		console.log(
			"ExecutableSelector: mainExePath prop changed to:",
			mainExePath,
		);
		internalMainExePath = mainExePath;
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
	$: if (internalMainExePath && internalMainExePath !== prevMainExePath) {
		prevMainExePath = internalMainExePath;
		gameIconFailed = false;
		if (haveGameExe) {
			(async () => {
				const icon = await loadExeIcon(internalMainExePath);
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

			<div class="input-group launcher-input-group">
				<input
					id="launcherExe"
					type="text"
					bind:value={internalLauncherPath}
					placeholder="Path to launcher.exe (main executable to run)..."
					class="input"
				/>
				<button on:click={handleBrowseLauncherClick} class="btn"
					>Browse</button
				>
			</div>
		</div>

		<p class="exe-note launcher-note">
			Main executable to launch. Required for game execution.
		</p>
	</div>

	<!-- Game Executable Toggle -->
	<SlideButton
		bind:checked={haveGameExe}
		label="Use Game Exe (for LSFG-VK)"
		subtitle="Configure different game exe for LSFG-VK profile"
	/>

	<!-- Game Executable Section (Conditional) -->
	{#if haveGameExe}
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

				<div class="input-group game-input-group">
					<input
						id="gameExe"
						type="text"
						bind:value={internalMainExePath}
						placeholder="Select game .exe file..."
						class="input"
					/>
					<button on:click={handleBrowseGameClick} class="btn"
						>Browse</button
					>
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
		gap: 12px;
		margin-bottom: 16px;
	}

	.launcher-exe-section,
	.game-exe-section {
		padding: 16px;
		background: var(--glass-surface);
		border-radius: 10px;
		border: 1px solid var(--glass-border);
	}

	.launcher-exe-section {
		padding: 20px 16px;
		background: var(--glass-surface);
		border-radius: 10px;
		border: 1px solid rgba(96, 165, 250, 0.3);
	}

	.game-exe-section {
		padding: 16px;
		background: var(--glass-surface);
		border-radius: 10px;
		border: 1px solid rgba(192, 132, 252, 0.3);
	}

	label {
		display: flex;
		align-items: center;
		font-size: 0.875rem;
		font-weight: 700;
		color: var(--text-muted);
		margin-bottom: 12px;
	}

	.launcher-exe-section label {
		color: #2563eb; /* Darker blue for light mode */
		:global([data-theme="dark"]) & {
			color: #60a5fa;
		}
	}

	.game-exe-section label {
		color: #9333ea; /* Darker purple for light mode */
		:global([data-theme="dark"]) & {
			color: #c084fc;
		}
	}

	.required-tag {
		font-size: 0.75rem;
		padding: 4px 10px;
		border-radius: 4px;
		margin-left: 8px;
		text-transform: uppercase;
		font-weight: 900;
		background: rgba(37, 99, 235, 0.1);
		color: #2563eb;
		:global([data-theme="dark"]) & {
			background: rgba(96, 165, 250, 0.25);
			color: #60a5fa;
		}
	}

	.optional-tag {
		font-size: 0.75rem;
		padding: 4px 10px;
		border-radius: 4px;
		margin-left: 8px;
		text-transform: uppercase;
		font-weight: 900;
		background: rgba(147, 51, 234, 0.1);
		color: #9333ea;
		:global([data-theme="dark"]) & {
			background: rgba(192, 132, 252, 0.25);
			color: #c084fc;
		}
	}

	.launcher-exe-wrapper,
	.game-exe-wrapper {
		display: flex;
		gap: 12px;
		align-items: flex-start;
	}

	.exe-icon-display {
		flex-shrink: 0;
		width: 56px;
		height: 56px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
		background: var(--glass-bg);
		border: 1px solid var(--glass-border);
		overflow: hidden;

		&.launcher-icon {
			background: rgba(59, 130, 246, 0.12);
			border-color: rgba(59, 130, 246, 0.3);
		}

		&.game-icon {
			background: rgba(168, 85, 247, 0.12);
			border-color: rgba(168, 85, 247, 0.3);
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
		color: var(--text-dim);
		width: 100%;
		height: 100%;
	}

	.launcher-input-group,
	.game-input-group {
		flex: 1;
		display: flex;
		gap: 10px;
	}

	.input {
		flex: 1;
		padding: 10px 12px;
		background: var(--glass-bg);
		border: 1px solid var(--glass-border);
		border-radius: 6px;
		color: var(--text-main);
		font-size: 0.875rem;
		font-family: inherit;

		&::placeholder {
			color: var(--text-dim);
		}

		&:focus {
			outline: none;
			border-color: var(--accent-primary);
			background: var(--glass-surface);
			box-shadow: 0 0 0 2px var(--glass-low-alpha);
		}
	}

	.btn {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 8px 16px;
		border-radius: 6px;
		font-weight: 800;
		font-size: 0.8rem;
		cursor: pointer;
		transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
		border: 1px solid var(--glass-border);
		background: var(--glass-hover);
		color: var(--text-main);
		white-space: nowrap;

		&:hover {
			background: var(--glass-border);
			border-color: var(--glass-border-bright);
		}
	}

	.exe-note {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text-dim);
		margin-top: 10px;
		font-style: italic;
	}

	.launcher-note {
		color: rgba(96, 165, 250, 0.9);
	}

	.game-note {
		color: rgba(192, 132, 252, 0.9);
	}
</style>
