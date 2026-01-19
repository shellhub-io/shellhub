import { describe, expect, it, vi, afterEach, beforeEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useInstanceStore from "@admin/store/modules/instance";
import SettingsAuthentication from "@admin/components/Settings/SettingsAuthentication.vue";
import { mockAuthSettings, mockAuthSettingsLocalOnly } from "../../mocks";

// Mock window.open
const mockWindowOpen = vi.fn();
Object.defineProperty(window, "open", {
  writable: true,
  value: mockWindowOpen,
});

// Mock URL methods
const mockCreateObjectURL = vi.fn(() => "blob:mock-url");
const mockRevokeObjectURL = vi.fn();
global.URL.createObjectURL = mockCreateObjectURL;
global.URL.revokeObjectURL = mockRevokeObjectURL;

// Mock document.createElement for download
const mockClick = vi.fn();
const mockCreateElement = document.createElement.bind(document);
document.createElement = vi.fn((tagName: string) => {
  const element = mockCreateElement(tagName);
  if (tagName === "a") element.click = mockClick;
  return element;
});

describe("SettingsAuthentication", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingsAuthentication>>;
  let instanceStore: ReturnType<typeof useInstanceStore>;

  const mountWrapper = async (authSettings = mockAuthSettings) => {
    wrapper = mountComponent(SettingsAuthentication, {
      piniaOptions: {
        initialState: {
          adminInstance: { authenticationSettings: authSettings },
        },
      },
    });

    instanceStore = useInstanceStore();
    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders page header", () => {
      const header = wrapper.find('[title-test-id="auth-header"]');
      expect(header.exists()).toBe(true);
      expect(header.text()).toContain("Authentication");
    });

    it("renders authentication status section", () => {
      const statusHeader = wrapper.find('[data-test="auth-status-header"]');
      expect(statusHeader.exists()).toBe(true);
      expect(statusHeader.text()).toBe("Authentication Status");
    });

    it("renders local authentication switch", () => {
      const localAuthSwitch = wrapper.find('[data-test="local-auth-switch"]');
      expect(localAuthSwitch.exists()).toBe(true);
    });

    it("renders SAML authentication switch", () => {
      const samlAuthSwitch = wrapper.find('[data-test="saml-auth-switch"]');
      expect(samlAuthSwitch.exists()).toBe(true);
    });

    it("shows SSO section when SAML is enabled", () => {
      const ssoHeader = wrapper.find('[data-test="sso-header"]');
      expect(ssoHeader.exists()).toBe(true);
      expect(ssoHeader.text()).toBe("Single Sign-on (SSO)");
    });

    it("displays IdP SignOn POST URL", () => {
      const postUrl = wrapper.find('[data-test="idp-signon-post-value"]');
      expect(postUrl.exists()).toBe(true);
      expect(postUrl.text()).toBe(mockAuthSettings.saml.idp.binding.post);
    });

    it("displays IdP SignOn Redirect URL", () => {
      const redirectUrl = wrapper.find('[data-test="idp-signon-redirect-value"]');
      expect(redirectUrl.exists()).toBe(true);
      expect(redirectUrl.text()).toBe(mockAuthSettings.saml.idp.binding.redirect);
    });

    it("displays IdP Entity ID", () => {
      const entityId = wrapper.find('[data-test="idp-entity-value"]');
      expect(entityId.exists()).toBe(true);
      expect(entityId.text()).toBe(mockAuthSettings.saml.idp.entity_id);
    });

    it("displays SP certificate download button when certificate exists", () => {
      const downloadBtn = wrapper.find('[data-test="download-certificate-btn"]');
      expect(downloadBtn.exists()).toBe(true);
    });

    it("hides SSO section when SAML is disabled", async () => {
      wrapper.unmount();
      await mountWrapper(mockAuthSettingsLocalOnly);

      const ssoHeader = wrapper.find('[data-test="sso-header"]');
      expect(ssoHeader.exists()).toBe(false);
    });
  });

  describe("initial data loading", () => {
    it("fetches authentication settings on mount", async () => {
      await mountWrapper();
      expect(instanceStore.fetchAuthenticationSettings).toHaveBeenCalled();
    });
  });

  describe("local authentication toggle", () => {
    beforeEach(() => mountWrapper());

    it("calls updateLocalAuthentication when toggling switch", async () => {
      const localAuthSwitch = wrapper.find('[data-test="local-auth-switch"] input');
      await localAuthSwitch.trigger("click");
      await flushPromises();

      expect(instanceStore.updateLocalAuthentication).toHaveBeenCalledWith(false);
    });

    it("shows error when trying to disable all authentication methods", async () => {
      vi.mocked(instanceStore.updateLocalAuthentication).mockRejectedValueOnce(
        createAxiosError(400, "Bad Request"),
      );

      const localAuthSwitch = wrapper.find('[data-test="local-auth-switch"] input');
      await localAuthSwitch.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "You cannot disable all authentication methods.",
      );
    });

    it("shows generic error for other failures", async () => {
      vi.mocked(instanceStore.updateLocalAuthentication).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const localAuthSwitch = wrapper.find('[data-test="local-auth-switch"] input');
      await localAuthSwitch.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "An error occurred while updating local authentication.",
      );
    });
  });

  describe("SAML authentication toggle", () => {
    beforeEach(() => mountWrapper());

    it("disables SAML when toggling off", async () => {
      const samlAuthSwitch = wrapper.find('[data-test="saml-auth-switch"] input');
      await samlAuthSwitch.trigger("click");
      await flushPromises();

      expect(instanceStore.updateSamlAuthentication).toHaveBeenCalledWith({
        enable: false,
        idp: {
          entity_id: "",
          binding: { post: "", redirect: "" },
          certificate: "",
        },
        sp: { sign_requests: false },
      });
    });

    it("opens SSO dialog when enabling SAML", async () => {
      wrapper.unmount();
      await mountWrapper(mockAuthSettingsLocalOnly);

      const samlAuthSwitch = wrapper.find('[data-test="saml-auth-switch"] input');
      await samlAuthSwitch.trigger("click");
      await flushPromises();

      const dialog = new DOMWrapper(document.body).find('[data-test="configure-sso-dialog"]');
      expect(dialog.exists()).toBe(true);
    });

    it("shows error when trying to disable all authentication methods", async () => {
      vi.mocked(instanceStore.updateSamlAuthentication).mockRejectedValueOnce(
        createAxiosError(400, "Bad Request"),
      );

      const samlAuthSwitch = wrapper.find('[data-test="saml-auth-switch"] input');
      await samlAuthSwitch.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "You cannot disable all authentication methods.",
      );
    });

    it("shows generic error for other failures", async () => {
      vi.mocked(instanceStore.updateSamlAuthentication).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const samlAuthSwitch = wrapper.find('[data-test="saml-auth-switch"] input');
      await samlAuthSwitch.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "An error occurred while updating local authentication.",
      );
    });
  });

  describe("SSO configuration", () => {
    beforeEach(() => mountWrapper());

    it("opens SSO dialog when clicking configure button", async () => {
      const configBtn = wrapper.find('[data-test="sso-config-btn"]');
      await configBtn.trigger("click");
      await flushPromises();

      const dialog = new DOMWrapper(document.body).find('[data-test="configure-sso-dialog"]');
      expect(dialog.exists()).toBe(true);
    });

    it("shows 'Edit' text when SAML is enabled", () => {
      const configBtn = wrapper.find('[data-test="sso-config-btn"]');
      expect(configBtn.text()).toBe("Edit");
    });
  });

  describe("SP certificate download", () => {
    beforeEach(() => mountWrapper());

    it("downloads certificate when clicking download button", async () => {
      const downloadBtn = wrapper.find('[data-test="download-certificate-btn"]');
      await downloadBtn.trigger("click");
      await flushPromises();

      expect(mockCreateObjectURL).toHaveBeenCalled();
      expect(mockClick).toHaveBeenCalled();
      expect(mockRevokeObjectURL).toHaveBeenCalledWith("blob:mock-url");
    });

    it("shows error when no certificate is available", async () => {
      wrapper.unmount();
      const settingsWithoutCert = {
        ...mockAuthSettings,
        saml: {
          ...mockAuthSettings.saml,
          sp: {
            ...mockAuthSettings.saml.sp,
            certificate: "",
          },
        },
      };
      await mountWrapper(settingsWithoutCert);

      const downloadBtn = wrapper.find('[data-test="download-certificate-btn"]');
      expect(downloadBtn.exists()).toBe(false);
    });
  });

  describe("authentication URL redirect", () => {
    it("opens authentication URL in new tab when clicking test button", async () => {
      await mountWrapper();
      const redirectBtn = wrapper.find('[data-test="redirect-auth-btn"]');
      await redirectBtn.trigger("click");
      await flushPromises();

      expect(mockWindowOpen).toHaveBeenCalledWith(mockAuthSettings.saml.auth_url, "_blank");
    });
  });
});
