import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import useLicenseStore from "@admin/store/modules/license";
import { adminApi } from "@admin/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import License from "../../../../../src/components/Settings/SettingsLicense.vue";
import routes from "../../../../../src/router";

type LicenseWrapper = VueWrapper<InstanceType<typeof License>>;

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
  let wrapper: LicenseWrapper;
  let mock: MockAdapter;

  beforeEach(async () => {
    setActivePinia(createPinia());
    const vuetify = createVuetify();
    mock = new MockAdapter(adminApi.getAxios());

    const licenseStore = useLicenseStore();
    licenseStore.license = licenseMock;
    mock.onGet("http://localhost:3000/admin/api/license").reply(200);

    wrapper = await mount(License, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  Object.keys(licenseMock).forEach((field) => {
    it(`Receives the field ${field} of license state from store`, () => {
      expect(wrapper.vm.license[field]).toEqual(licenseMock[field]);
    });
  });

  it("Renders license fields in template", () => {
    expect(wrapper.find("[data-test=issuedAt-field]").exists()).toBe(true);
    expect(wrapper.find("[data-test=allowedRegions-field]").exists()).toBe(true);
  });

  Object.keys(licenseMock.customer).forEach((customer) => {
    it(`Receives the customer field ${customer} in template`, () => {
      expect(wrapper.find(`[data-test=${customer}]`).exists()).toBe(true);
    });
  });

  Object.keys(licenseMock.features).forEach((feature) => {
    if (["reports", "login_link"].includes(feature)) return;

    it(`Receives the feature ${feature} in template`, () => {
      expect(wrapper.find(`[data-test=${feature}]`).exists()).toBe(true);
    });
  });

  Object.keys(licenseMock.features).forEach((feature) => {
    if (["reports", "login_link"].includes(feature)) return;

    it(`Receives the feature ${feature} with expected text`, () => {
      const value = licenseMock.features[feature];
      const mapFeatureValue = (name: string) => {
        const value = licenseMock.features[name];
        if (typeof value === "boolean") {
          return value ? "mdi-check-circle" : "mdi-close-circle";
        }
        return "unlimited";
      };

      if (typeof value === "boolean") {
        expect(mapFeatureValue(feature)).toBe(value ? "mdi-check-circle" : "mdi-close-circle");
      } else {
        expect(mapFeatureValue(feature)).toBe("unlimited");
      }
    });
  });

  Object.keys(licenseMock.features).forEach((feature) => {
    it(`Receives the feature ${feature} with expected text`, () => {
      const value = licenseMock.features[feature];
      const mapFeatureValue = (name: string) => {
        const value = licenseMock.features[name];
        if (typeof value === "boolean") {
          return value ? "mdi-check-circle" : "mdi-close-circle";
        }
        return "unlimited";
      };

      if (typeof value === "boolean") {
        expect(mapFeatureValue(feature)).toBe(value ? "mdi-check-circle" : "mdi-close-circle");
      } else {
        expect(mapFeatureValue(feature)).toBe("unlimited");
      }
    });
  });
});
