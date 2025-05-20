import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import useSnackbarStore from "@admin/store/modules/snackbar";
import { INotificationsSuccess } from "@admin/interfaces/INotifications";
import { SnackbarPlugin } from "@/plugins/snackbar";
import NamespaceEdit from "../../../../../src/components/Namespace/NamespaceEdit.vue";

type NamespaceEditWrapper = VueWrapper<InstanceType<typeof NamespaceEdit>>;

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

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
  devices_count: 2,
  max_devices: 10,
  members: [
    {
      id: "",
      username: "ossystems",
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

describe("Namespace Edit", () => {
  let wrapper: NamespaceEditWrapper;

  beforeEach(() => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    setActivePinia(createPinia());

    const vuetify = createVuetify();

    const namespaceStore = useNamespacesStore();
    const snackbarStore = useSnackbarStore();

    vi.spyOn(namespaceStore, "put").mockResolvedValue(undefined);
    vi.spyOn(namespaceStore, "refresh").mockResolvedValue(undefined);
    vi.spyOn(snackbarStore, "showSnackbarSuccessAction").mockImplementation(() => INotificationsSuccess.namespaceEdit);
    vi.spyOn(snackbarStore, "showSnackbarErrorDefault").mockImplementation(() => vi.fn());

    wrapper = mount(NamespaceEdit, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: {
        namespace,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Calls store methods on form submission", async () => {
    const namespaceStore = useNamespacesStore();
    const snackbarStore = useSnackbarStore();

    wrapper.vm.onSubmit();

    await flushPromises();

    expect(namespaceStore.put).toHaveBeenCalled();
    expect(namespaceStore.refresh).toHaveBeenCalled();
    expect(snackbarStore.showSnackbarSuccessAction).toHaveBeenCalled();
  });
});
