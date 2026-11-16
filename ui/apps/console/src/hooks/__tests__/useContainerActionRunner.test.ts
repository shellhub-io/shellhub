import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { renderHook } from "@testing-library/react";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import { getDevicesQueryKey, getStatsQueryKey } from "@/client";
import { useContainerActionRunner } from "../useContainerActionRunner";

const entity = { uid: "uid-1", name: "my-container" };
const wrapper = createTestWrapper();

beforeEach(() => {
  server.use(
    http.patch(
      "*/api/containers/:uid/:status",
      () => new HttpResponse(null, { status: 200 }),
    ),
    http.delete(
      "*/api/containers/:uid",
      () => new HttpResponse(null, { status: 200 }),
    ),
  );
});

describe("useContainerActionRunner", () => {
  it.each(["accept", "reject"] as const)(
    "sends %s as the container's new status",
    async (operation) => {
      let capturedStatus: string | undefined;
      server.use(
        http.patch("*/api/containers/:uid/:status", ({ params }) => {
          capturedStatus = params.status as string;
          return new HttpResponse(null, { status: 200 });
        }),
      );
      const { result } = renderHook(() => useContainerActionRunner(), {
        wrapper,
      });
      await result.current(entity, operation);
      expect(capturedStatus).toBe(operation);
    },
  );

  it("confirms a replacement in the accept body", async () => {
    let body: unknown;
    server.use(
      http.patch("*/api/containers/:uid/:status", async ({ request }) => {
        body = await request.json();
        return new HttpResponse(null, { status: 200 });
      }),
    );
    const { result } = renderHook(() => useContainerActionRunner(), {
      wrapper,
    });
    await result.current(entity, "accept", true);
    expect(body).toEqual({ replace: true });
  });

  it("calls deleteContainer for remove", async () => {
    let called = false;
    server.use(
      http.delete("*/api/containers/:uid", () => {
        called = true;
        return new HttpResponse(null, { status: 200 });
      }),
    );
    const { result } = renderHook(() => useContainerActionRunner(), {
      wrapper,
    });
    await result.current(entity, "remove");
    expect(called).toBe(true);
  });

  it.each(["accept", "reject", "remove"] as const)(
    "marks the device lists and the dashboard counts stale after %s",
    async (operation) => {
      const queryClient = createTestQueryClient();
      queryClient.setQueryData(getStatsQueryKey(), {});
      queryClient.setQueryData(getDevicesQueryKey(), []);
      const { result } = renderHook(() => useContainerActionRunner(), {
        wrapper: createTestWrapper({ queryClient }),
      });

      await result.current(entity, operation);

      expect(queryClient.getQueryState(getStatsQueryKey())?.isInvalidated).toBe(
        true,
      );
      expect(
        queryClient.getQueryState(getDevicesQueryKey())?.isInvalidated,
      ).toBe(true);
    },
  );
});
