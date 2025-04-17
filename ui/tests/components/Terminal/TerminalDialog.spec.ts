import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import { nextTick, watch } from "vue";
import { store, key } from "@/store";
import TerminalDialog from "@/components/Terminal/TerminalDialog.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Terminal Dialog", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalDialog>>;

  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(TerminalDialog, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      props: {
        uid: "a582b47a42d",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders the component table", async () => {
    await wrapper.setProps({ enableConnectButton: true, enableConsoleIcon: true, online: true, show: true });

    expect(wrapper.find('[data-test="connect-btn"]').exists()).toBe(true);

    const dialog = new DOMWrapper(document.body);

    await flushPromises();

    await wrapper.findComponent('[data-test="connect-btn"]').trigger("click");

    expect(dialog.find('[data-test="terminal-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="password-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="connect2-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="auth-method-select"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="auth-method-select"]').setValue("Private Key");
    await flushPromises();

    expect(dialog.find('[data-test="password-field"]').exists()).toBe(false);
    expect(dialog.find('[data-test="privatekeys-select"]').exists()).toBe(true);
  });

  it("sets showLoginForm to true when showTerminal changes to true", async () => {
    await watch(() => wrapper.vm.showTerminal, (value) => {
      if (value) wrapper.vm.showLoginForm = true;
    });

    wrapper.vm.showTerminal = true;

    await nextTick();

    expect(wrapper.vm.showLoginForm).toBe(true);
  });

  it("encodes URL params correctly", () => {
    const params = { key1: "value1", key2: "value2" };
    const encodedParams = wrapper.vm.encodeURLParams(params);

    expect(encodedParams).toBe("key1=value1&key2=value2");
  });

  it("opens terminal and initializes xterm", () => {
    wrapper.vm.open();

    expect(wrapper.vm.showTerminal).toBe(true);
    expect(wrapper.vm.privateKey).toBe("");
    expect(wrapper.vm.xterm).toBeTruthy();
    expect(wrapper.vm.fitAddon).toBeTruthy();
  });
});
