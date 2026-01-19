import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useUsersStore from "@admin/store/modules/users";
import UserDelete from "@admin/components/User/UserDelete.vue";
import { Router } from "vue-router";

const triggerButtonTemplate = `
  <template #default="{ openDialog }">
    <button 
      data-test="trigger-button" 
      @click="openDialog"
    >
      Delete
    </button>
  </template>
`;

describe("UserDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserDelete>>;
  let usersStore: ReturnType<typeof useUsersStore>;
  let router: Router;
  const mockUserId = "user-123";

  const mountWrapper = (props: { redirect?: boolean; showTooltip?: boolean } = {}) => {
    router = createCleanAdminRouter();

    wrapper = mountComponent(UserDelete, {
      global: { plugins: [router] },
      props: {
        id: mockUserId,
        ...props,
      },
      slots: { default: triggerButtonTemplate },
      attachTo: document.body,
    });

    usersStore = useUsersStore();
  };

  const openDialog = async () => {
    await wrapper.find('[data-test="trigger-button"]').trigger("click");
    return new DOMWrapper(document.body).find('[role="dialog"]');
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the slot content", () => {
      const trigger = wrapper.find('[data-test="trigger-button"]');
      expect(trigger.exists()).toBe(true);
      expect(trigger.text()).toBe("Delete");
    });

    it("does not show the dialog initially", () => {
      expect(new DOMWrapper(document.body).find('[role="dialog"]').exists()).toBe(false);
    });

    it("shows tooltip when showTooltip prop is true", () => {
      wrapper.unmount();
      mountWrapper({ showTooltip: true });

      const tooltip = wrapper.findComponent({ name: "VTooltip" });
      expect(tooltip.exists()).toBe(true);
      expect(tooltip.props("text")).toBe("Remove");
      expect(tooltip.props("disabled")).toBe(false);
    });

    it("disables tooltip when showTooltip prop is false", () => {
      const tooltip = wrapper.findComponent({ name: "VTooltip" });
      expect(tooltip.props("disabled")).toBe(true);
    });
  });

  describe("opening dialog", () => {
    beforeEach(() => mountWrapper());

    it("shows dialog when clicking the trigger", async () => {
      const dialog = await openDialog();

      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Are you sure?");
      expect(dialog.text()).toContain("You are about to remove this user");
    });
  });

  describe("deleting user without redirect", () => {
    beforeEach(() => mountWrapper({ redirect: false }));

    it("calls store actions and shows success message on confirm", async () => {
      const dialog = await openDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(usersStore.deleteUser).toHaveBeenCalledWith(mockUserId);
      expect(usersStore.fetchUsersList).toHaveBeenCalled();
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("User removed successfully.");
    });

    it("emits update event after successful deletion", async () => {
      const dialog = await openDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("shows error message when delete fails", async () => {
      vi.mocked(usersStore.deleteUser).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const dialog = await openDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove the user.");
      expect(wrapper.emitted("update")).toBeUndefined();
    });
  });

  describe("deleting user with redirect", () => {
    beforeEach(() => mountWrapper({ redirect: true }));

    it("redirects to users page after successful deletion", async () => {
      const pushSpy = vi.spyOn(router, "push");

      const dialog = await openDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(usersStore.deleteUser).toHaveBeenCalledWith(mockUserId);
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("User removed successfully.");
      expect(pushSpy).toHaveBeenCalledWith("/users");
    });

    it("does not redirect when delete fails", async () => {
      vi.mocked(usersStore.deleteUser).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );
      const pushSpy = vi.spyOn(router, "push");

      const dialog = await openDialog();
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove the user.");
      expect(pushSpy).not.toHaveBeenCalled();
    });
  });
});
