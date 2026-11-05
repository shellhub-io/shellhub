import { describe, it, expect } from "vitest";
import { buildFilter } from "../useDevices";
import { decodeB64url as decodeFilter } from "@/tests/decodeB64url";

describe("buildFilter", () => {
  describe("search only", () => {
    it("encodes a name or custom_fields filter", () => {
      const result = decodeFilter(buildFilter("my-device", []));
      expect(result).toEqual([
        {
          type: "property",
          params: { name: "name", operator: "contains", value: "my-device" },
        },
        { type: "operator", params: { name: "or" } },
        {
          type: "property",
          params: {
            name: "custom_fields",
            operator: "contains",
            value: "my-device",
          },
        },
      ]);
    });
  });

  describe("tags only", () => {
    it("encodes a tags filter", () => {
      const result = decodeFilter(buildFilter("", ["web", "prod"]));
      expect(result).toEqual([
        {
          type: "property",
          params: {
            name: "tags.name",
            operator: "contains",
            value: ["web", "prod"],
          },
        },
      ]);
    });
  });

  describe("search and tags combined", () => {
    it("requires the tags beside each search condition", () => {
      const tags = {
        type: "property",
        params: { name: "tags.name", operator: "contains", value: ["prod"] },
      };
      const and = { type: "operator", params: { name: "and" } };

      const result = decodeFilter(buildFilter("srv", ["prod"]));
      expect(result).toEqual([
        {
          type: "property",
          params: { name: "name", operator: "contains", value: "srv" },
        },
        and,
        tags,
        { type: "operator", params: { name: "or" } },
        {
          type: "property",
          params: { name: "custom_fields", operator: "contains", value: "srv" },
        },
        and,
        tags,
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
      const result = decodeFilter(buildFilter("¿", []));
      expect(result).toEqual([
        {
          type: "property",
          params: { name: "name", operator: "contains", value: "¿" },
        },
        { type: "operator", params: { name: "or" } },
        {
          type: "property",
          params: { name: "custom_fields", operator: "contains", value: "¿" },
        },
      ]);
    });
  });
});
