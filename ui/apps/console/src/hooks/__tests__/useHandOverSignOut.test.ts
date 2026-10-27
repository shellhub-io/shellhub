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
  it("opens the sign-in page with the notice and a sign-out for it to carry out", () => {
    const { result } = renderHookWithRouter(
      () => ({ handOver: useHandOverSignOut(), location: useLocation() }),
      { initialEntries: ["/account/security"] },
    );

    act(() => result.current.handOver("Password changed."));

    expect(result.current.location.pathname).toBe("/login");
    expect(result.current.location.state).toEqual({
      notice: "Password changed.",
      signOut: true,
    });
  });

  it("keeps the token for the sign-in page to end, and abandons its session", () => {
    const { result } = renderHookWithRouter(() => useHandOverSignOut());

    act(() => result.current("Password changed."));

    const { token, isFromAbandonedSession } = useAuthStore.getState();
    expect(token).toBe(VALID_JWT);
    expect(isFromAbandonedSession(VALID_JWT)).toBe(true);
  });
});
