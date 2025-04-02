import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import License from "../../../../../src/components/Settings/SettingsLicense.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

type LicenseWrapper = VueWrapper<InstanceType<typeof License>>;

const license = {
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
  },
};

const store = createStore({
  state: {
    license,
  },
  getters: {
    "license/license": (state) => state.license,
  },
  actions: {
    "license/get": vi.fn(),
  },
});

describe("License", () => {
  let wrapper: LicenseWrapper;

  beforeEach(async () => {
    const vuetify = createVuetify();

    wrapper = await mount(License, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  Object.keys(license).forEach((field) => {
    it(`Receives the field ${field} of license state from store`, () => {
      expect(wrapper.vm.license[field]).toEqual(license[field]);
    });
  });
  it("Renders license fields in template", () => {
    expect(wrapper.find("[data-test=issuedAt-field]").exists()).toBe(true);
    expect(wrapper.find("[data-test=allowedRegions-field]").exists()).toBe(true);
  });
  Object.keys(license.customer).forEach((customer) => {
    it(`Receives the customer field ${customer} in template`, () => {
      expect(wrapper.find(`[data-test=${customer}]`).exists()).toBe(true);
    });
  });
  Object.keys(license.features).forEach((feature) => {
    it(`Receives the feature ${feature} in template`, () => {
      expect(wrapper.find(`[data-test=${feature}]`).exists()).toBe(true);
    });
  });
  Object.keys(license.features).forEach((feature) => {
    it(`Receives the feature ${feature} with expected text`, () => {
      const value = license.features[feature];
      const mapFeatureValue = (name: string) => {
        const value = license.features[name];
        if (typeof value === "boolean") {
          return value ? "mdi-check-circle" : "mdi-close-circle";
        }
        return "unlimited";
      };
      if (typeof value === "boolean") {
        // eslint-disable-next-line no-unused-expressions
        value
          ? expect(mapFeatureValue(feature)).toBe("mdi-check-circle")
          : expect(mapFeatureValue(feature)).toBe("mdi-close-circle");
      } else {
        expect(mapFeatureValue(feature)).toBe("unlimited");
      }
    });
  });
});
