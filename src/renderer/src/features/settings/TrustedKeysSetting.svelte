<script lang="ts">
	import { onMount } from "svelte";
	import Button from "@components/shared/Button.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import Section from "@components/shared/Section.svelte";
	import SettingRow from "@components/shared/SettingRow.svelte";
	import { pluginsApi, type TrustedKey } from "@lib/features/pluginsApi";
	import { errorMessage } from "@lib/features/rpc";
	import { notifications } from "@stores/notificationStore";
	import { fingerprintLabel, trustedKeyInputError } from "./trustedKeys";

	let keys = $state<TrustedKey[]>([]);
	let loading = $state(true);
	let loadFailed = $state(false);
	let draft = $state("");
	let addError = $state("");
	let adding = $state(false);
	let confirming = $state<string | null>(null);
	let removing = $state<string | null>(null);

	async function load(): Promise<void> {
		loading = true;
		loadFailed = false;
		try {
			keys = await pluginsApi.trustedKeys();
		} catch {
			loadFailed = true;
		} finally {
			loading = false;
		}
	}

	async function addKey(): Promise<void> {
		const problem = trustedKeyInputError(draft);
		addError = problem ?? "";
		if (problem) return;
		adding = true;
		try {
			keys = await pluginsApi.addTrustedKey(draft.trim());
			draft = "";
			notifications.success("Trusted key added");
		} catch (err) {
			notifications.error(errorMessage(err));
		} finally {
			adding = false;
		}
	}

	async function removeKey(key: TrustedKey): Promise<void> {
		removing = key.base64;
		try {
			keys = await pluginsApi.removeTrustedKey(key.base64);
			confirming = null;
			notifications.success("Trusted key removed");
		} catch (err) {
			notifications.error(errorMessage(err));
		} finally {
			removing = null;
		}
	}

	function submitKey(event: SubmitEvent): void {
		event.preventDefault();
		void addKey();
	}

	onMount(() => {
		void load();
	});
</script>

<Section
	title="Trusted plugin keys"
	icon="key"
	help="Packages signed by one of these keys install without the unsigned warning."
>
	<Notice tone="warning" icon="gpp_maybe">
		<p>
			Signing proves a package came from this key. It does not prove the code is safe, so only add
			keys you trust.
		</p>
	</Notice>

	{#if loadFailed}
		<Notice tone="error" icon="error">
			<p>Could not load the trusted keys.</p>
		</Notice>
		<div class="retry">
			<Button variant="secondary" icon="refresh" onclick={() => void load()}>Try again</Button>
		</div>
	{:else if loading}
		<LoadingState label="Loading trusted keys..." />
	{:else if keys.length === 0}
		<p class="empty">No trusted keys yet. Unsigned packages still ask before they install.</p>
	{:else}
		<div class="keys">
			{#each keys as key (key.base64)}
				<SettingRow title={fingerprintLabel(key.fingerprint)} help="ed25519 public key">
					{#if confirming === key.base64}
						<div class="confirm">
							<Button
								variant="danger"
								size="sm"
								loading={removing === key.base64}
								onclick={() => void removeKey(key)}
							>
								Confirm remove
							</Button>
							<Button
								size="sm"
								disabled={removing === key.base64}
								onclick={() => (confirming = null)}
							>
								Cancel
							</Button>
						</div>
					{:else}
						<Button size="sm" icon="delete" onclick={() => (confirming = key.base64)}>
							Remove
						</Button>
					{/if}
				</SettingRow>
			{/each}
		</div>
	{/if}

	<SettingRow
		title="Add a trusted key"
		help="Paste the base64 ed25519 public key printed by llp keygen."
	>
		<form class="add" onsubmit={submitKey}>
			<input
				class="input"
				type="text"
				bind:value={draft}
				oninput={() => (addError = "")}
				placeholder="Base64 public key"
				aria-label="Base64 ed25519 public key"
				aria-invalid={Boolean(addError)}
				aria-describedby={addError ? "trusted-key-add-error" : undefined}
				spellcheck="false"
				autocomplete="off"
				disabled={loading}
			/>
			<Button
				type="submit"
				size="sm"
				icon="add"
				loading={adding}
				disabled={loading || !draft.trim()}
			>
				Add
			</Button>
		</form>
	</SettingRow>

	{#if addError}
		<p class="error" id="trusted-key-add-error" role="alert">{addError}</p>
	{/if}
</Section>

<style lang="scss">
	.keys {
		display: flex;
		flex-direction: column;
	}

	.empty {
		margin: 8px 0;
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.add {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 0;
		margin: 0;

		.input {
			width: clamp(220px, 26vw, 380px);
			min-width: 0;
			font-size: 0.85rem;
		}
	}

	.confirm {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.retry {
		margin-top: 4px;
	}

	.error {
		margin: 8px 0 0;
		color: var(--danger);
		font-size: 0.85rem;
	}

	p {
		margin: 0;
	}
</style>
