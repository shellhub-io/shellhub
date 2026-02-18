import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { usersApi, systemApi } from "@/api/http";
import useUsersStore from "@/store/modules/users";
import type {
  IUserSignUp,
  IUserPatch,
  IUserPatchPassword,
  IUserSetup,
  IUserUpdatePassword,
  IUserSystemInfo,
  IPremiumFeature,
} from "@/interfaces/IUser";

const mockPersistAuth = vi.fn();
vi.mock("@/store/modules/auth", () => ({
  default: () => ({ persistAuth: mockPersistAuth }),
}));

describe("Users Store", () => {
  let usersStore: ReturnType<typeof useUsersStore>;
  let mockUsersApi: MockAdapter;
  let mockSystemApi: MockAdapter;

  beforeEach(() => {
    mockPersistAuth.mockClear();
    setActivePinia(createPinia());
    usersStore = useUsersStore();
    mockUsersApi = new MockAdapter(usersApi.getAxios());
    mockSystemApi = new MockAdapter(systemApi.getAxios());
  });

  afterEach(() => {
    mockUsersApi.reset();
    mockSystemApi.reset();
    vi.clearAllMocks();
  });

  describe("Initial State", () => {
    it("should have showPaywall as false", () => {
      expect(usersStore.showPaywall).toBe(false);
    });

    it("should have signUpToken as undefined", () => {
      expect(usersStore.signUpToken).toBeUndefined();
    });

    it("should have empty systemInfo object", () => {
      expect(usersStore.systemInfo).toEqual({});
    });
  });

  describe("signUp", () => {
    const mockSignUpData: IUserSignUp = {
      name: "John Doe",
      email: "john@example.com",
      username: "johndoe",
      password: "password123",
      emailMarketing: true,
      sig: "signature",
    };
    const signUpUrl = "http://localhost:3000/api/register";

    it("should sign up user successfully and return token", async () => {
      const mockToken = "jwt-token-123";
      mockUsersApi.onPost(signUpUrl).reply(200, { token: mockToken });
      const result = await usersStore.signUp(mockSignUpData);

      expect(result).toBe(mockToken);
      expect(usersStore.signUpToken).toBe(mockToken);
      expect(mockPersistAuth).toHaveBeenCalledWith({ token: mockToken });
    });

    it("should return false when token is not provided", async () => {
      mockUsersApi.onPost(signUpUrl).reply(200, {});

      const result = await usersStore.signUp(mockSignUpData);

      expect(result).toBe(false);
      expect(usersStore.signUpToken).toBeUndefined();
      expect(mockPersistAuth).not.toHaveBeenCalled();
    });

    it("should throw error when signup fails with 400", async () => {
      mockUsersApi.onPost(signUpUrl).reply(400, { message: "User already exists" });

      await expect(usersStore.signUp(mockSignUpData)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should throw error when network fails", async () => {
      mockUsersApi.onPost(signUpUrl).networkError();

      await expect(usersStore.signUp(mockSignUpData)).rejects.toThrow();
    });
  });

  describe("patchData", () => {
    const mockPatchData: IUserPatch = {
      name: "Jane Doe",
      username: "janedoe",
      email: "jane@example.com",
      recovery_email: "recovery@example.com",
    };
    const patchDataUrl = "http://localhost:3000/api/users";

    it("should update user data successfully", async () => {
      mockUsersApi.onPatch(patchDataUrl).reply(200);

      await expect(usersStore.patchData(mockPatchData)).resolves.not.toThrow();
    });

    it("should throw error when update fails with 400", async () => {
      mockUsersApi.onPatch(patchDataUrl).reply(400, { message: "Invalid data" });

      await expect(usersStore.patchData(mockPatchData)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should throw error when network fails", async () => {
      mockUsersApi.onPatch(patchDataUrl).networkError();

      await expect(usersStore.patchData(mockPatchData)).rejects.toThrow();
    });
  });

  describe("setup", () => {
    const mockSetupData: IUserSetup = {
      sign: "signature",
      email: "admin@example.com",
      name: "Admin User",
      password: "adminpass123",
      username: "admin",
    };
    const setupUrl = "http://localhost:3000/api/setup?sign=signature";

    it("should setup system successfully", async () => {
      mockSystemApi.onPost(setupUrl).reply(200);

      await expect(usersStore.setup(mockSetupData)).resolves.not.toThrow();
    });

    it("should throw error when setup fails with 400", async () => {
      mockSystemApi.onPost(setupUrl).reply(400, { message: "Setup already completed" });

      await expect(usersStore.setup(mockSetupData)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should throw error when network fails", async () => {
      mockSystemApi.onPost(setupUrl).networkError();

      await expect(usersStore.setup(mockSetupData)).rejects.toThrow();
    });
  });

  describe("patchPassword", () => {
    const mockPasswordData: IUserPatchPassword = {
      name: "John Doe",
      username: "johndoe",
      email: "john@example.com",
      recovery_email: "recovery@example.com",
      currentPassword: "oldpass123",
      newPassword: "newpass456",
    };
    const patchPasswordUrl = "http://localhost:3000/api/users";

    it("should update password successfully", async () => {
      mockUsersApi.onPatch(patchPasswordUrl).reply(200);

      await expect(usersStore.patchPassword(mockPasswordData)).resolves.not.toThrow();
    });

    it("should throw error when password update fails with 403", async () => {
      mockUsersApi.onPatch(patchPasswordUrl).reply(403, { message: "Current password is incorrect" });

      await expect(usersStore.patchPassword(mockPasswordData)).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should throw error when network fails", async () => {
      mockUsersApi.onPatch(patchPasswordUrl).networkError();

      await expect(usersStore.patchPassword(mockPasswordData)).rejects.toThrow();
    });
  });

  describe("resendEmail", () => {
    const resendEmailUrl = "http://localhost:3000/api/user/resend_email";

    it("should resend email successfully", async () => {
      mockUsersApi.onPost(resendEmailUrl).reply(200);

      await expect(usersStore.resendEmail("johndoe")).resolves.not.toThrow();
    });

    it("should throw error when resend fails with 404", async () => {
      mockUsersApi.onPost(resendEmailUrl).reply(404, { message: "User not found" });

      await expect(usersStore.resendEmail("johndoe")).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw error when network fails", async () => {
      mockUsersApi.onPost(resendEmailUrl).networkError();

      await expect(usersStore.resendEmail("johndoe")).rejects.toThrow();
    });
  });

  describe("recoverPassword", () => {
    const recoverPasswordUrl = "http://localhost:3000/api/user/recover_password";

    it("should send password recovery email successfully", async () => {
      mockUsersApi.onPost(recoverPasswordUrl).reply(200);

      await expect(usersStore.recoverPassword("johndoe")).resolves.not.toThrow();
    });

    it("should throw error when recovery fails with 404", async () => {
      mockUsersApi.onPost(recoverPasswordUrl).reply(404, { message: "User not found" });

      await expect(usersStore.recoverPassword("johndoe")).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw error when network fails", async () => {
      mockUsersApi.onPost(recoverPasswordUrl).networkError();

      await expect(usersStore.recoverPassword("johndoe")).rejects.toThrow();
    });
  });

  describe("validateAccount", () => {
    const mockValidationData = {
      email: "john@example.com",
      token: "validation-token-123",
    };
    const validateAccountUrl = "http://localhost:3000/api/user/validation_account?email=john%40example.com&token=validation-token-123";

    it("should validate account successfully", async () => {
      mockUsersApi.onGet(validateAccountUrl).reply(200);

      await expect(usersStore.validateAccount(mockValidationData)).resolves.not.toThrow();
    });

    it("should throw error when validation fails with 400", async () => {
      mockUsersApi.onGet(validateAccountUrl).reply(400, { message: "Invalid token" });

      await expect(usersStore.validateAccount(mockValidationData)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should throw error when network fails", async () => {
      mockUsersApi.onGet(validateAccountUrl).networkError();

      await expect(usersStore.validateAccount(mockValidationData)).rejects.toThrow();
    });
  });

  describe("updatePassword", () => {
    const mockUpdatePasswordData: IUserUpdatePassword = {
      id: "user-id-123",
      token: "reset-token-456",
      password: "newpassword789",
    };
    const updatePasswordUrl = "http://localhost:3000/api/user/user-id-123/update_password";

    it("should update password with token successfully", async () => {
      mockUsersApi.onPost(updatePasswordUrl).reply(200);

      await expect(usersStore.updatePassword(mockUpdatePasswordData)).resolves.not.toThrow();
    });

    it("should throw error when password update fails with 400", async () => {
      mockUsersApi.onPost(updatePasswordUrl).reply(400, { message: "Invalid token" });

      await expect(usersStore.updatePassword(mockUpdatePasswordData)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should throw error when network fails", async () => {
      mockUsersApi.onPost(updatePasswordUrl).networkError();

      await expect(usersStore.updatePassword(mockUpdatePasswordData)).rejects.toThrow();
    });
  });

  describe("getPremiumContent", () => {
    const mockPremiumFeatures: IPremiumFeature[] = [
      {
        title: "Enterprise Features",
        features: ["Feature 1", "Feature 2"],
        button: {
          link: "https://example.com",
          label: "Learn More",
        },
      },
    ];

    it("should fetch premium content successfully", async () => {
      global.fetch = vi.fn().mockResolvedValue({
        json: () => mockPremiumFeatures,
      });

      const result = await usersStore.getPremiumContent();

      expect(result).toEqual(mockPremiumFeatures);
      expect(global.fetch).toHaveBeenCalledWith("https://static.shellhub.io/premium-features.v1.json");
    });

    it("should throw error when fetch fails", async () => {
      global.fetch = vi.fn().mockRejectedValue(new Error("Network error"));

      await expect(usersStore.getPremiumContent()).rejects.toThrow("Network error");
    });

    it("should throw error when JSON parsing fails", async () => {
      global.fetch = vi.fn().mockResolvedValue({
        json: () => {
          throw new Error("Invalid JSON");
        },
      });

      await expect(usersStore.getPremiumContent()).rejects.toThrow("Invalid JSON");
    });
  });

  describe("fetchSystemInfo", () => {
    const mockSystemInfo = {
      authentication: {
        local: false,
        saml: true,
      },
      endpoints: {
        api: "localhost:80",
        ssh: "localhost:22",
      },
      setup: true,
      version: "v0.21.4",
    };
    const systemInfoUrl = "http://localhost:3000/info";

    it("should fetch system info successfully", async () => {
      mockSystemApi.onGet(systemInfoUrl).reply(200, mockSystemInfo);

      await usersStore.fetchSystemInfo();

      expect(usersStore.systemInfo).toEqual(mockSystemInfo);
    });

    it("should update systemInfo state with fetched data", async () => {
      const updatedInfo: IUserSystemInfo = {
        ...mockSystemInfo,
        version: "1.0.0",
      };
      mockSystemApi.onGet(systemInfoUrl).reply(200, updatedInfo);

      await usersStore.fetchSystemInfo();

      expect(usersStore.systemInfo.version).toBe("1.0.0");
    });

    it("should throw error when network fails", async () => {
      mockSystemApi.onGet(systemInfoUrl).networkError();

      await expect(usersStore.fetchSystemInfo()).rejects.toThrow();
    });
  });

  describe("checkHealth", () => {
    const healthcheckUrl = "http://localhost:3000/healthcheck";

    it("should resolve successfully when the API is healthy", async () => {
      mockSystemApi.onGet(healthcheckUrl).reply(200);

      await expect(usersStore.checkHealth()).resolves.not.toThrow();
    });

    it("should throw error when the API returns 503", async () => {
      mockSystemApi.onGet(healthcheckUrl).reply(503);

      await expect(usersStore.checkHealth()).rejects.toBeAxiosErrorWithStatus(503);
    });

    it("should throw error when network fails", async () => {
      mockSystemApi.onGet(healthcheckUrl).networkError();

      await expect(usersStore.checkHealth()).rejects.toThrow();
    });
  });

  describe("getSamlUrl", () => {
    const samlUrl = "http://localhost:3000/api/user/saml/auth";

    it("should get SAML URL successfully", async () => {
      const mockSamlUrl = "https://saml.example.com/login";
      mockUsersApi.onGet(samlUrl).reply(200, { url: mockSamlUrl });

      const result = await usersStore.getSamlUrl();

      expect(result).toBe(mockSamlUrl);
    });

    it("should throw error when SAML URL fetch fails with 400", async () => {
      mockUsersApi.onGet(samlUrl).reply(400, { message: "SAML not configured" });
      await expect(usersStore.getSamlUrl()).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should throw error when network fails", async () => {
      mockUsersApi.onGet(samlUrl).networkError();
      await expect(usersStore.getSamlUrl()).rejects.toThrow();
    });
  });
});
