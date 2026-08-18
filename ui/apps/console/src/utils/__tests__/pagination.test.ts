import { describe, it, expect } from "vitest";
import { pageCount, PER_PAGE } from "@/utils/pagination";

describe("pageCount", () => {
  it("defaults to the shared page size", () => {
    expect(PER_PAGE).toBe(10);
    expect(pageCount(25)).toBe(pageCount(25, PER_PAGE));
  });

  it.each([
    [0, 10, 0],
    [1, 10, 1],
    [10, 10, 1],
    [11, 10, 2],
    [25, 10, 3],
    [30, 10, 3],
    [7, 5, 2],
  ])("splits %i items at %i per page into %i pages", (totalCount, perPage, expected) => {
    expect(pageCount(totalCount, perPage)).toBe(expected);
  });
});
