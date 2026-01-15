import { createVuetify } from "vuetify";
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import useLicenseStore from "@admin/store/modules/license";
import License from "@admin/components/Settings/SettingsLicense.vue";
import routes from "@admin/router";
import { adminApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const licenseMock = {
  id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  expired: false,
  about_to_expire: false,
  grace_period: false,
  issued_at: -1,
  starts_at: -1,
  expires_at: -1,
  allowed_regions: [],
  customer: {
    id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    name: "ShellHub",
    email: "contato@ossystems.com.br",
    company: "O.S. Systems",
  },
  features: {
    devices: -1,
    session_recording: true,
    firewall_rules: true,
    billing: false,
    reports: false,
    login_link: false,
  },
};

describe("License", () => {
  const mockAdminApi = new MockAdapter(adminApi.getAxios());
  setActivePinia(createPinia());
  const licenseStore = useLicenseStore();
  const vuetify = createVuetify();
  licenseStore.license = licenseMock;
  mockAdminApi.onGet("http://localhost:3000/admin/api/license").reply(200, licenseMock);

  const wrapper = mount(License, {
    global: {
      plugins: [vuetify, routes, SnackbarPlugin],
    },
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  Object.keys(licenseMock).forEach((field) => {
    it(`Receives the field ${field} of license state from store`, () => {
      expect(wrapper.vm.license[field]).toEqual(licenseMock[field]);
    });
  });

  Object.keys(licenseMock.customer).forEach((customer) => {
    it(`Receives the customer field ${customer} in template`, () => {
      expect(wrapper.find(`[data-test='${customer}']`).exists()).toBe(true);
    });
  });

  Object.keys(licenseMock.features).forEach((feature) => {
    if (["reports", "login_link"].includes(feature)) return;

    it(`Receives the feature ${feature} in template`, () => {
      expect(wrapper.find(`[data-test=${feature}]`).exists()).toBe(true);
    });
  });
});
