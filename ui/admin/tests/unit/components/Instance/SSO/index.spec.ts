import MockAdapter from "axios-mock-adapter";
import { beforeEach, describe, it, expect, vi } from "vitest";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { createPinia, setActivePinia } from "pinia";
import { adminApi } from "@admin/api/http";
import useInstanceStore from "@admin/store/modules/instance";
import ConfigureSSO from "@admin/components/Instance/SSO/ConfigureSSO.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type ConfigureSSOWrapper = VueWrapper<InstanceType<typeof ConfigureSSO>>;

describe("Configure SSO", () => {
  let wrapper: ConfigureSSOWrapper;
  let adminMock: MockAdapter;

  const vuetify = createVuetify();

  beforeEach(async () => {
    setActivePinia(createPinia());

    wrapper = mount(ConfigureSSO, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });

    wrapper.vm.showDialog = true;
    await flushPromises();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("resets fields when 'close' is clicked", async () => {
    wrapper.vm.useMetadataUrl = true;
    wrapper.vm.IdPMetadataURL = "https://example.com/metadata";

    await wrapper.findComponent("[data-test='close-btn']").trigger("click");

    expect(wrapper.vm.IdPMetadataURL).toBe("");
    expect(wrapper.vm.useMetadataUrl).toBe(false);
  });

  it("disables save button if required fields are empty", () => {
    wrapper.vm.useMetadataUrl = true;
    wrapper.vm.IdPMetadataURL = "";

    expect(wrapper.findComponent("[data-test='save-btn']").attributes("disabled")).toBeDefined();
  });

  it("adds a mapping when 'Add Mapping' is clicked", async () => {
    await wrapper.findComponent("[data-test='advanced-settings-title']").trigger("click");
    await wrapper.findComponent("[data-test='add-mapping-btn']").trigger("click");
    expect(wrapper.vm.mappings.length).toBe(1);
  });

  it("removes a mapping when 'Remove Mapping' button is clicked", async () => {
    wrapper.vm.mappings = [{ key: "Email", value: "test@example.com" }];

    await wrapper.findComponent("[data-test='advanced-settings-title']").trigger("click");
    await wrapper.findComponent("[data-test='remove-mapping-btn']").trigger("click");

    expect(wrapper.vm.mappings.length).toBe(0);
  });

  it("calls store action on save", async () => {
    adminMock = new MockAdapter(adminApi.getAxios());
    adminMock.onPut("http://localhost:3000/admin/api/authentication/saml").reply(200);

    const instanceStore = useInstanceStore();
    const storeSpy = vi.spyOn(instanceStore, "updateSamlAuthentication").mockResolvedValue();

    await wrapper.findComponent('[data-test="checkbox-idp-metadata"]').setValue(true);
    await wrapper.findComponent('[data-test="idp-metadata-url"]').setValue("https://example.co/metadata");
    await wrapper.findComponent("[data-test='save-btn']").trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      enable: true,
      idp: {
        metadata_url: "https://example.co/metadata",
      },
      sp: {
        sign_requests: false,
      },
    });
  });

  it("renders Advanced Settings when expansion panel is opened", async () => {
    await wrapper.findComponent("[data-test='advanced-settings-title']").trigger("click");
    expect(wrapper.findComponent("[data-test='saml-mappings-table']").exists()).toBe(true);
  });

  it("shows an error if X.509 certificate does not include BEGIN/END blocks", async () => {
    wrapper.vm.useMetadataUrl = false;

    await wrapper.findComponent("[data-test='advanced-settings-title']").trigger("click");

    const certificateWithoutBlocks = "MIIDdzCCAl+gAwIBAgIEb1Yc...";
    wrapper.vm.handleCertificateChange(certificateWithoutBlocks);

    expect(wrapper.vm.x509CertificateErrorMessage)
      .toBe("Certificate must include -----BEGIN CERTIFICATE----- and -----END CERTIFICATE----- blocks.");
  });

  it("shows an error if X.509 certificate has BEGIN/END blocks but is invalid", async () => {
    wrapper.vm.useMetadataUrl = false;

    const invalidCert = `
    -----BEGIN CERTIFICATE-----
    INVALIDCERTIFICATEDATA
    -----END CERTIFICATE-----
  `;

    await wrapper.vm.handleCertificateChange(invalidCert);

    expect(wrapper.vm.x509CertificateErrorMessage).toBe("Invalid X.509 certificate.");
  });
});
