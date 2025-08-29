import { beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { mfaApi, usersApi } from "@/api/http";
import useAuthStore from "@/store/modules/auth";

describe("Auth Pinia Store", () => {
  const mockMfaApi = new MockAdapter(mfaApi.getAxios());
  const mockUsersApi = new MockAdapter(usersApi.getAxios());
  let authStore: ReturnType<typeof useAuthStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    authStore = useAuthStore();
  });

  describe("Default Values", () => {
    it("should return the default authentication variables", () => {
      expect(authStore.username).toEqual("");
      expect(authStore.name).toEqual("");
      expect(authStore.tenantId).toEqual("");
      expect(authStore.email).toEqual("");
      expect(authStore.recoveryEmail).toEqual("");
      expect(authStore.loginTimeout).toEqual(0);
      expect(authStore.token).toEqual("");
      expect(authStore.isMfaEnabled).toEqual(false);
      expect(authStore.showRecoveryModal).toEqual(false);
    });
  });

  describe("MFA Actions", () => {
    it("should disable MFA", async () => {
      mockMfaApi.onPut("http://localhost:3000/api/user/mfa/disable").reply(200);

      await authStore.disableMfa({ code: "000000" });

      expect(authStore.isMfaEnabled).toEqual(false);
    });

    it("should enable MFA", async () => {
      const enableMfaResponse = { token: "token" };
      const enableMfaData = {
        code: "000000",
        secret: "OYDXN4MO2S2JTASNBG5AD54FVT7A5GVH",
        recovery_codes: ["HW2wlxV40B", "2xsmMUHHHb", "DTQgVsaVac", "KXPBoXvuWD", "QQYTPfotBi", "XWiKBEPyb4"],
      };

      mockMfaApi.onPut("http://localhost:3000/api/user/mfa/enable").reply(200, enableMfaResponse);

      await authStore.enableMfa(enableMfaData);
      await flushPromises();

      expect(authStore.isMfaEnabled).toEqual(true);
    });

    it("should validate MFA", async () => {
      const validateMfaResponse = { token: "token" };
      const verificationCode = "000000";

      // Set up MFA token first
      authStore.mfaToken = "test-mfa-token";

      mockMfaApi.onPost("http://localhost:3000/api/user/mfa/auth").reply(200, validateMfaResponse);

      await authStore.validateMfa(verificationCode);
      await flushPromises();

      expect(authStore.mfaToken).toEqual(validateMfaResponse.token);
    });

    it("should recover MFA", async () => {
      const recoveryMfaResponse = { token: "token" };
      const recoveryCode = "000000";

      mockMfaApi.onPost("http://localhost:3000/api/user/mfa/recover").reply(200, recoveryMfaResponse);

      await authStore.recoverMfa(recoveryCode);
      authStore.isMfaEnabled = true; // Simulate MFA being enabled
      await flushPromises();

      expect(authStore.mfaToken).toEqual(recoveryMfaResponse.token);
      expect(authStore.showRecoveryModal).toEqual(true);
    });

    it("should generate MFA", async () => {
      const generateMfaResponse = {
        secret: "secret-mfa",
        link: "qr-code-link",
        recovery_codes: ["HW2wlxV40B", "2xsmMUHHHb", "DTQgVsaVac", "KXPBoXvuWD", "QQYTPfotBi", "XWiKBEPyb4"],
      };

      mockMfaApi.onGet("http://localhost:3000/api/user/mfa/generate").reply(200, generateMfaResponse);

      const result = await authStore.generateMfa();

      expect(result.secret).toEqual(generateMfaResponse.secret);
      expect(result.link).toEqual(generateMfaResponse.link);
      expect(result.recovery_codes).toEqual(generateMfaResponse.recovery_codes);
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
        tenantId: "fake-tenant",
        role: "administrator",
        recovery_email: "recovery@test.com",
        auth_methods: ["local", "oauth"],
      };

      mockUsersApi.onGet("http://localhost:3000/api/auth/user").reply(200, getUserStatusResponse);

      await authStore.getUserInfo();

      expect(authStore.token).toEqual(getUserStatusResponse.token);
      expect(authStore.username).toEqual(getUserStatusResponse.user);
      expect(authStore.name).toEqual(getUserStatusResponse.name);
      expect(authStore.tenantId).toEqual(getUserStatusResponse.tenantId);
      expect(authStore.email).toEqual(getUserStatusResponse.email);
      expect(authStore.id).toEqual(getUserStatusResponse.id);
      expect(authStore.role).toEqual(getUserStatusResponse.role);
      expect(authStore.isMfaEnabled).toEqual(getUserStatusResponse.mfa);
      expect(authStore.recoveryEmail).toEqual(getUserStatusResponse.recovery_email);
      expect(authStore.authMethods).toEqual(getUserStatusResponse.auth_methods);
    });
  });
});
