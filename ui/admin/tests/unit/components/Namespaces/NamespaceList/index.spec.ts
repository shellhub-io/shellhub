import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { INamespace } from "@admin/interfaces/INamespace";
import { SnackbarPlugin } from "@/plugins/snackbar";
import routes from "../../../../../src/router";
import NamespaceList from "../../../../../src/components/Namespace/NamespaceList.vue";

type NamespaceListWrapper = VueWrapper<InstanceType<typeof NamespaceList>>;

const namespaces: INamespace[] = [
  {
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
    devices_pending_count: 0,
    devices_rejected_count: 1,
    max_devices: 10,
    members: [
      {
        id: "",
        role: "owner",
        username: "ossystems",
      },
    ],
    name: "ossystems",
    owner: "ossystems",
    settings: {
      session_record: true,
    },
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  },
  {
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
    devices_accepted_count: 10,
    devices_pending_count: 2,
    devices_rejected_count: 0,
    max_devices: 100,
    members: [
      {
        id: "",
        role: "owner",
        username: "ossystems",
      },
    ],
    name: "dev",
    owner: "dev",
    settings: {
      session_record: true,
    },
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  },
];

describe("Namespace List", () => {
  let wrapper: NamespaceListWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    const namespaceStore = useNamespacesStore();

    namespaceStore.namespaces = namespaces;
    namespaceStore.numberNamespaces = namespaces.length;

    vi.spyOn(namespaceStore, "fetch").mockResolvedValue(true);

    wrapper = mount(NamespaceList, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders data in the computed", () => {
    const store = useNamespacesStore();
    expect(store.list).toEqual(namespaces);
  });
});
