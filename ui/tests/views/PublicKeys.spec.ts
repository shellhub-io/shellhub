import { flushPromises, VueWrapper } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import createCleanRouter from "@tests/utils/router";
import PublicKeys from "@/views/PublicKeys.vue";
import * as publicKeysApi from "@/store/api/public_keys";
import { mockPublicKeys } from "@tests/views/mocks";

vi.mock("@/store/api/public_keys");

describe("Public Keys View", () => {
  let wrapper: VueWrapper<InstanceType<typeof PublicKeys>>;
  const router = createCleanRouter();

  const mountWrapper = async (hasKeys = true, mockError?: Error) => {
    const initialState = hasKeys
      ? {
        publicKeys: {
          publicKeys: mockPublicKeys,
          publicKeyCount: 1,
        },
      }
      : {
        publicKeys: {
          publicKeys: [],
          publicKeyCount: 0,
        },
      };

    if (mockError) {
      vi.spyOn(console, "error").mockImplementation(() => { });
      vi.mocked(publicKeysApi.fetchPublicKeys).mockRejectedValueOnce(mockError);
    }

    wrapper = mountComponent(PublicKeys, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState,
        stubActions: !mockError,
      },
    });

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when public keys exist", () => {
    beforeEach(async () => { await mountWrapper(); });

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="public-keys-title"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Public Keys");
    });

    it("displays add public key button in header", () => {
      const addBtn = wrapper.findComponent({ name: "PublicKeyAdd" });
      expect(addBtn.exists()).toBe(true);
    });

    it("displays the public keys list", () => {
      const publicKeysList = wrapper.findComponent({ name: "PublicKeysList" });
      expect(publicKeysList.exists()).toBe(true);
    });

    it("does not show the no items message", () => {
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
    });
  });

  describe("when no public keys exist", () => {
    beforeEach(async () => { await mountWrapper(false); });

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="public-keys-title"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Public Keys");
    });

    it("displays add public key button in header", () => {
      const addBtn = wrapper.findComponent({ name: "PublicKeyAdd" });
      expect(addBtn.exists()).toBe(true);
    });

    it("does not display the public keys list", () => {
      const publicKeysList = wrapper.findComponent({ name: "PublicKeysList" });
      expect(publicKeysList.exists()).toBe(false);
    });

    it("shows the no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      expect(noItemsMessage.exists()).toBe(true);
      expect(noItemsMessage.text()).toContain("Public Keys");
      expect(noItemsMessage.text()).toContain("SSH keys are more secure than passwords");
    });

    it("displays add public key button in no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      const addBtn = noItemsMessage.findComponent({ name: "PublicKeyAdd" });
      expect(addBtn.exists()).toBe(true);
    });
  });

  describe("when loading public keys fails", () => {
    it("displays error snackbar notification", async () => {
      await mountWrapper(false, new Error("Failed to fetch public keys"));
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load the public keys list.");
    });
  });
});
