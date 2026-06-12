import { describe, it, expect } from "vitest";
import { cn } from "../primitives/cn";

describe("cn", () => {
  it("merges tailwind classes — last conflicting class wins", () => {
    expect(cn("rounded-xl", "rounded-lg")).toBe("rounded-lg");
  });

  it("concatenates non-conflicting classes", () => {
    expect(cn("flex", "items-center")).toBe("flex items-center");
  });

  it("handles falsy values", () => {
    expect(cn("flex", false, undefined, null, "gap-4")).toBe("flex gap-4");
  });

  it("handles conditional objects", () => {
    expect(cn("flex", { "items-center": true, "items-start": false })).toBe(
      "flex items-center",
    );
  });

  it("merges custom font-size — text-2xs wins over text-xs", () => {
    expect(cn("text-xs", "text-2xs")).toBe("text-2xs");
  });
});
