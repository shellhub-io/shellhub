import { describe, it, expect } from "vitest";
import { setEdition } from "@/tests/edition";
import { canCreateNamespace } from "../namespaceCreation";

describe("canCreateNamespace", () => {
  it.each([
    ["community", { max_namespaces: -1 }, false],
    ["cloud", { max_namespaces: 0 }, false],
    ["cloud", undefined, false],
    ["enterprise", { max_namespaces: -1 }, true],
  ] as const)("%s with %o is %s", (edition, user, expected) => {
    setEdition(edition);
    expect(canCreateNamespace(user)).toBe(expected);
  });
});
