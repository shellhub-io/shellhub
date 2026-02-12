import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MemberEdit from "@/components/Team/Member/MemberEdit.vue";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockMember } from "@tests/mocks/namespace";
import useNamespacesStore from "@/store/modules/namespaces";

describe("MemberEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof MemberEdit>>;
  let store: ReturnType<typeof useNamespacesStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = (
    { hasAuthorization = true } = {},
  ) => {
    wrapper = mountComponent(MemberEdit, {
      props: {
        member: mockMember,
        hasAuthorization,
      },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          auth: { tenantId: "fake-tenant-data" },
        },
      },
    });
    store = useNamespacesStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  it("renders the list item with edit button", () => {
    expect(wrapper.find('[data-test="member-edit-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="member-edit-title"]').text()).toBe("Edit");
    expect(wrapper.find(".v-icon").classes()).toContain("mdi-pencil");
  });

  it("disables the list item when user doesn't have authorization", () => {
    mountWrapper({ hasAuthorization: false });

    const listItem = wrapper.find('[data-test="member-edit-btn"]');
    expect(listItem.classes()).toContain("v-list-item--disabled");
  });

  it("opens dialog when clicking the list item", async () => {
    await wrapper.find('[data-test="member-edit-btn"]').trigger("click");
    await flushPromises();

    const dialogEl = dialog.find('[data-test="member-edit-dialog"]');
    expect(dialogEl.exists()).toBe(true);
    expect(dialogEl.text()).toContain("Update member role");
  });

  it("displays role selector in dialog", async () => {
    await wrapper.find('[data-test="member-edit-btn"]').trigger("click");
    await flushPromises();

    const roleSelect = dialog.find('[data-test="role-select"]');
    expect(roleSelect.exists()).toBe(true);
  });

  it("successfully updates member role", async () => {
    await wrapper.find('[data-test="member-edit-btn"]').trigger("click");
    await flushPromises();

    const roleSelect = wrapper.findComponent({ name: "RoleSelect" });
    await roleSelect.setValue("administrator");
    await flushPromises();

    await dialog.find('[data-test="edit-btn"]').trigger("click");
    await flushPromises();

    expect(store.updateNamespaceMember).toHaveBeenCalledWith({
      user_id: mockMember.id,
      tenant_id: "fake-tenant-data",
      role: "administrator",
    });
    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Successfully updated user role.");
    expect(wrapper.emitted("update")).toBeTruthy();
  });

  it("handles 400 error when user isn't linked to namespace", async () => {
    vi.mocked(store.updateNamespaceMember).mockRejectedValueOnce(
      createAxiosError(400, "Bad Request"),
    );

    await wrapper.find('[data-test="member-edit-btn"]').trigger("click");
    await flushPromises();

    await dialog.find('[data-test="edit-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("The user isn't linked to the namespace.");
    expect(wrapper.emitted("update")).toBeFalsy();
  });

  it("handles 403 error when user doesn't have permission to assign role", async () => {
    vi.mocked(store.updateNamespaceMember).mockRejectedValueOnce(
      createAxiosError(403, "Forbidden"),
    );

    await wrapper.find('[data-test="member-edit-btn"]').trigger("click");
    await flushPromises();

    await dialog.find('[data-test="edit-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("You don't have permission to assign a role to the user.");
    expect(wrapper.emitted("update")).toBeFalsy();
  });

  it("handles 404 error when username doesn't exist", async () => {
    vi.mocked(store.updateNamespaceMember).mockRejectedValueOnce(
      createAxiosError(404, "Not Found"),
    );

    await wrapper.find('[data-test="member-edit-btn"]').trigger("click");
    await flushPromises();

    await dialog.find('[data-test="edit-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("The username doesn't exist.");
    expect(wrapper.emitted("update")).toBeFalsy();
  });

  it("handles generic error when updating member fails", async () => {
    vi.mocked(store.updateNamespaceMember).mockRejectedValueOnce(
      createAxiosError(500, "Internal Server Error"),
    );

    await wrapper.find('[data-test="member-edit-btn"]').trigger("click");
    await flushPromises();

    await dialog.find('[data-test="edit-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update user role.");
    expect(wrapper.emitted("update")).toBeFalsy();
  });

  it("closes dialog when clicking cancel button", async () => {
    await wrapper.find('[data-test="member-edit-btn"]').trigger("click");
    await flushPromises();

    await dialog.find('[data-test="close-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
  });

  it("emits update event and closes dialog after successful update", async () => {
    await wrapper.find('[data-test="member-edit-btn"]').trigger("click");
    await flushPromises();

    await dialog.find('[data-test="edit-btn"]').trigger("click");
    await flushPromises();

    expect(wrapper.emitted("update")).toHaveLength(1);

    expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
  });
});
