import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import TerminalConnectButton from "@/components/Terminal/TerminalConnectButton.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Terminal Connect Button", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalConnectButton>>;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(TerminalConnectButton, {
      global: {
        plugins: [router, vuetify, SnackbarPlugin],
      },
      props: {
        online: true,
        deviceUid: "a582b47a42d",
        deviceName: "test-device",
        sshid: "namespace.70-85-c2-08-60-2a@staging.shellhub.io",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the correct text based on online prop", async () => {
    const connectBtn = wrapper.find("[data-test='connect-btn']");
    expect(connectBtn.text()).toBe("Connect");
    await wrapper.setProps({ online: false });
    expect(connectBtn.text()).toBe("Offline");
  });

  it("Sets disabled attribute based on online prop", async () => {
    const connectBtn = wrapper.find("[data-test='connect-btn']");
    expect(connectBtn.attributes("disabled")).toBeUndefined();
    await wrapper.setProps({ online: false });
    expect(connectBtn.attributes("disabled")).toBeDefined();
  });

  it("Opens the Web Terminal dialog when Connect button is clicked", async () => {
    const connectBtn = wrapper.find("[data-test='connect-btn']");
    await connectBtn.trigger("click");
    expect(wrapper.vm.showWebTerminal).toBe(true);
  });

  it("Opens the Web Terminal dialog when 'Connect via web' menu item is clicked", async () => {
    const menuActivator = wrapper.findAllComponents({ name: "VBtn" }).at(1);
    await menuActivator?.trigger("click");

    const webItem = wrapper.findComponent("[data-test='Connect via web']");
    await webItem.trigger("click");

    expect(wrapper.vm.showWebTerminal).toBe(true);
  });

  it("Opens the Terminal Helper dialog when 'Connect via terminal' menu item is clicked", async () => {
    const menuActivator = wrapper.findAllComponents({ name: "VBtn" }).at(1);
    await menuActivator?.trigger("click");

    const terminalItem = wrapper.findComponent("[data-test='Connect via terminal']");
    await terminalItem.trigger("click");

    expect(wrapper.vm.showTerminalHelper).toBe(true);
  });
});
