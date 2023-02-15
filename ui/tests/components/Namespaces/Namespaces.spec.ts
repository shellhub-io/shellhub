import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Namespace from "../../../src/components/Namespace/Namespace.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { envVariables } from "../../../src/envVariables";

const inANamespace = true;

const namespace = {
  name: "namespace3",
  owner: "user1",
  member_names: ["user6", "user7", "user8"],
  tenant_id: "e359bf484715",
};

const store = createStore({
  state: {
    namespace,
  },
  getters: {
    "namespaces/get": (state) => state.namespace,
  },
  actions: {
    "namespaces/fetch": () => {},
    "namespaces/get": () => {},
    "namespaces/switchNamespace": () => {},
    "namespaces/setOwnerStatus": () => {},
    "snackbar/showSnackbarErrorLoading": () => {},
    "snackbar/showSnackbarErrorAssociation": () => {},
  },
});

describe("Namespace", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  ///////
  // In this case, check owner fields rendering in enterprise version.
  ///////

  describe("Enterprise version", () => {
    beforeEach(() => {
      wrapper = mount(Namespace, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        shallow: true,
      });
      localStorage.setItem("tenant", namespace.tenant_id);

      envVariables.isEnterprise = true;
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
    it("Compare data with default value", () => {
      expect(wrapper.vm.inANamespace).toEqual(false);
    });
    it("Process data in the computed", () => {
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.hasNamespace).toEqual(true);
      expect(wrapper.vm.tenant).toEqual(namespace.tenant_id);
    });
  });

  ///////
  // In this case, check owner fields rendering in open version
  // of the template.
  ///////

  describe("Open version", () => {
    beforeEach(() => {
      wrapper = mount(Namespace, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        shallow: true,
      });
      localStorage.setItem("tenant", namespace.tenant_id);

      envVariables.isEnterprise = false;
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
    it("Receives data in props", () => {
      expect(wrapper.vm.inANamespace).toEqual(!inANamespace);
    });
    it("Compare data with default value", () => {
      expect(wrapper.vm.inANamespace).toEqual(false);
    });
    it("Process data in the computed", () => {
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.hasNamespace).toEqual(true);
      expect(wrapper.vm.tenant).toEqual(namespace.tenant_id);
    });
  });

  describe("Request assertions", () => {
    beforeEach(() => {
      vi.spyOn(Storage.prototype, "getItem").mockReturnValue("");
    });

    it("Should switch to a namespace when the current namespace is not found", async () => {
      // TODO
    });
  });
});
