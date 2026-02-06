import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import TagRemove from "@/components/Tags/TagRemove.vue";
import useTagsStore from "@/store/modules/tags";
import handleError from "@/utils/handleError";

describe("TagRemove", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagRemove>>;
  let tagsStore: ReturnType<typeof useTagsStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openDialog = async () => {
    const listItem = wrapper.find('[data-test="open-tag-remove"]');
    await listItem.trigger("click");
    await flushPromises();
  };

  const triggerConfirmButton = async () => {
    const confirmBtn = dialog.find('[data-test="confirm-btn"]');
    await confirmBtn.trigger("click");
    await flushPromises();
  };

  const mountWrapper = ({ tagName = "tag-test", hasAuthorization = true } = {}) => {
    wrapper = mountComponent(TagRemove, {
      props: { tagName, hasAuthorization },
      attachTo: document.body,
    });
    tagsStore = useTagsStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Rendering", () => {
    it("renders remove list item", () => {
      const listItem = wrapper.find('[data-test="open-tag-remove"]');
      expect(listItem.exists()).toBe(true);
    });

    it("disables list item when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper({ hasAuthorization: false });
      const listItem = wrapper.find('[data-test="open-tag-remove"]');
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Tag removal", () => {
    beforeEach(() => openDialog());

    it("opens dialog when clicking remove item", () => {
      const messageDialog = dialog.find('[data-test="delete-tag-dialog"]');
      expect(messageDialog.exists()).toBe(true);
    });

    it("calls deleteTag when confirming removal", async () => {
      await triggerConfirmButton();

      expect(tagsStore.deleteTag).toHaveBeenCalledWith("tag-test");
    });

    it("shows success snackbar on successful removal", async () => {
      await triggerConfirmButton();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("tag-test was removed successfully.");
    });

    it("emits update event on successful removal", async () => {
      await triggerConfirmButton();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog when clicking cancel", async () => {
      const cancelBtn = dialog.find('[data-test="close-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      expect(tagsStore.deleteTag).not.toHaveBeenCalled();
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when removal fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      mountWrapper();
      vi.mocked(tagsStore.deleteTag).mockRejectedValueOnce(error);

      await openDialog();

      await triggerConfirmButton();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove tag.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
