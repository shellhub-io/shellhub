import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import NamespaceAdd from "../../../src/components/Namespace/NamespaceAdd.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { envVariables } from "../../../src/envVariables";

const show = true;
const firstNamespace = true;

const invalidNamespaces = [
  "'",
  '"',
  "!",
  "@",
  "#",
  "$",
  "%",
  "¨",
  "&",
  "*",
  "(",
  ")",
  "-",
  "_",
  "=",
  "+",
  "´",
  "`",
  "[",
  "{",
  "~",
  "^",
  "]",
  ",",
  "<",
  "..",
  ">",
  ";",
  ":",
  "/",
  "?",
];

const invalidMinAndMaxCharacters = [
  "s",
  "sh",
  "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "namespaces/switchNamespace": vi.fn(),
    "namespaces/post": vi.fn(),
    "namespaces/fetch": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
    "snackbar/showSnackbarErrorLoading": vi.fn(),
  },
});

describe("NamespaceAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceAdd>>;
  const vuetify = createVuetify();

  ///////
  // In this case, the rendering of the dialog is checked. In which
  // case with the input data it cannot take place.
  ///////

  describe("Doesn't render the dialog", () => {
    beforeEach(() => {
      wrapper = mount(NamespaceAdd, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        props: { show: !show, firstNamespace },
        // shallow: true,
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

    it("Receive data in props", () => {
      expect(wrapper.vm.showDialog).toEqual(!show);
      expect(wrapper.vm.firstNamespace).toEqual(firstNamespace);
    });
    it("Compare data with default value", () => {
      expect(wrapper.vm.showDialog).toEqual(false);
      expect(wrapper.vm.namespaceName).toEqual("");
    });
    it("Process data in the computed", () => {
      expect(wrapper.vm.showDialog).toEqual(false);
    });
    //////
    // HTML validation
    //////

    it("Renders the template with data", () => {
      expect(wrapper.find('[data-test="namespaceAdd-card"]').exists()).toBe(
        false,
      );
      expect(wrapper.find('[data-test="namespace-text"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="add-btn"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, the rendering of the dialog is checked. In which
  // case with the input data it cannot take place.
  ///////

  describe("Dialog", () => {
    beforeEach(() => {
      wrapper = mount(NamespaceAdd, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        props: { show, firstNamespace },
        shallow: true,
      });
      wrapper.vm.showDialog = true;

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

    it("Receive data in props", () => {
      expect(wrapper.vm.showDialog).toEqual(show);
      expect(wrapper.vm.firstNamespace).toEqual(firstNamespace);
    });
    it("Compare data with default value", () => {
      expect(wrapper.vm.namespaceName).toEqual("");
    });
    it("Process data in the computed", () => {
      expect(wrapper.vm.showDialog).toEqual(true);
    });

    //////
    // HTML validation
    //////

    //////
    // In this case, the empty fields are validated.
    //////

    it("Show validation messages", async () => {
      wrapper.vm.namespaceName = "ShelHub";
      wrapper.vm.namespaceName = undefined;
      await flushPromises();
      expect(wrapper.vm.namespaceNameError).toBe("this is a required field");
    });

    it("Show validation messages", async () => {
      wrapper.vm.namespaceName = "ShelHub..";
      await flushPromises();
      expect(wrapper.vm.namespaceNameError).toBe(
        "The name must not contain dots",
      );
    });

    it("Show validation messages", async () => {
      invalidMinAndMaxCharacters.forEach(async (character) => {
        wrapper.vm.namespaceName = character;
        await flushPromises();
        expect(wrapper.vm.namespaceNameError).toBe(
          "this must be at most 30 characters",
        );
      });
    });

    // TODO RFC1123
  });
});
