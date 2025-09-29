import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MemberEdit from "@/components/Team/Member/MemberEdit.vue";
import { namespacesApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespaceMember } from "@/interfaces/INamespace";

type MemberEditWrapper = VueWrapper<InstanceType<typeof MemberEdit>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const members = [
  {
    id: "xxxxxxxx",
    role: "owner" as const,
  },
] as INamespaceMember[];

describe("Member Edit", () => {
  let wrapper: MemberEditWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();

  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  beforeEach(async () => {
    authStore.tenantId = "fake-tenant-data";

    wrapper = mount(MemberEdit, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        member: members[0], hasAuthorization: true,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", async () => {
    await wrapper.findComponent('[data-test="member-edit-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Edit Member Error Validation", async () => {
    mockNamespacesApi.onPatch("http://localhost:3000/api/namespaces/fake-tenant-data/members/xxxxxxxx").reply(409);

    const storeSpy = vi.spyOn(namespacesStore, "updateNamespaceMember");

    await wrapper.findComponent('[data-test="member-edit-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="role-select"]').setValue("not-right-role");

    await wrapper.findComponent('[data-test="edit-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      role: "not-right-role",
      tenant_id: "fake-tenant-data",
      user_id: "xxxxxxxx",
    });

    expect(mockSnackbar.showError).toBeCalledWith("Failed to update user role.");
  });

  it("Edit Member Success Validation", async () => {
    mockNamespacesApi.onPatch("http://localhost:3000/api/namespaces/fake-tenant-data/members/xxxxxxxx").reply(200);

    const storeSpy = vi.spyOn(namespacesStore, "updateNamespaceMember");

    await wrapper.findComponent('[data-test="member-edit-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="role-select"]').setValue("administrator");

    await wrapper.findComponent('[data-test="edit-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      role: "administrator",
      tenant_id: "fake-tenant-data",
      user_id: "xxxxxxxx",
    });
  });
});
