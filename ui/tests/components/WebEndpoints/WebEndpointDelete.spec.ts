import { setActivePinia, createPinia } from "pinia";
import { flushPromises, mount, VueWrapper, DOMWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import WebEndpointDelete from "@/components/WebEndpoints/WebEndpointDelete.vue";
import { webEndpointsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useWebEndpointsStore from "@/store/modules/web_endpoints";

type WebEndpointDeleteWrapper = VueWrapper<InstanceType<typeof WebEndpointDelete>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("WebEndpointDelete.vue", () => {
  let wrapper: WebEndpointDeleteWrapper;
  const mockWebEndpointsApi = new MockAdapter(webEndpointsApi.getAxios());
  setActivePinia(createPinia());
  const webEndpointsStore = useWebEndpointsStore();
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(WebEndpointDelete, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        address: "fake-address",
        modelValue: true,
      },
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("opens and renders the dialog with correct elements", async () => {
    const dialog = new DOMWrapper(document.body);

    await wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="delete-btn"]').exists()).toBe(true);
  });

  it("emits update and shows success snackbar when delete succeeds", async () => {
    const storeSpy = vi.spyOn(webEndpointsStore, "deleteWebEndpoint");
    mockWebEndpointsApi
      .onDelete("http://localhost:3000/api/web-endpoints/fake-address")
      .reply(200);

    await wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="delete-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith("fake-address");

    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Web Endpoint deleted successfully.");
    expect(wrapper.emitted("update")).toBeTruthy();
  });

  it("shows error snackbar when delete fails", async () => {
    mockWebEndpointsApi
      .onDelete("http://localhost:3000/api/web-endpoints/fake-address")
      .reply(500);

    await wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="delete-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete Web Endpoint.");
  });
});
