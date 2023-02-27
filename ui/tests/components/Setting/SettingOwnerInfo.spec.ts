import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { createStore } from "vuex";
import SettingOwnerInfo from "../../../src/components/Setting/SettingOwnerInfo.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("SettingOwnerInfo", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingOwnerInfo>>;
  const vuetify = createVuetify();

  const isOwner = true;

  const namespace = {
    name: "namespace1",
    owner: "124",
    members: [
      { id: "124", username: "user4", role: "owner" },
      { id: "123", username: "user1", role: "operator" },
      { id: "125", username: "user5", role: "administrator" },
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

    ///////
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

    ///////
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
      const namespaceOwnerMessage = `Contact ${namespace.members[0].username} user for more information.`;
      expect(wrapper.find("[data-test=contactUser-p]").text()).toEqual(namespaceOwnerMessage);
    });
  });
});
