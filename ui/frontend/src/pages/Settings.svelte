<script lang="ts">
	import { settingsStore } from "@stores/settingsStore";
	import { PickFileCustom } from "@bindings/light-launcher/internal/app/app";
	import { notifications } from "@stores/notificationStore";

	let currentSettings = {
		theme: "light",
		transparency: 1.0,
		backgroundImagePath: "",
	};

	settingsStore.subscribe((val) => {
		currentSettings = val;
	});

	function handleThemeToggle() {
		settingsStore.update((s) => ({
			...s,
			theme: s.theme === "light" ? "dark" : "light",
		}));
	}

	function handleTransparencyChange(e: Event) {
		const val = parseFloat((e.target as HTMLInputElement).value);
		settingsStore.update((s) => ({
			...s,
			transparency: val,
		}));
	}

	async function handleBrowseBackground() {
		try {
			const path = await PickFileCustom("Select Background Image", [
				{
					DisplayName: "Images",
					Pattern: "*.png;*.jpg;*.jpeg;*.svg;*.webp",
				},
			]);
			if (path) {
				settingsStore.update((s) => ({
					...s,
					backgroundImagePath: path,
				}));
				notifications.add("Background image updated", "success");
			}
		} catch (err) {
			notifications.add("Failed to pick image", "error");
		}
	}

	function handleClearBackground() {
		settingsStore.update((s) => ({
			...s,
			backgroundImagePath: "",
		}));
		notifications.add("Background image cleared", "info");
	}
</script>

<div class="settings-container">
	<h1 class="page-title">Appearance & Settings</h1>

	<div class="settings-grid">
		<div class="settings-card glass">
			<div class="settings-section">
				<h3>Theme Mode</h3>
				<p class="desc">Switch between dark and light themes.</p>
				<button class="btn" on:click={handleThemeToggle}>
					<span class="material-icons mini-icon">contrast</span>
					<span
						>Switch to {currentSettings.theme === "light"
							? "Dark"
							: "Light"} Mode</span
					>
				</button>
			</div>
		</div>

		<div class="settings-card glass">
			<div class="settings-section">
				<h3>Window Transparency</h3>
				<p class="desc">
					Adjust OS window transparency.
				</p>
				<div class="slider-row">
					<input
						type="range"
						min="0.1"
						max="1.0"
						step="0.05"
						value={currentSettings.transparency}
						on:input={handleTransparencyChange}
						class="transparency-slider"
					/>
					<span class="pct-display"
						>{Math.round(currentSettings.transparency * 100)}%</span
					>
				</div>
			</div>
		</div>

		<div class="settings-card glass">
			<div class="settings-section">
				<h3>Background Image</h3>
				<p class="desc">
					Set a custom image background.
				</p>

				{#if currentSettings.backgroundImagePath}
					<div class="bg-path-display">
						<span class="material-icons mini-icon">wallpaper</span>
						<span
							class="path-text"
							title={currentSettings.backgroundImagePath}
							>{currentSettings.backgroundImagePath}</span
						>
					</div>
				{/if}

				<div class="actions-row">
					<button class="btn primary" on:click={handleBrowseBackground}>
						<span class="material-icons mini-icon">folder</span>
						Browse Image
					</button>
					{#if currentSettings.backgroundImagePath}
						<button class="btn danger" on:click={handleClearBackground}>
							<span class="material-icons mini-icon">delete</span>
							Clear
						</button>
					{/if}
				</div>
			</div>
		</div>
	</div>
</div>

<style lang="scss">
	.settings-container {
		padding: 32px;
		height: 100%;
		display: flex;
		flex-direction: column;
		overflow-y: auto;
	}

	.page-title {
		font-size: 2rem;
		font-weight: 800;
		color: var(--text-main);
		margin: 0 0 32px 0;
	}

	.settings-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 24px;
		width: 100%;
	}

	.settings-card {
		padding: 32px;
		border-radius: 20px;
		border: 1px solid var(--glass-border);
		background: var(--glass-surface);
		display: flex;
		flex-direction: column;
		gap: 28px;
		transition: all 0.2s ease;

		&:hover {
			background: var(--glass-hover);
			border-color: var(--glass-border-bright);
		}
	}

	.settings-section {
		display: flex;
		flex-direction: column;
		gap: 8px;

		h3 {
			margin: 0;
			font-size: 1.1rem;
			font-weight: 700;
			color: var(--text-main);
		}

		.desc {
			margin: 0 0 8px 0;
			font-size: 0.85rem;
			color: var(--text-dim);
		}

		.mini-icon {
			font-size: 18px;
			margin-right: 6px;
		}
	}



	.slider-row {
		display: flex;
		align-items: center;
		gap: 16px;

		.transparency-slider {
			flex: 1;
			accent-color: var(--accent-primary);
		}

		.pct-display {
			font-size: 0.9rem;
			font-weight: 700;
			color: var(--accent-primary);
			width: 40px;
			text-align: right;
		}
	}

	.bg-path-display {
		display: flex;
		align-items: center;
		padding: 10px 14px;
		background: var(--glass-bg);
		border: 1px solid var(--glass-border);
		border-radius: 10px;
		margin-bottom: 12px;
		color: var(--text-muted);

		.path-text {
			font-size: 0.8rem;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
	}

	.actions-row {
		display: flex;
		gap: 12px;

		button {
			display: inline-flex;
			align-items: center;
			justify-content: center;
		}
	}
</style>
