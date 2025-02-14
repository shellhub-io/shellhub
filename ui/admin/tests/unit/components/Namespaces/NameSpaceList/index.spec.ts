import { createVuetify } from "vuetify";
import { createStore } from "vuex";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NamespaceList from "../../../../../src/components/Namespace/NamespaceList.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

const namespaces = [
  {
    billing: null,
    created_at: "2022-04-13T11:42:49.578Z",
    devices_count: 2,
    maxDevices: 10,
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
    devices_count: 12,
    max_devices: 100,
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
    "namespaces/numberOfNamespaces": (state) => state.namespaces.length,
  },
  actions: {
    "namespaces/fetch": () => vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("Namespace Edit", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(NamespaceList, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it("Renders data in the computed", () => {
    const namespacesComputed = wrapper.vm.namespaces;
    expect(namespacesComputed).toEqual(namespaces);
  });
});
