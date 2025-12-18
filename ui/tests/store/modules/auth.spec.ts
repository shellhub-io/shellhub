import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { mfaApi, usersApi, namespacesApi } from "@/api/http";
import useAuthStore from "@/store/modules/auth";

describe("Auth Store", () => {
  let mockMfaApi: MockAdapter;
  let mockUsersApi: MockAdapter;
  let mockNamespacesApi: MockAdapter;
  let store: ReturnType<typeof useAuthStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockMfaApi = new MockAdapter(mfaApi.getAxios());
    mockUsersApi = new MockAdapter(usersApi.getAxios());
    mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
    store = useAuthStore();
    localStorage.clear();
  });

  afterEach(() => {
    mockMfaApi.reset();
    mockUsersApi.reset();
    mockNamespacesApi.reset();
    localStorage.clear();
  });

  describe("Initial State", () => {
    it("should have empty authentication variables", () => {
      expect(store.username).toBe("");
      expect(store.name).toBe("");
      expect(store.tenantId).toBe("");
      expect(store.email).toBe("");
      expect(store.recoveryEmail).toBe("");
      expect(store.loginTimeout).toBe(0);
      expect(store.token).toBe("");
      expect(store.isMfaEnabled).toBe(false);
    });

    it("should have isLoggedIn as false when no token", () => {
      expect(store.isLoggedIn).toBe(false);
    });

    it("should load token from localStorage if present", () => {
      localStorage.setItem("token", "stored-token");
      localStorage.setItem("user", "stored-user");

      setActivePinia(createPinia());
      const newStore = useAuthStore();

      expect(newStore.token).toBe("stored-token");
      expect(newStore.username).toBe("stored-user");
    });
  });

  describe("login", () => {
    const loginUrl = "http://localhost:3000/api/login";

    it("should login successfully and persist auth data", async () => {
      const loginData = {
        username: "testuser",
        password: "password123",
      };

      const mockResponse = {
        token: "test-token",
        user: "testuser",
        name: "Test User",
        email: "test@example.com",
        tenant: "test-tenant",
        id: "user-123",
        role: "administrator",
        mfa: false,
        auth_methods: ["local"],
      };

      mockUsersApi
        .onPost(loginUrl)
        .reply(200, mockResponse);

      await store.login(loginData);

      expect(store.token).toBe(mockResponse.token);
      expect(store.username).toBe(mockResponse.user);
      expect(store.name).toBe(mockResponse.name);
      expect(store.email).toBe(mockResponse.email);
      expect(store.tenantId).toBe(mockResponse.tenant);
      expect(store.id).toBe(mockResponse.id);
      expect(store.role).toBe(mockResponse.role);
      expect(store.isMfaEnabled).toBe(false);
      expect(store.isLoggedIn).toBe(true);
    });

    it("should persist auth data to localStorage", async () => {
      const mockResponse = {
        token: "test-token",
        user: "testuser",
        name: "Test User",
        email: "test@example.com",
        tenant: "test-tenant",
        id: "user-123",
        role: "administrator",
      };

      mockUsersApi
        .onPost(loginUrl)
        .reply(200, mockResponse);

      await store.login({ username: "testuser", password: "password123" });

      expect(localStorage.getItem("token")).toBe(mockResponse.token);
      expect(localStorage.getItem("user")).toBe(mockResponse.user);
      expect(localStorage.getItem("name")).toBe(mockResponse.name);
      expect(localStorage.getItem("email")).toBe(mockResponse.email);
      expect(localStorage.getItem("tenant")).toBe(mockResponse.tenant);
      expect(localStorage.getItem("id")).toBe(mockResponse.id);
      expect(localStorage.getItem("role")).toBe(mockResponse.role);
    });

    it("should handle MFA requirement and set mfaToken", async () => {
      const loginData = {
        username: "mfauser",
        password: "password123",
      };

      mockUsersApi
        .onPost(loginUrl)
        .reply(
          401,
          { message: "MFA required" },
          { "x-mfa-token": "mfa-challenge-token" },
        );

      await store.login(loginData);

      expect(store.isMfaEnabled).toBe(true);
      expect(store.mfaToken).toBe("mfa-challenge-token");
      expect(localStorage.getItem("mfa")).toBe("true");
    });

    it("should handle account lockout timeout", async () => {
      mockUsersApi
        .onPost(loginUrl)
        .reply(
          423,
          { message: "Account locked" },
          { "x-account-lockout": "300" },
        );

      await expect(
        store.login({ username: "user", password: "wrong" }),
      ).rejects.toBeAxiosErrorWithStatus(423);

      expect(store.loginTimeout).toBe("300");
    });

    it("should handle invalid credentials error", async () => {
      mockUsersApi
        .onPost(loginUrl)
        .reply(401, { message: "Invalid credentials" });

      await expect(
        store.login({ username: "user", password: "wrong" }),
      ).rejects.toBeAxiosErrorWithStatus(401);
    });
  });

  describe("loginWithToken", () => {
    const authUserUrl = "http://localhost:3000/api/auth/user";

    it("should login with token and get user info", async () => {
      const mockResponse = {
        token: "existing-token",
        user: "tokenuser",
        name: "Token User",
        email: "token@example.com",
        tenant: "token-tenant",
        id: "user-456",
        role: "observer",
      };

      mockUsersApi
        .onGet(authUserUrl)
        .reply(200, mockResponse);

      await store.loginWithToken("existing-token");

      expect(store.token).toBe(mockResponse.token);
      expect(store.username).toBe(mockResponse.user);
      expect(store.name).toBe(mockResponse.name);
      expect(store.tenantId).toBe(mockResponse.tenant);
    });

    it("should handle invalid token", async () => {
      mockUsersApi
        .onGet(authUserUrl)
        .reply(401, { message: "Invalid token" });

      await expect(
        store.loginWithToken("invalid-token"),
      ).rejects.toBeAxiosErrorWithStatus(401);
    });
  });

  describe("getUserInfo", () => {
    const authUserUrl = "http://localhost:3000/api/auth/user";

    it("should fetch and persist user information", async () => {
      const mockResponse = {
        token: "user-token",
        user: "username",
        name: "Full Name",
        email: "user@example.com",
        tenant: "user-tenant",
        id: "user-id",
        role: "administrator",
        mfa: true,
        recovery_email: "recovery@example.com",
        auth_methods: ["local", "oauth"],
      };

      mockUsersApi
        .onGet(authUserUrl)
        .reply(200, mockResponse);

      await store.getUserInfo();

      expect(store.token).toBe(mockResponse.token);
      expect(store.username).toBe(mockResponse.user);
      expect(store.name).toBe(mockResponse.name);
      expect(store.email).toBe(mockResponse.email);
      expect(store.tenantId).toBe(mockResponse.tenant);
      expect(store.id).toBe(mockResponse.id);
      expect(store.role).toBe(mockResponse.role);
      expect(store.isMfaEnabled).toBe(mockResponse.mfa);
      expect(store.recoveryEmail).toBe(mockResponse.recovery_email);
      expect(store.authMethods).toEqual(mockResponse.auth_methods);
    });
  });

  describe("MFA - generateMfa", () => {
    const generateMfaUrl = "http://localhost:3000/api/user/mfa/generate";

    it("should generate MFA secret and QR code", async () => {
      const mockResponse = {
        secret: "OYDXN4MO2S2JTASNBG5AD54FVT7A5GVH",
        link: "otpauth://totp/ShellHub:user@example.com?secret=SECRET&issuer=ShellHub",
        recovery_codes: ["CODE1", "CODE2", "CODE3"],
      };

      mockMfaApi
        .onGet(generateMfaUrl)
        .reply(200, mockResponse);

      const result = await store.generateMfa();

      expect(result.secret).toBe(mockResponse.secret);
      expect(result.link).toBe(mockResponse.link);
      expect(result.recovery_codes).toEqual(mockResponse.recovery_codes);
    });
  });

  describe("MFA - enableMfa", () => {
    const enableMfaUrl = "http://localhost:3000/api/user/mfa/enable";

    it("should enable MFA successfully", async () => {
      const enableData = {
        code: "123456",
        secret: "SECRET",
        recovery_codes: ["CODE1", "CODE2"],
      };

      mockMfaApi
        .onPut(enableMfaUrl)
        .reply(200);

      await store.enableMfa(enableData);

      expect(store.isMfaEnabled).toBe(true);
    });

    it("should handle invalid verification code", async () => {
      const enableData = {
        code: "000000",
        secret: "SECRET",
        recovery_codes: ["CODE1"],
      };

      mockMfaApi
        .onPut(enableMfaUrl)
        .reply(400, { message: "Invalid verification code" });

      await expect(
        store.enableMfa(enableData),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });
  });

  describe("MFA - disableMfa", () => {
    const disableMfaUrl = "http://localhost:3000/api/user/mfa/disable";

    it("should disable MFA successfully", async () => {
      store.isMfaEnabled = true;
      localStorage.setItem("mfa", "true");

      mockMfaApi
        .onPut(disableMfaUrl)
        .reply(200);

      await store.disableMfa({ code: "123456" });

      expect(store.isMfaEnabled).toBe(false);
      expect(localStorage.getItem("mfa")).toBe("false");
    });

    it("should handle invalid code when disabling MFA", async () => {
      mockMfaApi
        .onPut(disableMfaUrl)
        .reply(400, { message: "Invalid code" });

      await expect(
        store.disableMfa({ code: "000000" }),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });
  });

  describe("MFA - validateMfa", () => {
    const validateMfaUrl = "http://localhost:3000/api/user/mfa/auth";

    it("should validate MFA code successfully", async () => {
      store.mfaToken = "mfa-challenge-token";

      const mockResponse = {
        token: "authenticated-token",
        user: "mfauser",
        name: "MFA User",
        tenant: "mfa-tenant",
      };

      mockMfaApi
        .onPost(validateMfaUrl)
        .reply(200, mockResponse);

      await store.validateMfa("123456");

      expect(store.token).toBe(mockResponse.token);
      expect(store.mfaToken).toBe(mockResponse.token);
      expect(store.isMfaEnabled).toBe(true);
      expect(store.username).toBe(mockResponse.user);
    });

    it("should handle invalid MFA code", async () => {
      store.mfaToken = "mfa-challenge-token";

      mockMfaApi
        .onPost(validateMfaUrl)
        .reply(401, { message: "Invalid MFA code" });

      await expect(
        store.validateMfa("000000"),
      ).rejects.toBeAxiosErrorWithStatus(401);
    });
  });

  describe("MFA - recoverMfa", () => {
    const recoverMfaUrl = "http://localhost:3000/api/user/mfa/recover";

    it("should recover MFA with recovery code", async () => {
      localStorage.setItem("name", "testuser");

      const mockResponse = {
        token: "recovery-token",
        user: "recovereduser",
        name: "Recovered User",
      };

      mockMfaApi
        .onPost(recoverMfaUrl)
        .reply(200, mockResponse, {
          "x-expires-at": "1800",
        });

      await store.recoverMfa("RECOVERY-CODE-1");

      expect(store.token).toBe(mockResponse.token);
      expect(store.mfaToken).toBe(mockResponse.token);
      expect(store.isRecoveringMfa).toBe(true);
      expect(store.recoveryCode).toBe("RECOVERY-CODE-1");
      expect(store.disableTimeout).toBe("1800");
    });

    it("should handle invalid recovery code", async () => {
      localStorage.setItem("name", "testuser");

      mockMfaApi
        .onPost(recoverMfaUrl)
        .reply(401, { message: "Invalid recovery code" });

      await expect(
        store.recoverMfa("INVALID-CODE"),
      ).rejects.toBeAxiosErrorWithStatus(401);
    });
  });

  describe("MFA - requestMfaReset", () => {
    const requestMfaResetUrl = "http://localhost:3000/api/user/mfa/reset";

    it("should request MFA reset", async () => {
      localStorage.setItem("name", "testuser");

      mockMfaApi
        .onPost(requestMfaResetUrl)
        .reply(200);

      await expect(
        store.requestMfaReset(),
      ).resolves.not.toThrow();
    });
  });

  describe("MFA - resetMfa", () => {
    const generateResetMfaUrl = (id: string) => `http://localhost:3000/api/user/mfa/reset/${id}`;

    it("should reset MFA with validation", async () => {
      const resetData = {
        id: "testuser",
        main_email_code: "user@test.com",
        recovery_email_code: "recovery@test.com",
      };

      const mockResponse = {
        token: "new-token",
        user: "testuser",
        name: "Test User",
      };

      mockMfaApi
        .onPut(generateResetMfaUrl("testuser"))
        .reply(200, mockResponse);

      await store.resetMfa(resetData);

      expect(store.token).toBe(mockResponse.token);
      expect(store.username).toBe(mockResponse.user);
    });

    it("should handle invalid reset token", async () => {
      const resetData = {
        id: "testuser",
        main_email_code: "invalid@test.com",
        recovery_email_code: "recovery@test.com",
      };

      mockMfaApi
        .onPut(generateResetMfaUrl("testuser"))
        .reply(401, { message: "Invalid reset token" });

      await expect(
        store.resetMfa(resetData),
      ).rejects.toBeAxiosErrorWithStatus(401);
    });
  });

  describe("logout", () => {
    it("should clear all auth state and localStorage", () => {
      // Set up authenticated state
      store.token = "test-token";
      store.username = "testuser";
      store.name = "Test User";
      store.tenantId = "test-tenant";
      store.email = "test@example.com";
      store.id = "user-123";
      store.role = "administrator";
      store.isMfaEnabled = true;

      localStorage.setItem("token", "test-token");
      localStorage.setItem("user", "testuser");
      localStorage.setItem("mfa", "true");

      store.logout();

      expect(store.token).toBe("");
      expect(store.username).toBe("");
      expect(store.name).toBe("");
      expect(store.tenantId).toBe("");
      expect(store.email).toBe("");
      expect(store.id).toBe("");
      expect(store.role).toBe("");
      expect(store.isMfaEnabled).toBe(false);
      expect(store.isLoggedIn).toBe(false);

      expect(localStorage.getItem("token")).toBeNull();
      expect(localStorage.getItem("user")).toBeNull();
      expect(localStorage.getItem("mfa")).toBeNull();
    });
  });

  describe("deleteUser", () => {
    const deleteUserUrl = "http://localhost:3000/api/user";

    it("should delete user and logout", async () => {
      store.token = "test-token";
      store.username = "testuser";
      localStorage.setItem("token", "test-token");

      mockUsersApi
        .onDelete(deleteUserUrl)
        .reply(200);

      await store.deleteUser();

      expect(store.token).toBe("");
      expect(store.username).toBe("");
      expect(localStorage.getItem("token")).toBeNull();
    });

    it("should handle error deleting user", async () => {
      mockUsersApi
        .onDelete(deleteUserUrl)
        .reply(403, { message: "Forbidden" });

      await expect(
        store.deleteUser(),
      ).rejects.toBeAxiosErrorWithStatus(403);
    });
  });

  describe("updateUserData", () => {
    it("should update user data and localStorage", () => {
      const updateData = {
        name: "Updated Name",
        username: "updateduser",
        email: "updated@example.com",
        recovery_email: "recovery@example.com",
      };

      store.updateUserData(updateData);

      expect(store.name).toBe(updateData.name);
      expect(store.username).toBe(updateData.username);
      expect(store.email).toBe(updateData.email);
      expect(store.recoveryEmail).toBe(updateData.recovery_email);

      expect(localStorage.getItem("name")).toBe(updateData.name);
      expect(localStorage.getItem("user")).toBe(updateData.username);
      expect(localStorage.getItem("email")).toBe(updateData.email);
      expect(localStorage.getItem("recovery_email")).toBe(updateData.recovery_email);
    });

    it("should preserve existing values if not provided", () => {
      store.name = "Original Name";
      store.username = "originaluser";

      store.updateUserData({ email: "new@example.com" });

      expect(store.name).toBe("Original Name");
      expect(store.username).toBe("originaluser");
      expect(store.email).toBe("new@example.com");
    });
  });

  describe("enterInvitedNamespace", () => {
    const generateEnterNamespaceUrl = (tenantId: string) => `http://localhost:3000/api/auth/token/${tenantId}`;

    it("should switch to invited namespace", async () => {
      const targetTenantId = "invited-tenant";
      const mockResponse = {
        token: "new-token",
        user: "testuser",
        role: "observer",
      };

      mockNamespacesApi
        .onGet(generateEnterNamespaceUrl("invited-tenant"))
        .reply(200, mockResponse);

      await store.enterInvitedNamespace(targetTenantId);

      expect(store.token).toBe(mockResponse.token);
      expect(store.tenantId).toBe(targetTenantId);
      expect(store.role).toBe(mockResponse.role);
    });

    it("should handle error switching namespace", async () => {
      mockNamespacesApi
        .onGet(generateEnterNamespaceUrl("non-existent"))
        .reply(404, { message: "Namespace not found" });

      await expect(
        store.enterInvitedNamespace("non-existent"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });
  });

  describe("Computed Properties", () => {
    it("should compute showForceRecoveryMail when MFA enabled but no recovery email", () => {
      store.isMfaEnabled = true;
      store.recoveryEmail = "";

      expect(store.showForceRecoveryMail).toBe(true);
    });

    it("should not show force recovery mail when recovery email exists", () => {
      store.isMfaEnabled = true;
      store.recoveryEmail = "recovery@example.com";

      expect(store.showForceRecoveryMail).toBe(false);
    });

    it("should compute showRecoveryModal when recovering MFA", () => {
      store.isRecoveringMfa = true;
      store.isMfaEnabled = true;

      expect(store.showRecoveryModal).toBe(true);
    });
  });
});
