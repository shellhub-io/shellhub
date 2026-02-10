import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import { Router } from "vue-router";
import UserDelete from "@/components/User/UserDelete.vue";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import Login from "@/views/Login.vue";
import { createAxiosError } from "@tests/utils/axiosError";
import useAuthStore from "@/store/modules/auth";
import { createCleanRouter } from "@tests/utils/router";
import { routes } from "@/router";
import { mockNamespace } from "@tests/mocks";

const mockRoutes = [
  ...routes,
  // Add Login route without beforeEnter guard
  { path: "/login", name: "Login", meta: { layout: "LoginLayout", requiresAuth: false }, component: Login },
];

describe("UserDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserDelete>>;
  let dialog: DOMWrapper<Element>;
  let authStore: ReturnType<typeof useAuthStore>;
  let router: Router;

  const mountWrapper = (modelValue = false, hasNamespaces = false) => {
    router = createCleanRouter(mockRoutes);

    wrapper = mountComponent(UserDelete, {
      props: { modelValue },
      attachTo: document.body,
      global: { plugins: [router] },
      piniaOptions: {
        initialState: {
          namespaces: { namespaceList: hasNamespaces ? [mockNamespace] : [] },
        },
      },
    });

    authStore = useAuthStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => {
    mountWrapper();
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering when closed", () => {
    it("does not render dialog when modelValue is false", () => {
      expect(dialog.find('[data-test="user-delete-dialog"]').exists()).toBe(false);
    });
  });

  describe("Component rendering when open without namespaces", () => {
    beforeEach(async () => {
      wrapper.unmount();
      document.body.innerHTML = "";
      mountWrapper(true, false);
      await flushPromises();
    });

    it("renders MessageDialog with correct props", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.exists()).toBe(true);
      expect(messageDialog.props("title")).toBe("Confirm Account Deletion");
      expect(messageDialog.props("icon")).toBe("mdi-account-remove");
      expect(messageDialog.props("iconColor")).toBe("error");
      expect(messageDialog.props("cancelText")).toBe("Cancel");
      expect(messageDialog.props("confirmText")).toBe("Delete Account");
      expect(messageDialog.props("confirmColor")).toBe("error");
    });

    it("renders correct description when no namespaces", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("description")).toBe(
        "Are you sure you want to delete your account? This action cannot be undone.",
      );
    });

    it("enables delete button when no namespaces", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("confirmDisabled")).toBe(false);
    });

    it("does not show namespace warning when no namespaces", () => {
      expect(dialog.find('[data-test="namespace-warning"]').exists()).toBe(false);
    });

    it("renders close and delete buttons", () => {
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="delete-user-btn"]').exists()).toBe(true);
    });
  });

  describe("Component rendering when open with namespaces", () => {
    beforeEach(async () => {
      wrapper.unmount();
      document.body.innerHTML = "";
      mountWrapper(true, true);
      await flushPromises();
    });

    it("renders correct description when has namespaces", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("description")).toBe(
        "You cannot delete your account while you have active namespaces.",
      );
    });

    it("disables delete button when has namespaces", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("confirmDisabled")).toBe(true);
    });

    it("shows namespace warning alert", () => {
      const warning = dialog.find('[data-test="namespace-warning"]');
      expect(warning.exists()).toBe(true);
      expect(warning.text()).toContain("Warning:");
      expect(warning.text()).toContain("You cannot delete your account while you have active namespaces");
      expect(warning.text()).toContain("Please delete all your owned namespaces");
    });

    it("renders warning alert with correct styling", () => {
      const warning = dialog.find('[data-test="namespace-warning"]');
      expect(warning.classes()).toContain("v-alert");
    });
  });

  describe("User deletion success", () => {
    beforeEach(async () => {
      wrapper.unmount();
      document.body.innerHTML = "";
      mountWrapper(true, false);
      await flushPromises();
    });

    it("successfully deletes user account and redirects to login", async () => {
      const pushSpy = vi.spyOn(router, "push");

      const deleteBtn = dialog.find('[data-test="delete-user-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      expect(authStore.deleteUser).toHaveBeenCalled();
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Account deleted successfully.");
      expect(pushSpy).toHaveBeenCalledWith({ name: "Login" });
    });
  });

  describe("User deletion error handling", () => {
    beforeEach(async () => {
      wrapper.unmount();
      document.body.innerHTML = "";
      mountWrapper(true, false);
      await flushPromises();
    });

    it("shows specific error when user has active namespaces (403)", async () => {
      const error = createAxiosError(403, "Forbidden");
      vi.mocked(authStore.deleteUser).mockRejectedValueOnce(error);

      const deleteBtn = dialog.find('[data-test="delete-user-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "You cannot delete your account while you have active namespaces.",
      );
    });

    it("does not redirect when deletion fails with 403", async () => {
      const error = createAxiosError(403, "Forbidden");
      vi.mocked(authStore.deleteUser).mockRejectedValueOnce(error);
      const pushSpy = vi.spyOn(router, "push");

      const deleteBtn = dialog.find('[data-test="delete-user-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      expect(pushSpy).not.toHaveBeenCalled();
    });

    it("shows generic error for other errors", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(authStore.deleteUser).mockRejectedValueOnce(error);

      const deleteBtn = dialog.find('[data-test="delete-user-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete account.");
    });

    it("handles network errors", async () => {
      const error = createAxiosError(0, "Network Error");
      vi.mocked(authStore.deleteUser).mockRejectedValueOnce(error);

      const deleteBtn = dialog.find('[data-test="delete-user-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete account.");
    });
  });

  describe("Dialog close behavior", () => {
    beforeEach(async () => {
      wrapper.unmount();
      document.body.innerHTML = "";
      mountWrapper(true, false);
      await flushPromises();
    });

    it("closes dialog when Cancel button is clicked", async () => {
      const cancelBtn = dialog.find('[data-test="close-btn"]');

      await cancelBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")![0]).toEqual([false]);
    });

    it("emits update:modelValue with false when dialog closes", async () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });

      messageDialog.vm.$emit("cancel");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")![0]).toEqual([false]);
    });
  });
});
