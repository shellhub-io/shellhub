import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import NamespaceAdd from "@/components/Namespace/NamespaceAdd.vue";
import useNamespacesStore from "@/store/modules/namespaces";
import { envVariables } from "@/envVariables";
import handleError from "@/utils/handleError";

vi.mock("@/envVariables", () => ({
  envVariables: {
    isCommunity: false,
  },
}));

describe("NamespaceAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceAdd>>;
  let dialog: DOMWrapper<Element>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const setNamespaceNameFieldValue = async (value = "new-namespace") => {
    const textField = wrapper.findComponent({ name: "VTextField" });
    await textField.setValue(value);
    await flushPromises();
    return textField;
  };

  const mountWrapper = (isCommunity = false) => {
    envVariables.isCommunity = isCommunity;

    wrapper = mountComponent(NamespaceAdd, {
      props: { modelValue: true },
      attachTo: document.body,
    });

    namespacesStore = useNamespacesStore();
    dialog = new DOMWrapper(document.body);
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Dialog display", () => {
    beforeEach(() => mountWrapper());

    it("Renders FormDialog component", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.exists()).toBe(true);
    });

    it("Shows correct title in non-community version", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("New Namespace");
    });

    it("Shows correct icon", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("icon")).toBe("mdi-folder-plus");
    });

    it("Shows Submit and Close buttons in non-community version", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmText")).toBe("Submit");
      expect(formDialog.props("cancelText")).toBe("Close");
    });
  });

  describe("Dialog display - Community version", () => {
    beforeEach(() => mountWrapper(true));

    it("Shows CLI title in community version", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("Add a namespace using the CLI");
    });

    it("Hides submit button in community version", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmText")).toBe("");
    });

    it("Disables confirm button in community version", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("Shows footer helper text in community version", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("footerHelperText")).toBe("Learn more on");
      expect(formDialog.props("footerHelperLinkText")).toBe("ShellHub Administration Guide");
      expect(formDialog.props("footerHelperLink")).toBe("https://docs.shellhub.io/self-hosted/administration");
    });
  });

  describe("Namespace input field", () => {
    beforeEach(() => mountWrapper());

    it("Renders namespace input field", () => {
      const textField = wrapper.findComponent({ name: "VTextField" });
      expect(textField.exists()).toBe(true);
      expect(textField.props("label")).toBe("Namespace");
    });

    it("Shows validation rules in content", () => {
      expect(dialog.text()).toContain("lowercase alphanumeric characters and hyphens");
      expect(dialog.text()).toContain("minimum of 3 characters");
      expect(dialog.text()).toContain("maximum of 63 characters");
      expect(dialog.text()).toContain("cannot be changed after creation");
    });
  });

  describe("Community version content", () => {
    beforeEach(() => mountWrapper(true));

    it("Shows CLI instructions in community version", () => {
      expect(dialog.text()).toContain("Community Edition");
      expect(dialog.text()).toContain("administration CLI");
      expect(dialog.text()).toContain("ShellHub Administration Guide");
    });

    it("Does not show input field in community version", () => {
      const textField = wrapper.findComponent({ name: "VTextField" });
      expect(textField.exists()).toBe(false);
    });
  });

  describe("Form validation", () => {
    beforeEach(() => mountWrapper());

    it("Validates required namespace", async () => {
      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const textField = wrapper.findComponent({ name: "VTextField" });
      expect(textField.props("errorMessages")).toBeTruthy();
    });

    it("Validates minimum length of 3 characters", async () => {
      const textField = await setNamespaceNameFieldValue("ab");

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(textField.props("errorMessages")).toContain("Namespace should be at least 3 characters");
    });

    it("Validates maximum length of 30 characters", async () => {
      const textField = await setNamespaceNameFieldValue("a".repeat(31));

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(textField.props("errorMessages")).toContain("Namespace should be at most 30 characters");
    });

    it("Validates format with lowercase alphanumeric and hyphens", async () => {
      const textField = await setNamespaceNameFieldValue("Invalid-Name");

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(textField.props("errorMessages")).toContain("Invalid format");
    });

    it("Accepts valid namespace format", async () => {
      const textField = await setNamespaceNameFieldValue("valid-namespace");

      expect(textField.props("errorMessages")).toStrictEqual([]);
    });
  });

  describe("Create namespace", () => {
    beforeEach(() => mountWrapper());
    it("Calls createNamespace when confirmed", async () => {
      vi.mocked(namespacesStore.createNamespace).mockResolvedValueOnce("new-tenant-id");

      await setNamespaceNameFieldValue();

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.createNamespace).toHaveBeenCalledWith("new-namespace");
    });

    it("Switches to new namespace after creation", async () => {
      vi.mocked(namespacesStore.createNamespace).mockResolvedValueOnce("new-tenant-id");

      const switchNamespaceSpy = vi.spyOn(namespacesStore, "switchNamespace");

      await setNamespaceNameFieldValue();

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(switchNamespaceSpy).toHaveBeenCalledWith("new-tenant-id");
    });

    it("Reloads page after successful creation", async () => {
      vi.mocked(namespacesStore.createNamespace).mockResolvedValueOnce("new-tenant-id");
      const reloadSpy = vi.spyOn(window.location, "reload");

      await setNamespaceNameFieldValue();

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(reloadSpy).toHaveBeenCalled();
    });

    it("Closes dialog after successful creation", async () => {
      vi.mocked(namespacesStore.createNamespace).mockResolvedValueOnce("new-tenant-id");

      await setNamespaceNameFieldValue();

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });
  });

  describe("Error handling", () => {
    beforeEach(() => mountWrapper());

    it("Handles 400 error with validation message", async () => {
      vi.mocked(namespacesStore.createNamespace).mockRejectedValueOnce(createAxiosError(400, "Bad Request"));

      const textField = await setNamespaceNameFieldValue("invalid");

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(textField.props("errorMessages")).toContain("Your namespace should be 3-30 characters long");
      expect(mockSnackbar.showError).not.toHaveBeenCalled();
      expect(handleError).not.toHaveBeenCalled();
    });

    it("Handles 403 error with plan upgrade message", async () => {
      vi.mocked(namespacesStore.createNamespace).mockRejectedValueOnce(createAxiosError(403, "Forbidden"));

      const textField = await setNamespaceNameFieldValue("extra-namespace");

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(textField.props("errorMessages")).toContain("Update your plan to create more namespaces");
      expect(mockSnackbar.showError).not.toHaveBeenCalled();
      expect(handleError).not.toHaveBeenCalled();
    });

    it("Handles 409 error with already exists message", async () => {
      vi.mocked(namespacesStore.createNamespace).mockRejectedValueOnce(createAxiosError(409, "Conflict"));

      const textField = await setNamespaceNameFieldValue("existing-namespace");

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(textField.props("errorMessages")).toContain("Namespace already exists");
      expect(mockSnackbar.showError).not.toHaveBeenCalled();
      expect(handleError).not.toHaveBeenCalled();
    });

    it("Handles 500 error", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      vi.mocked(namespacesStore.createNamespace).mockRejectedValueOnce(error);

      await setNamespaceNameFieldValue();

      const confirmBtn = dialog.find('[data-test="add-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.createNamespace).toHaveBeenCalled();
      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while creating the namespace.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
