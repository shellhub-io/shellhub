import { createVuetify } from "vuetify";
import { createStore } from "vuex";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NamespaceExport from "../../../../../src/components/Namespace/NamespaceExport.vue";
import { key } from "../../../../../src/store";

type NamespaceExportWrapper = VueWrapper<InstanceType<typeof NamespaceExport>>;

const namespaces = [
  {
    billing: null,
    created_at: "2022-04-13T11:42:49.578Z",
    devices_count: 2,
    max_devices: 10,
    members: [
      {
        id: "",
        role: "admin",
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
    billing: null,
    created_at: "2022-04-13T11:42:49.578Z",
    devices_count: 1,
    max_devices: 16,
    members: [
      {
        id: "",
        role: "admin",
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

const store = createStore({
  state: {
    namespaces,
  },
  getters: {
    "namespaces/list": (state) => state.namespaces,
  },
  actions: {
    "namespaces/fetch": () => vi.fn(),
    "namespaces/exportNamespacesToCsv": () => vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("Namespace Export", () => {
  let wrapper: NamespaceExportWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(NamespaceExport, {
      global: {
        plugins: [[store, key], vuetify],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
