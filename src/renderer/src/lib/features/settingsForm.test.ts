import { describe, expect, it } from "vitest";
import type { JsonSchema } from "./types";
import {
	addItem,
	buildSettingsFields,
	buildSettingsSections,
	canResetToDefault,
	changedKeys,
	cloneRaw,
	collectSettings,
	moveItem,
	pickMethodFor,
	rawSettings,
	removeItem,
	serverErrors,
	unusedOptions,
	type RawSettings,
	type SettingsField
} from "./settingsForm";

const schema: JsonSchema = {
	type: "object",
	required: ["username", "retries"],
	properties: {
		username: { type: "string", title: "User name", description: "Shown in logs" },
		retries: { type: "integer", title: "Retries", minimum: 0, maximum: 5, default: 2 },
		speed: { type: "number", minimum: 0.5 },
		region: { type: "string", enum: ["eu", "us"], title: "Region" },
		verbose: { type: "boolean", default: true },
		nested: { type: "object" },
		folder: { type: "string", "x-ll-widget": "folder", "x-ll-group": "Files" },
		file: { type: "string", "x-ll-widget": "file", "x-ll-advanced": true },
		priorities: {
			type: "array",
			items: { type: "string", enum: ["a", "b", "c"] },
			uniqueItems: true,
			minItems: 1,
			maxItems: 3,
			"x-ll-group": "Files"
		},
		tags: { type: "array", items: { type: "string" }, default: ["x"] },
		label: {
			type: "string",
			minLength: 2,
			maxLength: 4,
			examples: ["abc", "de"],
			"x-ll-group": "Network"
		},
		proxy: { type: "string", "x-ll-advanced": true, "x-ll-group": "Network" }
	}
};

const fields = buildSettingsFields(schema);
const byKey = (key: string): SettingsField => {
	const field = fields.find((item) => item.key === key);
	if (!field) throw new Error(`no field ${key}`);
	return field;
};

/** A valid form state for every field, so a test can change one input at a time. */
function validRaw(): RawSettings {
	return {
		username: "ada",
		retries: "2",
		speed: "",
		region: "us",
		verbose: true,
		folder: "/data",
		file: "",
		priorities: ["a", "b"],
		tags: ["x"],
		label: "abc",
		proxy: ""
	};
}

describe("buildSettingsFields", () => {
	it("maps supported types and skips unsupported ones", () => {
		expect(fields.map((field) => [field.key, field.kind])).toEqual([
			["username", "text"],
			["retries", "number"],
			["speed", "number"],
			["region", "select"],
			["verbose", "boolean"],
			["folder", "path"],
			["file", "path"],
			["priorities", "list"],
			["tags", "list"],
			["label", "text"],
			["proxy", "text"]
		]);
	});

	it("reads titles, descriptions, examples and rules", () => {
		expect(byKey("username")).toMatchObject({
			label: "User name",
			description: "Shown in logs",
			required: true
		});
		expect(byKey("retries")).toMatchObject({ integer: true, min: 0, max: 5, required: true });
		expect(byKey("label")).toMatchObject({
			minLength: 2,
			maxLength: 4,
			examples: ["abc", "de"]
		});
	});

	it("reads the x-ll extensions", () => {
		expect(byKey("folder")).toMatchObject({ pick: "folder", group: "Files", advanced: false });
		expect(byKey("file")).toMatchObject({ pick: "file", advanced: true });
		expect(byKey("priorities")).toMatchObject({
			options: ["a", "b", "c"],
			unique: true,
			minItems: 1,
			maxItems: 3
		});
	});

	it("returns no fields without a schema", () => {
		expect(buildSettingsFields(null)).toEqual([]);
	});
});

describe("buildSettingsSections", () => {
	it("groups fields in first-appearance order and puts advanced fields last", () => {
		const sections = buildSettingsSections(fields);
		expect(sections.map((section) => [section.title, section.advanced])).toEqual([
			["", false],
			["Files", false],
			["Network", false],
			["Advanced", true]
		]);
		const advanced = sections[sections.length - 1];
		expect(advanced.fields.map((field) => field.key)).toEqual(["file", "proxy"]);
	});
});

describe("nested objects", () => {
	const nestedSchema: JsonSchema = {
		type: "object",
		required: ["username"],
		properties: {
			username: { type: "string" },
			proxy: {
				type: "object",
				title: "Proxy",
				required: ["host"],
				properties: {
					host: { type: "string", title: "Host" },
					port: { type: "integer", default: 8080 },
					auth: {
						type: "object",
						title: "Auth",
						properties: { token: { type: "string", "x-ll-advanced": true } }
					}
				}
			},
			limits: { type: "object", properties: { max: { type: "number" } } },
			unused: { type: "object" }
		}
	};
	const nestedFields = buildSettingsFields(nestedSchema);

	it("flattens nested fields with dotted keys and their object heading", () => {
		expect(nestedFields.map((field) => [field.key, field.path.join("/"), field.group])).toEqual([
			["username", "username", ""],
			["proxy.host", "proxy/host", "Proxy"],
			["proxy.port", "proxy/port", "Proxy"],
			["proxy.auth.token", "proxy/auth/token", "Auth"],
			["limits.max", "limits/max", "limits"]
		]);
	});

	it("puts nested fields under their object heading", () => {
		const sections = buildSettingsSections(nestedFields);
		expect(sections.map((section) => [section.title, section.advanced])).toEqual([
			["", false],
			["Proxy", false],
			["limits", false],
			["Advanced", true]
		]);
		expect(sections[3].fields.map((field) => field.key)).toEqual(["proxy.auth.token"]);
	});

	it("reads and writes nested saved values", () => {
		const saved = {
			username: "ada",
			proxy: { host: "localhost", port: 9000, auth: { token: "t" } },
			limits: { max: 5 }
		};
		const raw = rawSettings(nestedFields, saved);
		expect(raw).toEqual({
			username: "ada",
			"proxy.host": "localhost",
			"proxy.port": "9000",
			"proxy.auth.token": "t",
			"limits.max": "5"
		});

		const check = collectSettings(nestedFields, raw);
		expect(check.errors).toEqual({});
		expect(check.values).toEqual({
			username: "ada",
			proxy: { host: "localhost", port: 9000, auth: { token: "t" } },
			limits: { max: 5 }
		});
	});

	it("uses a nested field default and reports its required error by dotted key", () => {
		const raw = rawSettings(nestedFields, {});
		expect(raw["proxy.port"]).toBe("8080");
		const check = collectSettings(nestedFields, { ...raw, username: "ada", "proxy.host": "" });
		expect(check.errors).toEqual({ "proxy.host": "Required" });
	});
});

describe("rawSettings", () => {
	it("prefers saved values, then schema defaults", () => {
		const raw = rawSettings(fields, { username: "ada", speed: 1.5, priorities: ["c"] });
		expect(raw).toMatchObject({
			username: "ada",
			retries: "2",
			speed: "1.5",
			verbose: true,
			priorities: ["c"],
			tags: ["x"]
		});
	});
});

describe("collectSettings", () => {
	it("returns typed values when everything is valid", () => {
		const raw = { ...validRaw(), username: " ada ", retries: "3", region: "us", verbose: false };
		const result = collectSettings(fields, raw);
		expect(result.errors).toEqual({});
		expect(result.values).toMatchObject({
			username: "ada",
			retries: 3,
			region: "us",
			verbose: false,
			priorities: ["a", "b"],
			tags: ["x"],
			folder: "/data"
		});
		expect(result.values).not.toHaveProperty("speed");
	});

	it("reports required and out of range values", () => {
		const result = collectSettings(fields, {
			...validRaw(),
			username: "",
			retries: "9",
			speed: "0.1"
		});
		expect(result.errors).toMatchObject({
			username: "Required",
			retries: "Must be at most 5",
			speed: "Must be at least 0.5"
		});
	});

	it("rejects non numbers and fractions for integer fields", () => {
		expect(collectSettings(fields, { ...validRaw(), retries: "abc" }).errors.retries).toBe(
			"Enter a number"
		);
		expect(collectSettings(fields, { ...validRaw(), retries: "1.5" }).errors.retries).toBe(
			"Enter a whole number"
		);
	});

	it("rejects a select value that is not an option", () => {
		expect(collectSettings(fields, { ...validRaw(), region: "asia" }).errors.region).toBe(
			"Choose an option"
		);
	});

	it("checks list size, duplicates and allowed items", () => {
		const empty = collectSettings(fields, { ...validRaw(), priorities: [] });
		expect(empty.errors.priorities).toBe("Add at least 1 items");

		const duplicate = collectSettings(fields, { ...validRaw(), priorities: ["a", "a"] });
		expect(duplicate.errors.priorities).toBe("Remove the duplicate items");

		const unknown = collectSettings(fields, { ...validRaw(), priorities: ["z"] });
		expect(unknown.errors.priorities).toBe("Choose only listed items");

		const tooMany = collectSettings(fields, { ...validRaw(), priorities: ["a", "b", "c", "a"] });
		expect(tooMany.errors.priorities).toBeDefined();
	});

	it("drops blank list items", () => {
		const result = collectSettings(fields, { ...validRaw(), tags: ["  ", "x"] });
		expect(result.values.tags).toEqual(["x"]);
	});

	it("checks text length", () => {
		expect(collectSettings(fields, { ...validRaw(), label: "a" }).errors.label).toBe(
			"Must be at least 2 characters"
		);
		expect(collectSettings(fields, { ...validRaw(), label: "abcde" }).errors.label).toBe(
			"Must be at most 4 characters"
		);
	});
});

describe("dirty state and defaults", () => {
	it("lists only the fields whose form state changed", () => {
		const baseline = validRaw();
		const raw = { ...cloneRaw(baseline), username: "bob", priorities: ["b", "a"] };
		expect(changedKeys(fields, raw, baseline)).toEqual(["username", "priorities"]);
		expect(changedKeys(fields, cloneRaw(baseline), baseline)).toEqual([]);
	});

	it("cloneRaw copies lists, so edits do not reach the baseline", () => {
		const baseline = validRaw();
		const copy = cloneRaw(baseline);
		copy.priorities = removeItem(copy.priorities as string[], 0);
		expect(baseline.priorities).toEqual(["a", "b"]);
	});

	it("offers a reset only when the field has a default and differs from it", () => {
		const raw = validRaw();
		expect(canResetToDefault(byKey("retries"), { ...raw, retries: "4" })).toBe(true);
		expect(canResetToDefault(byKey("retries"), raw)).toBe(false);
		expect(canResetToDefault(byKey("speed"), raw)).toBe(false);
	});
});

describe("serverErrors", () => {
	it("maps each field to a readable message", () => {
		expect(
			serverErrors([
				{ field: "region", message: "must be one of: eu, us" },
				{ field: "username", message: "is required" },
				{ field: "extra", message: "is not a known setting" },
				{ field: "priorities", message: "item 2 must be one of: a, b" }
			])
		).toEqual({
			region: "Must be one of: eu, us",
			username: "Required",
			extra: "Not a known setting",
			priorities: "Item 2 must be one of: a, b"
		});
	});
});

describe("pickMethodFor", () => {
	it("uses the shell picker that matches the widget", () => {
		const shell = { pickFolder: async () => "/a", pickFile: async () => "/b" };
		expect(pickMethodFor(byKey("folder"), shell)).toBe("pickFolder");
		expect(pickMethodFor(byKey("file"), shell)).toBe("pickFile");
	});

	it("has no picker without a shell, or for fields that are not paths", () => {
		expect(pickMethodFor(byKey("folder"), undefined)).toBeUndefined();
		expect(pickMethodFor(byKey("folder"), { pickFile: async () => "" })).toBeUndefined();
		expect(pickMethodFor(byKey("username"), { pickFolder: async () => "" })).toBeUndefined();
	});
});

describe("list edits", () => {
	it("moves an item up and down, and keeps edges in place", () => {
		expect(moveItem(["a", "b", "c"], 1, -1)).toEqual(["b", "a", "c"]);
		expect(moveItem(["a", "b", "c"], 1, 1)).toEqual(["a", "c", "b"]);
		expect(moveItem(["a", "b"], 0, -1)).toEqual(["a", "b"]);
		expect(moveItem(["a", "b"], 1, 1)).toEqual(["a", "b"]);
	});

	it("removes an item and adds trimmed text", () => {
		expect(removeItem(["a", "b", "c"], 1)).toEqual(["a", "c"]);
		expect(addItem(["a"], "  b ")).toEqual(["a", "b"]);
		expect(addItem(["a"], "   ")).toEqual(["a"]);
	});

	it("lists the enum values that are not yet used", () => {
		expect(unusedOptions(byKey("priorities"), ["b"])).toEqual(["a", "c"]);
	});
});
