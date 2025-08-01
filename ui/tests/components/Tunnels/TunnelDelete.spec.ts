import { setActivePinia, createPinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import TunnelDelete from "@/components/Tunnels/TunnelDelete.vue";
import { router } from "@/router";
import { tunnelApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type TunnelDeleteWrapper = VueWrapper<InstanceType<typeof TunnelDelete>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Tunnel Delete", async () => {
  let wrapper: TunnelDeleteWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockTunnelsApi = new MockAdapter(tunnelApi.getAxios());

  beforeEach(async () => {
    wrapper = mount(TunnelDelete, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        uid: "fake-uid",
        address: "fake-address",
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

    expect(wrapper.find('[data-test="tunnel-delete-dialog-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="tunnel-delete-dialog-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="delete-btn"]').exists()).toBe(true);
  });

  it("Successfully delete tunnel", async () => {
    mockTunnelsApi.onDelete("http://localhost:3000/api/devices/fake-uid/tunnels/fake-address").reply(200);

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="tunnel-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="delete-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith("tunnels/delete", {
      uid: "fake-uid",
      address: "fake-address",
    });
  });

  it("Successfully delete tunnel", async () => {
    mockTunnelsApi.onDelete("http://localhost:3000/api/devices/fake-uid/tunnels/fake-address").reply(403);

    await wrapper.findComponent('[data-test="tunnel-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="delete-btn"]').trigger("click");

    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete tunnel.");
  });
});
