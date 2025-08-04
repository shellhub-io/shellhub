import { describe, it, expect, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useInstanceStore from "@admin/store/modules/instance";

describe("Instance Pinia Store", () => {
  let instanceStore: ReturnType<typeof useInstanceStore>;

  const mockAuthenticationSettings = {
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
        sign_requests: true,
        certificate: "cert",
      },
    },
  };

  beforeEach(() => {
    setActivePinia(createPinia());
    instanceStore = useInstanceStore();
  });

  describe("State", () => {
    it("should have the correct initial state", () => {
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
            sign_requests: false,
            certificate: "",
          },
        },
      });
    });
  });

  describe("Getters", () => {
    beforeEach(() => {
      instanceStore.authenticationSettings = mockAuthenticationSettings;
    });

    it("should return authenticationSettings", () => {
      const result = instanceStore.getAuthenticationSettings;
      expect(result).toEqual(mockAuthenticationSettings);
    });

    it("should return whether local authentication is enabled", () => {
      const result = instanceStore.isLocalAuthEnabled;
      expect(result).toBe(true);
    });

    it("should return whether SAML is enabled", () => {
      const result = instanceStore.isSamlEnabled;
      expect(result).toBe(true);
    });
  });
});
