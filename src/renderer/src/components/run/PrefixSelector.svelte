<script lang="ts">
	import Dropdown, { asOptions } from "@components/shared/Dropdown.svelte";

	$: prefixItems = asOptions([...availablePrefixes, "Custom..."]);
	import BrowseInput from "@components/shared/BrowseInput.svelte";
	import Button from "@components/shared/Button.svelte";

	export let availablePrefixes: string[] = [];
	export let selectedPrefixName: string = "Default";
	export let prefixPath: string = "";
	export let baseDir: string = "";
	export let onPrefixChange: (name: string) => Promise<void>;
	export let onBrowsePrefix: () => Promise<void>;
</script>

<div class="form-group">
	<label for="winePrefix">WINEPREFIX</label>
	<div class="input-group">
		<div class="dropdown-wrapper">
			<Dropdown
				options={prefixItems}
				bind:value={selectedPrefixName}
				onChange={onPrefixChange}
			/>
		</div>
		<Button onclick={onBrowsePrefix}>Browse</Button>
	</div>
	{#if selectedPrefixName === "Custom..." || !prefixPath.startsWith(baseDir)}
		<div style="margin-top: 12px;">
			<BrowseInput
				bind:value={prefixPath}
				type="folder"
				placeholder="Custom WINEPREFIX path..."
				browseHandler={onBrowsePrefix}
			/>
		</div>
	{/if}
</div>

<style lang="scss">
	.form-group label {
		display: block;
		font-size: 0.9rem;
		font-weight: 800;
		color: var(--accent-primary);
		text-transform: uppercase;
		letter-spacing: 1px;
		margin-bottom: 10px;
	}
	.input-group {
		display: flex;
		gap: 12px;
		width: 100%;
		.dropdown-wrapper {
			flex: 1;
		}
	}
</style>
