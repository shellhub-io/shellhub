import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import NamespaceInstructions from "../../../src/components/Namespace/NamespaceInstructions.vue";
import routes from "../../../src/router";
import { envVariables } from "../../../src/envVariables";

describe("NamespaceInstructions", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceInstructions>>;
  const vuetify = createVuetify();
  envVariables.isEnterprise = true;
  const show = true;

  ///////
  // In this case, check owner fields rendering in enterprise version of
  // the template.
  ///////

  describe("Enterprise Version renders", () => {
    beforeEach(() => {
      wrapper = mount(NamespaceInstructions, {
        global: {
          plugins: [routes, vuetify],
        },
        props: {
          show,
        },
        shallow: true,
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
    it("Receives data in props", () => {
      expect(wrapper.vm.show).toEqual(show);
    });
    it("Compare data with default value", () => {
      expect(wrapper.vm.dialogAdd).toEqual(false);
    });
    //////
    // HTML validation
    //////

    // Todo
  });

  ///////
  // In this case, check owner fields rendering in open version of
  // the template.
  ///////

  describe("", () => {
    beforeEach(() => {
      wrapper = mount(NamespaceInstructions, {
        global: {
          plugins: [routes, vuetify],
        },
        props: {
          show,
        },
        shallow: true,
      });
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
      expect(wrapper.vm.show).toEqual(show);
    });
    it("Compare data with default value", () => {
      expect(wrapper.vm.dialogAdd).toEqual(false);
    });
    //////
    // HTML validation
    //////

    // Todo
  });
});
