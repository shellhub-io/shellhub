import { describe, it, expect } from "vitest";
import { stripeErrorMessage } from "../stripeErrors";

describe("stripeErrorMessage", () => {
  it("returns the mapped message for known codes", () => {
    expect(stripeErrorMessage("card_declined")).toMatch(/declined/i);
    expect(stripeErrorMessage("expired_card")).toMatch(/expired/i);
    expect(stripeErrorMessage("incorrect_cvc")).toMatch(/cvc/i);
    expect(stripeErrorMessage("insufficient_funds")).toMatch(/insufficient/i);
    expect(stripeErrorMessage("authentication_required")).toMatch(
      /authentication/i,
    );
  });

  it("falls back to the given fallback when code is unknown", () => {
    expect(stripeErrorMessage("unknown_code", "Custom fallback")).toBe(
      "Custom fallback",
    );
  });

  it("uses the default fallback when code and explicit fallback are absent", () => {
    expect(stripeErrorMessage(undefined)).toMatch(/try again/i);
  });
});
