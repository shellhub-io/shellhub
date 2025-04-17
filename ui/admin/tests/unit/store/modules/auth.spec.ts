import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useAuthStore from "@admin/store/modules/auth";

describe("Auth", () => {
  let authStore: ReturnType<typeof useAuthStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    authStore = useAuthStore();

    localStorage.clear();
  });

  it("returns initial states", () => {
    expect(authStore.authStatus).toBe("");
    expect(authStore.token).toBe("");
    expect(authStore.currentUser).toBe("");
    expect(authStore.isLoggedIn).toBe(false);
  });

  it("handles login states correctly", async () => {
    const statusLoading = "loading";
    const statusError = "error";
    const statusSuccess = "success";
    const token = "eyJhbGciOiJSUzI1NiIsInR5c";
    const user = "user";

    authStore.status = statusLoading;
    expect(authStore.authStatus).toBe(statusLoading);

    authStore.status = statusError;
    expect(authStore.authStatus).toBe(statusError);

    authStore.status = statusSuccess;
    authStore.token = token;
    authStore.user = user;

    expect(authStore.authStatus).toBe(statusSuccess);
    expect(authStore.isLoggedIn).toBe(true);
    expect(authStore.currentUser).toBe(user);

    authStore.logout();

    expect(authStore.authStatus).toBe("");
    expect(authStore.isLoggedIn).toBe(false);
    expect(authStore.currentUser).toBe("");
    expect(authStore.token).toBe("");
  });
});
