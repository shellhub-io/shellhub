import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SettingOwnerInfo from "../../../src/components/Setting/SettingOwnerInfo.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("SettingOwnerInfo", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const isOwner = true;

  const namespace = {
    name: "namespace1",
    owner: "124",
    members: [
      { id: "124", name: "user4" },
      { id: "123", name: "user1" },
      { id: "125", name: "user5" },
    ],
    tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484713",
  };

  const store = createStore({
    state: {
      namespace,
    },
    getters: {
      "namespaces/get": (state) => state.namespace,
    },
    actions: {},
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is to check if the message no exists.
  ///////

  describe("Owner is true", () => {
    beforeEach(() => {
      wrapper = mount(SettingOwnerInfo, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        props: {
          isOwner,
        },
      });
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

    ///////s
    // Data checking
    //////
    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", () => {
      expect(wrapper.find('[data-test="message-div"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is to check if the message exists.
  ///////

  describe("Owner is false", () => {
    beforeEach(() => {
      wrapper = mount(SettingOwnerInfo, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        props: {
          isOwner: false,
        },
      });
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

    ///////s
    // Data checking
    //////
    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", () => {
      expect(wrapper.find('[data-test="message-div"]').exists()).toBe(true);
      const namespaceOwnerMessage = `Contact ${namespace.members[0].name} user for more information.`;
      expect(wrapper.find('[data-test=contactUser-p]').text()).toEqual(namespaceOwnerMessage);
    });
  });
});
