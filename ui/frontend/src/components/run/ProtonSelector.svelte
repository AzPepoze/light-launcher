<script lang="ts">
	import Dropdown from "@components/shared/Dropdown.svelte";
	
	export let protonOptions: string[] = [];
	export let selectedProton: string = "";
	export let isLoadingProton: boolean = true;
	export let onProtonChange: (value: string) => void;
	export let disabled: boolean = false;
</script>

<div class="form-group" class:disabled>
	<label for="protonVersion">Proton Version</label>
	<div id="protonVersion">
		<Dropdown
			options={protonOptions}
			bind:value={selectedProton}
			placeholder={isLoadingProton ? "Scanning..." : "Select Version"}
			disabled={isLoadingProton || disabled}
			onChange={onProtonChange}
		/>
	</div>
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
		transition: color 0.2s ease, opacity 0.2s ease;
	}
	.form-group.disabled label {
		color: var(--text-muted, rgba(255, 255, 255, 0.4));
		opacity: 0.7;
	}
</style>
