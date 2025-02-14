import { describe, it, expect } from "vitest";
import { store } from "../../../../src/store";

const mockAuthenticationSettings = {
  local: {
    enabled: true,
  },
  saml: {
    enabled: true,
    auth_url: "https://example.com/auth",
    idp: {
      entity_id: "entity123",
      signon_url: "https://example.com/signon",
      certificates: ["cert123"],
    },
    sp: {
      sign_requests: true,
    },
  },
};

describe("Instance Vuex Module", () => {
  describe("State", () => {
    it("should have the correct initial state", () => {
      expect(store.state.instance.authenticationSettings).toEqual({
        local: {
          enabled: false,
        },
        saml: {
          enabled: false,
          auth_url: "",
          idp: {
            entity_id: "",
            signon_url: "",
            certificates: [],
          },
          sp: {
            sign_requests: false,
          },
        },
      });
    });
  });

  describe("Getters", () => {
    it("should return authenticationSettings", () => {
      store.commit("instance/setAuthenticationSettings", mockAuthenticationSettings);
      const result = store.getters["instance/authenticationSettings"];
      expect(result).toEqual(mockAuthenticationSettings);
    });

    it("should return whether local authentication is enabled", () => {
      store.commit("instance/setAuthenticationSettings", mockAuthenticationSettings);
      const result = store.getters["instance/isLocalAuthEnabled"];
      expect(result).toBe(true);
    });

    it("should return whether SAML is enabled", () => {
      store.commit("instance/setAuthenticationSettings", mockAuthenticationSettings);
      const result = store.getters["instance/isSamlEnabled"];
      expect(result).toBe(true);
    });
  });

  describe("Mutations", () => {
    it("should set authentication settings", () => {
      store.commit("instance/setAuthenticationSettings", mockAuthenticationSettings);
      expect(store.state.instance.authenticationSettings).toEqual(mockAuthenticationSettings);
    });
  });
});
