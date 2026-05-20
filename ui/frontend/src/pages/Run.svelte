<script lang="ts">
	import ExecutableSelector from "@components/run/ExecutableSelector.svelte";
	import LaunchButton from "@components/run/LaunchButton.svelte";
	import MissingDependenciesModal from "@components/run/MissingDependenciesModal.svelte";
	import PrefixSelector from "@components/run/PrefixSelector.svelte";
	import ProtonSelector from "@components/run/ProtonSelector.svelte";
	import ConfigForm from "@components/shared/ConfigForm.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import { RunPageState } from "@components/run/RunPageState.svelte";
	import { onMount } from "svelte";

	const state = new RunPageState();

	onMount(() => {
		state.initialize();
	});
</script>

<div class="run-container">
	<PageHeader title="Launch Configuration" icon="rocket_launch" />

	<div class="zones-stack">
		<!-- Zone 1: General & Paths -->
		<div class="zone-card">
			<div class="zone-header">
				<span class="material-icons">folder</span>
				<h2>General & Paths</h2>
			</div>

			<div class="form-group profile-name-group">
				<label for="profileName">Profile Name</label>
				<input
					id="profileName"
					type="text"
					class="input profile-input"
					bind:value={state.options.Name}
					placeholder="Enter a name for this profile..."
				/>
			</div>

			<ExecutableSelector
				launcherPath={state.options.LauncherPath}
				gamePath={state.options.GamePath}
				bind:useGamePath={state.useGamePath}
				bind:launcherIcon={state.launcherIcon}
				bind:gameIcon={state.gameIcon}
				onBrowseLauncher={() => state.handleBrowseLauncher()}
				onBrowseGame={() => state.handleBrowseGame()}
			/>
		</div>

		<!-- Zone 2: Environment & Compatibility -->
		<div class="zone-card">
			<div class="zone-header">
				<span class="material-icons">layers</span>
				<h2>Environment & Compatibility</h2>
			</div>

			<PrefixSelector
				bind:availablePrefixes={state.availablePrefixes}
				bind:selectedPrefixName={state.selectedPrefixName}
				bind:prefixPath={state.prefixPath}
				baseDir={state.baseDir}
				onPrefixChange={(name) => state.handlePrefixChange(name)}
				onBrowsePrefix={() => state.handleBrowsePrefix()}
			/>

			<div class="form-group">
				<SlideButton
					bind:checked={state.options.UseCustomProton}
					label="Use Custom Proton"
					subtitle="Prevent prefix changes from overwriting your proton selection"
				/>
			</div>

			<ProtonSelector
				bind:protonOptions={state.proton.protonOptions}
				bind:selectedProton={state.proton.selectedProton}
				bind:isLoadingProton={state.proton.isLoadingProton}
				onProtonChange={(val) => state.proton.handleProtonChange(val)}
				disabled={!state.options.UseCustomProton}
				disabledMessage="Enable 'Use Custom Proton' to change version"
			/>
		</div>

		<!-- Zone 3: Advanced Integrations -->
		<div class="zone-card">
			<div class="zone-header">
				<span class="material-icons">tune</span>
				<h2>Advanced Integrations</h2>
			</div>
			<ConfigForm
				bind:options={state.options}
				bind:showLogsWindow={state.showLogsWindow}
			/>
		</div>
	</div>

	<MissingDependenciesModal
		show={state.showValidationModal}
		missingTools={state.missingToolsList}
		onClose={() => (state.showValidationModal = false)}
		onConfirm={() => state.proceedToLaunch()}
	/>

	<div class="actions-row">
		<button
			class="btn secondary bg-launch-btn"
			on:click={() => state.handleLaunch(false)}
			title="Run in Background (Keep Launcher Open)"
		>
			<span class="material-icons">play_arrow</span>
		</button>
		<div class="launch-wrapper">
			<LaunchButton onLaunch={() => state.handleLaunch(true)} />
		</div>
		<button
			class="btn secondary save-btn"
			on:click={() => state.handleSave()}
			disabled={state.isSaving}
			title="Save Configuration"
		>
			<span class="material-icons"
				>{state.isSaving ? "sync" : "save"}</span
			>
		</button>
	</div>
</div>

<style lang="scss">
	.run-container {
		display: flex;
		flex-direction: column;
		padding: 40px 48px;
	}

	.zones-stack {
		display: flex;
		flex-direction: column;
		gap: 32px;
		margin-top: 24px;
	}

	.zone-card {
		background: var(--bg-surface);
		border-radius: var(--radius-lg);
		border: 2px solid rgba(255, 255, 255, 0.05);
		padding: 32px;
		display: flex;
		flex-direction: column;
		gap: 28px;
		transition:
			border-color var(--transition-fast),
			box-shadow var(--transition-fast);

		&:hover {
			border-color: rgba(255, 255, 255, 0.08);
			box-shadow: 0 12px 40px rgba(0, 0, 0, 0.25);
		}

		.zone-header {
			display: flex;
			align-items: center;
			gap: 12px;
			border-bottom: 2px solid rgba(255, 255, 255, 0.05);
			padding-bottom: 16px;
			margin-bottom: 4px;

			.material-icons {
				font-size: 24px;
				color: var(--accent-primary);
			}

			h2 {
				margin: 0;
				font-size: 1.15rem;
				font-weight: 800;
				color: var(--text-main);
				text-transform: uppercase;
				letter-spacing: 1px;
			}
		}
	}

	.profile-name-group {
		display: flex;
		flex-direction: column;
		gap: 10px;

		label {
			display: block;
			font-size: 0.9rem;
			font-weight: 800;
			color: var(--accent-primary);
			text-transform: uppercase;
			letter-spacing: 1px;
		}

		.profile-input {
			width: 100%;
			font-size: 1.15rem;
			font-weight: 700;
			background: var(--bg-input);
			border: 2px solid var(--glass-border);
			padding: 12px 20px;
			border-radius: var(--radius-md);
			color: var(--text-main);
			outline: none;
			transition:
				border-color var(--transition-fast),
				transform var(--transition-fast);

			&:focus {
				border-color: var(--accent-primary);
				transform: scale(1.005);
			}
		}
	}

	.actions-row {
		position: sticky;
		bottom: -32px;
		margin: 48px -32px -32px -32px;
		padding: 32px;
		z-index: 10;
		background: linear-gradient(to top, var(--bg-base) 80%, transparent);
		display: flex;
		align-items: center;
		gap: 16px;
		pointer-events: none;
	}

	.launch-wrapper {
		flex: 1;
		pointer-events: auto;
	}

	.bg-launch-btn {
		pointer-events: auto;
		width: 60px;
		height: 60px;
		padding: 0;
		flex-shrink: 0;
		transform: none !important;
		transition: color var(--transition-fast), border-color var(--transition-fast) !important;

		&:hover:not(:disabled) {
			color: var(--accent-primary);
			border-color: var(--accent-primary);
			transform: none !important;
		}

		&:active:not(:disabled) {
			transform: none !important;
		}

		.material-icons {
			font-size: 24px;
		}
	}

	.save-btn {
		pointer-events: auto;
		width: 60px;
		height: 60px;
		padding: 0;
		flex-shrink: 0;

		&:hover:not(:disabled) {
			transform: scale(1.1);
		}

		&:active:not(:disabled) {
			transform: scale(0.95);
		}

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
			.material-icons {
				animation: spin 2s linear infinite;
			}
		}

		.material-icons {
			font-size: 24px;
		}
	}

	@keyframes spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}
</style>
