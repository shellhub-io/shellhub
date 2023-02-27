import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import TerminalDialog from "../../../src/components/Terminal/TerminalDialog.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const uid = "a582b47a";
const username = "user";
const password = "pass";
const tabActive = "Password";

const privateKeys = [
  {
    name: "shellhub",
    data: "BBGVvbmF",
  },
  {
    name: "shellhub",
    data: "AbGVvbmF",
  },
];

const store = createStore({
  state: {
    terminal: uid,
    privateKeys,
  },
  getters: {
    "modals/terminal": (state) => state.terminal,
    "privateKey/list": (state) => state.privateKeys,
  },
  actions: {
    "modals/toggleTerminal": vi.fn(),
  },
});

describe("TerminalDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalDialog>>;
  const vuetify = createVuetify();

  ///////
  // In this case, the rendering of the console icon is tested.
  // For this test to work, the uid in props is an empty string.
  ///////

  describe("Button", () => {
    beforeEach(() => {
      wrapper = mount(TerminalDialog, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        props: {
          uid: "",
          show: false,
          enableConnectButton: true,
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
    // Data and Props checking
    //////

    it("Receive data in props", () => {
      expect(wrapper.vm.uid).toEqual("");
      expect(wrapper.vm.show).toEqual(false);
    });
    it("Compare data with default value", () => {
      expect(wrapper.vm.username).toEqual("");
      expect(wrapper.vm.password).toEqual("");
      expect(wrapper.vm.showLoginForm).toEqual(true);
      expect(wrapper.vm.tabActive).toEqual(tabActive);
    });
    it("Receive data in computed", () => {
      expect(wrapper.vm.showTerminal).toEqual(false);
      expect(wrapper.vm.show).toEqual(false);
      expect(wrapper.vm.getListPrivateKeys).toEqual(privateKeys);
    });
    it("Check the watch action", async () => {
      wrapper.vm.username = username;
      wrapper.vm.password = password;

      expect(wrapper.vm.username).toEqual(username);
      expect(wrapper.vm.password).toEqual(password);
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", () => {
      expect(wrapper.find('[data-test="connect-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="console-icon"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="console-item"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="terminal-dialog"]').exists()).toBe(
        false,
      );
    });
  });

  describe("Icon", () => {
    beforeEach(() => {
      wrapper = mount(TerminalDialog, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        props: {
          uid: "",
          show: false,
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
    // Data and Props checking
    //////

    it("Receive data in props", () => {
      expect(wrapper.vm.uid).toEqual("");
      expect(wrapper.vm.show).toEqual(false);
    });

    it("Compare data with default value", () => {
      expect(wrapper.vm.username).toEqual("");
      expect(wrapper.vm.password).toEqual("");
      expect(wrapper.vm.showLoginForm).toEqual(true);
      expect(wrapper.vm.tabActive).toEqual(tabActive);
    });
    it("Receive data in computed", () => {
      expect(wrapper.vm.showTerminal).toEqual(false);
      expect(wrapper.vm.show).toEqual(false);
      expect(wrapper.vm.getListPrivateKeys).toEqual(privateKeys);
    });
    it("Check the watch action", async () => {
      wrapper.vm.username = username;
      wrapper.vm.password = password;

      expect(wrapper.vm.username).toEqual(username);
      expect(wrapper.vm.password).toEqual(password);
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", () => {
      expect(wrapper.find('[data-test="connect-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="console-icon"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="console-item"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="terminal-dialog"]').exists()).toBe(
        false,
      );
    });
  });

  ///////
  // In this case the dialog is opened
  ///////

  describe("Dialog opened", () => {
    beforeEach(() => {
      wrapper = mount(TerminalDialog, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        props: {
          uid,
          show: true,
          enableConnectButton: true,
        },
        data() {
          return {
            showTerminal: false,
          };
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
    // Data and Props checking
    //////

    it("Receive data in props", () => {
      expect(wrapper.vm.uid).toEqual(uid);
      expect(wrapper.vm.show).toEqual(true);
      expect(wrapper.vm.enableConnectButton).toEqual(true);
    });

    it("Compare data with default value", () => {
      expect(wrapper.vm.username).toEqual("");
      expect(wrapper.vm.password).toEqual("");
      expect(wrapper.vm.showLoginForm).toEqual(true);
      expect(wrapper.vm.tabActive).toEqual(tabActive);
    });
    it("Receive data in computed", () => {
      expect(wrapper.vm.showTerminal).toEqual(false);
      expect(wrapper.vm.show).toEqual(true);
      expect(wrapper.vm.getListPrivateKeys).toEqual(privateKeys);
    });
    it("Check the watch action", async () => {
      wrapper.vm.username = username;
      wrapper.vm.password = password;

      expect(wrapper.vm.username).toEqual(username);
      expect(wrapper.vm.password).toEqual(password);
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", () => {
      // TODO
      expect(wrapper.find('[data-test="connect-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="console-icon"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="console-item"]').exists()).toBe(false);
    });
  });
});
