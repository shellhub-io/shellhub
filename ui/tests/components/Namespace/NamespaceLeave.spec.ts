import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import { mockNamespace } from "@tests/mocks";
import NamespaceLeave from "@/components/Namespace/NamespaceLeave.vue";
import useNamespacesStore from "@/store/modules/namespaces";
import { Router } from "vue-router";
import { createCleanRouter } from "@tests/utils/router";
import handleError from "@/utils/handleError";

describe("NamespaceLeave", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceLeave>>;
  let dialog: DOMWrapper<Element>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let router: Router;

  const mountWrapper = () => {
    localStorage.setItem("tenant", mockNamespace.tenant_id);
    router = createCleanRouter();
    vi.spyOn(router, "go").mockImplementation(() => {});
    wrapper = mountComponent(NamespaceLeave, {
      global: { plugins: [router] },
      props: { modelValue: true },
      attachTo: document.body,
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

  describe("Dialog display", () => {
    it("Renders MessageDialog component", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.exists()).toBe(true);
    });

    it("Shows correct title", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("Leave Namespace");
    });

    it("Shows warning icon", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("icon")).toBe("mdi-exit-to-app");
      expect(messageDialog.props("iconColor")).toBe("warning");
    });

    it("Shows Leave and Close buttons", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("confirmText")).toBe("Leave");
      expect(messageDialog.props("confirmColor")).toBe("error");
      expect(messageDialog.props("cancelText")).toBe("Close");
    });
  });

  describe("Leave namespace", () => {
    it("Calls leaveNamespace when confirmed", async () => {
      const confirmBtn = dialog.find('[data-test="leave-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.leaveNamespace).toHaveBeenCalledWith(mockNamespace.tenant_id);
    });

    it("Reloads page after successful leave", async () => {
      const routerGoSpy = vi.spyOn(router, "go").mockImplementation(() => {});

      const confirmBtn = dialog.find('[data-test="leave-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(routerGoSpy).toHaveBeenCalledWith(0);
    });
  });

  describe("Error handling", () => {
    it("Handles server error", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(namespacesStore.leaveNamespace).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="leave-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.leaveNamespace).toHaveBeenCalled();
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to leave the namespace.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
