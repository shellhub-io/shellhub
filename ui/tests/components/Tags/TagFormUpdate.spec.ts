import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockTag } from "@tests/mocks/tag";
import TagFormUpdate from "@/components/Tags/TagFormUpdate.vue";
import useTagsStore from "@/store/modules/tags";
import handleError from "@/utils/handleError";

describe("TagFormUpdate", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagFormUpdate>>;
  let tagsStore: ReturnType<typeof useTagsStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openDialog = async () => {
    const button = wrapper.find('[data-test="open-tags-btn"]');
    await button.trigger("click");
    await flushPromises();
  };

  const mountWrapper = ({
    deviceUid = "device-123",
    tagsList = [{ name: "test-tag" }],
    hasAuthorization = true,
    tags = [mockTag],
  } = {}) => {
    wrapper = mountComponent(TagFormUpdate, {
      props: {
        deviceUid,
        tagsList,
        hasAuthorization,
      },
      piniaOptions: {
        initialState: {
          tags: {
            tags,
            tagCount: tags.length,
          },
        },
      },
      attachTo: document.body,
    });

    tagsStore = useTagsStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => {
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders list item button", () => {
      const button = wrapper.find('[data-test="open-tags-btn"]');
      expect(button.exists()).toBe(true);
    });

    it("shows 'Add Tags' when no tags exist", () => {
      wrapper.unmount();
      mountWrapper({ tagsList: [] });
      const title = wrapper.find('[data-test="has-tags-verification"]');
      expect(title.text()).toBe("Add Tags");
    });

    it("shows 'Edit tags' when tags exist", () => {
      const title = wrapper.find('[data-test="has-tags-verification"]');
      expect(title.text()).toBe("Edit tags");
    });

    it("disables button when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper({ hasAuthorization: false });
      const button = wrapper.find('[data-test="open-tags-btn"]');
      expect(button.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Dialog display", () => {
    it("opens dialog when button is clicked", async () => {
      mountWrapper();
      await openDialog();

      const dialogElement = dialog.find('[data-test="tags-form-dialog"]');
      expect(dialogElement.exists()).toBe(true);
    });

    it("loads tags when dialog opens", async () => {
      mountWrapper();
      await openDialog();

      expect(tagsStore.fetchTagList).toHaveBeenCalled();
    });

    it("shows autocomplete when dialog is open", async () => {
      mountWrapper();
      await openDialog();

      const autocomplete = dialog.find('[data-test="device-tags-autocomplete"]');
      expect(autocomplete.exists()).toBe(true);
    });
  });

  describe("Tag selection", () => {
    it("displays selected tags as chips", async () => {
      mountWrapper({ tagsList: [{ name: "tag1" }, { name: "tag2" }] });
      await openDialog();

      const selectedTags = dialog.findAll('[data-test="selected-tag-chip"]');
      expect(selectedTags.length).toBe(2);
    });

    it("limits selection to 3 tags", async () => {
      mountWrapper();
      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:modelValue", ["tag1", "tag2", "tag3", "tag4"]);
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Maximum of 3 tags allowed.");
    });

    it("allows removing tags", async () => {
      mountWrapper({ tagsList: [{ name: "tag1" }, { name: "tag2" }] });
      await openDialog();

      const removeBtn = dialog.find('[data-test="selected-tag-chip"] .v-chip__close');
      await removeBtn.trigger("click");
      await flushPromises();

      expect(wrapper.findAll('[data-test="selected-tag-chip"]').length).toBeLessThan(2);
    });
  });

  describe("Tag creation", () => {
    it("shows create button when valid new tag is entered", async () => {
      mountWrapper({ tags: [] });
      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:menu", true);
      await autocomplete.vm.$emit("update:search", "new-tag");
      await flushPromises();

      const createBtn = dialog.find('[data-test="create-new-tag-btn"]');
      expect(createBtn.exists()).toBe(true);
    });

    it("creates new tag and adds to selection", async () => {
      mountWrapper({ tags: [], tagsList: [] });
      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:menu", true);
      await autocomplete.vm.$emit("update:search", "new-tag");
      await flushPromises();

      const createBtn = dialog.find('[data-test="create-new-tag-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      expect(tagsStore.createTag).toHaveBeenCalledWith("new-tag");
    });

    it("shows error when tag creation fails", async () => {
      mountWrapper({ tags: [], tagsList: [] });
      vi.mocked(tagsStore.createTag).mockRejectedValueOnce(
        createAxiosError(500, "Error"),
      );

      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:menu", true);
      await autocomplete.vm.$emit("update:search", "new-tag");
      await flushPromises();

      const createBtn = dialog.find('[data-test="create-new-tag-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create tag.");
    });
  });

  describe("Saving tags", () => {
    it("calls addTagToDevice for newly added tags", async () => {
      mountWrapper({ tagsList: [] });
      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:modelValue", ["new-tag"]);
      await flushPromises();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(tagsStore.addTagToDevice).toHaveBeenCalledWith("device-123", "new-tag");
    });

    it("calls removeTagFromDevice for removed tags", async () => {
      mountWrapper({ tagsList: [{ name: "tag1" }, { name: "tag2" }] });
      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:modelValue", ["tag1"]);
      await flushPromises();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(tagsStore.removeTagFromDevice).toHaveBeenCalledWith("device-123", "tag2");
    });

    it("shows success message on save", async () => {
      mountWrapper({ tagsList: [] });
      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:modelValue", ["new-tag"]);
      await flushPromises();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Tags updated successfully.");
    });

    it("emits update event on successful save", async () => {
      mountWrapper({ tagsList: [] });
      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:modelValue", ["new-tag"]);
      await flushPromises();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog on successful save", async () => {
      mountWrapper({ tagsList: [] });
      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:modelValue", ["new-tag"]);
      await flushPromises();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
    });

    it("shows error message on save failure", async () => {
      mountWrapper({ tagsList: [] });
      vi.mocked(tagsStore.addTagToDevice).mockRejectedValueOnce(
        createAxiosError(500, "Error"),
      );

      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:modelValue", ["new-tag"]);
      await flushPromises();

      const saveBtn = dialog.find('[data-test="confirm-btn"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update tags.");
    });
  });

  describe("Error handling", () => {
    it("shows error when loading tags fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      mountWrapper();
      vi.mocked(tagsStore.fetchTagList).mockRejectedValueOnce(error);

      await openDialog();

      const autocomplete = wrapper.findComponent({ name: "VAutocomplete" });
      await autocomplete.vm.$emit("update:search", "new-tag");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load tags.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
