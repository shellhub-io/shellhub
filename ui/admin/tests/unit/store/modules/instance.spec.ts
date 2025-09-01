import { describe, it, expect } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useInstanceStore from "@admin/store/modules/instance";

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
      sign_auth_requests: true,
      certificate: "cert",
    },
  },
};

describe("Instance Pinia Store", () => {
  setActivePinia(createPinia());
  const instanceStore = useInstanceStore();

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
          sign_auth_requests: false,
          certificate: "",
        },
      },
    });
  });

  it("should return authenticationSettings", () => {
    instanceStore.authenticationSettings = mockAuthenticationSettings;
    expect(instanceStore.authenticationSettings).toEqual(mockAuthenticationSettings);
  });

  it("should return whether local authentication is enabled", () => {
    expect(instanceStore.isLocalAuthEnabled).toBe(true);
  });

  it("should return whether SAML is enabled", () => {
    expect(instanceStore.isSamlEnabled).toBe(true);
  });
});
