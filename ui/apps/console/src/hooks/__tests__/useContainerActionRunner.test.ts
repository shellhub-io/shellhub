import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { mockSdkResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { useContainerActionRunner } from "../useContainerActionRunner";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    updateContainerStatus: vi.fn(),
    deleteContainer: vi.fn(),
  }),
);

const entity = { uid: "uid-1", name: "my-container" };
const wrapper = createTestWrapper();

beforeEach(() => {
  vi.clearAllMocks();
  sdk.updateContainerStatus.mockResolvedValue(mockSdkResponse(undefined));
  sdk.deleteContainer.mockResolvedValue(mockSdkResponse(undefined));
});

describe("useContainerActionRunner", () => {
  it("calls updateContainerStatus with accept for accept", async () => {
    const { result } = renderHook(() => useContainerActionRunner(), { wrapper });
    await result.current(entity, "accept");
    expect(sdk.updateContainerStatus).toHaveBeenCalledWith(
      expect.objectContaining({ path: { uid: "uid-1", status: "accept" } }),
    );
  });

  it("calls updateContainerStatus with reject for reject", async () => {
    const { result } = renderHook(() => useContainerActionRunner(), { wrapper });
    await result.current(entity, "reject");
    expect(sdk.updateContainerStatus).toHaveBeenCalledWith(
      expect.objectContaining({ path: { uid: "uid-1", status: "reject" } }),
    );
  });

  it("calls deleteContainer for remove", async () => {
    const { result } = renderHook(() => useContainerActionRunner(), { wrapper });
    await result.current(entity, "remove");
    expect(sdk.deleteContainer).toHaveBeenCalledWith(
      expect.objectContaining({ path: { uid: "uid-1" } }),
    );
  });
});
