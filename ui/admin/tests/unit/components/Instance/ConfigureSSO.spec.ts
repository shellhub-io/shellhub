import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useInstanceStore from "@admin/store/modules/instance";
import ConfigureSSO from "@admin/components/Instance/SSO/ConfigureSSO.vue";

// Mock certificate from https://mocksaml.com/
const validCertificate = `
-----BEGIN CERTIFICATE-----
MIIC4jCCAcoCCQC33wnybT5QZDANBgkqhkiG9w0BAQsFADAyMQswCQYDVQQGEwJV
SzEPMA0GA1UECgwGQm94eUhRMRIwEAYDVQQDDAlNb2NrIFNBTUwwIBcNMjIwMjI4
MjE0NjM4WhgPMzAyMTA3MDEyMTQ2MzhaMDIxCzAJBgNVBAYTAlVLMQ8wDQYDVQQK
DAZCb3h5SFExEjAQBgNVBAMMCU1vY2sgU0FNTDCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBALGfYettMsct1T6tVUwTudNJH5Pnb9GGnkXi9Zw/e6x45DD0
RuRONbFlJ2T4RjAE/uG+AjXxXQ8o2SZfb9+GgmCHuTJFNgHoZ1nFVXCmb/Hg8Hpd
4vOAGXndixaReOiq3EH5XvpMjMkJ3+8+9VYMzMZOjkgQtAqO36eAFFfNKX7dTj3V
pwLkvz6/KFCq8OAwY+AUi4eZm5J57D31GzjHwfjH9WTeX0MyndmnNB1qV75qQR3b
2/W5sGHRv+9AarggJkF+ptUkXoLtVA51wcfYm6hILptpde5FQC8RWY1YrswBWAEZ
NfyrR4JeSweElNHg4NVOs4TwGjOPwWGqzTfgTlECAwEAATANBgkqhkiG9w0BAQsF
AAOCAQEAAYRlYflSXAWoZpFfwNiCQVE5d9zZ0DPzNdWhAybXcTyMf0z5mDf6FWBW
5Gyoi9u3EMEDnzLcJNkwJAAc39Apa4I2/tml+Jy29dk8bTyX6m93ngmCgdLh5Za4
khuU3AM3L63g7VexCuO7kwkjh/+LqdcIXsVGO6XDfu2QOs1Xpe9zIzLpwm/RNYeX
UjbSj5ce/jekpAw7qyVVL4xOyh8AtUW1ek3wIw1MJvEgEPt0d16oshWJpoS1OT8L
r/22SvYEo3EmSGdTVGgk3x3s+A0qWAqTcyjr7Q4s/GKYRFfomGwz0TZ4Iw1ZN99M
m0eo2USlSRTVl7QHRTuiuSThHpLKQQ==
-----END CERTIFICATE-----
`;

describe("ConfigureSSO", () => {
  let wrapper: VueWrapper<InstanceType<typeof ConfigureSSO>>;
  let instanceStore: ReturnType<typeof useInstanceStore>;

  const mountWrapper = (isSamlEnabled = false) => {
    wrapper = mountComponent(ConfigureSSO, {
      props: { modelValue: true },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          adminInstance: {
            authenticationSettings: {
              local: { enabled: true },
              saml: {
                enabled: isSamlEnabled,
                idp: {
                  entity_id: "",
                  binding: { post: "", redirect: "" },
                  certificates: [],
                  mappings: undefined,
                },
                sp: { sign_auth_requests: false },
              },
            },
          },
        },
      },
    });

    instanceStore = useInstanceStore();
  };

  const getDialog = () => new DOMWrapper(document.body);

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering and dialog visibility", () => {
    beforeEach(() => mountWrapper());

    it("renders the dialog when modelValue is true", async () => {
      await flushPromises();
      const dialog = getDialog();
      expect(dialog.find('[role="dialog"]').exists()).toBe(true);
    });

    it("displays the dialog title", async () => {
      await flushPromises();
      const dialog = getDialog();
      expect(dialog.text()).toContain("Configure Single Sign-on");
    });
  });

  describe("metadata URL mode", () => {
    beforeEach(() => mountWrapper());

    it("shows metadata URL field when checkbox is enabled", async () => {
      await flushPromises();
      const dialog = getDialog();

      const checkbox = dialog.find('[data-test="checkbox-idp-metadata"] input');
      await checkbox.setValue(true);
      await flushPromises();

      expect(dialog.find('[data-test="idp-metadata-url"]').exists()).toBe(true);
      expect(dialog.find('[data-test="idp-manual-section"]').exists()).toBe(false);
    });

    it("disables save button when metadata URL is empty", async () => {
      await flushPromises();
      const dialog = getDialog();

      const checkbox = dialog.find('[data-test="checkbox-idp-metadata"] input');
      await checkbox.setValue(true);
      await flushPromises();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      expect(saveButton.attributes("disabled")).toBeDefined();
    });

    it("enables save button when valid metadata URL is provided", async () => {
      await flushPromises();
      const dialog = getDialog();

      const checkbox = dialog.find('[data-test="checkbox-idp-metadata"] input');
      await checkbox.setValue(true);
      await flushPromises();

      const urlField = dialog.find('[data-test="idp-metadata-url"] input');
      await urlField.setValue("https://example.com/metadata");
      await flushPromises();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      expect(saveButton.attributes("disabled")).toBeUndefined();
    });

    it("saves configuration with metadata URL", async () => {
      await flushPromises();
      const dialog = getDialog();

      vi.mocked(instanceStore.updateSamlAuthentication).mockResolvedValueOnce();

      const checkbox = dialog.find('[data-test="checkbox-idp-metadata"] input');
      await checkbox.setValue(true);
      await flushPromises();

      const urlField = dialog.find('[data-test="idp-metadata-url"] input');
      await urlField.setValue("https://example.com/metadata");
      await flushPromises();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      await saveButton.trigger("click");
      await flushPromises();

      expect(instanceStore.updateSamlAuthentication).toHaveBeenCalledWith({
        enable: true,
        idp: { metadata_url: "https://example.com/metadata" },
        sp: { sign_requests: false },
      });
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Successfully updated SAML configuration.");
    });
  });

  describe("manual configuration mode", () => {
    beforeEach(() => mountWrapper());

    it("shows manual configuration fields by default", async () => {
      await flushPromises();
      const dialog = getDialog();

      expect(dialog.find('[data-test="idp-manual-section"]').exists()).toBe(true);
      expect(dialog.find('[data-test="idp-signon-post-url"]').exists()).toBe(true);
      expect(dialog.find('[data-test="idp-signon-redirect-url"]').exists()).toBe(true);
      expect(dialog.find('[data-test="idp-entity-id"]').exists()).toBe(true);
      expect(dialog.find('[data-test="idp-x509-certificate"]').exists()).toBe(true);
    });

    it("shows warning when no URLs are provided", async () => {
      await flushPromises();
      const dialog = getDialog();

      expect(dialog.find('[data-test="manual-config-info"]').exists()).toBe(true);
      expect(dialog.text()).toContain("You need to provide at least one of the following URLs");
    });

    it("disables save button when required fields are empty", async () => {
      await flushPromises();
      const dialog = getDialog();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      expect(saveButton.attributes("disabled")).toBeDefined();
    });

    it("enables save button when all required manual fields are filled", async () => {
      await flushPromises();
      const dialog = getDialog();

      const postUrlField = dialog.find('[data-test="idp-signon-post-url"] input');
      await postUrlField.setValue("https://idp.example.com/sso/post");
      await flushPromises();

      const entityIdField = dialog.find('[data-test="idp-entity-id"] input');
      await entityIdField.setValue("https://idp.example.com");
      await flushPromises();

      const certField = dialog.find('[data-test="idp-x509-certificate"] textarea');
      await certField.setValue(validCertificate);
      await flushPromises();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      expect(saveButton.attributes("disabled")).toBeUndefined();
    });

    it("accepts redirect URL instead of POST URL", async () => {
      await flushPromises();
      const dialog = getDialog();

      const redirectUrlField = dialog.find('[data-test="idp-signon-redirect-url"] input');
      await redirectUrlField.setValue("https://idp.example.com/sso/redirect");
      await flushPromises();

      const entityIdField = dialog.find('[data-test="idp-entity-id"] input');
      await entityIdField.setValue("https://idp.example.com");
      await flushPromises();

      const certField = dialog.find('[data-test="idp-x509-certificate"] textarea');
      await certField.setValue(validCertificate);
      await flushPromises();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      expect(saveButton.attributes("disabled")).toBeUndefined();
    });

    it("saves configuration with manual settings", async () => {
      await flushPromises();
      const dialog = getDialog();

      vi.mocked(instanceStore.updateSamlAuthentication).mockResolvedValueOnce();

      const postUrlField = dialog.find('[data-test="idp-signon-post-url"] input');
      await postUrlField.setValue("https://idp.example.com/sso/post");
      await flushPromises();

      const redirectUrlField = dialog.find('[data-test="idp-signon-redirect-url"] input');
      await redirectUrlField.setValue("https://idp.example.com/sso/redirect");
      await flushPromises();

      const entityIdField = dialog.find('[data-test="idp-entity-id"] input');
      await entityIdField.setValue("https://idp.example.com");
      await flushPromises();

      const certField = dialog.find('[data-test="idp-x509-certificate"] textarea');
      await certField.setValue(validCertificate);
      await flushPromises();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      await saveButton.trigger("click");
      await flushPromises();

      expect(instanceStore.updateSamlAuthentication).toHaveBeenCalledWith({
        enable: true,
        idp: {
          entity_id: "https://idp.example.com",
          binding: {
            post: "https://idp.example.com/sso/post",
            redirect: "https://idp.example.com/sso/redirect",
          },
          certificate: expect.stringContaining("BEGIN CERTIFICATE"),
        },
        sp: { sign_requests: false },
      });
    });
  });

  describe("certificate validation", () => {
    beforeEach(() => mountWrapper());

    it("shows error when certificate is missing BEGIN/END blocks", async () => {
      await flushPromises();
      const dialog = getDialog();

      const certField = dialog.find('[data-test="idp-x509-certificate"] textarea');
      await certField.setValue("INVALIDCERTIFICATEDATA");
      await flushPromises();

      expect(dialog.text()).toContain("Certificate must include -----BEGIN CERTIFICATE----- and -----END CERTIFICATE----- blocks");
    });

    it("shows error when certificate format is invalid", async () => {
      await flushPromises();
      const dialog = getDialog();

      const invalidCert = `-----BEGIN CERTIFICATE-----
INVALIDCERTIFICATEDATA
-----END CERTIFICATE-----`;

      const certField = dialog.find('[data-test="idp-x509-certificate"] textarea');
      await certField.setValue(invalidCert);
      await flushPromises();

      expect(dialog.text()).toContain("Invalid X.509 certificate");
    });

    it("accepts valid certificate", async () => {
      await flushPromises();
      const dialog = getDialog();

      const certField = dialog.find('[data-test="idp-x509-certificate"] textarea');
      await certField.setValue(validCertificate);
      await flushPromises();

      // No error message should be displayed
      const certFieldWrapper = dialog.find('[data-test="idp-x509-certificate"]');
      expect(certFieldWrapper.text()).not.toContain("Invalid X.509 certificate");
      expect(certFieldWrapper.text()).not.toContain("Certificate must include");
    });
  });

  describe("advanced settings - SAML mappings", () => {
    beforeEach(() => mountWrapper());

    it("shows SAML mappings table when advanced settings is expanded", async () => {
      await flushPromises();
      const dialog = getDialog();

      const advancedSettings = dialog.find('[data-test="advanced-settings-title"]');
      await advancedSettings.trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="saml-mappings-table"]').exists()).toBe(true);
      expect(dialog.find('[data-test="add-mapping-btn"]').exists()).toBe(true);
    });

    it("adds a mapping when Add Mapping button is clicked", async () => {
      await flushPromises();
      const dialog = getDialog();

      const advancedSettings = dialog.find('[data-test="advanced-settings-title"]');
      await advancedSettings.trigger("click");
      await flushPromises();

      const addButton = dialog.find('[data-test="add-mapping-btn"]');
      await addButton.trigger("click");
      await flushPromises();

      expect(dialog.findAll('[data-test="saml-mapping-key"]')).toHaveLength(1);
    });

    it("allows adding up to 2 mappings", async () => {
      await flushPromises();
      const dialog = getDialog();

      const advancedSettings = dialog.find('[data-test="advanced-settings-title"]');
      await advancedSettings.trigger("click");
      await flushPromises();

      const addButton = dialog.find('[data-test="add-mapping-btn"]');
      await addButton.trigger("click");
      await flushPromises();
      await addButton.trigger("click");
      await flushPromises();

      expect(dialog.findAll('[data-test="saml-mapping-key"]')).toHaveLength(2);
      expect(addButton.attributes("disabled")).toBeDefined();
    });

    it("removes a mapping when remove button is clicked", async () => {
      await flushPromises();
      const dialog = getDialog();

      const advancedSettings = dialog.find('[data-test="advanced-settings-title"]');
      await advancedSettings.trigger("click");
      await flushPromises();

      const addButton = dialog.find('[data-test="add-mapping-btn"]');
      await addButton.trigger("click");
      await flushPromises();

      const removeButton = dialog.find('[data-test="remove-mapping-btn"]');
      await removeButton.trigger("click");
      await flushPromises();

      expect(dialog.findAll('[data-test="saml-mapping-key"]')).toHaveLength(0);
    });

    it("sends mappings in save request when configured", async () => {
      await flushPromises();
      const dialog = getDialog();

      vi.mocked(instanceStore.updateSamlAuthentication).mockResolvedValueOnce();

      // Fill required fields
      const checkbox = dialog.find('[data-test="checkbox-idp-metadata"] input');
      await checkbox.setValue(true);
      await flushPromises();

      const urlField = dialog.find('[data-test="idp-metadata-url"] input');
      await urlField.setValue("https://example.com/metadata");
      await flushPromises();

      // Add mappings
      const advancedSettings = dialog.find('[data-test="advanced-settings-title"]');
      await advancedSettings.trigger("click");
      await flushPromises();

      const addButton = dialog.find('[data-test="add-mapping-btn"]');
      await addButton.trigger("click");
      await flushPromises();

      const mappingValue = dialog.find('[data-test="saml-mapping-value"] input');
      await mappingValue.setValue("custom.email");
      await flushPromises();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      await saveButton.trigger("click");
      await flushPromises();

      expect(instanceStore.updateSamlAuthentication).toHaveBeenCalledWith(
        expect.objectContaining({
          idp: expect.objectContaining({
            mappings: expect.any(Object),
          }),
        }),
      );
    });

    it("shows sign request checkbox in advanced settings", async () => {
      await flushPromises();
      const dialog = getDialog();

      const advancedSettings = dialog.find('[data-test="advanced-settings-title"]');
      await advancedSettings.trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="sign-request-checkbox"]').exists()).toBe(true);
    });

    it("sends sign_requests setting when enabled", async () => {
      await flushPromises();
      const dialog = getDialog();

      vi.mocked(instanceStore.updateSamlAuthentication).mockResolvedValueOnce();

      const checkbox = dialog.find('[data-test="checkbox-idp-metadata"] input');
      await checkbox.setValue(true);
      await flushPromises();

      const urlField = dialog.find('[data-test="idp-metadata-url"] input');
      await urlField.setValue("https://example.com/metadata");
      await flushPromises();

      const advancedSettings = dialog.find('[data-test="advanced-settings-title"]');
      await advancedSettings.trigger("click");
      await flushPromises();

      const signRequestCheckbox = dialog.find('[data-test="sign-request-checkbox"] input');
      await signRequestCheckbox.setValue(true);
      await flushPromises();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      await saveButton.trigger("click");
      await flushPromises();

      expect(instanceStore.updateSamlAuthentication).toHaveBeenCalledWith(
        expect.objectContaining({
          sp: { sign_requests: true },
        }),
      );
    });
  });

  describe("editing existing SAML configuration", () => {
    it("populates fields when opening with existing SAML config", async () => {
      wrapper = mountComponent(ConfigureSSO, {
        props: { modelValue: false },
        attachTo: document.body,
        piniaOptions: {
          initialState: {
            adminInstance: {
              authenticationSettings: {
                local: { enabled: true },
                saml: {
                  enabled: true,
                  idp: {
                    entity_id: "https://existing.idp.com",
                    binding: {
                      post: "https://existing.idp.com/sso/post",
                      redirect: "https://existing.idp.com/sso/redirect",
                    },
                    certificates: [validCertificate],
                    mappings: {
                      email: "custom.email",
                      name: "custom.name",
                    },
                  },
                  sp: { sign_auth_requests: true },
                },
              },
            },
          },
        },
      });

      instanceStore = useInstanceStore();

      // Open dialog
      await wrapper.setProps({ modelValue: true });
      await flushPromises();

      const dialog = getDialog();

      const postUrlField = dialog.find('[data-test="idp-signon-post-url"] input').element as HTMLInputElement;
      const redirectUrlField = dialog.find('[data-test="idp-signon-redirect-url"] input').element as HTMLInputElement;
      const entityIdField = dialog.find('[data-test="idp-entity-id"] input').element as HTMLInputElement;

      // Check that fields are populated
      expect(postUrlField.value).toBe("https://existing.idp.com/sso/post");
      expect(redirectUrlField.value).toBe("https://existing.idp.com/sso/redirect");
      expect(entityIdField.value).toBe("https://existing.idp.com");

      // Check advanced settings
      const advancedSettings = dialog.find('[data-test="advanced-settings-title"]');
      await advancedSettings.trigger("click");
      await flushPromises();

      expect(dialog.findAll('[data-test="saml-mapping-key"]')).toHaveLength(2);
    });
  });

  describe("error handling", () => {
    beforeEach(() => mountWrapper());

    it("shows error snackbar when save fails", async () => {
      await flushPromises();
      const dialog = getDialog();

      vi.mocked(instanceStore.updateSamlAuthentication).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const checkbox = dialog.find('[data-test="checkbox-idp-metadata"] input');
      await checkbox.setValue(true);
      await flushPromises();

      const urlField = dialog.find('[data-test="idp-metadata-url"] input');
      await urlField.setValue("https://example.com/metadata");
      await flushPromises();

      const saveButton = dialog.find('[data-test="confirm-btn"]');
      await saveButton.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update SAML configuration.");
    });
  });

  describe("closing dialog", () => {
    beforeEach(() => mountWrapper());

    it("closes dialog when cancel button is clicked", async () => {
      await flushPromises();
      const dialog = getDialog();

      const cancelButton = dialog.find('[data-test="cancel-btn"]');
      await cancelButton.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });
  });
});
