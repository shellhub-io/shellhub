import { beforeEach, describe, it, expect, vi } from "vitest";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import useInstanceStore from "@admin/store/modules/instance";
import SettingsAuthentication from "@admin/components/Settings/SettingsAuthentication.vue";
import routes from "@admin/router";
import { adminApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

window.matchMedia = vi.fn().mockImplementation((query) => ({
  matches: false,
  media: query,
  onchange: null,
  addListener: vi.fn(),
  removeListener: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
  dispatchEvent: vi.fn(),
}));

type SettingsAuthenticationWrapper = VueWrapper<InstanceType<typeof SettingsAuthentication>>;

const authData = {
  local: {
    enabled: false,
  },
  saml: {
    enabled: true,
    auth_url: "https://auth.example.com",
    assertion_url: "http://example/api/user/saml/auth",
    idp: {
      entity_id: "entity-id-example",
      binding: {
        post: "https://example.com/signon-post",
        redirect: "https://example.com/signon-redirect",
      },
      certificates: ["certificate-string"],
      mappings: {
        email: "emailAddress",
        name: "displayName",
      },
    },
    sp: {
      sign_auth_requests: true,
      certificate: "test",
    },
  },
};

describe("Authentication", () => {
  let wrapper: SettingsAuthenticationWrapper;
  const mockAdminApi = new MockAdapter(adminApi.getAxios());
  setActivePinia(createPinia());
  const instanceStore = useInstanceStore();
  const vuetify = createVuetify();

  beforeEach(() => {
    mockAdminApi.onGet("http://localhost:3000/admin/api/authentication").reply(200, authData);

    vi.spyOn(instanceStore, "fetchAuthenticationSettings").mockResolvedValue(undefined);
    vi.spyOn(instanceStore, "updateLocalAuthentication").mockResolvedValue(undefined);
    vi.spyOn(instanceStore, "updateSamlAuthentication").mockResolvedValue(undefined);

    instanceStore.authenticationSettings = authData;

    wrapper = mount(SettingsAuthentication, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("renders correctly", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("shows the SSO dialog when 'Configure' is clicked", async () => {
    await wrapper.find("[data-test='sso-config-btn']").trigger("click");
    expect(wrapper.vm.showSSODialog).toBe(true);
  });

  it("calls updateLocalAuthentication when clicking switch", async () => {
    const spy = vi.spyOn(instanceStore, "updateLocalAuthentication");

    await wrapper.find("[data-test='local-auth-switch']").trigger("click");

    expect(spy).toHaveBeenCalledWith(true);
  });

  it("renders SAML settings when enabled", () => {
    expect(wrapper.find("[data-test='idp-signon-post-value']").exists()).toBe(true);
    expect(wrapper.find("[data-test='idp-signon-redirect-value']").exists()).toBe(true);
    expect(wrapper.find("[data-test='idp-entity-value']").exists()).toBe(true);
  });

  it("renders SP certificate button when certificate exists", () => {
    expect(wrapper.find("[data-test='download-certificate-btn']").exists()).toBe(true);
  });

  it("opens authentication URL in new tab when 'Redirect' is clicked", async () => {
    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    await wrapper.find("[data-test='redirect-auth-btn']").trigger("click");
    expect(openSpy).toHaveBeenCalledWith(authData.saml.auth_url, "_blank");
  });
});
