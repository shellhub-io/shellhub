import { describe, it, expect } from "vitest";
import { consoleUrl, docsUrl, githubUrl, loginUrl, signupUrl } from "@/links";

describe("links", () => {
  it("loginUrl is consoleUrl + /login", () => {
    expect(loginUrl).toBe(`${consoleUrl}/login`);
  });

  it("signupUrl is consoleUrl + /sign-up", () => {
    expect(signupUrl).toBe(`${consoleUrl}/sign-up`);
  });

  it("consoleUrl is a well-formed URL", () => {
    expect(() => new URL(consoleUrl)).not.toThrow();
  });

  it("docsUrl is a well-formed URL", () => {
    expect(() => new URL(docsUrl)).not.toThrow();
  });

  it("githubUrl is a well-formed URL", () => {
    expect(() => new URL(githubUrl)).not.toThrow();
  });

  it("loginUrl is a well-formed URL", () => {
    expect(() => new URL(loginUrl)).not.toThrow();
  });

  it("signupUrl is a well-formed URL", () => {
    expect(() => new URL(signupUrl)).not.toThrow();
  });
});
