import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import TerminalConnectButton from "@/components/Terminal/TerminalConnectButton.vue";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";

describe("TerminalConnectButton", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalConnectButton>>;

  const mountWrapper = (online = true) => {
    wrapper = mountComponent(TerminalConnectButton, {
      global: { plugins: [createCleanRouter()] },
      props: {
        online,
        deviceUid: "a582b47a42d",
        deviceName: "test-device",
        sshid: "namespace.70-85-c2-08-60-2a@staging.shellhub.io",
      },
    });
  };

  beforeEach(() => mountWrapper());

  afterEach(() => wrapper?.unmount());

  it("renders connect button with correct text when online", () => {
    const connectBtn = wrapper.find('[data-test="connect-btn"]');
    expect(connectBtn.text()).toBe("Connect");
    expect(connectBtn.attributes("disabled")).toBeUndefined();
  });

  it("renders offline button when device is offline", () => {
    mountWrapper(false);

    const connectBtn = wrapper.find('[data-test="connect-btn"]');
    expect(connectBtn.text()).toBe("Offline");
    expect(connectBtn.attributes("disabled")).toBeDefined();
  });

  it("applies green border when device is online", () => {
    const btnGroup = wrapper.find(".v-btn-group");
    expect(btnGroup.classes()).toContain("green-border");
  });

  it("does not apply green border when device is offline", () => {
    mountWrapper(false);

    const btnGroup = wrapper.find(".v-btn-group");
    expect(btnGroup.classes()).not.toContain("green-border");
  });

  it("disables menu dropdown when device is offline", () => {
    mountWrapper(false);

    const menuButtons = wrapper.findAllComponents({ name: "VBtn" });
    const dropdownButton = menuButtons[1];

    expect(dropdownButton.attributes("disabled")).toBeDefined();
  });

  it("opens TerminalDialog when connect button is clicked", async () => {
    const connectBtn = wrapper.find('[data-test="connect-btn"]');
    await connectBtn.trigger("click");
    await flushPromises();

    const terminalDialog = wrapper.findComponent({ name: "TerminalDialog" });
    expect(terminalDialog.props("modelValue")).toBe(true);
  });

  it("passes correct props to TerminalDialog", () => {
    const terminalDialog = wrapper.findComponent({ name: "TerminalDialog" });

    expect(terminalDialog.props("deviceUid")).toBe("a582b47a42d");
    expect(terminalDialog.props("deviceName")).toBe("test-device");
    expect(terminalDialog.props("sshid")).toBe("namespace.70-85-c2-08-60-2a@staging.shellhub.io");
  });

  it("opens TerminalDialog when 'Connect via web' menu item is clicked", async () => {
    const menuActivator = wrapper.find('[data-test="dropdown-btn"]');
    await menuActivator.trigger("click");
    await flushPromises();

    const menu = new DOMWrapper(document.body).find(".v-menu");

    const webItem = menu.find('[data-test="connect-via-web"]');
    await webItem.trigger("click");
    await flushPromises();

    const terminalDialog = wrapper.findComponent({ name: "TerminalDialog" });
    expect(terminalDialog.props("modelValue")).toBe(true);
  });

  it("opens SSHIDHelper when 'Connect via terminal' menu item is clicked", async () => {
    const menuActivator = wrapper.find('[data-test="dropdown-btn"]');
    await menuActivator.trigger("click");
    await flushPromises();

    const body = new DOMWrapper(document.body);
    const menu = body.find(".v-menu");

    const terminalItem = menu.find('[data-test="connect-via-terminal"]');
    await terminalItem.trigger("click");
    await flushPromises();

    const sshidHelper = body.find('[data-test="sshid-helper"]');
    expect(sshidHelper.exists()).toBe(true);
  });

  it("passes correct sshid prop to SSHIDHelper", () => {
    const sshidHelper = wrapper.findComponent({ name: "SSHIDHelper" });
    expect(sshidHelper.props("sshid")).toBe("namespace.70-85-c2-08-60-2a@staging.shellhub.io");
  });

  it("displays both menu items with correct icons", async () => {
    const menuActivator = wrapper.find('[data-test="dropdown-btn"]');
    await menuActivator.trigger("click");
    await flushPromises();

    const menu = new DOMWrapper(document.body).find(".v-menu");

    const webItem = menu.find('[data-test="connect-via-web"]');
    const terminalItem = menu.find('[data-test="connect-via-terminal"]');

    expect(webItem.text()).toContain("Connect via web");
    expect(webItem.find(".v-icon").classes()).toContain("mdi-application-outline");

    expect(terminalItem.text()).toContain("Connect via terminal");
    expect(terminalItem.find(".v-icon").classes()).toContain("mdi-console");
  });
});
