<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import { tooltip } from "@components/shared/tooltip";
	import { UNSIGNED_BADGE, UNSIGNED_TOOLTIP } from "./pluginCopy";
</script>

<span use:tooltip={UNSIGNED_TOOLTIP}>
	<Badge tone="warning" icon="gpp_maybe">{UNSIGNED_BADGE}</Badge>
</span>
