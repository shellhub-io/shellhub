import { describe, it, expect, beforeEach } from "vitest";
import { act } from "@testing-library/react";
import { useLocation } from "react-router-dom";
import { useHandOverSignOut } from "@/hooks/useHandOverSignOut";
import { useAuthStore } from "@/stores/authStore";
import { renderHookWithRouter } from "@/tests/renderHookWithRouter";
import { VALID_JWT } from "@/tests/seedAuthStore";

beforeEach(() => {
  useAuthStore.setState({ token: VALID_JWT, signOutHandedOver: false });
});

describe("useHandOverSignOut", () => {
  it("opens the sign-in page with the notice and a sign-out for it to carry out", async () => {
    const { result } = renderHookWithRouter(
      () => ({ handOver: useHandOverSignOut(), location: useLocation() }),
      { initialEntries: ["/account/security"] },
    );

    await act(() =>
      result.current.handOver(() => Promise.resolve(), "Password changed."),
    );

    expect(result.current.location.pathname).toBe("/login");
    expect(result.current.location.state).toEqual({
      notice: "Password changed.",
      signOut: true,
    });
  });

  it("keeps the token for the sign-in page to end, and abandons its session", async () => {
    const { result } = renderHookWithRouter(() => useHandOverSignOut());

    await act(() => result.current(() => Promise.resolve(), "Password changed."));

    const { token, isFromAbandonedSession } = useAuthStore.getState();
    expect(token).toBe(VALID_JWT);
    expect(isFromAbandonedSession(VALID_JWT)).toBe(true);
  });

  it("abandons the session before the revoking request resolves", async () => {
    const { result } = renderHookWithRouter(() => useHandOverSignOut());
    let abandonedWhileRevoking = false;

    await act(() =>
      result.current(() => {
        abandonedWhileRevoking =
          useAuthStore.getState().isFromAbandonedSession(VALID_JWT);
        return Promise.resolve();
      }, "Password changed."),
    );

    expect(abandonedWhileRevoking).toBe(true);
  });

  it("withdraws the hand-over and stays on the page when the revoking request fails", async () => {
    const { result } = renderHookWithRouter(
      () => ({ handOver: useHandOverSignOut(), location: useLocation() }),
      { initialEntries: ["/account/security"] },
    );
    const failure = new Error("forbidden");

    await act(() =>
      expect(
        result.current.handOver(() => Promise.reject(failure), "Password changed."),
      ).rejects.toBe(failure),
    );

    expect(result.current.location.pathname).toBe("/account/security");
    expect(useAuthStore.getState().isFromAbandonedSession(VALID_JWT)).toBe(
      false,
    );
  });
});
