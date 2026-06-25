import { describe, it, expect } from "vitest";
import { useSearchParams } from "react-router-dom";
import { renderHookWithRouter } from "../renderHookWithRouter";

describe("renderHookWithRouter", () => {
  it("resolves useSearchParams without throwing", () => {
    expect(() => {
      renderHookWithRouter(() => useSearchParams());
    }).not.toThrow();
  });

  it("returns the search params object from the hook", () => {
    const { result } = renderHookWithRouter(() => useSearchParams());
    const [searchParams] = result.current;
    expect(searchParams).toBeInstanceOf(URLSearchParams);
  });

  it("seeds the URL from initialEntries so params are readable", () => {
    const { result } = renderHookWithRouter(() => useSearchParams(), {
      initialEntries: ["/?foo=bar"],
    });
    const [searchParams] = result.current;
    expect(searchParams.get("foo")).toBe("bar");
  });

  it("defaults to an empty search string when no initialEntries are provided", () => {
    const { result } = renderHookWithRouter(() => useSearchParams());
    const [searchParams] = result.current;
    expect(searchParams.toString()).toBe("");
  });

  it("accepts multiple initialEntries and starts at the last one", () => {
    const { result } = renderHookWithRouter(() => useSearchParams(), {
      initialEntries: ["/first", "/?page=2"],
    });
    const [searchParams] = result.current;
    expect(searchParams.get("page")).toBe("2");
  });
});
