import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NamespaceList from "../../../src/components/Namespace/NamespaceList.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { envVariables } from "../../../src/envVariables";

const namespace = {
  name: "namespace3",
  owner: "user1",
  member_names: ["user6", "user7", "user8"],
  tenant_id: "e359bf484715",
};

const namespaces = [
  {
    name: "namespace1",
    owner: "user1",
    member_names: ["user3", "user4", "user5"],
    tenant_id: "xxxxxxxx",
  },
  {
    name: "namespace2",
    owner: "user1",
    member_names: ["user3", "user4"],
    tenant_id: "xxxxxxxy",
  },
];

const store = createStore({
  state: {
    namespace,
    namespaces,
  },
  getters: {
    "namespaces/list": (state) => state.namespaces,
    "namespaces/get": (state) => state.namespace,
  },
  actions: {
    "namespaces/switchNamespace": () => {},
    "snackbar/showSnackbarErrorLoading": () => {},
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
      wrapper = mount(NamespaceList, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
      });

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
    it("Process data in the computed", () => {
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.namespaces).toEqual(
        namespaces.filter((el) => el.name !== namespace.name)
      );
    });
    //////
    // HTML validation
    //////

    it("Renders the template with data", async () => {
      const namespacesLocal = namespaces.filter(
        (el) => el.name !== namespace.name
      );

      Object.keys(namespacesLocal).forEach((item) => {
        expect(
          wrapper
            // @ts-ignore
            .find(`[data-test="${namespacesLocal[item].name}-namespace"]`)
            .text()
          // @ts-ignore
        ).toEqual(namespacesLocal[item].name);
      });
    });
  });

  ///////
  // In this case, check owner fields rendering in open version
  // of the template.
  ///////

  describe("Open version", () => {
    beforeEach(() => {
      wrapper = mount(NamespaceList, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
      });

      envVariables.isEnterprise = false;
      vi.spyOn(Storage.prototype, 'getItem').mockReturnValue(namespace.tenant_id);

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
    it("Process data in the computed", () => {
      expect(wrapper.vm.namespace).toEqual(namespace);
      expect(wrapper.vm.namespaces).toEqual(
        namespaces.filter((el) => el.name !== namespace.name)
      );
    });
    //////
    // HTML validation
    //////

    it("Renders the template with data", async () => {
      const namespacesLocal = namespaces.filter(
        (el) => el.name !== namespace.name
      );

      Object.keys(namespacesLocal).forEach((item) => {
        expect(
          wrapper
            // @ts-ignore
            .find(`[data-test="${namespacesLocal[item].name}-namespace"]`)
            .text()
          // @ts-ignore
        ).toEqual(namespacesLocal[item].name);
      });
    });
  });
});
