<script lang="ts">
	import Dropdown from "@components/shared/Dropdown.svelte";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import * as core from "@shared";

	export let options: core.LaunchOptions;

	const scalers = ["auto", "integer", "fit", "fill", "stretch"];
	const filters = ["linear", "nearest", "fsr", "nis", "pixel"];
	const windowModes = ["fullscreen", "borderless", "windowed"];
</script>

<div class="gamescope-config-form">
	<!-- Resolution Row (Nested / Game Resolution) -->
	<div class="form-row">
		<div class="form-group">
			<label for="gsW">Game Width (px)</label>
			<input
				id="gsW"
				type="text"
				class="input"
				bind:value={options.Extras.Gamescope.Width}
				placeholder="e.g. 1920"
			/>
			<p class="help-text">Internal resolution exposed to the game (-w)</p>
		</div>
		<div class="form-group">
			<label for="gsH">Game Height (px)</label>
			<input
				id="gsH"
				type="text"
				class="input"
				bind:value={options.Extras.Gamescope.Height}
				placeholder="e.g. 1080"
			/>
			<p class="help-text">Internal resolution exposed to the game (-h)</p>
		</div>
	</div>

	<!-- Output Resolution Row -->
	<div class="form-row">
		<div class="form-group">
			<label for="gsOutW">Output Width (px)</label>
			<input
				id="gsOutW"
				type="text"
				class="input"
				bind:value={options.Extras.Gamescope.OutputWidth}
				placeholder="e.g. 1920"
			/>
			<p class="help-text">Real screen / window target width (-W)</p>
		</div>
		<div class="form-group">
			<label for="gsOutH">Output Height (px)</label>
			<input
				id="gsOutH"
				type="text"
				class="input"
				bind:value={options.Extras.Gamescope.OutputHeight}
				placeholder="e.g. 1080"
			/>
			<p class="help-text">Real screen / window target height (-H)</p>
		</div>
	</div>

	<!-- Performance Settings -->
	<div class="form-row">
		<div class="form-group">
			<label for="gsR">Refresh Rate (Hz)</label>
			<input
				id="gsR"
				type="text"
				class="input"
				bind:value={options.Extras.Gamescope.RefreshRate}
				placeholder="e.g. 60"
			/>
			<p class="help-text">Target nested refresh rate (-r)</p>
		</div>
		<div class="form-group">
			<label for="gsFRLimit">Framerate Limit (FPS)</label>
			<input
				id="gsFRLimit"
				type="text"
				class="input"
				bind:value={options.Extras.Gamescope.FramerateLimit}
				placeholder="e.g. 60"
			/>
			<p class="help-text">Limits Gamescope output fps (--framerate-limit)</p>
		</div>
	</div>

	<!-- Window & Scaling Mode -->
	<div class="form-row">
		<div class="form-group">
			<label for="gsWindowMode">Window Mode</label>
			<div id="gsWindowMode">
				<Dropdown
					options={windowModes}
					bind:value={options.Extras.Gamescope.WindowMode}
					placeholder="Select window mode"
				/>
			</div>
			<p class="help-text">Gamescope window type (-f or -b)</p>
		</div>
		<div class="form-group">
			<label for="gsScaler">Scaler</label>
			<div id="gsScaler">
				<Dropdown
					options={scalers}
					bind:value={options.Extras.Gamescope.Scaler}
					placeholder="auto"
				/>
			</div>
			<p class="help-text">Scaling strategy (-S)</p>
		</div>
	</div>

	<!-- Filter & Sharpness -->
	<div class="form-row">
		<div class="form-group">
			<label for="gsFilter">Filter</label>
			<div id="gsFilter">
				<Dropdown
					options={filters}
					bind:value={options.Extras.Gamescope.Filter}
					placeholder="linear"
				/>
			</div>
			<p class="help-text">Upscaling filter type (-F)</p>
		</div>
		<div class="form-group">
			<label for="gsSharpness">Sharpness (0-20)</label>
			<input
				id="gsSharpness"
				type="text"
				class="input"
				bind:value={options.Extras.Gamescope.Sharpness}
				placeholder="e.g. 2"
			/>
			<p class="help-text">FSR/NIS scaling sharpness (0 is max, 20 is min)</p>
		</div>
	</div>

	<!-- Custom Arguments -->
	<div class="form-group">
		<label for="gsCustomArgs">Custom Arguments</label>
		<input
			id="gsCustomArgs"
			type="text"
			class="input"
			bind:value={options.Extras.Gamescope.CustomArgs}
			placeholder="e.g. --prefer-vk-device 1002:7300 --force-orientation left"
		/>
		<p class="help-text">Additional raw command line options passed directly to Gamescope</p>
	</div>

	<div class="divider"></div>

	<!-- Toggles Grid -->
	<div class="toggles-grid">
		<SlideButton
			bind:checked={options.Extras.Gamescope.HDR}
			label="HDR Output"
			subtitle="Enable HDR support (--hdr-enabled)"
		/>
		<SlideButton
			bind:checked={options.Extras.Gamescope.AdaptiveSync}
			label="Adaptive Sync (VRR)"
			subtitle="Variable Refresh Rate (--adaptive-sync)"
		/>
		<SlideButton
			bind:checked={options.Extras.Gamescope.Mangoapp}
			label="MangoApp Overlay"
			subtitle="MangoHud built into Gamescope (--mangoapp)"
		/>
	</div>
</div>

<style lang="scss">
	.gamescope-config-form {
		display: flex;
		flex-direction: column;
		gap: 16px;
		max-width: 900px;
		width: 100%;
		margin: 0 auto;
	}

	.form-row {
		display: flex;
		gap: 16px;
		width: 100%;

		.form-group {
			flex: 1;
		}
	}

	.form-group {
		display: flex;
		flex-direction: column;

		label {
			display: block;
			font-size: 0.875rem;
			font-weight: 600;
			color: var(--text-muted);
			margin-bottom: 8px;
			display: flex;
			align-items: center;
			gap: 12px;
		}

		.help-text {
			font-size: 0.7rem;
			color: var(--text-dim);
			margin-top: 4px;
			font-style: italic;
		}
	}

	.input {
		padding: 12px 20px;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-pill);
		color: var(--text-main);
		font-size: 0.9rem;
		font-weight: 700;
		outline: none;
		transition: border-color var(--transition-fast), transform var(--transition-spring);

		&:focus {
			border-color: var(--accent-primary);
			transform: scale(1.005);
		}
	}

	.toggles-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
		gap: 16px;
	}

	.divider {
		height: 1px;
		background: var(--glass-border);
		margin: 16px 0;
	}
</style>
