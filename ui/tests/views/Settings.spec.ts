import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import Settings from "../../src/views/Settings.vue";
import { key } from "../../src/store";
import routes from "../../src/router";
import { envVariables } from "../../src/envVariables";

describe("Settings", () => {
  let wrapper: VueWrapper<InstanceType<typeof Settings>>;
  const vuetify = createVuetify();

  const numberNamespaces = 1;

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
      "stats/stats": () => [],
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
        mocks: {
          envVariables: {
            ...envVariables,
            isCloud: true,
            billingEnable: true,
          },
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
      Object.values(items).forEach((item) => {
        expect(
          wrapper.find(`[data-test="${item.title}-tab"]`).exists(),
        ).toEqual(true);
      });
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
        mocks: {
          envVariables: {
            ...envVariables,
            isCloud: false,
            billingEnable: true,
          },
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
          wrapper.find(`[data-test="${items[item].title}-tab"]`).text(),
        ).toEqual(items[item].title);
      });
    });
  });
});
