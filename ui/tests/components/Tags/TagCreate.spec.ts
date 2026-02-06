import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import TagCreate from "@/components/Tags/TagCreate.vue";
import useTagsStore from "@/store/modules/tags";
import handleError from "@/utils/handleError";

describe("TagCreate", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagCreate>>;
  let tagsStore: ReturnType<typeof useTagsStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = () => {
    wrapper = mountComponent(TagCreate, { attachTo: document.body, props: { modelValue: true } });
    tagsStore = useTagsStore();
    dialog = new DOMWrapper(document.body).find('[data-test="tag-create-dialog"]');
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Dialog display", () => {
    it("renders dialog when modelValue is true", () => {
      expect(dialog.exists()).toBe(true);
    });

    it("shows dialog title", () => {
      const title = dialog.find('[data-test="window-dialog-titlebar"]');
      expect(title.text()).toBe("Create Tag");
    });
  });

  describe("Form interaction", () => {
    it("has a name input field", () => {
      const nameInput = dialog.find('[data-test="tag-field"]');
      expect(nameInput.exists()).toBe(true);
    });

    it("has a create button", () => {
      const createBtn = dialog.find('[data-test="create-btn"]');
      expect(createBtn.exists()).toBe(true);
    });

    it("has a cancel button", () => {
      const cancelBtn = dialog.find('[data-test="close-btn"]');
      expect(cancelBtn.exists()).toBe(true);
    });

    it("disables create button when name is empty", () => {
      const createBtn = dialog.find('[data-test="create-btn"]');
      expect(createBtn.attributes("disabled")).toBeDefined();
    });

    it("enables create button when name is valid", async () => {
      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("test-tag");
      await flushPromises();

      const createBtn = dialog.find('[data-test="create-btn"]');
      expect(createBtn.attributes("disabled")).toBeUndefined();
    });
  });

  describe("Tag creation", () => {
    it("calls createTag when submitting valid form", async () => {
      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("new-tag");
      await flushPromises();

      const createBtn = dialog.find('[data-test="create-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      expect(tagsStore.createTag).toHaveBeenCalledWith("new-tag");
    });

    it("shows success snackbar on successful creation", async () => {
      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("new-tag");
      await flushPromises();

      const createBtn = dialog.find('[data-test="create-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Successfully created tag");
    });

    it("emits update event on successful creation", async () => {
      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("new-tag");
      await flushPromises();

      const createBtn = dialog.find('[data-test="create-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog on successful creation", async () => {
      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("new-tag");
      await flushPromises();

      const createBtn = dialog.find('[data-test="create-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when creation fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(tagsStore.createTag).mockRejectedValueOnce(error);

      const nameInput = dialog.find('[data-test="tag-field"] input');
      await nameInput.setValue("new-tag");
      await flushPromises();

      const createBtn = dialog.find('[data-test="create-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create tag.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
