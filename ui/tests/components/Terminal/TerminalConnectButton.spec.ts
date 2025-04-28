import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import { store, key } from "@/store";
import TerminalConnectButton from "@/components/Terminal/TerminalConnectButton.vue";
import { router } from "@/router";

describe("Terminal Connect Button", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalConnectButton>>;

  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(TerminalConnectButton, {
      global: {
        plugins: [[store, key], router, vuetify],
      },
      props: {
        online: true,
        deviceUid: "a582b47a42d",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the correct text based on prop value", async () => {
    const connectBtn = wrapper.find("[data-test='connect-btn']");
    expect(connectBtn.text()).toBe("Connect");
    await wrapper.setProps({ online: false });
    expect(connectBtn.text()).toBe("Offline");
  });

  it("Sets disabled attribute based on prop value", async () => {
    const connectBtn = wrapper.find("[data-test='connect-btn']");
    expect(connectBtn.attributes("disabled")).toBeFalsy();
    await wrapper.setProps({ online: false });
    expect(connectBtn.attributes("disabled")).toBeDefined();
  });

  it("Opens dialog when clicked", async () => {
    const connectBtn = wrapper.find("[data-test='connect-btn']");
    await connectBtn.trigger("click");
    expect(wrapper.vm.showDialog).toBe(true);
  });
});
