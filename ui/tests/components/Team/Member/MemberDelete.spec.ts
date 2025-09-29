import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MemberDelete from "@/components/Team/Member/MemberDelete.vue";
import { namespacesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespaceMember } from "@/interfaces/INamespace";

type MemberDeleteWrapper = VueWrapper<InstanceType<typeof MemberDelete>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Member Delete", () => {
  let wrapper: MemberDeleteWrapper;
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  const members = [
    {
      id: "xxxxxxxx",
      role: "owner" as const,
    },
  ] as INamespaceMember[];

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant-data");

    wrapper = mount(MemberDelete, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        member: members[0],
        hasAuthorization: true,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", async () => {
    await wrapper.findComponent('[data-test="member-delete-dialog-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Delete Member Error Validation", async () => {
    mockNamespacesApi.onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/members/xxxxxxxx").reply(403);

    const storeSpy = vi.spyOn(namespacesStore, "removeMemberFromNamespace");

    await wrapper.findComponent('[data-test="member-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="member-delete-remove-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      tenant_id: "fake-tenant-data",
      user_id: "xxxxxxxx",
    });

    expect(mockSnackbar.showError).toBeCalledWith("Failed to remove user from namespace.");
  });

  it("Delete Member Success Validation", async () => {
    mockNamespacesApi.onDelete("http://localhost:3000/api/namespaces/fake-tenant-data/members/xxxxxxxx").reply(200);

    const storeSpy = vi.spyOn(namespacesStore, "removeMemberFromNamespace");

    await wrapper.findComponent('[data-test="member-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="member-delete-remove-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      tenant_id: "fake-tenant-data",
      user_id: "xxxxxxxx",
    });

    expect(mockSnackbar.showSuccess).toBeCalledWith("Successfully removed user from namespace.");
  });
});
