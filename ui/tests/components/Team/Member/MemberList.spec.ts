import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MemberList from "@/components/Team/Member/MemberList.vue";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockMembers } from "@tests/mocks/namespace";
import useNamespacesStore from "@/store/modules/namespaces";
import * as hasPermissionModule from "@/utils/permission";

describe("MemberList", () => {
  let wrapper: VueWrapper<InstanceType<typeof MemberList>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const mountWrapper = ({
    hasMembers = true,
    canEditMember = true,
    canRemoveMember = true,
  } = {}) => {
    vi.spyOn(hasPermissionModule, "default").mockImplementation((permission: string) => {
      if (permission === "namespace:editMember") return canEditMember;
      if (permission === "namespace:removeMember") return canRemoveMember;
      return false;
    });

    wrapper = mountComponent(MemberList, {
      piniaOptions: {
        initialState: {
          auth: { tenantId: "fake-tenant-data" },
          namespaces: { currentNamespace: { members: hasMembers ? mockMembers : null } },
        },
      },
    });
    namespacesStore = useNamespacesStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  it("renders the member table", () => {
    expect(wrapper.find('[data-test="member-table"]').exists()).toBe(true);
  });

  it("renders table headers", () => {
    const headers = wrapper.findAll('[data-test="member-table-headers"]');
    expect(headers).toHaveLength(3);
    expect(headers[0].text()).toContain("Email");
    expect(headers[1].text()).toContain("Role");
    expect(headers[2].text()).toContain("Actions");
  });

  it("renders member rows", () => {
    const rows = wrapper.find('[data-test="member-table-rows"]').findAll("tr");
    expect(rows).toHaveLength(mockMembers.length);
  });

  it("displays member email with icon", () => {
    const firstRow = wrapper.find('[data-test="member-table-rows"]').findAll("tr")[0];
    expect(firstRow.text()).toContain(mockMembers[0].email);
    expect(firstRow.find(".v-icon").classes()).toContain("mdi-account");
  });

  it("displays member role capitalized", () => {
    const rows = wrapper.find('[data-test="member-table-rows"]').findAll("tr");
    rows.forEach((row, index) => {
      const roleCell = row.findAll("td")[1];
      expect(roleCell.classes()).toContain("text-capitalize");
      expect(roleCell.text()).toBe(mockMembers[index].role);
    });
  });

  it("displays tooltip with added_at date when member was added", () => {
    const tooltip = new DOMWrapper(document.body).find('[data-test="added-at-tooltip"]');

    expect(tooltip.exists()).toBe(true);
    expect(tooltip.text()).toContain("This member was added on");
  });

  it("displays actions menu for non-owner members", () => {
    const nonOwnerRow = wrapper.find('[data-test="member-table-rows"]').findAll("tr")[1];
    const actionsButton = nonOwnerRow.find('[data-test="namespace-member-actions"]');
    expect(actionsButton.exists()).toBe(true);
    expect(actionsButton.find(".v-icon").classes()).toContain("mdi-format-list-bulleted");
  });

  it("does not display actions menu for owner members", () => {
    const ownerRow = wrapper.find('[data-test="member-table-rows"]').findAll("tr")[0];
    const actionsButton = ownerRow.find('[data-test="namespace-member-actions"]');
    expect(actionsButton.exists()).toBe(false);
  });

  it("displays tooltip for owner member indicating no modification allowed", () => {
    const tooltip = new DOMWrapper(document.body).find('[data-test="owner-actions-tooltip"]');

    expect(tooltip.exists()).toBe(true);
    expect(tooltip.text()).toContain("No one can modify the owner of this namespace.");
  });

  it("displays no data message when members list is empty", () => {
    wrapper.unmount();
    mountWrapper({ hasMembers: false });

    expect(wrapper.text()).toContain("No data available");
  });

  it("successfully fetches namespace members", async () => {
    await wrapper.findComponent({ name: "MemberDelete" }).vm.$emit("update");
    await flushPromises();

    expect(namespacesStore.fetchNamespace).toHaveBeenCalledWith("fake-tenant-data");
  });

  it("handles 403 error when user doesn't have permission to view namespace", async () => {
    vi.mocked(namespacesStore.fetchNamespace).mockRejectedValueOnce(
      createAxiosError(403, "Forbidden"),
    );

    await wrapper.findComponent({ name: "MemberDelete" }).vm.$emit("update");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("You don't have permission to view this namespace.");
  });

  it("handles generic error when fetching namespace fails", async () => {
    vi.mocked(namespacesStore.fetchNamespace).mockRejectedValueOnce(
      createAxiosError(500, "Internal Server Error"),
    );

    await wrapper.findComponent({ name: "MemberDelete" }).vm.$emit("update");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch namespace members.");
  });

  it("emits clickSortableIcon event when clicking sortable header", async () => {
    // Note: Current headers are not sortable, but testing the event emission mechanism
    // In case sortable headers are added in the future
    wrapper.vm.$emit("clickSortableIcon", "email");
    await flushPromises();

    expect(wrapper.emitted("clickSortableIcon")).toBeTruthy();
    expect(wrapper.emitted("clickSortableIcon")?.[0]).toEqual(["email"]);
  });

  it("renders MemberEdit component for non-owner members", () => {
    const nonOwnerRow = wrapper.find('[data-test="member-table-rows"]').findAll("tr")[1];
    const actionsButton = nonOwnerRow.find('[data-test="namespace-member-actions"]');
    expect(actionsButton.exists()).toBe(true);
  });

  it("renders MemberDelete component for non-owner members", () => {
    const nonOwnerRow = wrapper.find('[data-test="member-table-rows"]').findAll("tr")[1];
    const actionsButton = nonOwnerRow.find('[data-test="namespace-member-actions"]');
    expect(actionsButton.exists()).toBe(true);
  });

  it("disables edit action when user doesn't have editMember permission", async () => {
    mountWrapper({ canEditMember: false, canRemoveMember: true });

    const nonOwnerRow = wrapper.find('[data-test="member-table-rows"]').findAll("tr")[1];
    const actionsButton = nonOwnerRow.find('[data-test="namespace-member-actions"]');
    await actionsButton.trigger("click");
    await flushPromises();

    const editButton = wrapper.findComponent({ name: "MemberEdit" });
    expect(editButton.find('[data-test="member-edit-btn"]').classes()).toContain("v-list-item--disabled");
  });

  it("disables remove action when user doesn't have removeMember permission", async () => {
    mountWrapper({ canEditMember: true, canRemoveMember: false });

    const nonOwnerRow = wrapper.find('[data-test="member-table-rows"]').findAll("tr")[1];
    const actionsButton = nonOwnerRow.find('[data-test="namespace-member-actions"]');
    await actionsButton.trigger("click");
    await flushPromises();

    const deleteButton = wrapper.findComponent({ name: "MemberDelete" });
    expect(deleteButton.find('[data-test="member-delete-dialog-btn"]').classes()).toContain("v-list-item--disabled");
  });
});
