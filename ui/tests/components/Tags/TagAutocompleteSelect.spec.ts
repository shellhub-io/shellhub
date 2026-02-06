import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockTag } from "@tests/mocks/tag";
import TagAutocompleteSelect from "@/components/Tags/TagAutocompleteSelect.vue";
import useTagsStore from "@/store/modules/tags";
import handleError from "@/utils/handleError";

describe("TagAutocompleteSelect", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagAutocompleteSelect>>;
  let tagsStore: ReturnType<typeof useTagsStore>;

  const mountWrapper = ({
    selectedTags = [] as string[],
    tags = [mockTag],
    tagCount = tags.length,
  } = {}) => {
    wrapper = mountComponent(TagAutocompleteSelect, {
      attachTo: document.body,
      props: {
        selectedTags,
        tagSelectorErrorMessage: "",
      },
      piniaOptions: {
        initialState: {
          tags: { tags, tagCount },
        },
      },
    });

    tagsStore = useTagsStore();
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders autocomplete selector", () => {
      const selector = wrapper.find('[data-test="tags-selector"]');
      expect(selector.exists()).toBe(true);
    });

    it("displays correct label", () => {
      const selectorLabel = wrapper.find('[data-test="tags-selector"] label');
      expect(selectorLabel.text()).toBe("Tags");
    });
  });

  describe("Initial data loading", () => {
    it("calls fetchTagList on mount with initial parameters", () => {
      mountWrapper();

      expect(tagsStore.fetchTagList).toHaveBeenCalledWith({
        filter: "",
        perPage: 10,
      });
    });

    it("validates tags on mount with empty selection", async () => {
      mountWrapper();
      await flushPromises();

      const tagsSelector = wrapper.find('[data-test="tags-selector"]');
      expect(tagsSelector.find(".v-messages__message").text()).toBe("You must choose at least one tag");
    });
  });

  describe("Tag selection and validation", () => {
    it("emits update event when tags are selected", async () => {
      mountWrapper();

      const selector = wrapper.findComponent({ name: "VAutocomplete" });
      await selector.vm.$emit("update:modelValue", [mockTag.name]);
      await flushPromises();

      expect(wrapper.emitted("update:selectedTags")).toBeTruthy();
    });

    it("shows error when more than 3 tags are selected", async () => {
      const fourTags = ["tag1", "tag2", "tag3", "tag4"];
      mountWrapper({ selectedTags: fourTags });
      await flushPromises();

      const tagsSelector = wrapper.find('[data-test="tags-selector"]');
      expect(tagsSelector.find(".v-messages__message").text()).toBe("You can select up to three tags only");
    });

    it("clears error when 1 tag is selected", async () => {
      mountWrapper({ selectedTags: [mockTag.name] });
      await flushPromises();

      const tagsSelector = wrapper.find('[data-test="tags-selector"]');
      expect(tagsSelector.find(".v-messages__message").exists()).toBe(false);
    });
  });

  describe("Tag display", () => {
    it("displays selected tags", () => {
      mountWrapper({ selectedTags: [mockTag.name] });

      const selector = wrapper.findComponent({ name: "VAutocomplete" });
      expect(selector.props("modelValue")).toContain(mockTag.name);
    });

    it("allows multiple tag selection", () => {
      mountWrapper({
        selectedTags: ["tag1", "tag2"],
        tags: [
          { ...mockTag, name: "tag1" },
          { ...mockTag, name: "tag2" },
        ],
      });

      const selector = wrapper.findComponent({ name: "VAutocomplete" });
      expect(selector.props("multiple")).toBe(true);
    });

    it("shows tags from store in items list", () => {
      const multipleTags = [
        { ...mockTag, name: "tag1" },
        { ...mockTag, name: "tag2" },
        { ...mockTag, name: "tag3" },
      ];
      mountWrapper({ tags: multipleTags });

      const selector = wrapper.findComponent({ name: "VAutocomplete" });
      const items = selector.props("items");
      expect(items).toHaveLength(3);
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when fetchTagList fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      mountWrapper();
      vi.mocked(tagsStore.fetchTagList).mockRejectedValueOnce(error);
      await flushPromises();

      const selector = wrapper.findComponent({ name: "VAutocomplete" });
      await selector.vm.$emit("update:search", "test");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load tags.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Pagination and infinite scroll", () => {
    it("loads initial page with 10 items", () => {
      mountWrapper();

      expect(tagsStore.fetchTagList).toHaveBeenCalledWith({
        filter: "",
        perPage: 10,
      });
    });

    it("renders sentinel element for intersection observer", async () => {
      mountWrapper();

      const selector = wrapper.findComponent({ name: "VAutocomplete" });
      await selector.vm.$emit("update:menu", true);
      await flushPromises();

      const sentinel = new DOMWrapper(document.body).find('[data-test="tags-sentinel"]');
      expect(sentinel.exists()).toBe(true);
    });

    it("resets pagination when searching", async () => {
      mountWrapper();

      const selector = wrapper.findComponent({ name: "VAutocomplete" });

      await selector.vm.$emit("update:search", "first");
      await flushPromises();

      await selector.vm.$emit("update:search", "second");
      await flushPromises();

      expect(tagsStore.fetchTagList).toHaveBeenCalledWith({
        filter: expect.any(String),
        perPage: 10,
      });
    });

    it("encodes filter correctly when searching", async () => {
      mountWrapper();

      const selector = wrapper.findComponent({ name: "VAutocomplete" });
      await selector.find("input").setValue("test-tag");
      await selector.vm.$emit("update:search", "test-tag");
      await flushPromises();

      const calls = vi.mocked(tagsStore.fetchTagList).mock.calls;
      const lastCall = calls[calls.length - 1][0];

      expect(lastCall?.filter).toBeTruthy();
      const decoded = JSON.parse(Buffer.from(lastCall?.filter as string, "base64").toString());
      expect(decoded[0].params.value).toBe("test-tag");
    });

    it("handles empty search filter", async () => {
      mountWrapper();

      const selector = wrapper.findComponent({ name: "VAutocomplete" });
      await selector.vm.$emit("update:search", "");
      await flushPromises();

      const calls = vi.mocked(tagsStore.fetchTagList).mock.calls;
      const lastCall = calls[calls.length - 1][0];

      expect(lastCall?.filter).toBe("");
    });

    it("opens menu when user interacts with autocomplete", async () => {
      mountWrapper();

      const selector = wrapper.findComponent({ name: "VAutocomplete" });
      await selector.vm.$emit("update:menu", true);
      await flushPromises();

      expect(selector.emitted("update:menu")).toBeTruthy();
      expect(selector.emitted("update:menu")?.[0]).toEqual([true]);
    });
  });
});
