import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Settings from "../../src/views/Settings.vue";
import { createStore } from "vuex";
import { key } from "../../src/store";
import routes from "../../src/router";
import { envVariables } from "../../src/envVariables";

describe("Settings", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  let numberNamespaces = 1;

  const items = [
    {
      title: "Profile",
      path: "/settings",
    },
    {
      title: "Namespace",
      path: "/settings/namespace-manager",
    },
    {
      title: "Private Keys",
      path: "/settings/private-keys",
    },
    {
      title: "Billing",
      path: "/settings/billing",
    },
  ];

  const store = createStore({
    state: {
      numberNamespaces,
    },
    getters: {
      "namespaces/getNumberNamespaces": (state) => state.numberNamespaces,
    },
    actions: {
      "box/setStatus": vi.fn(),
      "sessions/resetPagePerpage": vi.fn(),
      "sessions/refresh": vi.fn(),
      "stats/get": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });

  ///////
  // In this case, the billing tab can only be rendered in the cloud.
  // Checks if this tab has been rendered when user has namespace.
  ///////

  describe("Cloud is true", () => {
    beforeEach(() => {
      wrapper = mount(Settings, {
        global: {
          plugins: [[store, key], vuetify, routes],
        },
      });

      envVariables.isCloud = true;
      envVariables.billingEnable = true;
    });

    ///////
    // Component Rendering
    //////

    it("Is a Vue instance", () => {
      expect(wrapper).toBeTruthy();
    });
    it("Renders the component", () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data checking
    //////
    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", async () => {
      for (const [key, value] of Object.entries(items)) {
        expect(
          wrapper.find(`[data-test="${value.title}-tab"]`).attributes("href")
        ).toEqual(value.path);
      }
    });
  });

  ///////
  // In this case, the billing tab can only be rendered in the cloud.
  // Checks if this tab is not rendered when user has no namespace.
  ///////

  describe("Cloud is false", () => {
    beforeEach(() => {
      wrapper = mount(Settings, {
        global: {
          plugins: [[store, key], vuetify, routes],
        },
      });

      envVariables.isCloud = false;
      envVariables.billingEnable = true;
    });

    ///////
    // Component Rendering
    //////

    it("Is a Vue instance", () => {
      expect(wrapper).toBeTruthy();
    });
    it("Renders the component", () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data checking
    //////
    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", async () => {
      Object.keys(items.slice(0, -1)).forEach((item) => {
        expect(
          // @ts-ignore
          wrapper.find(`[data-test="${items[item].title}-tab"]`).text()
          // @ts-ignore
        ).toEqual(items[item].title);
      });
    });
  });
});
