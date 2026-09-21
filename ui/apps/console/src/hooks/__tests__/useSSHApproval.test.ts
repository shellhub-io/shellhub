import { describe, it, expect, beforeEach } from "vitest";
import { act, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { renderHookWithClient } from "@/tests/wrapper";
import { server } from "@/tests/msw";
import { defaultHandlers } from "@/tests/handlers";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useSSHApproval } from "../useSSHApproval";

const pending = {
  code: "WXYZ2K7Q",
  kind: "identity",
  state: "pending",
  device_name: "device",
  username: "gustavo",
  ip_address: "10.0.0.1",
  fingerprint: "SHA256:aaa",
  expires_in_seconds: 90,
};

beforeEach(() => {
  server.use(
    ...defaultHandlers,
    http.get("*/api/ssh-approvals/:code", () =>
      HttpResponse.json(pending),
    ),
  );
  seedAuthStore();
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
    server.use(
      http.post("*/api/ssh-approvals/:code/confirm", () =>
        HttpResponse.json({ confirmation_code: "CONF7788" }),
      ),
    );

    const { result, returned } = await decided("confirm");

    expect(returned).toBe("CONF7788");
    expect(result.current.phase).toBe("confirmed");
    expect(result.current.confirmationCode).toBe("CONF7788");
  });

  it("hands back an empty code on a reject, which is what dismisses the login", async () => {
    server.use(
      http.post("*/api/ssh-approvals/:code/reject", () =>
        HttpResponse.json({}),
      ),
    );

    const { result, returned } = await decided("reject");

    expect(returned).toBe("");
    expect(result.current.phase).toBe("rejected");
  });

  it("hands back nothing when the request is already gone, and says so", async () => {
    server.use(
      http.post("*/api/ssh-approvals/:code/confirm", () =>
        HttpResponse.json(null, { status: 404 }),
      ),
    );

    const { result, returned } = await decided("confirm");

    expect(returned).toBeNull();
    expect(result.current.phase).toBe("expired");
  });
});
