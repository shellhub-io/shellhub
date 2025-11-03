import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import SettingOwnerInfo from "@/components/Setting/SettingOwnerInfo.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespaceMember } from "@/interfaces/INamespace";

type SettingOwnerInfoWrapper = VueWrapper<InstanceType<typeof SettingOwnerInfo>>;

const members = [
  {
    id: "507f1f77bcf86cd799439011",
    role: "owner" as const,
  },
  {
    id: "507f1f77bcf86cd799745632",
    role: "operator" as const,
  },
] as INamespaceMember[];

const namespaceData = {
  name: "test",
  owner: "507f1f77bcf86cd799439011",
  tenant_id: "fake-tenant-data",
  members,
  max_devices: 3,
  devices_count: 3,
  created_at: "",
  billing: null,
  settings: {
    session_record: true,
  },
  devices_accepted_count: 3,
  devices_rejected_count: 0,
  devices_pending_count: 0,
  type: "team" as const,
};

describe("Setting Owner Info", () => {
  let wrapper: SettingOwnerInfoWrapper;
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();

  beforeEach(() => {
    namespacesStore.currentNamespace = namespaceData;

    wrapper = mount(SettingOwnerInfo, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Displays message when user is not the owner", () => {
    expect(wrapper.find('[data-test="message-div"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="contactUser-p"]').text()).toContain("Contact the owner for more information.");
  });
});
