import { createPinia, setActivePinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { store, key } from "@/store";
import { router } from "@/router";
import { containersApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import ContainerRejectedList from "@/components/Containers/ContainerRejectedList.vue";

type ContainerRejectedListWrapper = VueWrapper<InstanceType<typeof ContainerRejectedList>>;

const containers = [
  {
    uid: "a582b47a42d",
    name: "39-5e-2a",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T18:58:53.276Z",
    online: false,
    namespace: "user",
    status: "accepted",
    tags: ["test"],
  },
  {
    uid: "a582b47a42e",
    name: "39-5e-2b",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T19:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted",
    tags: ["test"],
  },
];

describe("Container Rejected List", () => {
  let wrapper: ContainerRejectedListWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockContainersApi = new MockAdapter(containersApi.getAxios());

  beforeEach(async () => {
    mockContainersApi.onGet("http://localhost:3000/api/containers?filter=&page=1&per_page=10&status=rejected").reply(200, containers);

    wrapper = mount(ContainerRejectedList, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component HTML", async () => {
    expect(wrapper.findComponent('[data-test="container-table"]').exists()).toBe(true);
  });
});
