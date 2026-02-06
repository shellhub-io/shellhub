import { flushPromises, VueWrapper } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import PrivateKeys from "@/views/PrivateKeys.vue";
import { mockPrivateKeys } from "../mocks";

describe("Private Keys View", () => {
  let wrapper: VueWrapper<InstanceType<typeof PrivateKeys>>;

  const mountWrapper = async (hasKeys = true) => {
    const initialState = { privateKeys: { privateKeys: hasKeys ? mockPrivateKeys : [] } };

    wrapper = mountComponent(PrivateKeys, {
      global: {
        stubs: {
          "v-file-upload": true,
          "v-file-upload-item": true,
        },
      },
      piniaOptions: { initialState },
    });

    await flushPromises();
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("when private keys exist", () => {
    beforeEach(() => mountWrapper());

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="private-keys-page-header"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Private Keys");
    });

    it("displays add private key button in header", () => {
      expect(wrapper.find('[data-test="add-private-key-btn"]').exists()).toBe(true);
    });

    it("displays the private keys list", () => {
      expect(wrapper.find('[data-test="private-keys-list"]').exists()).toBe(true);
    });

    it("does not show the no items message", () => {
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
    });

    it("opens add private key dialog when button is clicked", async () => {
      const addBtn = wrapper.find('[data-test="add-private-key-btn"]');
      await addBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "PrivateKeyAdd" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("modelValue")).toBe(true);
    });
  });

  describe("when no private keys exist", () => {
    beforeEach(async () => await mountWrapper(false));

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="private-keys-page-header"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Private Keys");
    });

    it("does not display add private key button in header", () => {
      expect(wrapper.find('[data-test="add-private-key-btn"]').exists()).toBe(false);
    });

    it("does not display the private keys list", () => {
      expect(wrapper.find('[data-test="private-keys-list"]').exists()).toBe(false);
    });

    it("shows the no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      expect(noItemsMessage.exists()).toBe(true);
      expect(noItemsMessage.text()).toContain("Private Keys");
    });

    it("displays add private key button in no items message", () => {
      expect(wrapper.find('[data-test="no-items-add-private-key-btn"]').exists()).toBe(true);
    });

    it("opens add private key dialog when no items button is clicked", async () => {
      const addBtn = wrapper.find('[data-test="no-items-add-private-key-btn"]');
      await addBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "PrivateKeyAdd" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("modelValue")).toBe(true);
    });
  });
});
