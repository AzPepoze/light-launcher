import type { SkipReason, SkippedUpload } from "../../lib/features/pluginUpload";

export const LEGAL_NOTICE = "Use only accounts you own and content you are allowed to get.";

export const UNSIGNED_WARNING = "This plugin is not signed. Only install code you trust.";

/** Empty until a community plugin list exists. The link shows as plain text while empty. */
export const COMMUNITY_PLUGINS_URL = "";

export const PERMISSION_HINT = "The plugin will be able to:";

export const SITES_HEADING = "Connect to these sites:";

export const UPDATE_PERMISSION_HINT = "This update asks for new permissions:";

export const NO_NEW_PERMISSIONS = "This update asks for no new permissions.";

export const REPOSITORY_CHANGED_WARNING =
	"This update names a different repository than the installed version.";

export const CHECK_UPDATES_LABEL = "Check for updates";

export const CHECKING_UPDATES_LABEL = "Checking...";

export const SLEEPING_BADGE = "Sleeping";

export const DROP_HINT = "Drop plugin zips or folders here";

export const DROP_OVERLAY_TEXT = "Drop to check the plugins";

export const CHOOSE_FILE_LABEL = "Choose files";

export const CHOOSE_FOLDER_LABEL = "Choose folder";

export const UPLOADING_TEXT = "Uploading plugin...";

export const DISMISS_LABEL = "Dismiss";

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
