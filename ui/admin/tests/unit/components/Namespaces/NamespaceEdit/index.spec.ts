import { createVuetify } from "vuetify";
import { createStore } from "vuex";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NamespaceEdit from "../../../../../src/components/Namespace/NamespaceEdit.vue";
import { key } from "../../../../../src/store";

type NamespaceEditWrapper = VueWrapper<InstanceType<typeof NamespaceEdit>>;

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "namespaces/put": () => vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

const namespace = {
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
};

describe("Namespace Edit", () => {
  let wrapper: NamespaceEditWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(NamespaceEdit, {
      global: {
        plugins: [[store, key], vuetify],
      },
      props: {
        namespace,
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
