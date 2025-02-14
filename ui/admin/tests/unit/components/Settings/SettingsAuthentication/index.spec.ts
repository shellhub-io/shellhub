import { beforeEach, describe, it, expect, vi } from "vitest";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { adminApi } from "../../../../../src/api/http";
import { store, key } from "../../../../../src/store";
import SettingsAuthentication from "../../../../../src/components/Settings/SettingsAuthentication.vue";

// Mock for clipboard
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn(),
  },
});

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
      signon_url: "https://signon.example.com",
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

  const vuetify = createVuetify();
  let mockAdminApi: MockAdapter;

  beforeEach(async () => {
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

    mockAdminApi = new MockAdapter(adminApi.getAxios());
    mockAdminApi.onGet("http://localhost:3000/admin/api/authentication").reply(200, authData);

    wrapper = mount(SettingsAuthentication, {
      global: {
        plugins: [[store, key], vuetify],
      },
    });

    await flushPromises();
  });

  it("is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("renders correctly", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("shows the SSO dialog when 'Configure' is clicked", async () => {
    await wrapper.findComponent("[data-test='sso-config-btn']").trigger("click");
    expect(wrapper.vm.dialogSSO).toBe(true);
  });

  it("disables Local Authentication switch when action fails", async () => {
    const errorSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent("[data-test='local-auth-switch']").trigger("click");

    expect(errorSpy).toHaveBeenCalledWith("instance/updateLocalAuthentication", true);
  });

  it("renders SAML settings when enabled", async () => {
    wrapper.vm.samlEnabled = true;
    await flushPromises();

    expect(wrapper.findComponent("[data-test='idp-signon-value']").exists()).toBe(true);
    expect(wrapper.findComponent("[data-test='idp-entity-value']").exists()).toBe(true);
  });

  it("renders SP certificate when it has value", async () => {
    expect(wrapper.findComponent("[data-test='download-certificate-btn']").exists()).toBe(true);
  });

  it("copies Assertion URL to clipboard when 'Copy URL' button is clicked", async () => {
    const copyBtn = wrapper.find("[data-test='copy-assertion-btn']");
    expect(copyBtn.exists()).toBe(true);

    await copyBtn.trigger("click");

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(authData.saml.assertion_url);
  });

  it("redirects to Authentication URL when 'Redirect' button is clicked", async () => {
    const redirectBtn = wrapper.find("[data-test='redirect-auth-btn']");
    expect(redirectBtn.exists()).toBe(true);

    const windowOpenSpy = vi.spyOn(window, "open");
    await redirectBtn.trigger("click");

    expect(windowOpenSpy).toHaveBeenCalledWith(authData.saml.auth_url, "_blank");
  });
});
