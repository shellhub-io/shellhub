import { flushPromises, mount, VueWrapper, DOMWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import WebEndpointDelete from "@/components/WebEndpoints/WebEndpointDelete.vue";
import { webEndpointsApi, namespacesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type WebEndpointDeleteWrapper = VueWrapper<InstanceType<typeof WebEndpointDelete>>;

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const members = [
  {
    id: "xxxxxxxx",
    username: "test",
    role: "owner",
  },
];

const namespaceData = {
  name: "test",
  owner: "xxxxxxxx",
  tenant_id: "fake-tenant-data",
  members,
  max_devices: 3,
  devices_count: 3,
  devices: 2,
  created_at: "",
};

const authData = {
  status: "",
  token: "",
  user: "test",
  name: "test",
  tenant: "fake-tenant-data",
  email: "test@test.com",
  id: "xxxxxxxx",
  role: "owner",
};

describe("WebEndpointDelete.vue", () => {
  let wrapper: WebEndpointDeleteWrapper;

  let mockNamespace: MockAdapter;
  let mockWebEndpoints: MockAdapter;

  const vuetify = createVuetify();

  beforeEach(() => {
    const el = document.createElement("div");
    document.body.appendChild(el);

    localStorage.setItem("tenant", "fake-tenant-data");

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockWebEndpoints = new MockAdapter(webEndpointsApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    store.commit("auth/authSuccess", {
      role: "owner",
      tenant: "fake-tenant-data",
    });

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(WebEndpointDelete, {
      global: {
        plugins: [[store, key], vuetify],
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
    const spy = vi.spyOn(store, "dispatch");
    mockWebEndpoints
      .onDelete("http://localhost:3000/api/web-endpoints/fake-address")
      .reply(200);

    await wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="delete-btn"]').trigger("click");
    await flushPromises();

    expect(spy).toHaveBeenCalledWith("webEndpoints/delete", {
      address: "fake-address",
    });

    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Web Endpoint deleted successfully.");
    expect(wrapper.emitted("update")).toBeTruthy();
  });

  it("shows error snackbar when delete fails", async () => {
    mockWebEndpoints
      .onDelete("http://localhost:3000/api/web-endpoints/fake-address")
      .reply(500);

    await wrapper.find('[data-test="web-endpoint-delete-dialog-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="delete-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete Web Endpoint.");
  });
});
