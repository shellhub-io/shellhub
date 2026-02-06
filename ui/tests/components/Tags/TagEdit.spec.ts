import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import TagEdit from "@/components/Tags/TagEdit.vue";
import useTagsStore from "@/store/modules/tags";
import handleError from "@/utils/handleError";

describe("TagEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagEdit>>;
  let tagsStore: ReturnType<typeof useTagsStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openDialog = async () => {
    const listItem = wrapper.find('[data-test="open-tag-edit"]');
    await listItem.trigger("click");
    await flushPromises();
  };

  const mountWrapper = ({ tagName = "tag-test", hasAuthorization = true } = {}) => {
    wrapper = mountComponent(TagEdit, {
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
    it("renders edit list item", () => {
      const listItem = wrapper.find('[data-test="open-tag-edit"]');
      expect(listItem.exists()).toBe(true);
    });

    it("disables list item when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper({ hasAuthorization: false });
      const listItem = wrapper.find('[data-test="open-tag-edit"]');
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Tag update", () => {
    it("calls updateTag when submitting valid form", async () => {
      await openDialog();

      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("updated-tag");
      await flushPromises();

      const updateBtn = dialog.find('[data-test="edit-btn"]');
      await updateBtn.trigger("click");
      await flushPromises();

      expect(tagsStore.updateTag).toHaveBeenCalledWith("tag-test", { name: "updated-tag" });
    });

    it("shows success snackbar on successful update", async () => {
      await openDialog();

      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("updated-tag");
      await flushPromises();

      const updateBtn = dialog.find('[data-test="edit-btn"]');
      await updateBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Tag updated successfully.");
    });

    it("emits update event on successful update", async () => {
      await openDialog();

      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("updated-tag");
      await flushPromises();

      const updateBtn = dialog.find('[data-test="edit-btn"]');
      await updateBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when update fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      mountWrapper();
      vi.mocked(tagsStore.updateTag).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("updated-tag");
      await flushPromises();

      const updateBtn = dialog.find('[data-test="edit-btn"]');
      await updateBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update tag.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
