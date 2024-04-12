import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { flushPromises } from "@vue/test-utils";
import { store } from "@/store";
import { mfaApi, usersApi, apiKeysApi } from "@/api/http";

describe("Auth Store Actions", () => {
  let mockMfa: MockAdapter;
  let mockUser: MockAdapter;
  let mockApiKeys: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    mockMfa = new MockAdapter(mfaApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockApiKeys = new MockAdapter(apiKeysApi.getAxios());
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mockMfa.reset();
    mockUser.reset();
    mockApiKeys.reset();
  });

  describe("Default Values", () => {
    it("should return the default authentication variables", () => {
      expect(store.getters["auth/currentUser"]).toEqual("");
      expect(store.getters["auth/currentName"]).toEqual("");
      expect(store.getters["auth/tenant"]).toEqual("");
      expect(store.getters["auth/email"]).toEqual("");
      expect(store.getters["auth/recoveryEmail"]).toEqual("");
      expect(store.getters["auth/getLoginTimeout"]).toEqual(0);
      expect(store.getters["auth/stateToken"]).toEqual("");
      expect(store.getters["auth/authStatus"]).toEqual("");
      expect(store.getters["auth/link_mfa"]).toEqual("");
      expect(store.getters["auth/isMfa"]).toEqual(false);
      expect(store.getters["auth/recoveryCodes"]).toEqual([]);
      expect(store.getters["auth/secret"]).toEqual("");
      expect(store.getters["auth/showRecoveryModal"]).toEqual(false);
    });
  });

  describe("MFA Actions", () => {
    it("should disable MFA", async () => {
      const dispatchSpy = vi.spyOn(store, "dispatch");

      mockMfa.onPut("http://localhost:3000/api/user/mfa/disable").reply(200);

      await store.dispatch("auth/disableMfa", { code: "000000" });

      expect(dispatchSpy).toHaveBeenCalledWith("auth/disableMfa", { code: "000000" });
      expect(store.getters["auth/isMfa"]).toEqual(false);
    });

    it("should enable MFA", async () => {
      const enableMfaResponse = { token: "token" };
      const enableMfaData = {
        code: "000000",
        secret: "OYDXN4MO2S2JTASNBG5AD54FVT7A5GVH",
        recovery_codes: ["HW2wlxV40B", "2xsmMUHHHb", "DTQgVsaVac", "KXPBoXvuWD", "QQYTPfotBi", "XWiKBEPyb4"],
      };

      const dispatchSpy = vi.spyOn(store, "dispatch");

      mockMfa.onPut("http://localhost:3000/api/user/mfa/enable").reply(200, enableMfaResponse);

      await store.dispatch("auth/enableMfa", enableMfaData);
      await flushPromises();

      expect(dispatchSpy).toHaveBeenCalledWith("auth/enableMfa", enableMfaData);
    });

    it("should validate MFA", async () => {
      const validateMfaResponse = { token: "token" };
      const validateMfaData = { code: "000000" };

      const dispatchSpy = vi.spyOn(store, "dispatch");

      mockMfa.onPost("http://localhost:3000/api/user/mfa/auth").reply(200, validateMfaResponse);

      await store.dispatch("auth/validateMfa", validateMfaData);
      await flushPromises();

      expect(dispatchSpy).toHaveBeenCalledWith("auth/validateMfa", validateMfaData);
      expect(store.getters["auth/stateToken"]).toEqual(validateMfaResponse.token);
    });

    it("should recover MFA", async () => {
      const recoveryMfaResponse = { token: "token" };
      const recoveryMfaData = { code: "000000" };

      const dispatchSpy = vi.spyOn(store, "dispatch");

      mockMfa.onPost("http://localhost:3000/api/user/mfa/recover").reply(200, recoveryMfaResponse);

      await store.dispatch("auth/recoverLoginMfa", recoveryMfaData);
      await flushPromises();

      expect(dispatchSpy).toHaveBeenCalledWith("auth/recoverLoginMfa", recoveryMfaData);
      expect(store.getters["auth/stateToken"]).toEqual(recoveryMfaResponse.token);
      expect(store.getters["auth/showRecoveryModal"]).toEqual(true);
    });

    it("should generate MFA", async () => {
      const generateMfaResponse = {
        secret: "secret-mfa",
        link: "link-mfa",
        recovery_codes: ["HW2wlxV40B", "2xsmMUHHHb", "DTQgVsaVac", "KXPBoXvuWD", "QQYTPfotBi", "XWiKBEPyb4"],
      };

      const dispatchSpy = vi.spyOn(store, "dispatch");

      mockMfa.onGet("http://localhost:3000/api/user/mfa/generate").reply(200, generateMfaResponse);

      await store.dispatch("auth/generateMfa");

      expect(dispatchSpy).toHaveBeenCalledWith("auth/generateMfa");
      expect(store.getters["auth/link_mfa"]).toEqual(generateMfaResponse.link);
      expect(store.getters["auth/secret"]).toEqual(generateMfaResponse.secret);
      expect(store.getters["auth/recoveryCodes"]).toEqual(generateMfaResponse.recovery_codes);
    });
  });

  describe("User Actions", () => {
    it("should get user status", async () => {
      const getUserStatusResponse = {
        mfa: true,
        token: "token",
        id: "userId",
        user: "username",
        name: "testname",
        email: "test@test.com",
        tenant: "fake-tenant",
        role: "administrator",
      };

      mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, getUserStatusResponse);

      const dispatchSpy = vi.spyOn(store, "dispatch");

      await store.dispatch("auth/getUserInfo");

      expect(dispatchSpy).toHaveBeenCalledWith("auth/getUserInfo");
      expect(store.getters["auth/stateToken"]).toEqual(getUserStatusResponse.token);
      expect(store.getters["auth/currentUser"]).toEqual(getUserStatusResponse.user);
      expect(store.getters["auth/currentName"]).toEqual(getUserStatusResponse.name);
      expect(store.getters["auth/tenant"]).toEqual(getUserStatusResponse.tenant);
      expect(store.getters["auth/email"]).toEqual(getUserStatusResponse.email);
      expect(store.getters["auth/id"]).toEqual(getUserStatusResponse.id);
      expect(store.getters["auth/role"]).toEqual(getUserStatusResponse.role);
      expect(store.getters["auth/isMfa"]).toEqual(getUserStatusResponse.mfa);
    });
  });
});
