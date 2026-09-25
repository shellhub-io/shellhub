import { describe, it, expect, beforeEach } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore, VALID_JWT } from "@/tests/seedAuthStore";
import { useVaultStore } from "@/stores/vaultStore";
import { useSwitchNamespace } from "../useNamespaceMutations";

beforeEach(() => {
  seedAuthStore({ tenant: "tenant-home" });
  useVaultStore.setState({ status: "unlocked" });
  server.use(
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: VALID_JWT, role: "owner" }),
    ),
  );
});

describe("useSwitchNamespace", () => {
  it("leaves the vault unlocked when switching to the namespace already active", async () => {
    const { result } = renderHook(() => useSwitchNamespace(), {
      wrapper: createTestWrapper({ initialEntries: ["/devices"] }),
    });

    await act(async () => {
      await result.current.mutateAsync({ tenantId: "tenant-home" });
    });

    expect(useVaultStore.getState().status).toBe("unlocked");
  });
});
