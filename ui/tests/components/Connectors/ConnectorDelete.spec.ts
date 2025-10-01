import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ConnectorDelete from "@/components/Connector/ConnectorDelete.vue";
import { namespacesApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useConnectorStore from "@/store/modules/connectors";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type ConnectorDeleteWrapper = VueWrapper<InstanceType<typeof ConnectorDelete>>;

describe("Connector Delete", () => {
  let wrapper: ConnectorDeleteWrapper;
  setActivePinia(createPinia());
  const connectorStore = useConnectorStore();
  const vuetify = createVuetify();

  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  beforeEach(async () => {
    wrapper = mount(ConnectorDelete, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        uid: "fake-fingerprint",
        hasAuthorization: true,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Successfully removes a connector", async () => {
    await wrapper.setProps({ uid: "fake-fingerprint" });
    await wrapper.findComponent('[data-test="connector-remove-btn"]').trigger("click");
    mockNamespacesApi.onDelete("http://localhost:3000/api/connector/fake-fingerprint").reply(200);
    const storeSpy = vi.spyOn(connectorStore, "deleteConnector");
    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    expect(storeSpy).toHaveBeenCalledWith("fake-fingerprint");
  });

  it("Shows error snackbar if removing a connector fails", async () => {
    await wrapper.setProps({ uid: "fake-fingerprint" });
    await wrapper.findComponent('[data-test="connector-remove-btn"]').trigger("click");
    mockNamespacesApi.onDelete("http://localhost:3000/api/connector/fake-fingerprint").reply(404); // non-existent key
    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove connector.");
  });
});
