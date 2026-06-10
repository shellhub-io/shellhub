import { describe, it, expect } from "vitest";
import { buildFilter } from "../useDevices";
import { decodeB64url as decodeFilter } from "@/test/decodeB64url";

describe("buildFilter", () => {
  describe("search only", () => {
    it("encodes a name OR custom_fields filter", () => {
      const result = decodeFilter(buildFilter("my-device", []));
      expect(result).toEqual([
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "name", operator: "contains", value: "my-device" } },
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "custom_fields", operator: "contains", value: "my-device" } },
      ]);
    });
  });

  describe("tags only", () => {
    it("encodes a tags filter", () => {
      const result = decodeFilter(buildFilter("", ["web", "prod"]));
      expect(result).toEqual([
        { type: "property", params: { name: "tags.name", operator: "contains", value: ["web", "prod"] } },
      ]);
    });
  });

  describe("search and tags combined", () => {
    it("encodes both filters in the same array", () => {
      const result = decodeFilter(buildFilter("srv", ["prod"]));
      expect(result).toEqual([
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "name", operator: "contains", value: "srv" } },
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "custom_fields", operator: "contains", value: "srv" } },
        { type: "property", params: { name: "tags.name", operator: "contains", value: ["prod"] } },
      ]);
    });
  });

  describe("empty inputs", () => {
    it("returns an empty filter array when both are empty", () => {
      const result = decodeFilter(buildFilter("", []));
      expect(result).toEqual([]);
    });
  });

  describe("non-ASCII search value", () => {
    it("round-trips a non-ASCII search through base64url encoding", () => {
      // "¿" encodes to a base64url string containing "_" (the url-safe replacement for "/"),
      // which atob() cannot decode. This test verifies the polyfill-safe decode path.
      const result = decodeFilter(buildFilter("¿", []));
      expect(result).toEqual([
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "name", operator: "contains", value: "¿" } },
        { type: "operator", params: { name: "or" } },
        { type: "property", params: { name: "custom_fields", operator: "contains", value: "¿" } },
      ]);
    });
  });
});
