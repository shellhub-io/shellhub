import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import { mockNamespace } from "@tests/mocks";
import NamespaceDelete from "@/components/Namespace/NamespaceDelete.vue";
import useNamespacesStore from "@/store/modules/namespaces";
import useAuthStore from "@/store/modules/auth";
import { createCleanRouter } from "@tests/utils/router";
import { Router } from "vue-router";
import handleError from "@/utils/handleError";

vi.mock("@/utils/permission", () => ({ default: vi.fn(() => true) }));

vi.mock("@/envVariables", () => ({ envVariables: { isCloud: false } }));

describe("NamespaceDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceDelete>>;
  let dialog: DOMWrapper<Element>;
  let router: Router;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let authStore: ReturnType<typeof useAuthStore>;

  const mountWrapper = (isBillingActive = false) => {
    router = createCleanRouter();
    wrapper = mountComponent(NamespaceDelete, {
      global: { plugins: [router] },
      props: {
        modelValue: true,
        tenant: mockNamespace.tenant_id,
      },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          namespaces: { currentNamespace: mockNamespace },
          billing: {
            billing: { active: isBillingActive },
            auth: { token: "test-token" },
          },
        },
      },
    });

    namespacesStore = useNamespacesStore();
    authStore = useAuthStore();
    dialog = new DOMWrapper(document.body);
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Dialog display", () => {
    it("Renders MessageDialog component", () => {
      mountWrapper();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.exists()).toBe(true);
    });

    it("Shows correct title", () => {
      mountWrapper();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("Namespace Deletion");
    });

    it("Shows error icon", () => {
      mountWrapper();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("icon")).toBe("mdi-delete-alert");
      expect(messageDialog.props("iconColor")).toBe("error");
    });

    it("Shows Remove and Close buttons when billing is inactive", () => {
      mountWrapper();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("confirmText")).toBe("Remove");
      expect(messageDialog.props("confirmColor")).toBe("error");
      expect(messageDialog.props("cancelText")).toBe("Close");
    });

    it("Disables confirm button when billing is active", () => {
      mountWrapper(true);

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("confirmDisabled")).toBe(true);
      expect(messageDialog.props("confirmText")).toBe("");
    });
  });

  describe("Content - without active billing", () => {
    beforeEach(() => mountWrapper());

    it("Shows warning about permanent deletion", () => {
      const content = dialog.find('[data-test="content-text"]');
      expect(content.exists()).toBe(true);
      expect(content.text()).toContain("This action cannot be undone");
      expect(content.text()).toContain("permanently delete");
    });

    it("Displays namespace name", () => {
      const content = dialog.find('[data-test="content-text"]');
      expect(content.text()).toContain(mockNamespace.name);
    });
  });

  describe("Content - with active billing", () => {
    beforeEach(() => mountWrapper(true));

    it("Shows billing restriction message when billing is active", () => {
      const content = dialog.find('[data-test="content-subscription-text"]');
      expect(content.exists()).toBe(true);
      expect(content.text()).toContain("active subscription or an unpaid invoice");
    });

    it("Shows message about settling invoices", () => {
      const content = dialog.find('[data-test="content-subscription-text"]');
      expect(content.text()).toContain("all outstanding invoices are settled");
    });

    it("Does not show regular deletion message when billing is active", () => {
      const content = dialog.find('[data-test="content-text"]');
      expect(content.exists()).toBe(false);
    });
  });

  describe("Delete namespace", () => {
    beforeEach(() => {
      mountWrapper();
      vi.spyOn(router, "push").mockResolvedValue();
    });

    it("Calls deleteNamespace when confirmed", async () => {
      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.deleteNamespace).toHaveBeenCalledWith(mockNamespace.tenant_id);
    });

    it("Logs out user after successful deletion", async () => {
      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(authStore.logout).toHaveBeenCalled();
    });

    it("Redirects to login page after deletion", async () => {
      const routerPushSpy = vi.spyOn(router, "push").mockResolvedValue();

      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith({ name: "Login" });
    });

    it("Closes dialog after successful deletion", async () => {
      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("Closes dialog when Cancel is clicked", async () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Error handling", () => {
    beforeEach(() => mountWrapper());

    it("Handles 402 payment required error and emits billing-in-debt", async () => {
      const error = createAxiosError(402, "Payment required");
      vi.mocked(namespacesStore.deleteNamespace).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("billing-in-debt")).toBeTruthy();
      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while deleting the namespace.");
      expect(handleError).toHaveBeenCalledWith(error);
    });

    it("Handles generic server error", async () => {
      const error = createAxiosError(500, "Internal server error");
      vi.mocked(namespacesStore.deleteNamespace).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="remove-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.deleteNamespace).toHaveBeenCalledWith(mockNamespace.tenant_id);
      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while deleting the namespace.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
