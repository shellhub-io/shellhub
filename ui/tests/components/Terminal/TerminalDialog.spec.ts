import MockAdapter from "axios-mock-adapter";
import axios from "axios";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi, afterEach } from "vitest";
import { createRouter, createWebHistory } from "vue-router";
import { createPinia, setActivePinia } from "pinia";
import { store, key } from "@/store";
import TerminalDialog from "@/components/Terminal/TerminalDialog.vue";
import { routes } from "@/router";
import { TerminalAuthMethods } from "@/interfaces/ITerminal";

vi.mock("xterm", () => ({
  Terminal: vi.fn().mockImplementation(() => ({
    open: vi.fn(),
    focus: vi.fn(),
    write: vi.fn(),
    onData: vi.fn(),
    onResize: vi.fn(),
    loadAddon: vi.fn(),
    cols: 80,
    rows: 24,
  })),
}));

describe("Terminal Dialog", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalDialog>>;
  let dialog: DOMWrapper<HTMLElement>;
  setActivePinia(createPinia());

  const router = createRouter({
    history: createWebHistory(),
    routes,
  });

  const vuetify = createVuetify();
  beforeEach(async () => {
    wrapper = mount(TerminalDialog, {
      global: {
        plugins: [[store, key], vuetify, router],
      },
      props: {
        modelValue: true,
        deviceUid: "a582b47a42d",
      },
    });

    wrapper.vm.showDialog = true;
    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Renders form or terminal based on showLoginForm", async () => {
    expect(dialog.find("[data-test='terminal-container']").exists()).toBe(false);
    expect(dialog.find("[data-test='terminal-login-form']").exists()).toBe(true);

    wrapper.vm.showLoginForm = false;
    await flushPromises();

    expect(dialog.find("[data-test='terminal-container']").exists()).toBe(true);
    expect(dialog.find("[data-test='terminal-login-form']").exists()).toBe(false);
  });

  it("Shows X button when terminal is open", async () => {
    wrapper.vm.showLoginForm = false;

    await flushPromises();

    const closeBtn = dialog.find('[data-test="close-terminal-btn"]');
    expect(closeBtn.exists()).toBe(true);
  });

  it("Closes the terminal dialog on double ESC press", async () => {
    wrapper.vm.showLoginForm = false;
    const escEvent = new KeyboardEvent("keyup", { key: "Escape", bubbles: true });

    document.dispatchEvent(escEvent);
    expect(wrapper.vm.showDialog).toBe(true);

    document.dispatchEvent(escEvent);
    expect(wrapper.vm.showDialog).toBe(false);
  });

  it("sets token and closes login form on successful connect", async () => {
    const mockAxios = new MockAdapter(axios);
    mockAxios.onPost("ws/ssh").reply(200, { token: "fake-token" });

    wrapper.vm.handleSubmit({
      authenticationMethod: TerminalAuthMethods.Password,
      username: "test-user",
      password: "test-pass",
    });

    await flushPromises();

    expect(wrapper.vm.token).toBe("fake-token");
    expect(wrapper.vm.showLoginForm).toBe(false);
  });

  it("opens the terminal when route matches /devices/:deviceUid/terminal", async () => {
    wrapper.vm.showDialog = false;
    await flushPromises();

    expect(wrapper.vm.showDialog).toBe(false);

    await router.push("/devices/a582b47a42d/terminal");
    await flushPromises();

    expect(wrapper.vm.showDialog).toBe(true);
  });
});
