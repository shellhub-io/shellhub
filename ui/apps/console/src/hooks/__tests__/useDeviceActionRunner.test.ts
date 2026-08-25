import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { mockSdkResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { useDeviceActionRunner } from "../useDeviceActionRunner";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    acceptDevice: vi.fn(),
    updateDeviceStatus: vi.fn(),
    deleteDevice: vi.fn(),
  }),
);

const entity = { uid: "uid-1", name: "my-device" };
const wrapper = createTestWrapper();

beforeEach(() => {
  vi.clearAllMocks();
  sdk.acceptDevice.mockResolvedValue(mockSdkResponse(undefined));
  sdk.updateDeviceStatus.mockResolvedValue(mockSdkResponse(undefined));
  sdk.deleteDevice.mockResolvedValue(mockSdkResponse(undefined));
});

describe("useDeviceActionRunner", () => {
  it("calls acceptDevice for accept", async () => {
    const { result } = renderHook(() => useDeviceActionRunner(), { wrapper });
    await result.current(entity, "accept");
    expect(sdk.acceptDevice).toHaveBeenCalledWith(
      expect.objectContaining({ path: { uid: "uid-1" } }),
    );
  });

  it("calls updateDeviceStatus with reject for reject", async () => {
    const { result } = renderHook(() => useDeviceActionRunner(), { wrapper });
    await result.current(entity, "reject");
    expect(sdk.updateDeviceStatus).toHaveBeenCalledWith(
      expect.objectContaining({ path: { uid: "uid-1", status: "reject" } }),
    );
  });

  it("calls deleteDevice for remove", async () => {
    const { result } = renderHook(() => useDeviceActionRunner(), { wrapper });
    await result.current(entity, "remove");
    expect(sdk.deleteDevice).toHaveBeenCalledWith(
      expect.objectContaining({ path: { uid: "uid-1" } }),
    );
  });
});
