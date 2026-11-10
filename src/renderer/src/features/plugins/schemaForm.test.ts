import { describe, expect, it } from "vitest";
import type { JsonSchema } from "../../lib/features/types";
import { buildSchemaFields, collectValues, rawFormValues } from "./schemaForm";

const schema: JsonSchema = {
	type: "object",
	required: ["username", "retries"],
	properties: {
		username: { type: "string", title: "User name", description: "Shown in logs" },
		retries: { type: "integer", title: "Retries", minimum: 0, maximum: 5, default: 2 },
		speed: { type: "number", minimum: 0.5 },
		region: { type: "string", enum: ["eu", "us"], title: "Region" },
		verbose: { type: "boolean", default: true },
		nested: { type: "object" }
	}
};

describe("buildSchemaFields", () => {
	it("maps supported property types and skips unsupported ones", () => {
		const fields = buildSchemaFields(schema);
		expect(fields.map((field) => [field.key, field.kind])).toEqual([
			["username", "text"],
			["retries", "number"],
			["speed", "number"],
			["region", "select"],
			["verbose", "boolean"]
		]);
	});

	it("reads labels, required flags and integer rules", () => {
		const [username, retries] = buildSchemaFields(schema);
		expect(username).toMatchObject({
			label: "User name",
			description: "Shown in logs",
			required: true
		});
		expect(retries).toMatchObject({
			label: "Retries",
			integer: true,
			min: 0,
			max: 5,
			required: true
		});
	});

	it("returns no fields without a schema", () => {
		expect(buildSchemaFields(null)).toEqual([]);
	});
});

describe("rawFormValues", () => {
	it("prefers stored values, then schema defaults", () => {
		const fields = buildSchemaFields(schema);
		const raw = rawFormValues(fields, { username: "ada", speed: 1.5 });
		expect(raw).toMatchObject({ username: "ada", retries: "2", speed: "1.5", verbose: true });
	});
});

describe("collectValues", () => {
	const fields = buildSchemaFields(schema);

	it("returns typed values when everything is valid", () => {
		const result = collectValues(fields, {
			username: " ada ",
			retries: "3",
			speed: "",
			region: "us",
			verbose: false
		});
		expect(result.errors).toEqual({});
		expect(result.values).toEqual({ username: "ada", retries: 3, region: "us", verbose: false });
	});

	it("reports required and out of range values", () => {
		const result = collectValues(fields, {
			username: "",
			retries: "9",
			speed: "0.1",
			region: "eu",
			verbose: true
		});
		expect(result.errors).toEqual({
			username: "Required",
			retries: "Must be at most 5",
			speed: "Must be at least 0.5"
		});
	});

	it("rejects non numbers and fractions for integer fields", () => {
		const bad = collectValues(fields, {
			username: "x",
			retries: "abc",
			speed: "",
			region: "eu",
			verbose: true
		});
		expect(bad.errors.retries).toBe("Enter a number");
		const fraction = collectValues(fields, {
			username: "x",
			retries: "1.5",
			speed: "",
			region: "eu",
			verbose: true
		});
		expect(fraction.errors.retries).toBe("Enter a whole number");
	});

	it("rejects a select value that is not an option", () => {
		const result = collectValues(fields, {
			username: "x",
			retries: "1",
			speed: "",
			region: "asia",
			verbose: true
		});
		expect(result.errors.region).toBe("Choose an option");
	});
});
