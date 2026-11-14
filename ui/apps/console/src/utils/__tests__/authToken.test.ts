import { describe, expect, it } from "vitest";
import { isSignInLanding, isTokenExpired } from "@/utils/authToken";
import { EXPIRED_JWT, VALID_JWT } from "@/tests/seedAuthStore";

describe("isTokenExpired", () => {
  it.each([
    ["a token whose exp is in the future", VALID_JWT, false],
    ["a token whose exp has passed", EXPIRED_JWT, true],
    ["a value that is not a JWT", "not-a-jwt", true],
  ])("reports %s", (_, token, expired) => {
    expect(isTokenExpired(token)).toBe(expired);
  });
});

describe("isSignInLanding", () => {
  it.each([
    ["?token=abc", true],
    ["?sso_code=abc", true],
    ["?sso=start", true],
    ["?sso=other", false],
    ["?redirect=%2Fdevices", false],
    ["", false],
  ])("reports %s as %s", (search, landing) => {
    expect(isSignInLanding(search)).toBe(landing);
  });
});
