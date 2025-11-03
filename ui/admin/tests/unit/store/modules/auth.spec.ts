import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useAuthStore from "@admin/store/modules/auth";

describe("Auth", () => {
  setActivePinia(createPinia());
  const authStore = useAuthStore();

  beforeEach(() => {
    localStorage.clear();
  });

  it("returns initial states", () => {
    expect(authStore.status).toBe("");
    expect(authStore.token).toBe("");
    expect(authStore.currentUser).toBe("");
    expect(authStore.isLoggedIn).toBe(false);
  });

  it("handles login states correctly", () => {
    const statusLoading = "loading";
    const statusError = "error";
    const statusSuccess = "success";
    const token = "eyJhbGciOiJSUzI1NiIsInR5c";
    const user = "user";

    authStore.status = statusLoading;
    expect(authStore.status).toBe(statusLoading);

    authStore.status = statusError;
    expect(authStore.status).toBe(statusError);

    authStore.status = statusSuccess;
    authStore.token = token;
    authStore.currentUser = user;

    expect(authStore.status).toBe(statusSuccess);
    expect(authStore.isLoggedIn).toBe(true);
    expect(authStore.currentUser).toBe(user);

    authStore.logout();

    expect(authStore.status).toBe("");
    expect(authStore.isLoggedIn).toBe(false);
    expect(authStore.currentUser).toBe("");
    expect(authStore.token).toBe("");
  });
});
