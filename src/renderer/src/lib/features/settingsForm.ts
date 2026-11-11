import type { JsonSchema, JsonSchemaProperty } from "./types";

/** How a field is drawn: a text box, a path with a browse button, a number, a switch, a select or a reorderable list. */
export type FieldKind = "text" | "path" | "number" | "boolean" | "select" | "list";
export type FieldOption = string | number;
export type PathPick = "folder" | "file";

/**
 * What the form holds for one field: text for text, path, number and select fields,
 * a boolean for a switch, and the item texts for a list.
 */
export type FieldInput = string | boolean | string[];
export type RawSettings = Record<string, FieldInput>;

export interface SettingsField {
	key: string;
	label: string;
	description: string;
	examples: string[];
	kind: FieldKind;
	/** The x-ll-group name, or "" when the field has no group. */
	group: string;
	advanced: boolean;
	required: boolean;
	integer: boolean;
	/** Allowed values: the enum of a select, or the enum of a list's items. */
	options: FieldOption[];
	min?: number;
	max?: number;
	minLength?: number;
	maxLength?: number;
	minItems?: number;
	maxItems?: number;
	unique: boolean;
	pick?: PathPick;
	defaultValue?: unknown;
}

export interface SettingsSection {
	/** The group name, or "Advanced" for the advanced section. Empty for fields without a group. */
	title: string;
	advanced: boolean;
	fields: SettingsField[];
}

export interface SettingsCheck {
	values: Record<string, unknown>;
	errors: Record<string, string>;
}

/** One validation problem reported by the server for one field. */
export interface ServerFieldError {
	field: string;
	message: string;
}

export type PickMethod = "pickFolder" | "pickFile";

/** The picker functions that window.llShell may provide. */
export interface ShellPickers {
	pickFolder?: () => Promise<string>;
	pickFile?: () => Promise<string>;
}

interface FieldRead {
	value?: unknown;
	error?: string;
}

const ADVANCED_TITLE = "Advanced";

function listKind(items: JsonSchemaProperty | undefined): FieldKind | null {
	if (items?.enum?.length || items?.type === "string") return "list";
	return null;
}

function fieldKind(property: JsonSchemaProperty): FieldKind | null {
	if (property.type === "array") return listKind(property.items);
	if (property.enum?.length) return "select";
	switch (property.type) {
		case "string":
			return property["x-ll-widget"] ? "path" : "text";
		case "number":
		case "integer":
			return "number";
		case "boolean":
			return "boolean";
		default:
			return null;
	}
}

function optionsOf(property: JsonSchemaProperty, kind: FieldKind): FieldOption[] {
	if (kind === "list") return property.items?.enum ?? [];
	return property.enum ?? [];
}

function fieldFrom(
	key: string,
	property: JsonSchemaProperty,
	kind: FieldKind,
	required: boolean
): SettingsField {
	return {
		key,
		label: property.title ?? key,
		description: property.description ?? "",
		examples: (property.examples ?? []).map(String),
		kind,
		group: property["x-ll-group"] ?? "",
		advanced: property["x-ll-advanced"] === true,
		required,
		integer: property.type === "integer",
		options: optionsOf(property, kind),
		min: property.minimum,
		max: property.maximum,
		minLength: property.minLength,
		maxLength: property.maxLength,
		minItems: property.minItems,
		maxItems: property.maxItems,
		unique: property.uniqueItems === true,
		pick: property["x-ll-widget"],
		defaultValue: property.default
	};
}

/** Turns the top-level properties of a settings schema into form fields. Unsupported types are skipped. */
export function buildSettingsFields(schema: JsonSchema | null): SettingsField[] {
	if (!schema?.properties) return [];
	const required = new Set(schema.required ?? []);
	const fields: SettingsField[] = [];
	for (const [key, property] of Object.entries(schema.properties)) {
		const kind = fieldKind(property);
		if (kind) fields.push(fieldFrom(key, property, kind, required.has(key)));
	}
	return fields;
}

/** Groups fields by x-ll-group in the order the groups first appear. Advanced fields go to one last section. */
export function buildSettingsSections(fields: SettingsField[]): SettingsSection[] {
	const sections: SettingsSection[] = [];
	const byGroup = new Map<string, SettingsSection>();
	const advanced: SettingsField[] = [];
	for (const field of fields) {
		if (field.advanced) {
			advanced.push(field);
			continue;
		}
		let section = byGroup.get(field.group);
		if (!section) {
			section = { title: field.group, advanced: false, fields: [] };
			byGroup.set(field.group, section);
			sections.push(section);
		}
		section.fields.push(field);
	}
	if (advanced.length > 0) {
		sections.push({ title: ADVANCED_TITLE, advanced: true, fields: advanced });
	}
	return sections;
}

function inputFromValue(field: SettingsField, value: unknown): FieldInput {
	switch (field.kind) {
		case "boolean":
			return value === true;
		case "list":
			return Array.isArray(value) ? value.map(String) : [];
		default:
			return value === undefined || value === null ? "" : String(value);
	}
}

/** The form state for saved values, with the schema default used for any value that is not saved. */
export function rawSettings(fields: SettingsField[], saved: Record<string, unknown>): RawSettings {
	const raw: RawSettings = {};
	for (const field of fields) {
		raw[field.key] = inputFromValue(field, saved[field.key] ?? field.defaultValue);
	}
	return raw;
}

/** The form state that holds the schema default, or undefined when the field has no default. */
export function defaultInput(field: SettingsField): FieldInput | undefined {
	if (field.defaultValue === undefined) return undefined;
	return inputFromValue(field, field.defaultValue);
}

export function sameInput(a: FieldInput | undefined, b: FieldInput | undefined): boolean {
	if (Array.isArray(a) && Array.isArray(b)) {
		return a.length === b.length && a.every((item, index) => item === b[index]);
	}
	return a === b;
}

/** True when the field has a default and the form holds something else. */
export function canResetToDefault(field: SettingsField, raw: RawSettings): boolean {
	const fallback = defaultInput(field);
	return fallback !== undefined && !sameInput(raw[field.key], fallback);
}

/** The keys of the fields whose form state differs from the baseline. */
export function changedKeys(
	fields: SettingsField[],
	raw: RawSettings,
	baseline: RawSettings
): string[] {
	return fields
		.filter((field) => !sameInput(raw[field.key], baseline[field.key]))
		.map((field) => field.key);
}

export function cloneRaw(raw: RawSettings): RawSettings {
	const copy: RawSettings = {};
	for (const [key, input] of Object.entries(raw)) {
		copy[key] = Array.isArray(input) ? [...input] : input;
	}
	return copy;
}

function asText(input: FieldInput | undefined): string {
	return typeof input === "string" ? input.trim() : "";
}

function matchOption(options: FieldOption[], text: string): FieldOption | undefined {
	return options.find((option) => String(option) === text);
}

function readText(field: SettingsField, text: string): FieldRead {
	if (text === "") return field.required ? { error: "Required" } : { value: "" };
	const length = [...text].length;
	if (field.minLength !== undefined && length < field.minLength) {
		return { error: `Must be at least ${field.minLength} characters` };
	}
	if (field.maxLength !== undefined && length > field.maxLength) {
		return { error: `Must be at most ${field.maxLength} characters` };
	}
	return { value: text };
}

function readNumber(field: SettingsField, text: string): FieldRead {
	if (text === "") return field.required ? { error: "Required" } : {};
	const value = Number(text);
	if (!Number.isFinite(value)) return { error: "Enter a number" };
	if (field.integer && !Number.isInteger(value)) return { error: "Enter a whole number" };
	if (field.min !== undefined && value < field.min) {
		return { error: `Must be at least ${field.min}` };
	}
	if (field.max !== undefined && value > field.max) {
		return { error: `Must be at most ${field.max}` };
	}
	return { value };
}

function readSelect(field: SettingsField, text: string): FieldRead {
	const option = matchOption(field.options, text);
	if (option !== undefined) return { value: option };
	if (text !== "") return { error: "Choose an option" };
	return field.required ? { error: "Required" } : {};
}

function readList(field: SettingsField, texts: string[]): FieldRead {
	const items = texts.map((text) => text.trim()).filter((text) => text !== "");
	if (items.length === 0 && field.required) return { error: "Add at least one item" };
	if (field.minItems !== undefined && items.length < field.minItems) {
		return { error: `Add at least ${field.minItems} items` };
	}
	if (field.maxItems !== undefined && items.length > field.maxItems) {
		return { error: `Use at most ${field.maxItems} items` };
	}
	if (field.unique && new Set(items).size !== items.length) {
		return { error: "Remove the duplicate items" };
	}
	const values: FieldOption[] = [];
	for (const item of items) {
		const value = field.options.length > 0 ? matchOption(field.options, item) : item;
		if (value === undefined) return { error: "Choose only listed items" };
		values.push(value);
	}
	return { value: values };
}

function readField(field: SettingsField, input: FieldInput | undefined): FieldRead {
	switch (field.kind) {
		case "boolean":
			return { value: input === true };
		case "list":
			return readList(field, Array.isArray(input) ? input : []);
		case "number":
			return readNumber(field, asText(input));
		case "select":
			return readSelect(field, asText(input));
		default:
			return readText(field, asText(input));
	}
}

/** Turns the form state into plugin settings, or one error message per invalid field. */
export function collectSettings(fields: SettingsField[], raw: RawSettings): SettingsCheck {
	const values: Record<string, unknown> = {};
	const errors: Record<string, string> = {};
	for (const field of fields) {
		const read = readField(field, raw[field.key]);
		if (read.error) {
			errors[field.key] = read.error;
		} else if (read.value !== undefined) {
			values[field.key] = read.value;
		}
	}
	return { values, errors };
}

/** Lowers the first letter, so a server message reads as a form hint. */
function serverMessageText(message: string): string {
	if (message === "is required") return "Required";
	if (message === "is not a known setting") return "Not a known setting";
	return message.charAt(0).toUpperCase() + message.slice(1);
}

/** Maps the server's field errors to the form's error messages. */
export function serverErrors(problems: ServerFieldError[]): Record<string, string> {
	const errors: Record<string, string> = {};
	for (const problem of problems) {
		errors[problem.field] = serverMessageText(problem.message);
	}
	return errors;
}

/** The picker to open for a path field, or undefined when the shell has none (the text box is used then). */
export function pickMethodFor(
	field: SettingsField,
	shell: ShellPickers | undefined
): PickMethod | undefined {
	if (field.kind !== "path" || !field.pick) return undefined;
	const method: PickMethod = field.pick === "folder" ? "pickFolder" : "pickFile";
	return shell?.[method] ? method : undefined;
}

/** Moves the item at index one place up (-1) or down (1). Items at the edge stay where they are. */
export function moveItem(items: string[], index: number, step: -1 | 1): string[] {
	const target = index + step;
	if (target < 0 || target >= items.length) return items;
	const next = [...items];
	next[index] = items[target];
	next[target] = items[index];
	return next;
}

export function removeItem(items: string[], index: number): string[] {
	return items.filter((_, position) => position !== index);
}

/** Adds the trimmed text as the last item. Blank text adds nothing. */
export function addItem(items: string[], text: string): string[] {
	const trimmed = text.trim();
	return trimmed === "" ? items : [...items, trimmed];
}

/** The enum values a list can still take. */
export function unusedOptions(field: SettingsField, items: string[]): string[] {
	return field.options.map(String).filter((option) => !items.includes(option));
}
