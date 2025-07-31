import { createPinia, setActivePinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import TunnelCreate from "@/components/Tunnels/TunnelCreate.vue";
import { router } from "@/router";
import { tunnelApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type TunnelCreateWrapper = VueWrapper<InstanceType<typeof TunnelCreate>>;

const tunnelResponse = {
  address: "9a8df9321368d567cfac8679cec7848c",
  namespace: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
  device: "13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a",
  host: "127.0.0.1",
  port: 8080,
};

describe("Tunnel Create", async () => {
  let wrapper: TunnelCreateWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockTunnelsApi = new MockAdapter(tunnelApi.getAxios());

  beforeEach(async () => {
    wrapper = mount(TunnelCreate, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      props: {
        uid: "fake-uid",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component table", async () => {
    const dialog = new DOMWrapper(document.body);

    expect(wrapper.find('[data-test="create-icon"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="tunnel-create-dialog-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.find('[data-test="tunnel-create-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="create-dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tunnel-create-alert"]').exists()).toBe(false);
    expect(dialog.find('[data-test="tunnel-create-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="timeout-combobox"]').exists()).toBe(true);
    expect(dialog.find('[data-test="address-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="port-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="custom-timeout"]').exists()).toBe(false);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="create-tunnel-btn"]').exists()).toBe(true);
  });

  it("Successfully added tunnel", async () => {
    mockTunnelsApi.onPost("http://localhost:3000/api/devices/fake-uid/tunnels").reply(200, tunnelResponse);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="tunnel-create-dialog-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="address-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue(8080);
    await wrapper.findComponent('[data-test="timeout-combobox"]').setValue(-1);

    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("tunnels/create", {
      uid: "fake-uid",
      host: "127.0.0.1",
      ttl: -1,
      port: 8080,
    });
  });

  it("Successfully added tunnel (custom expiration)", async () => {
    mockTunnelsApi.onPost("http://localhost:3000/api/devices/fake-uid/tunnels").reply(200, tunnelResponse);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="tunnel-create-dialog-btn"]').trigger("click");

    await flushPromises();

    await wrapper.findComponent('[data-test="address-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue(8080);
    await wrapper.findComponent('[data-test="timeout-combobox"]').setValue("custom");
    await wrapper.findComponent('[data-test="custom-timeout"]').setValue(6000);

    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("tunnels/create", {
      uid: "fake-uid",
      host: "127.0.0.1",
      ttl: 6000,
      port: 8080,
    });
  });

  it("Failed to add tunnel", async () => {
    mockTunnelsApi.onPost("http://localhost:3000/api/devices/fake-uid/tunnels").reply(403);

    await wrapper.findComponent('[data-test="tunnel-create-dialog-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="address-text"]').setValue("bad-address");
    await wrapper.findComponent('[data-test="port-text"]').setValue("bad-port");
    await wrapper.findComponent('[data-test="timeout-combobox"]').setValue(-1);

    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(wrapper
      .findComponent('[data-test="tunnel-create-alert"]').text()).toBe("This device has reached the maximum allowed number of tunnels");
  });
});
