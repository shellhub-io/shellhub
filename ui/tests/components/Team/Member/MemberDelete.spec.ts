import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MemberDelete from "@/components/Team/Member/MemberDelete.vue";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockMember } from "@tests/mocks/namespace";
import useNamespacesStore from "@/store/modules/namespaces";
import handleError from "@/utils/handleError";

describe("MemberDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof MemberDelete>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = (hasAuthorization = true) => {
    wrapper = mountComponent(MemberDelete, {
      props: { member: mockMember, hasAuthorization },
      attachTo: document.body,
      piniaOptions: { initialState: { auth: { tenantId: "fake-tenant-data" } } },
    });

    namespacesStore = useNamespacesStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  it("renders the list item with delete button", () => {
    expect(wrapper.find('[data-test="member-delete-dialog-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="member-delete-dialog-btn"]').text()).toBe("Remove");
    expect(wrapper.find(".v-icon").classes()).toContain("mdi-delete");
  });

  it("disables the list item when user doesn't have authorization", () => {
    mountWrapper(false);

    const listItem = wrapper.find("[data-test='member-delete-dialog-btn']");
    expect(listItem.classes()).toContain("v-list-item--disabled");
  });

  it("opens dialog when clicking the list item", async () => {
    await wrapper.find("[data-test='member-delete-dialog-btn']").trigger("click");
    await flushPromises();

    const dialogCard = dialog.find('[data-test="member-delete-card"]');
    expect(dialogCard.exists()).toBe(true);
    expect(dialogCard.text()).toContain("Are you sure?");
    expect(dialogCard.text()).toContain("You are about to remove this user from the namespace");
  });

  it("closes dialog when clicking the cancel button", async () => {
    await wrapper.find("[data-test='member-delete-dialog-btn']").trigger("click");
    await flushPromises();

    await dialog.find('[data-test="member-delete-close-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
  });

  it("successfully removes member from namespace", async () => {
    await wrapper.find("[data-test='member-delete-dialog-btn']").trigger("click");
    await flushPromises();

    await dialog.find('[data-test="member-delete-remove-btn"]').trigger("click");
    await flushPromises();

    expect(namespacesStore.removeMemberFromNamespace).toHaveBeenCalledWith({
      user_id: mockMember.id,
      tenant_id: "fake-tenant-data",
    });
    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Successfully removed user from namespace.");
    expect(wrapper.emitted("update")).toBeTruthy();
  });

  it("handles generic error when removing member fails", async () => {
    const error = createAxiosError(500, "Internal Server Error");
    vi.mocked(namespacesStore.removeMemberFromNamespace).mockRejectedValueOnce(error);

    await wrapper.find("[data-test='member-delete-dialog-btn']").trigger("click");
    await flushPromises();

    await dialog.find('[data-test="member-delete-remove-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove user from namespace.");
    expect(wrapper.emitted("update")).toBeFalsy();
    expect(handleError).toHaveBeenCalledWith(error);
  });

  it("emits update event and closes dialog after successful removal", async () => {
    await wrapper.find("[data-test='member-delete-dialog-btn']").trigger("click");
    await flushPromises();

    await dialog.find('[data-test="member-delete-remove-btn"]').trigger("click");
    await flushPromises();

    expect(wrapper.emitted("update")).toHaveLength(1);
    expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
  });
});
