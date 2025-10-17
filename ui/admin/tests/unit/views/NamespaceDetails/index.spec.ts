import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import routes from "@admin/router";
import NamespaceDetails from "@admin/views/NamespaceDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type NamespaceDetailsWrapper = VueWrapper<InstanceType<typeof NamespaceDetails>>;

const namespaceDetail: IAdminNamespace = {
  name: "dev",
  owner: "6256b739302b50b6cc5eafcc",
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  members: [
    {
      id: "6256b739302b50b6cc5eafcc",
      role: "owner",
    },
    {
      id: "7326b239302b50b6cc5eafdd",
      role: "administrator",
    },
  ],
  settings: {
    session_record: true,
  },
  max_devices: 0,
  devices_accepted_count: 1,
  devices_pending_count: 0,
  devices_rejected_count: 0,
  created_at: "2022-04-13T11:43:24.668Z",
  billing: undefined,
} as IAdminNamespace;

const devicesCount = namespaceDetail.devices_accepted_count
  + namespaceDetail.devices_pending_count
  + namespaceDetail.devices_rejected_count;

const mockRoute = {
  params: {
    id: namespaceDetail.tenant_id,
  },
};

describe("Namespace Details", () => {
  let wrapper: NamespaceDetailsWrapper;
  const pinia = createPinia();
  setActivePinia(pinia);
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();
  namespacesStore.fetchNamespaceById = vi.fn().mockResolvedValue(namespaceDetail);

  beforeEach(async () => {
    wrapper = mount(NamespaceDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: {
          $route: mockRoute,
        },
      },
    });
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
    expect(wrapper.find('[data-test="namespace-devices-count"').text()).toContain(devicesCount);
    expect(wrapper.find(`[data-test='${namespaceDetail.owner}']`).text()).toContain(namespaceDetail.owner);
    expect(wrapper.find(`[data-test='${namespaceDetail.tenant_id}']`).text()).toContain(namespaceDetail.tenant_id);
    expect(wrapper.find(`[data-test='${namespaceDetail.settings.session_record}']`).text())
      .toContain(String(namespaceDetail.settings.session_record));
  });

  it("Should render the correct members list", () => {
    wrapper.findAll("ul").forEach((ul) => {
      ul.findAll("li").forEach((li) => {
        const fieldName = li.find("[data-test$='-item']");
        const fieldValue = li.find("[data-test$='-value']");
        expect(fieldName.exists()).toBeTruthy();
        expect(fieldValue.exists()).toBeTruthy();
      });
    });
    expect(wrapper.findAll("ul").length).toEqual(namespaceDetail.members.length);
  });
});
