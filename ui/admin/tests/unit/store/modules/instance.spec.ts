import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import useInstanceStore from "@admin/store/modules/instance";
import { IAdminAuth, IAdminUpdateSAML } from "@admin/interfaces/IInstance";

const mockAuthenticationSettings: IAdminAuth = {
  local: {
    enabled: true,
  },
  saml: {
    enabled: true,
    auth_url: "https://example.com/auth",
    assertion_url: "https://example.com/assertion",
    idp: {
      entity_id: "entity123",
      binding: {
        post: "https://example.com/signon-post",
        redirect: "https://example.com/signon-redirect",
      },
      certificates: ["cert123"],
    },
    sp: {
      sign_auth_requests: true,
      certificate: "cert-sp-123",
    },
  },
};

const mockSAMLUpdate: IAdminUpdateSAML = {
  enable: false,
  idp: {
    entity_id: "new-entity",
    binding: {
      post: "https://new-url.com/post",
      redirect: "https://new-url.com/redirect",
    },
  },
  sp: {
    sign_requests: false,
  },
};

describe("Admin Instance Store", () => {
  let instanceStore: ReturnType<typeof useInstanceStore>;
  let mockAdminApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    instanceStore = useInstanceStore();
    mockAdminApi = new MockAdapter(adminApi.getAxios());
  });

  afterEach(() => { mockAdminApi.reset(); });

  describe("Initial State", () => {
    it("should have default authentication settings", () => {
      expect(instanceStore.authenticationSettings).toEqual({
        local: { enabled: false },
        saml: {
          enabled: false,
          auth_url: "",
          assertion_url: "",
          idp: {
            entity_id: "",
            binding: {
              post: "",
              redirect: "",
            },
            certificates: [],
          },
          sp: {
            sign_auth_requests: false,
            certificate: "",
          },
        },
      });
    });

    it("should have isLocalAuthEnabled as false", () => {
      expect(instanceStore.isLocalAuthEnabled).toBe(false);
    });

    it("should have isSamlEnabled as false", () => {
      expect(instanceStore.isSamlEnabled).toBe(false);
    });
  });

  describe("Computed Properties", () => {
    it("should compute isLocalAuthEnabled correctly when enabled", () => {
      instanceStore.authenticationSettings = {
        ...instanceStore.authenticationSettings,
        local: { enabled: true },
      };

      expect(instanceStore.isLocalAuthEnabled).toBe(true);
    });

    it("should compute isSamlEnabled correctly when enabled", () => {
      instanceStore.authenticationSettings = {
        ...instanceStore.authenticationSettings,
        saml: {
          ...instanceStore.authenticationSettings.saml,
          enabled: true,
        },
      };

      expect(instanceStore.isSamlEnabled).toBe(true);
    });
  });

  describe("fetchAuthenticationSettings", () => {
    const baseUrl = "http://localhost:3000/admin/api/authentication";

    it("should fetch authentication settings successfully and update state", async () => {
      mockAdminApi.onGet(baseUrl).reply(200, mockAuthenticationSettings);

      await instanceStore.fetchAuthenticationSettings();

      expect(instanceStore.authenticationSettings).toEqual(mockAuthenticationSettings);
      expect(instanceStore.isLocalAuthEnabled).toBe(true);
      expect(instanceStore.isSamlEnabled).toBe(true);
    });

    it("should throw on server error when fetching settings", async () => {
      mockAdminApi.onGet(baseUrl).reply(500);

      await expect(instanceStore.fetchAuthenticationSettings()).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when fetching settings", async () => {
      mockAdminApi.onGet(baseUrl).networkError();

      await expect(instanceStore.fetchAuthenticationSettings()).rejects.toThrow("Network Error");
    });
  });

  describe("updateLocalAuthentication", () => {
    const updateUrl = "http://localhost:3000/admin/api/authentication/local";
    const fetchUrl = "http://localhost:3000/admin/api/authentication";

    it("should update local authentication and refresh settings", async () => {
      const updatedSettings = {
        ...mockAuthenticationSettings,
        local: { enabled: false },
      };

      mockAdminApi.onPut(updateUrl, { enable: false }).reply(200);
      mockAdminApi.onGet(fetchUrl).reply(200, updatedSettings);

      await instanceStore.updateLocalAuthentication(false);

      expect(instanceStore.authenticationSettings).toEqual(updatedSettings);
      expect(instanceStore.isLocalAuthEnabled).toBe(false);
    });

    it("should throw on server error when updating local auth", async () => {
      mockAdminApi.onPut(updateUrl, { enable: true }).reply(500);

      await expect(instanceStore.updateLocalAuthentication(true)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when updating local auth", async () => {
      mockAdminApi.onPut(updateUrl, { enable: true }).networkError();

      await expect(instanceStore.updateLocalAuthentication(true)).rejects.toThrow("Network Error");
    });
  });

  describe("updateSamlAuthentication", () => {
    const updateUrl = "http://localhost:3000/admin/api/authentication/saml";
    const fetchUrl = "http://localhost:3000/admin/api/authentication";

    it("should update SAML authentication and refresh settings", async () => {
      const updatedSettings = {
        ...mockAuthenticationSettings,
        saml: {
          ...mockSAMLUpdate,
        },
      };

      // Payload sends "enable", but response uses "enabled"
      const updatedSettingsResponse = { ...updatedSettings, saml: { ...updatedSettings.saml, enabled: false } };

      mockAdminApi.onPut(updateUrl, mockSAMLUpdate).reply(200);
      mockAdminApi.onGet(fetchUrl).reply(200, updatedSettingsResponse);

      await instanceStore.updateSamlAuthentication(mockSAMLUpdate);

      expect(instanceStore.authenticationSettings).toEqual(updatedSettingsResponse);
      expect(instanceStore.isSamlEnabled).toBe(false);
    });

    it("should throw on server error when updating SAML auth", async () => {
      mockAdminApi.onPut(updateUrl, mockSAMLUpdate).reply(500);

      await expect(instanceStore.updateSamlAuthentication(mockSAMLUpdate)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when updating SAML auth", async () => {
      mockAdminApi.onPut(updateUrl, mockSAMLUpdate).networkError();

      await expect(instanceStore.updateSamlAuthentication(mockSAMLUpdate)).rejects.toThrow("Network Error");
    });
  });
});
