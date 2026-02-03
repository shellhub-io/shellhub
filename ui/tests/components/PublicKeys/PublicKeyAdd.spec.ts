import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import PublicKeyAdd from "@/components/PublicKeys/PublicKeyAdd.vue";
import usePublicKeysStore from "@/store/modules/public_keys";
import handleError from "@/utils/handleError";

vi.mock("@/utils/permission", () => ({ default: () => true }));

const VALID_PUBLIC_KEY = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGJw8VPZSH0w7mNRBcDJE4I0HfnWAl2qDtYpTr8g5F8N user@example.com";

describe("PublicKeyAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof PublicKeyAdd>>;
  let dialog: DOMWrapper<Element>;
  let publicKeysStore: ReturnType<typeof usePublicKeysStore>;

  const openDialog = async () => {
    const addBtn = wrapper.find('[data-test="public-key-add-btn"]');
    await addBtn.trigger("click");
    await flushPromises();
  };

  beforeEach(() => {
    wrapper = mountComponent(PublicKeyAdd, {
      props: { size: "default" },
      attachTo: document.body,
      global: { stubs: ["v-file-upload", "v-file-upload-item"] },
    });

    publicKeysStore = usePublicKeysStore();
    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Add button", () => {
    it("Renders add button", () => {
      const addBtn = wrapper.find('[data-test="public-key-add-btn"]');
      expect(addBtn.exists()).toBe(true);
    });

    it("Shows 'Add Public Key' text", () => {
      const addBtn = wrapper.find('[data-test="public-key-add-btn"]');
      expect(addBtn.text()).toBe("Add Public Key");
    });

    it("Opens dialog when clicked", async () => {
      await openDialog();

      const formDialog = dialog.find('[data-test="public-key-add-dialog"]');
      expect(formDialog.exists()).toBe(true);
    });
  });

  describe("Dialog display", () => {
    beforeEach(() => openDialog());

    it("Shows FormDialog with correct props", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("New Public Key");
      expect(formDialog.props("icon")).toBe("mdi-key-outline");
      expect(formDialog.props("confirmText")).toBe("Save");
      expect(formDialog.props("cancelText")).toBe("Cancel");
    });

    it("Renders name field", () => {
      const nameField = dialog.find('[data-test="name-field"]');
      expect(nameField.exists()).toBe(true);
    });

    it("Renders username restriction field", () => {
      const usernameField = dialog.find('[data-test="username-restriction-field"]');
      expect(usernameField.exists()).toBe(true);
    });

    it("Renders filter restriction field", () => {
      const filterField = dialog.find('[data-test="filter-restriction-field"]');
      expect(filterField.exists()).toBe(true);
    });

    it("Renders public key data field", () => {
      const dataField = wrapper.findComponent({ name: "FileTextComponent" });
      expect(dataField.exists()).toBe(true);
    });

    it("Does not show rule field by default", () => {
      const ruleField = dialog.find('[data-test="rule-field"]');
      expect(ruleField.exists()).toBe(false);
    });

    it("Does not show hostname field by default", () => {
      const hostnameField = dialog.find('[data-test="hostname-field"]');
      expect(hostnameField.exists()).toBe(false);
    });
  });

  describe("Form validation - Name field", () => {
    beforeEach(() => openDialog());

    it("Disables confirm button when name is empty", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("Enables confirm button when form is valid", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("test-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PUBLIC_KEY);
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(false);
    });
  });

  describe("Form validation - Public key data", () => {
    beforeEach(() => openDialog());

    it("Disables confirm button when key data is empty", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("test-key");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });
  });

  describe("Username restriction", () => {
    beforeEach(() => openDialog());

    it("Shows rule field when username restriction is selected", async () => {
      const usernameSelect = wrapper.findAllComponents({ name: "VSelect" })[0];
      usernameSelect.vm.$emit("update:modelValue", "username");
      await flushPromises();

      const ruleField = dialog.find('[data-test="rule-field"]');
      expect(ruleField.exists()).toBe(true);
    });
  });

  describe("Filter restriction", () => {
    beforeEach(() => openDialog());

    it("Shows hostname field when hostname filter is selected", async () => {
      const filterSelect = wrapper.findAllComponents({ name: "VSelect" })[1];
      filterSelect.vm.$emit("update:modelValue", "hostname");
      await flushPromises();

      const hostnameField = dialog.find('[data-test="hostname-field"]');
      expect(hostnameField.exists()).toBe(true);
    });

    it("Shows tag selector when tags filter is selected", async () => {
      const filterSelect = wrapper.findAllComponents({ name: "VSelect" })[1];
      filterSelect.vm.$emit("update:modelValue", "tags");
      await flushPromises();

      const tagSelector = wrapper.findComponent({ name: "TagAutocompleteSelect" });
      expect(tagSelector.exists()).toBe(true);
    });
  });

  describe("Create public key", () => {
    beforeEach(() => openDialog());

    it("Calls createPublicKey when form is valid", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PUBLIC_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-add-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(publicKeysStore.createPublicKey).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "my-key",
          username: ".*",
          filter: { hostname: ".*" },
        }),
      );
    });

    it("Creates key with hostname filter when specified", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PUBLIC_KEY);

      const filterSelect = wrapper.findAllComponents({ name: "VSelect" })[1];
      filterSelect.vm.$emit("update:modelValue", "hostname");
      await flushPromises();

      const hostnameField = dialog.find('[data-test="hostname-field"] input');
      await hostnameField.setValue("^server");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-add-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(publicKeysStore.createPublicKey).toHaveBeenCalledWith(
        expect.objectContaining({
          filter: { hostname: "^server" },
        }),
      );
    });

    it("Creates key with username restriction when specified", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PUBLIC_KEY);

      const usernameSelect = wrapper.findAllComponents({ name: "VSelect" })[0];
      usernameSelect.vm.$emit("update:modelValue", "username");
      await flushPromises();

      const ruleField = dialog.find('[data-test="rule-field"] input');
      await ruleField.setValue("admin");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-add-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(publicKeysStore.createPublicKey).toHaveBeenCalledWith(
        expect.objectContaining({
          username: "admin",
        }),
      );
    });

    it("Shows success message after creation", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PUBLIC_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-add-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Public key created successfully.");
    });

    it("Emits update event after creation", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PUBLIC_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-add-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("Closes dialog after creation", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PUBLIC_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-add-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Error handling", () => {
    beforeEach(() => openDialog());

    it("Handles duplicate public key error (409)", async () => {
      vi.mocked(publicKeysStore.createPublicKey).mockRejectedValueOnce(
        createAxiosError(409, "Duplicate key"),
      );

      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("existing-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PUBLIC_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-add-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const errorMessage = dialog.find('[data-test="ftc-file-error"]');
      expect(errorMessage.text()).toContain("Public Key data already exists");
    });

    it("Handles generic error", async () => {
      const error = new Error("Unknown error");
      vi.mocked(publicKeysStore.createPublicKey).mockRejectedValueOnce(error);

      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PUBLIC_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-add-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create the public key.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Dialog actions", () => {
    beforeEach(() => openDialog());

    it("Closes dialog when cancel is clicked", async () => {
      const cancelBtn = dialog.find('[data-test="pk-add-cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });

    it("Resets form when dialog is closed", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("test");
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="pk-add-cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      await openDialog();

      const nameFieldAfter = dialog.find('[data-test="name-field"] input');
      expect((nameFieldAfter.element as HTMLInputElement).value).toBe("");
    });
  });
});
