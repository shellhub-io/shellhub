import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceDelete from "@admin/components/Namespace/NamespaceDelete.vue";
import { Router } from "vue-router";

describe("NamespaceDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceDelete>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let router: Router;
  const mockTenantId = "tenant-123";
  const mockNamespaceName = "test-namespace";

  const mountWrapper = async (routeName: string = "namespaceDetails") => {
    const params = routeName === "namespaceDetails" ? { id: mockTenantId } : {};
    router = createCleanAdminRouter();
    await router.push({ name: routeName, params });
    await router.isReady();

    wrapper = mountComponent(NamespaceDelete, {
      global: { plugins: [router] },
      props: {
        tenant: mockTenantId,
        name: mockNamespaceName,
        modelValue: true,
      },
      attachTo: document.body,
    });

    namespacesStore = useNamespacesStore();
  };

  const getDialog = () => new DOMWrapper(document.body).find('[role="dialog"]');

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("shows the dialog when modelValue is true", async () => {
      await flushPromises();
      const dialog = getDialog();

      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Namespace Deletion");
    });

    it("displays the namespace name in the content", async () => {
      await flushPromises();
      const dialog = getDialog();
      const content = dialog.find('[data-test="content-text"]');

      expect(content.text()).toContain("This action cannot be undone");
      expect(content.text()).toContain(mockNamespaceName);
    });

    it("shows remove and close buttons", async () => {
      await flushPromises();
      const dialog = getDialog();

      expect(dialog.find('[data-test="remove-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });
  });

  describe("deleting namespace from details page", () => {
    beforeEach(() => mountWrapper("namespaceDetails"));

    it("calls store action, shows success, and redirects when confirm is clicked", async () => {
      await flushPromises();
      const dialog = getDialog();
      const pushSpy = vi.spyOn(router, "push");

      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.deleteNamespace).toHaveBeenCalledWith(mockTenantId);
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespace deleted successfully.");
      expect(pushSpy).toHaveBeenCalledWith({ name: "namespaces" });
    });

    it("shows error message when delete fails", async () => {
      await flushPromises();
      vi.mocked(namespacesStore.deleteNamespace).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const dialog = getDialog();
      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while deleting the namespace.");
      expect(wrapper.emitted("update:modelValue")?.[0]).toBeUndefined();
    });
  });

  describe("deleting namespace from list page", () => {
    beforeEach(() => mountWrapper("namespaces"));

    it("calls store action, shows success, and emits update when confirm is clicked", async () => {
      await flushPromises();
      const dialog = getDialog();
      const pushSpy = vi.spyOn(router, "push");

      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.deleteNamespace).toHaveBeenCalledWith(mockTenantId);
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespace deleted successfully.");
      expect(pushSpy).not.toHaveBeenCalled();
      expect(wrapper.emitted("update")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    });

    it("does not emit update when delete fails", async () => {
      await flushPromises();
      vi.mocked(namespacesStore.deleteNamespace).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const dialog = getDialog();
      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while deleting the namespace.");
      expect(wrapper.emitted("update")).toBeUndefined();
    });
  });

  describe("closing dialog", () => {
    beforeEach(() => mountWrapper());

    it("closes dialog when cancel button is clicked", async () => {
      await flushPromises();
      const dialog = getDialog();

      const cancelBtn = dialog.find('[data-test="close-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });
  });
});
