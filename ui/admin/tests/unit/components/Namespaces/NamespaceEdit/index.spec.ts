import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceEdit from "@admin/components/Namespace/NamespaceEdit.vue";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

const namespace = {
  billing: {
    active: true,
    current_period_end: "",
    customer_id: "",
    payment_failed: null,
    payment_method_id: "",
    price_id: "",
    state: "",
    sub_item_id: "",
    subscription_id: "",
  },
  created_at: "2022-04-13T11:42:49.578Z",
  devices_accepted_count: 1,
  devices_pending_count: 1,
  devices_rejected_count: 0,
  max_devices: 10,
  members: [
    {
      id: "",
      role: "owner" as const,
    },
  ],
  name: "ossystems",
  owner: "ossystems",
  settings: {
    session_record: true,
  },
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
};

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Namespace Edit", () => {
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();
  namespacesStore.updateNamespace = vi.fn();
  namespacesStore.fetchNamespaceList = vi.fn();

  const wrapper = mount(NamespaceEdit, {
    global: {
      plugins: [createVuetify()],
      provide: { [SnackbarInjectionKey]: mockSnackbar },
    },
    props: { namespace: namespace as IAdminNamespace },
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Has the correct initial data", () => {
    expect(wrapper.vm.name).toBe(namespace.name);
    expect(wrapper.vm.maxDevices).toBe(namespace.max_devices);
    expect(wrapper.vm.sessionRecord).toBe(namespace.settings.session_record);
  });

  it("Calls namespace store and snackbar on form submission", async () => {
    wrapper.vm.submitForm();
    await flushPromises();
    expect(namespacesStore.updateNamespace).toHaveBeenCalled();
    expect(namespacesStore.fetchNamespaceList).toHaveBeenCalled();
    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespace updated successfully.");
  });
});
