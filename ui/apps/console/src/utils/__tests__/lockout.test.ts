import { describe, it, expect } from "vitest";
import { formatTimeLeft } from "../lockout";

describe("formatTimeLeft", () => {
  it.each([
    [0.4, "0 seconds"],
    [1, "1 second"],
    [59.9, "59 seconds"],
    [60, "1 minute"],
    [119, "1 minute"],
    [600, "10 minutes"],
  ])("%s seconds left reads %s", (seconds, expected) => {
    expect(formatTimeLeft(seconds)).toBe(expected);
  });
});
