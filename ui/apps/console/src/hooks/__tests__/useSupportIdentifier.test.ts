import { describe, it, expect, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { renderHookWithClient } from "@/tests/wrapper";
import { server } from "@/tests/msw";
import { defaultHandlers } from "@/tests/handlers";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useSupportIdentifier } from "../useSupportIdentifier";

beforeEach(() => {
  server.use(...defaultHandlers);
  seedAuthStore();
});

describe("useSupportIdentifier", () => {
  describe("when enabled=false", () => {
    it("never fires the query and returns null identifier", () => {
      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", false),
      );

      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isError).toBe(false);
    });
  });

  describe("when tenantId is empty", () => {
    it("does not fire the query when tenantId is empty string", () => {
      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("", true),
      );

      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });

    it("does not fire the query when tenantId is null", () => {
      const { result } = renderHookWithClient(() =>
        useSupportIdentifier(null, true),
      );

      expect(result.current.identifier).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });
  });

  describe("when enabled with a valid tenant", () => {
    it("returns the identifier from the response", async () => {
      server.use(
        http.get("*/api/namespaces/:tenant/support", () =>
          HttpResponse.json({ identifier: "abc123" }),
        ),
      );

      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", true),
      );

      await waitFor(() => expect(result.current.identifier).toBe("abc123"));
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isError).toBe(false);
    });
  });

  describe("retry policy", () => {
    it("retries the query exactly once on failure (transient blip recovery)", async () => {
      let callCount = 0;
      server.use(
        http.get("*/api/namespaces/:tenant/support", () => {
          callCount++;
          return HttpResponse.json(null, { status: 500 });
        }),
      );

      const { result } = renderHookWithClient(() =>
        useSupportIdentifier("tenant-123", true),
      );

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(callCount).toBe(2);
    });
  });
});
