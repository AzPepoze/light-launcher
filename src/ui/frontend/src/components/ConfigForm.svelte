<script lang="ts">
	import SlideButton from "./SlideButton.svelte";
	import Modal from "./Modal.svelte";
	import RangeSlider from "./RangeSlider.svelte";
	import LsfgConfigForm from "./LsfgConfigForm.svelte";
	import { PickFileCustom, GetTotalRam } from "@bindings/light-launcher-wails/backend/app";
	import * as core from "@bindings/light-launcher/pkg/core/models";
	import { onMount } from "svelte";
	import { loadLsfgResources, parseMemoryValue } from "../lib/formService";

	export let options: core.LaunchOptions;
	let showLsfgModal = false;
	let showGamescopeModal = false;
	let showMemoryModal = false;

	let memorySliderValue = 4;
	let systemRamTotal = 16;
	let gpuList: string[] = ["Auto (Detect)"];

	onMount(async () => {
		try {
			const ram = await GetTotalRam();
			if (ram > 0) systemRamTotal = ram;

			if (options.MemoryMinValue) {
				const numericMatch = options.MemoryMinValue.match(/([\d.]+)/);
				if (numericMatch) {
					const val = parseMemoryValue(options.MemoryMinValue);
					if (options.MemoryMinValue.endsWith("M")) {
						memorySliderValue = val;
					} else {
						memorySliderValue = val;
					}
				}
			}

			// Load GPUs and auto-detect DLL
			const { gpus, dll } = await loadLsfgResources();
			if (gpus.length > 0) {
				gpuList = ["Auto (Detect)", ...gpus];
			}
			if (dll && !options.LsfgDllPath) {
				options.LsfgDllPath = dll;
				console.log("[ConfigForm] Auto-detected DLL:", dll);
			}
		} catch (e) {
			console.error(e);
		}
	});

	async function handleBrowseDll() {
		try {
			const path = await PickFileCustom("Select Lossless.dll", [
				{ DisplayName: "Lossless.dll", Pattern: "Lossless.dll" },
			]);
			if (path) options.LsfgDllPath = path;
		} catch (err) {
			console.error(err);
		}
	}
</script>

<div class="config-form">
	<div class="form-group">
		<label for="customArgs">Custom Arguments</label>
		<input
			id="customArgs"
			type="text"
			class="input"
			bind:value={options.CustomArgs}
			placeholder="e.g. -windowed -novid"
		/>
	</div>

	<div class="toggles-grid">
		<SlideButton bind:checked={options.EnableMangoHud} label="MangoHud" subtitle="Performance overlay" />
		<SlideButton bind:checked={options.EnableGamemode} label="GameMode" subtitle="Optimize priorities" />
		<SlideButton
			bind:checked={options.EnableLsfgVk}
			label="LSFG-VK"
			subtitle="Lossless Scaling Frame Generation"
			hasConfig={true}
			onConfig={() => (showLsfgModal = true)}
		/>
		<SlideButton
			bind:checked={options.EnableGamescope}
			label="Gamescope"
			subtitle="Micro-compositor"
			hasConfig={true}
			onConfig={() => (showGamescopeModal = true)}
		/>
		<SlideButton
			bind:checked={options.EnableMemoryMin}
			label="Memory Protect"
			subtitle="Prevent Swap (Min RAM)"
			hasConfig={true}
			onConfig={() => (showMemoryModal = true)}
		/>
	</div>

	<!-- LSFG Settings Modal -->
	<Modal
		show={showLsfgModal}
		title="LSFG-VK Configuration"
		fullscreen={true}
		onClose={() => (showLsfgModal = false)}
	>
		<LsfgConfigForm {options} {gpuList} onDllBrowse={handleBrowseDll} />
	</Modal>

	<!-- Gamescope Settings Modal -->
	<Modal show={showGamescopeModal} title="Gamescope Configuration" onClose={() => (showGamescopeModal = false)}>
		<div class="modal-form">
			<div class="form-group">
				<label for="gamescopeWidth">Width (px)</label>
				<input
					id="gamescopeWidth"
					type="text"
					class="input"
					bind:value={options.GamescopeW}
					placeholder="e.g. 1920"
				/>
			</div>
			<div class="form-group">
				<label for="gamescopeHeight">Height (px)</label>
				<input
					id="gamescopeHeight"
					type="text"
					class="input"
					bind:value={options.GamescopeH}
					placeholder="e.g. 1080"
				/>
			</div>
			<div class="form-group">
				<label for="gamescopeRefresh">Refresh Rate (Hz)</label>
				<input
					id="gamescopeRefresh"
					type="text"
					class="input"
					bind:value={options.GamescopeR}
					placeholder="e.g. 60"
				/>
			</div>
			<p class="note">Note: Mouse visibility fix enabled automatically.</p>
		</div>
	</Modal>

	<!-- Memory Settings Modal -->
	<Modal show={showMemoryModal} title="Memory Protection" onClose={() => (showMemoryModal = false)}>
		<div class="modal-form">
			<div class="form-group">
				<label for="memorySlider">Minimum RAM Allocation</label>
				<RangeSlider
					value={memorySliderValue}
					max={systemRamTotal}
					snapValues={[2, 4, 6, 8, 12, 16, 24, 32, 48, 64]}
					onChange={(changedValue) => {
						memorySliderValue = changedValue;
						options.MemoryMinValue = changedValue + "G";
					}}
				/>
			</div>
			<p class="note">
				Guarantees {options.MemoryMinValue} of physical RAM for the game process, preventing swapping.
				<br />Values in Red Zone (>75%) might cause system instability.
			</p>
		</div>
	</Modal>
</div>

<style lang="scss">
	.toggles-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 16px;
		margin-top: 8px;
	}
	.modal-form {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.form-group label {
		display: block;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-muted);
		margin-bottom: 8px;
	}

	.form-group .input {
		width: 100%;
		display: block;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		color: var(--text-main);
	}
	.note {
		font-size: 0.75rem;
		color: var(--text-muted);
		font-style: italic;
		margin-top: 8px;
	}
</style>
