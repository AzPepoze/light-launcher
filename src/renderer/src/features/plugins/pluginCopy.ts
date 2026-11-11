import type { UploadProblem } from "../../lib/features/pluginUpload";

export const LEGAL_NOTICE = "Use only accounts you own and content you are allowed to get.";

export const UNSIGNED_WARNING = "This plugin is not signed. Only install code you trust.";

/** Empty until a community plugin list exists. The link shows as plain text while empty. */
export const COMMUNITY_PLUGINS_URL = "";

export const PERMISSION_HINT = "The plugin will be able to:";

export const UPDATE_PERMISSION_HINT = "This update asks for new permissions:";

export const NO_NEW_PERMISSIONS = "This update asks for no new permissions.";

export const DROP_HINT = "Drop a plugin zip or folder here";

export const DROP_OVERLAY_TEXT = "Drop to check the plugin";

export const CHOOSE_FILE_LABEL = "Choose file";

export const CHOOSE_FOLDER_LABEL = "Choose folder";

export const UPLOADING_TEXT = "Uploading plugin...";

export const UPLOAD_PROBLEM_COPY: Record<UploadProblem, string> = {
	"not-a-plugin": "Drop a .llplugin or .zip file, or a folder that has plugin.json.",
	"several-plugins": "Drop one plugin at a time."
};
