import MockAdapter from "axios-mock-adapter";
import { beforeEach, describe, it, expect, vi } from "vitest";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { adminApi } from "@admin/api/http";
import { store, key } from "../../../../../src/store";
import ConfigureSSO from "../../../../../src/components/Instance/SSO/ConfigureSSO.vue";

type ConfigureSSOWrapper = VueWrapper<InstanceType<typeof ConfigureSSO>>;

describe("Configure SSO", () => {
  let wrapper: ConfigureSSOWrapper;

  let adminMock: MockAdapter;

  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(ConfigureSSO, {
      global: {
        plugins: [[store, key], vuetify],
      },
    });

    wrapper.vm.dialog = true;
    await flushPromises();
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("resets fields when 'close' is clicked", async () => {
    wrapper.vm.checkbox = true;
    wrapper.vm.IdPMetadataURL = "https://example.com/metadata";

    await wrapper.findComponent("[data-test='close-btn']").trigger("click");
    expect(wrapper.vm.IdPMetadataURL).toBe("");
    expect(wrapper.vm.checkbox).toBe(false);
  });

  it("disables save button if required fields are empty", async () => {
    wrapper.vm.checkbox = true;
    wrapper.vm.IdPMetadataURL = "";

    expect(wrapper.findComponent("[data-test='save-btn']").attributes("disabled")).toBeDefined();
  });

  it("adds a mapping when 'Add Mapping' is clicked", async () => {
    await wrapper.findComponent("[data-test='advanced-settings-title']").trigger("click");
    await wrapper.findComponent("[data-test='add-mapping-btn']").trigger("click");
    expect(wrapper.vm.mappings.length).toBe(2);
  });

  it("removes a mapping when 'Remove Mapping' button is clicked", async () => {
    wrapper.vm.mappings = [{ key: "Email", value: "test@example.com" }];

    await wrapper.findComponent("[data-test='advanced-settings-title']").trigger("click");
    await wrapper.findComponent("[data-test='remove-mapping-btn']").trigger("click");
    expect(wrapper.vm.mappings.length).toBe(0);
  });

  it("calls Vuex action on save", async () => {
    adminMock = new MockAdapter(adminApi.getAxios());
    adminMock.onPut("http://localhost:3000/admin/api/authentication/saml").reply(200);
    await wrapper.findComponent('[data-test="checkbox-idp-metadata"]').setValue(true);

    await wrapper.findComponent('[data-test="idp-metadata-url"]').setValue("https://example.co/metadata");

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent("[data-test='save-btn']").trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith("instance/updateSamlAuthentication", {
      enable: true,
      idp: {
        mappings: {
          "": "",
        },
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
});
