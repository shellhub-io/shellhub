import { describe, it, expect } from "vitest";
import { renderHook, render } from "@testing-library/react";
import { useQueryClient } from "@tanstack/react-query";
import { useLocation } from "react-router-dom";
import { createTestWrapper, renderHookWithClient } from "../wrapper";

describe("createTestWrapper", () => {
  it("provides a QueryClient to children", () => {
    const Wrapper = createTestWrapper();
    const { result } = renderHook(() => useQueryClient(), { wrapper: Wrapper });
    expect(result.current).toBeDefined();
  });

  it("provides a MemoryRouter when initialEntries is set", () => {
    let pathname = "";
    function Probe() {
      pathname = useLocation().pathname;
      return null;
    }

    const Wrapper = createTestWrapper({ initialEntries: ["/devices"] });
    render(<Probe />, { wrapper: Wrapper });
    expect(pathname).toBe("/devices");
  });
});

describe("renderHookWithClient", () => {
  it("provides a QueryClient without needing an explicit wrapper", () => {
    const { result } = renderHookWithClient(() => useQueryClient());
    expect(result.current).toBeDefined();
  });
});
