// admin/tests/unit/components/Namespaces/NamespaceDelete/index.spec.ts
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Mock } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceDelete from "@admin/components/Namespace/NamespaceDelete.vue";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import handleError from "@/utils/handleError";

// --- Mocks ---

const mockRouter = {
  push: vi.fn(),
};

const mockRoute = {
  name: "namespaceDetails" as string,
};

vi.mock("vue-router", async () => {
  const actual = await vi.importActual<typeof import("vue-router")>("vue-router");
  return {
    ...actual,
    useRouter: () => mockRouter,
    useRoute: () => mockRoute,
  };
});

vi.mock("@/utils/handleError", () => ({
  __esModule: true,
  default: vi.fn(),
}));

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

// --- Test setup ---

describe("Namespace Delete", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceDelete>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let updateModelValue: ReturnType<typeof vi.fn>;
  let pinia: ReturnType<typeof createPinia>;

  const tenantId = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx";
  const namespaceName = "namespace"; // < 10 chars → no truncation

  const mountComponent = () => {
    updateModelValue = vi.fn();

    wrapper = mount(NamespaceDelete, {
      global: {
        plugins: [createVuetify(), pinia],
        provide: {
          [SnackbarInjectionKey]: mockSnackbar,
        },
        stubs: {
          // Stub MessageDialog to avoid dealing with v-dialog + teleport
          MessageDialog: {
            template: `
              <div data-test="message-dialog">
                <slot />
                <button data-test="remove-btn" @click="$emit('confirm')" />
                <button data-test="close-btn" @click="$emit('cancel')" />
              </div>
            `,
            props: [
              "modelValue",
              "title",
              "icon",
              "iconColor",
              "confirmColor",
              "confirmText",
              "confirmLoading",
              "cancelText",
              "confirmDataTest",
              "cancelDataTest",
            ],
          },
        },
      },
      props: {
        tenant: tenantId,
        name: namespaceName,
        modelValue: true,
        "onUpdate:modelValue": updateModelValue,
        onUpdate: vi.fn(),
      },
    });
  };

  beforeEach(() => {
    // Single shared Pinia instance
    pinia = createPinia();
    setActivePinia(pinia);

    namespacesStore = useNamespacesStore();
    namespacesStore.deleteNamespace = vi.fn().mockResolvedValue(undefined);

    mockRouter.push.mockReset();
    mockRoute.name = "namespaceDetails";

    mockSnackbar.showSuccess.mockReset();
    mockSnackbar.showError.mockReset();
    (handleError as Mock).mockReset();

    mountComponent();
  });

  it("Renders the dialog content with the namespace name", () => {
    const content = wrapper.get('[data-test="content-text"]');

    expect(content.text()).toContain("This action cannot be undone.");
    expect(content.text()).toContain(namespaceName);
  });

  it("Deletes namespace, shows success, redirects and closes when on namespaceDetails route", async () => {
    const removeButton = wrapper.get('[data-test="remove-btn"]');
    await removeButton.trigger("click");
    await flushPromises();

    expect(namespacesStore.deleteNamespace).toHaveBeenCalledTimes(1);
    expect(namespacesStore.deleteNamespace).toHaveBeenCalledWith(tenantId);

    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespace deleted successfully.");
    expect(mockSnackbar.showError).not.toHaveBeenCalled();

    expect(mockRouter.push).toHaveBeenCalledWith({ name: "namespaces" });

    expect(updateModelValue).toHaveBeenCalledWith(false);

    expect(wrapper.vm.isLoading).toBe(false);

    expect(wrapper.emitted("update")).toBeUndefined();
  });

  it("Deletes namespace, shows success and emits update when not on namespaceDetails route", async () => {
    mockRoute.name = "namespaces";

    mountComponent();

    const removeButton = wrapper.get('[data-test="remove-btn"]');
    await removeButton.trigger("click");
    await flushPromises();

    expect(namespacesStore.deleteNamespace).toHaveBeenCalledTimes(1);
    expect(namespacesStore.deleteNamespace).toHaveBeenCalledWith(tenantId);

    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespace deleted successfully.");
    expect(mockSnackbar.showError).not.toHaveBeenCalled();

    expect(mockRouter.push).not.toHaveBeenCalled();

    const updateEvents = wrapper.emitted("update");
    expect(updateEvents).toBeTruthy();
    expect(updateEvents?.length).toBe(1);

    expect(updateModelValue).toHaveBeenCalledWith(false);

    expect(wrapper.vm.isLoading).toBe(false);
  });

  it("Shows error snackbar, calls handleError and keeps dialog open when deletion fails", async () => {
    (namespacesStore.deleteNamespace as Mock).mockRejectedValueOnce(
      new Error("delete failed"),
    );

    const removeButton = wrapper.get('[data-test="remove-btn"]');
    await removeButton.trigger("click");
    await flushPromises();

    expect(namespacesStore.deleteNamespace).toHaveBeenCalledTimes(1);

    expect(mockSnackbar.showError).toHaveBeenCalledWith(
      "An error occurred while deleting the namespace.",
    );
    expect(mockSnackbar.showSuccess).not.toHaveBeenCalled();

    expect(handleError).toHaveBeenCalledTimes(1);

    expect(mockRouter.push).not.toHaveBeenCalled();

    expect(wrapper.emitted("update")).toBeUndefined();

    expect(updateModelValue).not.toHaveBeenCalledWith(false);

    expect(wrapper.vm.isLoading).toBe(false);
  });
});
