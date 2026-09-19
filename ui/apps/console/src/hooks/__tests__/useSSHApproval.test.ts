import { describe, it, expect, vi, beforeEach } from "vitest";
import { act, waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getSshApproval: vi.fn(),
    confirmSshApproval: vi.fn(),
    rejectSshApproval: vi.fn(),
  }),
);

import { useSSHApproval } from "../useSSHApproval";

const pending = {
  code: "WXYZ2K7Q",
  kind: "identity",
  state: "pending",
  device: { name: "device" },
  username: "gustavo",
  ip_address: "10.0.0.1",
  fingerprint: "SHA256:aaa",
  expires_in_seconds: 90,
};

beforeEach(() => {
  vi.clearAllMocks();
  sdk.getSshApproval.mockResolvedValue(mockSdkResponse(pending));
});

async function decided(decision: "confirm" | "reject") {
  const { result } = renderHookWithClient(() => useSSHApproval("WXYZ2K7Q"));

  await waitFor(() => expect(result.current.phase).toBe("pending"));

  let returned: string | null = "untouched";
  await act(async () => {
    returned = await result.current[decision]();
  });

  return { result, returned };
}

describe("useSSHApproval decide", () => {
  it("hands back the confirmation code so the terminal can be answered with it", async () => {
    sdk.confirmSshApproval.mockResolvedValue(
      mockSdkResponse({ confirmation_code: "CONF7788" }),
    );

    const { result, returned } = await decided("confirm");

    expect(returned).toBe("CONF7788");
    expect(result.current.phase).toBe("confirmed");
    expect(result.current.confirmationCode).toBe("CONF7788");
  });

  it("hands back an empty code on a reject, which is what dismisses the login", async () => {
    sdk.rejectSshApproval.mockResolvedValue(mockSdkResponse({}));

    const { result, returned } = await decided("reject");

    expect(returned).toBe("");
    expect(result.current.phase).toBe("rejected");
  });

  it("hands back nothing when the request is already gone, and says so", async () => {
    sdk.confirmSshApproval.mockRejectedValue(makeSdkError(404));

    const { result, returned } = await decided("confirm");

    expect(returned).toBeNull();
    expect(result.current.phase).toBe("expired");
  });
});
