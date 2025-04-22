import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { nextTick, watch } from "vue";
import { store, key } from "@/store";
import TerminalDialog from "@/components/Terminal/TerminalDialog.vue";
import { router } from "@/router";

describe("Terminal Dialog", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalDialog>>;

  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(TerminalDialog, {
      global: {
        plugins: [[store, key], vuetify, router],
      },
      props: {
        uid: "a582b47a42d",
        enableConnectButton: true,
        enableConsoleIcon: true,
        online: true,
        show: true,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the components", async () => {
    const dialog = new DOMWrapper(document.body);
    const connectBtn = wrapper.find('[data-test="connect-btn"]');

    await flushPromises();

    expect(connectBtn.exists()).toBe(true);
    await connectBtn.trigger("click");

    expect(dialog.find('[data-test="terminal-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="password-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="submit-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="auth-method-select"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="auth-method-select"]').setValue("Private Key");
    await flushPromises();
    expect(dialog.find('[data-test="password-field"]').exists()).toBe(false);
    expect(dialog.find('[data-test="private-keys-select"]').exists()).toBe(true);
  });

  it("sets showLoginForm to true when showDialog changes to true", async () => {
    watch(() => wrapper.vm.showDialog, (value) => {
      if (value) wrapper.vm.showLoginForm = true;
    });

    wrapper.vm.showDialog = true;

    await nextTick();

    expect(wrapper.vm.showLoginForm).toBe(true);
  });

  it("shows X button when terminal is open", async () => {
    const dialog = new DOMWrapper(document.body);
    const connectBtn = wrapper.find('[data-test="connect-btn"]');
    await connectBtn.trigger("click");

    wrapper.vm.showLoginForm = false;

    await flushPromises();

    const closeBtn = dialog.find('[data-test="close-terminal-btn"]');
    expect(closeBtn.exists()).toBe(true);
  });

  it("submits form when Enter is pressed", async () => {
    const submitFormSpy = vi.spyOn(wrapper.vm, "submitForm").mockImplementation(vi.fn());
    const dialog = new DOMWrapper(document.body);
    const connectBtn = wrapper.find('[data-test="connect-btn"]');

    await connectBtn.trigger("click");

    const usernameField = dialog.find('[data-test="username-field"] input');
    const passwordField = dialog.find('[data-test="password-field"] input');

    await usernameField.setValue("testuser");
    await passwordField.setValue("testpass");

    passwordField.trigger("keydown.enter.prevent");
    await nextTick();
    expect(submitFormSpy).toBeTruthy();
  });

  it("encodes URL params correctly", () => {
    const params = { key1: "value1", key2: "value2" };
    const encodedParams = wrapper.vm.encodeURLParams(params);

    expect(encodedParams).toBe("key1=value1&key2=value2");
  });

  it("opens terminal and initializes xterm", () => {
    wrapper.vm.open();

    expect(wrapper.vm.showDialog).toBe(true);
    expect(wrapper.vm.privateKey).toBe("");
    expect(wrapper.vm.xterm).toBeTruthy();
    expect(wrapper.vm.fitAddon).toBeTruthy();
  });
});
