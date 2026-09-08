import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { renderHook } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
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
  it("calls updateContainerStatus with accept for accept", async () => {
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
    await result.current(entity, "accept");
    expect(capturedStatus).toBe("accept");
  });

  it("calls updateContainerStatus with reject for reject", async () => {
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
    await result.current(entity, "reject");
    expect(capturedStatus).toBe("reject");
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
});
