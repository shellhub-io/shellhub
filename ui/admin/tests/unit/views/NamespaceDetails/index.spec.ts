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
  type: "team",
  members: [
    {
      id: "6256b739302b50b6cc5eafcc",
      role: "owner",
      email: "owner@example.com",
      status: "active",
      added_at: "2022-04-13T11:43:24.668Z",
      expires_at: "0001-01-01T00:00:00Z",
    },
    {
      id: "7326b239302b50b6cc5eafdd",
      role: "administrator",
      email: "admin@example.com",
      status: "pending",
      added_at: "2022-04-14T11:43:24.668Z",
      expires_at: "0001-01-01T00:00:00Z",
    },
  ],
  settings: {
    session_record: true,
    connection_announcement: "New connection",
  },
  max_devices: 10,
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

  beforeEach(() => {
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
    const nameField = wrapper.get('[data-test="namespace-name-field"]');
    expect(nameField.text()).toContain(namespaceDetail.name);

    const tenantField = wrapper.get('[data-test="namespace-tenant-id-field"]');
    expect(tenantField.text()).toContain(namespaceDetail.tenant_id);

    const ownerField = wrapper.get('[data-test="namespace-owner-field"]');
    expect(ownerField.text()).toContain(namespaceDetail.members[0].email);

    const devicesField = wrapper.get('[data-test="namespace-devices-field"]');
    expect(devicesField.text()).toContain(String(devicesCount));

    const breakdown = wrapper.get('[data-test="namespace-devices-breakdown"]');
    expect(
      breakdown.get('[data-test="namespace-devices-accepted"]').text(),
    ).toContain(String(namespaceDetail.devices_accepted_count));
    expect(
      breakdown.get('[data-test="namespace-devices-pending"]').text(),
    ).toContain(String(namespaceDetail.devices_pending_count));
    expect(
      breakdown.get('[data-test="namespace-devices-rejected"]').text(),
    ).toContain(String(namespaceDetail.devices_rejected_count));

    const maxDevicesField = wrapper.get('[data-test="namespace-max-devices-field"]');
    expect(maxDevicesField.text()).toContain(String(namespaceDetail.max_devices));

    const sessionRecordField = wrapper.get('[data-test="namespace-session-record-field"]');
    expect(sessionRecordField.text()).toContain("Enabled");
  });

  it("Should render the props of the Namespace on screen", () => {
    const nameField = wrapper.get('[data-test="namespace-name-field"]');
    expect(nameField.text()).toContain(namespaceDetail.name);

    const tenantField = wrapper.get('[data-test="namespace-tenant-id-field"]');
    expect(tenantField.text()).toContain(namespaceDetail.tenant_id);

    const ownerField = wrapper.get('[data-test="namespace-owner-field"]');
    expect(ownerField.text()).toContain(namespaceDetail.members[0].email);

    const devicesField = wrapper.get('[data-test="namespace-devices-field"]');
    expect(devicesField.text()).toContain(String(devicesCount));

    const breakdown = wrapper.get('[data-test="namespace-devices-breakdown"]');
    expect(
      breakdown.get('[data-test="namespace-devices-accepted"]').text(),
    ).toContain(String(namespaceDetail.devices_accepted_count));
    expect(
      breakdown.get('[data-test="namespace-devices-pending"]').text(),
    ).toContain(String(namespaceDetail.devices_pending_count));
    expect(
      breakdown.get('[data-test="namespace-devices-rejected"]').text(),
    ).toContain(String(namespaceDetail.devices_rejected_count));

    const maxDevicesField = wrapper.get('[data-test="namespace-max-devices-field"]');
    expect(maxDevicesField.text()).toContain(String(namespaceDetail.max_devices));

    const sessionRecordField = wrapper.get('[data-test="namespace-session-record-field"]');
    expect(sessionRecordField.text()).toContain("Enabled");
  });
});
