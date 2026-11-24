<script lang="ts">
	import { onDestroy, onMount } from "svelte";
	import type { Partition, SetupEntry } from "@shared";
	import DiskList from "@components/drives/DiskList.svelte";
	import SetupList from "@components/drives/SetupList.svelte";
	import BrowseInput from "@components/shared/BrowseInput.svelte";
	import Button from "@components/shared/Button.svelte";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import Section from "@components/shared/Section.svelte";
	import {
		cloneEntries,
		isSameSetup,
		letterOptions,
		lettersAvailableTo,
		nextFreeLetter
	} from "@lib/driveSetup";
	import { loadDriveSetup, loadPartitions, saveDriveSetup } from "@lib/prefixService";
	import { leaveGuard } from "@stores/leaveGuard";
	import { notifications } from "@stores/notificationStore";

	const FOLDER_INPUT_ID = "drive-folder";

	let partitions = $state<Partition[]>([]);
	let entries = $state<SetupEntry[]>([]);
	/** The setup as last loaded or saved; the Save button and the leave guard compare against it. */
	let savedEntries = $state<SetupEntry[]>([]);
	let isLoading = $state(true);
	/** True when the saved setup could not be read. Saving then would overwrite it with an empty list. */
	let loadFailed = $state(false);
	let isSaving = $state(false);
	let newFolder = $state("");
	let pickedLetter = $state("");

	const hasChanges = $derived(!isSameSetup(entries, savedEntries));
	const nextLetter = $derived(nextFreeLetter(entries));
	const folderLetter = $derived(
		lettersAvailableTo(entries, "").includes(pickedLetter) ? pickedLetter : nextLetter
	);
	const folderLetterOptions = $derived(letterOptions(lettersAvailableTo(entries, "")));

	onMount(async () => {
		leaveGuard.set({ isDirty: () => hasChanges, save });
		const [setup, disks] = await Promise.all([loadDriveSetup(), loadPartitions()]);
		if (setup === null) {
			loadFailed = true;
		} else {
			entries = setup;
			savedEntries = cloneEntries(setup);
		}
		partitions = disks;
		isLoading = false;
	});

	onDestroy(() => leaveGuard.set(null));

	function addDisk(partition: Partition) {
		if (!nextLetter) return;
		entries = [...entries, { letter: nextLetter, path: partition.mount }];
	}

	function addFolder() {
		const path = newFolder.trim();
		if (!path) {
			notifications.error("Pick a folder first.");
			return;
		}
		if (!folderLetter) {
			notifications.error("No drive letter is free.");
			return;
		}
		entries = [...entries, { letter: folderLetter, path }];
		newFolder = "";
		pickedLetter = "";
	}

	function focusFolderInput() {
		document.getElementById(FOLDER_INPUT_ID)?.focus();
	}

	async function save(): Promise<boolean> {
		if (loadFailed) {
			notifications.error("The saved setup could not be read. Reopen this page to try again.");
			return false;
		}
		if (entries.some((entry) => !entry.path.trim())) {
			notifications.error("Every drive needs a folder.");
			return false;
		}
		isSaving = true;
		const saved = await saveDriveSetup(cloneEntries(entries));
		isSaving = false;
		if (saved) savedEntries = cloneEntries(entries);
		return saved;
	}
</script>

<div class="drive-page">
	<PageHeader title="Drive setup" subtitle="Drives here are added to every prefix when a game starts." />

	{#if isLoading}
		<LoadingState label="Loading drives..." />
	{:else}
		<Section title="Detected disks" help="Pick a disk to add it as a drive.">
			<DiskList {partitions} {entries} {nextLetter} onAdd={addDisk} />
		</Section>

		<Section title="Drives for every prefix" help="Saved here, linked into each prefix on launch.">
			{#if entries.length === 0}
				<EmptyState icon="hard_drive" title="No drives set up" hint="Add a disk above or a folder below.">
					<Button size="sm" icon="create_new_folder" onclick={focusFolderInput}>Add folder</Button>
				</EmptyState>
			{:else}
				<SetupList bind:entries />
			{/if}

			<div class="add-folder">
				<h3>Add a folder</h3>
				<div class="add-row">
					<div class="folder-field">
						<BrowseInput id={FOLDER_INPUT_ID} label="Folder" bind:value={newFolder} />
					</div>
					<div class="letter-field">
						<Dropdown
							options={folderLetterOptions}
							value={folderLetter}
							onChange={(letter) => (pickedLetter = letter)}
						/>
					</div>
					<Button icon="add" onclick={addFolder}>Add</Button>
				</div>
			</div>
		</Section>

		<footer class="page-footer">
			<Button variant="primary" disabled={!hasChanges || loadFailed} loading={isSaving} onclick={save}>Save</Button>
		</footer>
	{/if}
</div>

<style lang="scss">
	.drive-page {
		display: flex;
		flex-direction: column;
		gap: 20px;
		padding-bottom: 24px;
	}

	.add-folder {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-top: 20px;
		padding-top: 20px;
		border-top: 2px solid rgba(255, 255, 255, 0.05);

		h3 {
			margin: 0;
			color: var(--text-main);
			font-size: 0.95rem;
			font-weight: 700;
		}
	}

	.add-row {
		display: flex;
		align-items: flex-end;
		gap: 12px;
	}

	.folder-field {
		flex: 1;
		min-width: 0;
	}

	.letter-field {
		flex: 0 0 130px;
	}

	.page-footer {
		display: flex;
		justify-content: flex-end;
	}
</style>
