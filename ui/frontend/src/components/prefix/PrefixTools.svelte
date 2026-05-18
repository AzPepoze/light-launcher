<script lang="ts">
	import ToolButton from "@components/shared/ToolButton.svelte";
	
	export let runningToolName: string = "";
	export let onRunTool: (tool: string) => void;
</script>

<div class="tools-grid">
	<ToolButton
		icon="settings"
		title="Winecfg"
		subtitle="Settings"
		loading={runningToolName === "winecfg"}
		onClick={() => onRunTool("winecfg")}
	/>
	<ToolButton
		icon="edit"
		title="Regedit"
		subtitle="Registry"
		loading={runningToolName === "regedit"}
		onClick={() => onRunTool("regedit")}
	/>
	<ToolButton
		icon="terminal"
		title="CMD"
		subtitle="Terminal"
		loading={runningToolName === "cmd"}
		onClick={() => onRunTool("cmd")}
	/>
	<ToolButton
		icon="auto_fix_high"
		title="Winetricks"
		subtitle="Extras"
		loading={runningToolName === "winetricks"}
		onClick={() => onRunTool("winetricks")}
	/>
	<ToolButton
		icon="bar_chart"
		title="TaskMgr"
		subtitle="Processes"
		loading={runningToolName === "taskmgr"}
		onClick={() => onRunTool("taskmgr")}
	/>
	<ToolButton
		icon="folder"
		title="Explorer"
		subtitle="Files"
		loading={runningToolName === "explorer"}
		onClick={() => onRunTool("explorer")}
	/>
</div>

<style lang="scss">
	.tools-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 20px;
		margin-bottom: 8px;
		flex-shrink: 0;
	}
</style>
