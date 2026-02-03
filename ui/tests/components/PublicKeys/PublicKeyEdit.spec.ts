import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import PublicKeyEdit from "@/components/PublicKeys/PublicKeyEdit.vue";
import { mockPublicKey } from "@tests/mocks/publicKey";
import usePublicKeysStore from "@/store/modules/public_keys";
import handleError from "@/utils/handleError";
import { createAxiosError } from "@tests/utils/axiosError";

describe("PublicKeyEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof PublicKeyEdit>>;
  let dialog: DOMWrapper<Element>;
  let publicKeysStore: ReturnType<typeof usePublicKeysStore>;

  const mountWrapper = (props = {}) => {
    wrapper = mountComponent(PublicKeyEdit, {
      global: { stubs: ["v-file-upload", "v-file-upload-item"] },
      props: {
        publicKey: mockPublicKey,
        hasAuthorization: true,
        ...props,
      },
      attachTo: document.body,
    });
  };

  const openDialog = async () => {
    const editBtn = wrapper.find('[data-test="public-key-edit-title-btn"]');
    await editBtn.trigger("click");
    await flushPromises();
  };

  beforeEach(() => {
    mountWrapper();
    publicKeysStore = usePublicKeysStore();
    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Edit button", () => {
    it("Renders edit button", () => {
      const editBtn = wrapper.find('[data-test="public-key-edit-title-btn"]');
      expect(editBtn.exists()).toBe(true);
    });

    it("Shows edit icon", () => {
      const icon = wrapper.find('[data-test="public-key-edit-icon"]');
      expect(icon.exists()).toBe(true);
      const vIcon = icon.findComponent({ name: "VIcon" });
      expect(vIcon.props("icon")).toBe("mdi-pencil");
    });

    it("Shows 'Edit' text", () => {
      const editBtn = wrapper.find('[data-test="public-key-edit-title-btn"]');
      expect(editBtn.text()).toBe("Edit");
    });

    it("Is disabled when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper({ hasAuthorization: false });

      const editBtn = wrapper.find('[data-test="public-key-edit-title-btn"]');
      expect(editBtn.classes()).toContain("v-list-item--disabled");
    });

    it("Opens dialog when clicked", async () => {
      await openDialog();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(true);
    });
  });

  describe("Dialog display", () => {
    beforeEach(() => openDialog());

    it("Shows FormDialog with correct props", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("Edit Public Key");
      expect(formDialog.props("icon")).toBe("mdi-key-outline");
      expect(formDialog.props("confirmText")).toBe("Save");
      expect(formDialog.props("cancelText")).toBe("Cancel");
    });

    it("Initializes name field with existing value", () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      expect((nameField.element as HTMLInputElement).value).toBe(mockPublicKey.name);
    });

    it("Shows public key data field as disabled", () => {
      const dataField = wrapper.findComponent({ name: "FileTextComponent" });
      expect(dataField.props("disabled")).toBe(true);
    });

    it("Renders username restriction field", () => {
      const usernameField = dialog.find('[data-test="username-restriction-field"]');
      expect(usernameField.exists()).toBe(true);
    });

    it("Renders filter restriction field", () => {
      const filterField = dialog.find('[data-test="filter-restriction-field"]');
      expect(filterField.exists()).toBe(true);
    });
  });

  describe("Form initialization", () => {
    beforeEach(() => openDialog());

    it("Initializes with all username when username is .*", () => {
      const usernameSelects = wrapper.findAllComponents({ name: "VSelect" });
      expect(usernameSelects[0].props("modelValue")).toBe("all");
    });

    it("Initializes with all filter when hostname is .*", () => {
      const filterSelects = wrapper.findAllComponents({ name: "VSelect" });
      expect(filterSelects[1].props("modelValue")).toBe("all");
    });

    it("Shows hostname field when publicKey has hostname filter", async () => {
      wrapper.unmount();
      mountWrapper({
        publicKey: {
          ...mockPublicKey,
          filter: { hostname: "^server" },
        },
      });

      await openDialog();

      dialog = new DOMWrapper(document.body);
      const hostnameField = dialog.find('[data-test="hostname-field"]');
      expect(hostnameField.exists()).toBe(true);
    });

    it("Shows rule field when publicKey has username restriction", async () => {
      wrapper.unmount();
      mountWrapper({
        publicKey: {
          ...mockPublicKey,
          username: "admin",
        },
      });

      await openDialog();

      dialog = new DOMWrapper(document.body);
      const ruleField = dialog.find('[data-test="rule-field"]');
      expect(ruleField.exists()).toBe(true);
    });
  });

  describe("Form validation", () => {
    beforeEach(() => openDialog());

    it("Disables confirm button when name is empty", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });
  });

  describe("Edit public key", () => {
    beforeEach(() => openDialog());

    it("Calls updatePublicKey with updated data", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("updated-key");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-edit-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(publicKeysStore.updatePublicKey).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "updated-key",
          fingerprint: mockPublicKey.fingerprint,
        }),
      );
    });

    it("Updates hostname filter when changed", async () => {
      const filterSelect = wrapper.findAllComponents({ name: "VSelect" })[1];
      filterSelect.vm.$emit("update:modelValue", "hostname");
      await flushPromises();

      const hostnameField = dialog.find('[data-test="hostname-field"] input');
      await hostnameField.setValue("^server");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-edit-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(publicKeysStore.updatePublicKey).toHaveBeenCalledWith(
        expect.objectContaining({
          filter: { hostname: "^server" },
        }),
      );
    });

    it("Updates username restriction when changed", async () => {
      const usernameSelect = wrapper.findAllComponents({ name: "VSelect" })[0];
      usernameSelect.vm.$emit("update:modelValue", "username");
      await flushPromises();

      const ruleField = dialog.find('[data-test="rule-field"] input');
      await ruleField.setValue("admin");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="pk-edit-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(publicKeysStore.updatePublicKey).toHaveBeenCalledWith(
        expect.objectContaining({
          username: "admin",
        }),
      );
    });

    it("Shows success message after edit", async () => {
      const confirmBtn = dialog.find('[data-test="pk-edit-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Public key updated successfully.");
    });

    it("Emits update event after edit", async () => {
      const confirmBtn = dialog.find('[data-test="pk-edit-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("Closes dialog after edit", async () => {
      const confirmBtn = dialog.find('[data-test="pk-edit-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Error handling", () => {
    beforeEach(() => openDialog());

    it("Handles generic error", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(publicKeysStore.updatePublicKey).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="pk-edit-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update public key.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Dialog actions", () => {
    beforeEach(() => openDialog());

    it("Closes dialog when cancel is clicked", async () => {
      const cancelBtn = dialog.find('[data-test="pk-edit-cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });

    it("Resets form to original values when dialog is reopened", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("changed-name");
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="pk-edit-cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      await openDialog();

      dialog = new DOMWrapper(document.body);
      const nameFieldAfter = dialog.find('[data-test="name-field"] input');
      expect((nameFieldAfter.element as HTMLInputElement).value).toBe(mockPublicKey.name);
    });
  });
});
