import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import ContainerList from "@/components/Containers/ContainerList.vue";
import { router } from "@/router";
import { containersApi, tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type ContainerListWrapper = VueWrapper<InstanceType<typeof ContainerList>>;

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
    tags: [{
      tenant_id: "fake-tenant-data",
      name: "test-tag",
      created_at: "",
      updated_at: "",
    }],
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
    tags: [
      {
        tenant_id: "fake-tenant-data",
        name: "test-tag",
        created_at: "",
        updated_at: "",
      },
    ],
  },
];

describe("Container List", () => {
  let wrapper: ContainerListWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockContainersApi = new MockAdapter(containersApi.getAxios());
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  localStorage.setItem("tenant", "fake-tenant-data");
  mockContainersApi.onGet("http://localhost:3000/api/containers?page=1&per_page=10&status=accepted").reply(200, containers);
  mockTagsApi
    .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
    .reply(200, []);

  beforeEach(() => {
    wrapper = mount(ContainerList, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
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
