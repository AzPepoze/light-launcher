import type { JsonSchema, JsonSchemaProperty } from "../../lib/features/types";

export type FieldKind = "text" | "number" | "boolean" | "select";

export interface SchemaField {
	key: string;
	label: string;
	description: string;
	kind: FieldKind;
	integer: boolean;
	required: boolean;
	options: (string | number)[];
	min?: number;
	max?: number;
	defaultValue?: string | number | boolean;
}

/** Raw form state: text for text, number and select fields; a boolean for checkboxes. */
export type RawValues = Record<string, string | boolean>;

export interface CollectResult {
	values: Record<string, unknown>;
	errors: Record<string, string>;
}

function fieldKind(property: JsonSchemaProperty): FieldKind | null {
	if (property.enum?.length) return "select";
	switch (property.type) {
		case "string":
			return "text";
		case "number":
		case "integer":
			return "number";
		case "boolean":
			return "boolean";
		default:
			return null;
	}
}

export function buildSchemaFields(schema: JsonSchema | null): SchemaField[] {
	if (!schema?.properties) return [];
	const required = new Set(schema.required ?? []);
	const fields: SchemaField[] = [];
	for (const [key, property] of Object.entries(schema.properties)) {
		const kind = fieldKind(property);
		if (!kind) continue;
		fields.push({
			key,
			label: property.title ?? key,
			description: property.description ?? "",
			kind,
			integer: property.type === "integer",
			required: required.has(key),
			options: property.enum ?? [],
			min: property.minimum,
			max: property.maximum,
			defaultValue: property.default
		});
	}
	return fields;
}

function rawValue(field: SchemaField, stored: Record<string, unknown>): string | boolean {
	const value = stored[field.key] ?? field.defaultValue;
	if (field.kind === "boolean") return value === true;
	return value === undefined || value === null ? "" : String(value);
}

export function rawFormValues(fields: SchemaField[], stored: Record<string, unknown>): RawValues {
	const raw: RawValues = {};
	for (const field of fields) raw[field.key] = rawValue(field, stored);
	return raw;
}

function parseNumber(field: SchemaField, text: string): { value?: number; error?: string } {
	if (text === "") return field.required ? { error: "Required" } : {};
	const value = Number(text);
	if (!Number.isFinite(value)) return { error: "Enter a number" };
	if (field.integer && !Number.isInteger(value)) return { error: "Enter a whole number" };
	if (field.min !== undefined && value < field.min)
		return { error: `Must be at least ${field.min}` };
	if (field.max !== undefined && value > field.max)
		return { error: `Must be at most ${field.max}` };
	return { value };
}

function parseSelect(
	field: SchemaField,
	text: string
): { value?: string | number; error?: string } {
	const option = field.options.find((item) => String(item) === text);
	if (option !== undefined) return { value: option };
	if (text === "" && !field.required) return {};
	return { error: "Choose an option" };
}

function parseText(field: SchemaField, text: string): { value?: string; error?: string } {
	if (text === "" && field.required) return { error: "Required" };
	return { value: text };
}

function parseField(field: SchemaField, text: string): { value?: unknown; error?: string } {
	switch (field.kind) {
		case "number":
			return parseNumber(field, text);
		case "select":
			return parseSelect(field, text);
		default:
			return parseText(field, text);
	}
}

/** Turns raw form state into typed plugin settings, or per-field error messages. */
export function collectValues(fields: SchemaField[], raw: RawValues): CollectResult {
	const values: Record<string, unknown> = {};
	const errors: Record<string, string> = {};

	for (const field of fields) {
		const input = raw[field.key];
		if (field.kind === "boolean") {
			values[field.key] = input === true;
			continue;
		}
		const text = typeof input === "string" ? input.trim() : "";
		const parsed = parseField(field, text);
		if (parsed.error) {
			errors[field.key] = parsed.error;
		} else if (parsed.value !== undefined) {
			values[field.key] = parsed.value;
		}
	}

	return { values, errors };
}
