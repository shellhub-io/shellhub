import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import ApiKeyForm from "@/components/Team/ApiKeys/ApiKeyForm.vue";
import { BasicRole } from "@/interfaces/INamespace";

describe("ApiKeyForm", () => {
  let wrapper: VueWrapper<InstanceType<typeof ApiKeyForm>>;

  const mountWrapper = ({
    mode = "create" as "create" | "edit",
    initialKeyName = "",
    initialRole = "administrator" as BasicRole,
    canManageRoles = true,
  } = {}) => {
    wrapper = mountComponent(ApiKeyForm, {
      props: {
        mode,
        initialKeyName,
        initialRole,
        canManageRoles,
      },
    });
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Create mode", () => {
    beforeEach(() => mountWrapper({ mode: "create" }));

    it("renders key name field", () => {
      const nameField = wrapper.find('[data-test="key-name-text"]');
      expect(nameField.exists()).toBe(true);
    });

    it("renders expiration date selector in create mode", () => {
      const expirationField = wrapper.find('[data-test="api-key-expiration-date"]');
      expect(expirationField.exists()).toBe(true);
    });

    it("renders role selector when canManageRoles is true", () => {
      const roleField = wrapper.find('[data-test="api-key-role"]');
      expect(roleField.exists()).toBe(true);
    });

    it("hides role selector when canManageRoles is false", () => {
      wrapper.unmount();
      mountWrapper({ mode: "create", canManageRoles: false });

      const roleField = wrapper.find('[data-test="api-key-role"]');
      expect(roleField.exists()).toBe(false);
    });

    it("shows create mode hint for name field", () => {
      const nameField = wrapper.find('[data-test="key-name-text"]');
      expect(nameField.html()).toContain("Provide a descriptive name for this key");
    });
  });

  describe("Edit mode", () => {
    beforeEach(() => mountWrapper({ mode: "edit", initialKeyName: "test-key" }));

    it("renders key name field", () => {
      const nameField = wrapper.find('[data-test="key-name-text"]');
      expect(nameField.exists()).toBe(true);
    });

    it("does not render expiration date selector in edit mode", () => {
      const expirationField = wrapper.find('[data-test="api-key-expiration-date"]');
      expect(expirationField.exists()).toBe(false);
    });

    it("renders role selector when canManageRoles is true", () => {
      const roleField = wrapper.find('[data-test="api-key-role"]');
      expect(roleField.exists()).toBe(true);
    });

    it("shows edit mode hint for name field", () => {
      const nameField = wrapper.find('[data-test="key-name-text"]');
      expect(nameField.html()).toContain("Please note that the new name must be unique");
    });

    it("populates initial key name", () => {
      const nameInput = wrapper.find('[data-test="key-name-text"] input').element as HTMLInputElement;
      expect(nameInput.value).toBe("test-key");
    });
  });

  describe("Form validation", () => {
    beforeEach(() => mountWrapper());

    it("shows error when key name is empty", async () => {
      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("test");
      await nameInput.setValue("");
      await nameInput.trigger("blur");
      await flushPromises();

      expect(wrapper.text()).toContain("Key name is required");
    });

    it("shows error when key name is too short", async () => {
      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("ab");
      await nameInput.trigger("blur");
      await flushPromises();

      expect(wrapper.text()).toContain("Key name must be at least 3 characters");
    });

    it("shows error when key name is too long", async () => {
      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("a".repeat(21));
      await nameInput.trigger("blur");
      await flushPromises();

      expect(wrapper.text()).toContain("Key name must be at most 20 characters");
    });

    it("shows error when key name contains spaces", async () => {
      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my key");
      await nameInput.trigger("blur");
      await flushPromises();

      expect(wrapper.text()).toContain("This field cannot contain any blank spaces");
    });

    it("emits update:valid with false when form is invalid", async () => {
      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("test");
      await nameInput.setValue("");
      await nameInput.trigger("blur");
      await flushPromises();

      const validEvents = wrapper.emitted("update:valid") as boolean[][];
      expect(validEvents[validEvents.length - 1]).toEqual([false]);
    });

    it("emits update:valid with true when form is valid", async () => {
      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("valid-key");
      await flushPromises();

      const validEvents = wrapper.emitted("update:valid") as boolean[][];
      expect(validEvents[validEvents.length - 1]).toEqual([true]);
    });
  });

  describe("Form submission", () => {
    it("emits submit event with correct data in create mode", async () => {
      mountWrapper({ mode: "create" });

      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      wrapper.vm.submitForm();
      await flushPromises();

      expect(wrapper.emitted("submit")).toBeTruthy();
      expect(wrapper.emitted("submit")?.[0]).toEqual([{
        name: "my-api-key",
        role: "administrator",
        expires_in: 30,
      }]);
    });

    it("emits submit event without expires_in in edit mode", async () => {
      mountWrapper({ mode: "edit", initialKeyName: "test-key" });

      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("updated-key");
      await flushPromises();

      wrapper.vm.submitForm();
      await flushPromises();

      expect(wrapper.emitted("submit")).toBeTruthy();
      const submitData = wrapper.emitted("submit")?.[0]?.[0] as Record<string, unknown>;
      expect(submitData.name).toBe("updated-key");
      expect(submitData.role).toBe("administrator");
      expect(submitData.expires_in).toBeUndefined();
    });

    it("includes selected role in submission", async () => {
      mountWrapper({ mode: "create", initialRole: "observer" });

      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      wrapper.vm.submitForm();
      await flushPromises();

      expect(wrapper.emitted("submit")?.[0]).toEqual([{
        name: "my-api-key",
        role: "observer",
        expires_in: 30,
      }]);
    });
  });

  describe("Expiration date selection", () => {
    beforeEach(() => mountWrapper({ mode: "create" }));

    it("defaults to 30 days expiration", () => {
      const expirationField = wrapper.find('[data-test="api-key-expiration-date"]');
      expect(expirationField.html()).toContain("30 days");
    });

    it("updates expiration hint when date changes", () => {
      const expirationSelect = wrapper.find('[data-test="api-key-expiration-date"]');

      // Check initial hint
      expect(expirationSelect.html()).toContain("Expires in");
    });

    it("includes selected expiration in submission", async () => {
      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      wrapper.vm.submitForm();
      await flushPromises();

      const submitData = wrapper.emitted("submit")?.[0]?.[0] as Record<string, unknown>;
      expect(submitData.expires_in).toBe(30);
    });
  });

  describe("Form reset", () => {
    beforeEach(() => mountWrapper({ mode: "create" }));

    it("clears form fields when reset is called", async () => {
      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      wrapper.vm.reset();
      await flushPromises();

      const inputElement = nameInput.element as HTMLInputElement;
      expect(inputElement.value).toBe("");
    });

    it("resets to initial values", async () => {
      wrapper.unmount();
      mountWrapper({ mode: "edit", initialKeyName: "initial-key", initialRole: "observer" });

      const nameInput = wrapper.find('[data-test="key-name-text"] input');
      await nameInput.setValue("changed-key");
      await flushPromises();

      wrapper.vm.reset();
      await flushPromises();

      const inputElement = nameInput.element as HTMLInputElement;
      expect(inputElement.value).toBe("initial-key");
    });
  });
});
