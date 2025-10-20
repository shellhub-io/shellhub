import { nextTick } from "vue";
import MockAdapter from "axios-mock-adapter";
import { describe, it, expect, vi } from "vitest";
import { DOMWrapper, flushPromises, mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { createPinia, setActivePinia } from "pinia";
import { adminApi } from "@admin/api/http";
import useInstanceStore from "@admin/store/modules/instance";
import ConfigureSSO from "@admin/components/Instance/SSO/ConfigureSSO.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Configure SSO", async () => {
  const mockAdminApi = new MockAdapter(adminApi.getAxios());
  const vuetify = createVuetify();
  setActivePinia(createPinia());

  const wrapper = mount(ConfigureSSO, {
    global: { plugins: [vuetify, SnackbarPlugin] },
    props: { modelValue: true },
  });

  await flushPromises();

  it("Renders the component", async () => {
    const dialog = new DOMWrapper(document.body);
    await nextTick();
    expect(dialog.html()).toMatchSnapshot();
  });

  it("disables save button if required fields are empty", () => {
    wrapper.vm.useMetadataUrl = true;
    wrapper.vm.IdPMetadataURL = "";

    expect(wrapper.findComponent("[data-test='confirm-btn']").attributes("disabled")).toBeDefined();
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
    mockAdminApi.onPut("http://localhost:3000/admin/api/authentication/saml").reply(200);

    const instanceStore = useInstanceStore();
    const storeSpy = vi.spyOn(instanceStore, "updateSamlAuthentication").mockResolvedValue();

    await wrapper.findComponent('[data-test="checkbox-idp-metadata"]').setValue(true);
    await wrapper.findComponent('[data-test="idp-metadata-url"]').setValue("https://example.co/metadata");
    await wrapper.findComponent("[data-test='confirm-btn']").trigger("click");

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
