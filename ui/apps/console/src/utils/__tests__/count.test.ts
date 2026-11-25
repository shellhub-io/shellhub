import { describe, it, expect } from "vitest";
import { pluralize } from "../count";

describe("pluralize", () => {
  it.each([
    [0, "0 minutes"],
    [1, "1 minute"],
    [2, "2 minutes"],
  ])("%i reads %s", (count, expected) => {
    expect(pluralize(count, "minute")).toBe(expected);
  });
});
