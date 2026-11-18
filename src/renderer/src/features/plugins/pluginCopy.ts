import type { PluginStatusKind } from "../../lib/features/pluginStatus";
import type { SkipReason, SkippedUpload } from "../../lib/features/pluginUpload";

export const LEGAL_NOTICE = "Use only accounts you own and content you are allowed to get.";

export const UNSIGNED_WARNING = "This plugin is not signed. Only install code you trust.";

export const COMMUNITY_TITLE = "Community plugins";

export const COMMUNITY_OPEN_LABEL = "Community";

export const COMMUNITY_BROWSE_LABEL = "Browse community plugins";

export const COMMUNITY_LOADING_LABEL = "Loading community plugins...";

export const COMMUNITY_NOT_CONFIGURED_COPY = "Community list is not configured.";

export const COMMUNITY_NOT_CONFIGURED_HINT =
	"Set pluginRegistryURL in settings.json to the https address of a community index.";

export const COMMUNITY_LOAD_FAILED_COPY = "Community list could not be loaded.";

export const COMMUNITY_RETRY_LABEL = "Retry";

export const COMMUNITY_NOTHING_AVAILABLE_COPY = "No community plugins are available to install.";

export const COMMUNITY_INSTALL_LABEL = "Install";

export const COMMUNITY_INSTALL_HINT =
	"Installing opens the usual permission check before anything is added.";

export const COMMUNITY_UNKNOWN_AUTHOR = "Unknown author";

export const COMMUNITY_HOMEPAGE_LABEL = "Homepage";

export const PERMISSION_HINT = "The plugin will be able to:";

export function sitesTitle(count: number): string {
	return count === 1 ? "Connects to 1 site" : `Connects to ${count} sites`;
}

export const UPDATE_PERMISSION_HINT = "This update asks for new permissions:";

export const NO_NEW_PERMISSIONS = "This update asks for no new permissions.";

export const REPOSITORY_CHANGED_WARNING =
	"This update names a different repository than the installed version.";

export const CHECK_UPDATES_LABEL = "Check for updates";

export const CHECKING_UPDATES_LABEL = "Checking...";

export const SLEEPING_BADGE = "Sleeping";

export const UNSIGNED_BADGE = "Unsigned";

export const UNSIGNED_TOOLTIP = "This plugin is not signed by its author";

export const STATUS_COPY: Record<PluginStatusKind, string> = {
	error: "Error",
	disabled: "Disabled",
	sleeping: SLEEPING_BADGE,
	unsigned: UNSIGNED_BADGE,
	enabled: "Enabled"
};

export const DROP_HINT = "Drop plugin zips or folders here";

export const DROP_OVERLAY_TEXT = "Drop to check the plugins";

export const CHOOSE_FILE_LABEL = "Choose files";

export const CHOOSE_FOLDER_LABEL = "Choose folder";

export const UPLOADING_TEXT = "Uploading plugin...";

export const SKIPPED_HEADING = "These items were skipped:";

export const NOTHING_TO_INSTALL_COPY =
	"Drop a .llplugin or .zip file, or a folder that has plugin.json.";

export const QUEUE_BUSY_COPY = "Finish installing the plugins you dropped first.";

const SKIP_REASON_COPY: Record<SkipReason, string> = {
	"not-a-plugin": "Not a plugin. Needs a .llplugin or .zip file, or a folder with plugin.json.",
	"nested-plugins": "Holds several plugins. Drop each plugin folder on its own."
};

export function skippedText(item: SkippedUpload): string {
	return `${item.name}: ${SKIP_REASON_COPY[item.reason]}`;
}

export function pluginProgress(position: number, total: number): string {
	return `Plugin ${position} of ${total}`;
}

export function installedSummary(installed: number, total: number): string {
	return `Installed ${installed} of ${total} plugins`;
}

export const DEV_BADGE = "Dev";

export const LINK_FOLDER_LABEL = "Link folder";

export const LINK_FOLDER_HINT =
	"Runs a plugin you are building from its folder. Restarts on save. Not signed.";

export const DEV_WARNING =
	"This plugin runs from a linked folder. It is not signed, and its files are not checked.";

export const UNLINK_LABEL = "Unlink";

export function linkedSummary(name: string): string {
	return `Linked ${name} from its folder`;
}

export function unlinkedSummary(name: string): string {
	return `Unlinked ${name}. The installed copy runs again, if there is one.`;
}
