<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { onMount } from "svelte";
	import ConfigForm from "@components/shared/ConfigForm.svelte";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import ExecutableSelector from "@components/run/ExecutableSelector.svelte";
	import PrefixSelector from "@components/run/PrefixSelector.svelte";
	import ProtonSelector from "@components/run/ProtonSelector.svelte";
	import LaunchButton from "@components/run/LaunchButton.svelte";
	import MissingDependenciesModal from "@components/run/MissingDependenciesModal.svelte";
	import { RunPageState } from "@lib/RunPageState.svelte";

	const state = new RunPageState();

	onMount(() => {
		state.initialize();
	});
</script>

<div class="run-container">
	<PageHeader title="Launch Configuration" icon="rocket_launch" />

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

	<!-- Executable Selector Component -->
	<ExecutableSelector
		launcherPath={state.options.LauncherPath}
		gamePath={state.options.GamePath}
		bind:useGamePath={state.useGamePath}
		bind:launcherIcon={state.launcherIcon}
		bind:gameIcon={state.gameIcon}
		onBrowseLauncher={() => state.handleBrowseLauncher()}
		onBrowseGame={() => state.handleBrowseGame()}
	/>

	<!-- Main Form Container -->
	<div class="form-container">
		<PrefixSelector
			bind:availablePrefixes={state.availablePrefixes}
			bind:selectedPrefixName={state.selectedPrefixName}
			bind:prefixPath={state.prefixPath}
			baseDir={state.baseDir}
			onPrefixChange={(name) => state.handlePrefixChange(name)}
			onBrowsePrefix={() => state.handleBrowsePrefix()}
		/>

		<div class="form-group" style="margin-bottom: 20px;">
			<SlideButton
				bind:checked={state.options.UseCustomProton}
				label="Use Custom Proton"
				subtitle="Prevent prefix changes from overwriting your proton selection"
			/>
		</div>

		<ProtonSelector
			bind:protonOptions={state.protonOptions}
			bind:selectedProton={state.selectedProton}
			bind:isLoadingProton={state.isLoadingProton}
			onProtonChange={(val) => state.handleProtonChange(val)}
			disabled={!state.options.UseCustomProton}
			disabledMessage="Enable 'Use Custom Proton' to change version"
		/>

		<div class="divider"></div>

		<ConfigForm bind:options={state.options} />

		<div class="form-group">
			<SlideButton
				bind:checked={state.showLogsWindow}
				label="Show Logs"
				subtitle="Open logs in terminal"
			/>
		</div>

		<MissingDependenciesModal
			show={state.showValidationModal}
			missingTools={state.missingToolsList}
			onClose={() => (state.showValidationModal = false)}
			onConfirm={() => state.proceedToLaunch()}
		/>

		<div class="actions-row">
			<div class="launch-wrapper">
				<LaunchButton onLaunch={() => state.handleLaunch()} />
			</div>
			<button class="icon-btn save-btn" on:click={() => state.handleSave()} disabled={state.isSaving} title="Save Configuration">
				<span class="material-icons">{state.isSaving ? "sync" : "save"}</span>
			</button>
		</div>
	</div>
</div>

<style lang="scss">
	.run-container {
		display: flex;
		flex-direction: column;
		padding: 40px 48px;
	}
	.form-container {
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 28px;
	}

	.profile-name-group {
		margin-bottom: 28px;
		padding: 20px;
		background: var(--bg-surface);
		border-radius: var(--radius-lg);
		border: 2px solid rgba(255, 255, 255, 0.05);

		label {
			display: block;
			font-size: 0.9rem;
			font-weight: 800;
			color: var(--accent-primary);
			text-transform: uppercase;
			letter-spacing: 1px;
			margin-bottom: 10px;
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
			transition: border-color var(--transition-fast), transform var(--transition-fast);

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

	.save-btn {
		pointer-events: auto;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 60px;
		height: 60px;
		border-radius: var(--radius-pill);
		background: var(--bg-surface);
		color: var(--text-muted);
		border: 2px solid rgba(255, 255, 255, 0.05);
		cursor: pointer;
		transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);

		&:hover:not(:disabled) {
			background: var(--bg-elevated);
			color: var(--text-main);
			border-color: var(--accent-secondary);
			transform: scale(1.1);
		}

		&:active {
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
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	.divider {
		height: 2px;
		background: rgba(255, 255, 255, 0.05);
		margin: 12px 0;
		border-radius: var(--radius-pill);
	}
</style>
