<script lang="ts">
	import Dropdown from "@components/shared/Dropdown.svelte";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import MultiplierInput from "@components/shared/MultiplierInput.svelte";
	import * as core from "@bindings/light-launcher/pkg/core/models";

	export let options: core.LaunchOptions;
	export let gpuList: string[] = [];
	export let onDllBrowse: (() => void) | null = null;
	export let compact: boolean = false; // For modal vs full-page mode

	$: showDllSection = !compact; // Only show DLL in modal/full form, not in Edit LSFG page

	function handleDllInput(e: Event) {
		const input = e.target as HTMLInputElement;
		options.LsfgDllPath = input.value;
	}
</script>

<div class="lsfg-config-form">
	<div class="form-group">
		<label for="lsfgDll">
			DLL Path
			{#if options.LsfgDllPath}
				<span class="status-badge success">✓ Found</span>
			{:else}
				<span class="status-badge error">
					<span class="material-icons" style="font-size: 14px;"
						>warning</span
					>
					Not Set
				</span>
			{/if}
		</label>
		<div class="input-group">
			<input
				id="lsfgDll"
				type="text"
				class="input sm"
				value={options.LsfgDllPath}
				on:input={handleDllInput}
				placeholder="Path to Lossless.dll..."
				readonly={!compact}
			/>
			{#if onDllBrowse && showDllSection}
				<button class="btn sm" on:click={onDllBrowse}>Browse</button
				>
			{/if}
		</div>
		<p class="help-text">Lossless Scaling DLL path for LSFG-VK</p>
	</div>

	<div class="form-group">
		<label for="lsfgMultiplier">FPS Multiplier</label>
		<div id="lsfgMultiplier">
			<MultiplierInput
				value={options.LsfgMultiplier}
				onChange={(val) => (options.LsfgMultiplier = val)}
			/>
		</div>
	</div>

	<div class="form-group">
		<SlideButton
			bind:checked={options.LsfgPerfMode}
			label="Performance Mode"
			subtitle="Faster frame generation, slight quality loss"
		/>
	</div>

	{#if !compact}
		<div class="divider"></div>
		<div class="advanced-title">Advanced Settings</div>
	{/if}

	<div class="form-group">
		<label for="lsfgGpu">GPU</label>
		<div id="lsfgGpu">
			<Dropdown
				options={gpuList}
				bind:value={options.LsfgGpu}
				onChange={(val) =>
					(options.LsfgGpu = val === "Auto (Detect)" ? "" : val)}
			/>
		</div>
		<p class="help-text">
			Select GPU for frame generation. Auto-detects if available GPUs
			are found.
		</p>
	</div>

	<div class="form-group">
		<label for="lsfgFlowScale">Flow Scale</label>
		<input
			id="lsfgFlowScale"
			type="text"
			class="input"
			bind:value={options.LsfgFlowScale}
			placeholder="e.g. 0.8"
		/>
		<p class="help-text">
			Optical flow scale for frame interpolation (0.0-2.0). Default:
			0.8
		</p>
	</div>

	<div class="form-group">
		<label for="lsfgPacing">Pacing</label>
		<div id="lsfgPacing">
			<Dropdown
				options={["none", "monitor", "monitor_external"]}
				bind:value={options.LsfgPacing}
				onChange={(val) => (options.LsfgPacing = val)}
			/>
		</div>
		<p class="help-text">Frame timing strategy for LSFG-VK.</p>
	</div>

	<div class="form-group">
		<SlideButton
			bind:checked={options.LsfgAllowFp16}
			label="Allow FP16"
			subtitle="Use 16-bit floating point (faster but less precise)"
		/>
	</div>
</div>

<style lang="scss">
	.lsfg-config-form {
		display: flex;
		flex-direction: column;
		gap: 16px;
		max-width: 900px;
		width: 100%;
		margin: 0 auto;
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

	.input-group {
		display: flex;
		gap: 12px;
		width: 100%;

		.input {
			flex: 1;
		}
	}

	.input {
		padding: 8px 12px;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 6px;
		color: var(--text-main);
		font-size: 0.95rem;

		&:focus {
			outline: none;
			border-color: var(--accent-primary);
			box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
		}

		&.sm {
			padding: 6px 10px;
			font-size: 0.875rem;
		}
	}

	.btn {
		padding: 8px 16px;
		background: var(--accent-primary);
		border: none;
		border-radius: 6px;
		color: var(--glass-bg);
		font-weight: 600;
		cursor: pointer;
		transition: background 0.2s;

		&:hover {
			background: var(
				--accent-primary-hover,
				color-mix(in srgb, var(--accent-primary) 80%, black)
			);
		}

		&.sm {
			padding: 6px 12px;
			font-size: 0.8rem;
		}
	}

	.status-badge {
		font-size: 0.65rem;
		padding: 2px 6px;
		border-radius: 4px;
		margin-left: auto;
		text-transform: uppercase;
		font-weight: bold;
		display: flex;
		align-items: center;
		gap: 4px;

		.material-icons {
			font-size: 10px;
		}

		&.success {
			background: rgba(16, 185, 129, 0.2);
			color: #10b981;
		}

		&.error {
			background: rgba(239, 68, 68, 0.2);
			color: #ef4444;
		}
	}

	.divider {
		height: 1px;
		background: var(--glass-border);
		margin: 16px 0;
	}

	.advanced-title {
		font-size: 0.875rem;
		font-weight: 700;
		color: var(--text-muted);
		margin-bottom: 12px;
		text-transform: uppercase;
		letter-spacing: 1px;
	}
</style>
