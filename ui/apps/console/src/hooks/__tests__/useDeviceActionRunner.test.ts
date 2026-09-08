import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { renderHook } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { useDeviceActionRunner } from "../useDeviceActionRunner";

const entity = { uid: "uid-1", name: "my-device" };
const wrapper = createTestWrapper();

beforeEach(() => {
  server.use(
    http.patch(
      "*/api/devices/:uid/accept",
      () => new HttpResponse(null, { status: 200 }),
    ),
    http.patch(
      "*/api/devices/:uid/:status",
      () => new HttpResponse(null, { status: 200 }),
    ),
    http.delete(
      "*/api/devices/:uid",
      () => new HttpResponse(null, { status: 200 }),
    ),
  );
});

describe("useDeviceActionRunner", () => {
  it("calls acceptDevice for accept", async () => {
    let called = false;
    server.use(
      http.patch("*/api/devices/:uid/accept", () => {
        called = true;
        return new HttpResponse(null, { status: 200 });
      }),
    );
    const { result } = renderHook(() => useDeviceActionRunner(), { wrapper });
    await result.current(entity, "accept");
    expect(called).toBe(true);
  });

  it("calls updateDeviceStatus with reject for reject", async () => {
    let capturedStatus: string | undefined;
    server.use(
      http.patch("*/api/devices/:uid/:status", ({ params }) => {
        capturedStatus = params.status as string;
        return new HttpResponse(null, { status: 200 });
      }),
    );
    const { result } = renderHook(() => useDeviceActionRunner(), { wrapper });
    await result.current(entity, "reject");
    expect(capturedStatus).toBe("reject");
  });

  it("calls deleteDevice for remove", async () => {
    let called = false;
    server.use(
      http.delete("*/api/devices/:uid", () => {
        called = true;
        return new HttpResponse(null, { status: 200 });
      }),
    );
    const { result } = renderHook(() => useDeviceActionRunner(), { wrapper });
    await result.current(entity, "remove");
    expect(called).toBe(true);
  });
});
