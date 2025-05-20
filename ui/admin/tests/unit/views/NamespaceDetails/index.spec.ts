import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import useSnackbarStore from "@admin/store/modules/snackbar";
import { INamespace } from "@admin/interfaces/INamespace";
import { SnackbarPlugin } from "@/plugins/snackbar";
import routes from "../../../../src/router";
import NamespaceDetails from "../../../../src/views/NamespaceDetails.vue";

type NamespaceDetailsWrapper = VueWrapper<InstanceType<typeof NamespaceDetails>>;

const namespaceDetail: INamespace = {
  name: "dev",
  owner: "6256b739302b50b6cc5eafcc",
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  members: [
    {
      id: "6256b739302b50b6cc5eafcc",
      username: "antony",
      role: "owner",
    },
    {
      id: "7326b239302b50b6cc5eafdd",
      username: "test",
      role: "administrator",
    },
  ],
  settings: {
    session_record: true,
  },
  max_devices: 0,
  devices_count: 1,
  created_at: "2022-04-13T11:43:24.668Z",
  billing: undefined,
};

const mockRoute = {
  params: {
    id: namespaceDetail.tenant_id,
  },
};

describe("Namespace Details", () => {
  let wrapper: NamespaceDetailsWrapper;

  beforeEach(async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const namespaceStore = useNamespacesStore();
    namespaceStore.get = vi.fn().mockResolvedValue(undefined);
    namespaceStore.namespace = namespaceDetail;

    const snackbarStore = useSnackbarStore();
    snackbarStore.showSnackbarErrorAction = vi.fn();

    const vuetify = createVuetify();

    wrapper = mount(NamespaceDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: {
          $route: mockRoute,
        },
      },
    });

    await namespaceStore.get(mockRoute.params.id);
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Has the correct data", () => {
    expect(wrapper.vm.namespace).toEqual(namespaceDetail);
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toEqual("Namespace Details");
  });

  it("Should render the props of the Namespace on screen", () => {
    expect(wrapper.find(`[data-test='${namespaceDetail.name}']`).text()).toContain(namespaceDetail.name);
    expect(wrapper.find(`[data-test='${namespaceDetail.devices_count}']`).text()).toContain(String(namespaceDetail.devices_count));
    expect(wrapper.find(`[data-test='${namespaceDetail.owner}']`).text()).toContain(namespaceDetail.owner);
    expect(wrapper.find(`[data-test='${namespaceDetail.tenant_id}']`).text()).toContain(namespaceDetail.tenant_id);
    expect(wrapper.find(`[data-test='${namespaceDetail.settings.session_record}']`).text())
      .toContain(String(namespaceDetail.settings.session_record));
  });

  it("Should render the correct members list", () => {
    wrapper.findAll("ul").forEach((ul) => {
      ul.findAll("li").forEach((li) => {
        const fieldName = li.find("span");
        const fieldValue = li.find("span.field-value");
        expect(fieldName.exists()).toBeTruthy();
        expect(fieldValue.exists()).toBeTruthy();
      });
    });
    expect(wrapper.findAll("ul").length).toEqual(namespaceDetail.members.length);
  });
});
