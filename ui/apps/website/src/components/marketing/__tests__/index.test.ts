import { describe, it, expect } from "vitest";
import * as marketing from "@/components/marketing";

describe("marketing barrel (index.ts)", () => {
  it("exports Section", () => {
    expect(marketing.Section).toBeDefined();
    expect(typeof marketing.Section).toBe("function");
  });

  it("exports SectionHeader", () => {
    expect(marketing.SectionHeader).toBeDefined();
    expect(typeof marketing.SectionHeader).toBe("function");
  });

  it("exports FeatureListItem", () => {
    expect(marketing.FeatureListItem).toBeDefined();
    expect(typeof marketing.FeatureListItem).toBe("function");
  });
});
