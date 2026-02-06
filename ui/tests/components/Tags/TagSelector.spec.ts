import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { mockTag } from "@tests/mocks/tag";
import TagSelector from "@/components/Tags/TagSelector.vue";
import useTagsStore from "@/store/modules/tags";

describe("TagSelector", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagSelector>>;
  let tagsStore: ReturnType<typeof useTagsStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openMenu = async () => {
    const button = wrapper.find('[data-test="tags-btn"]');
    await button.trigger("click");
    await flushPromises();
  };

  const mountWrapper = ({
    variant = "device" as "device" | "container",
    tags = [mockTag],
    selectedTags = [] as typeof mockTag[],
  } = {}) => {
    wrapper = mountComponent(TagSelector, {
      props: { variant },
      piniaOptions: {
        initialState: {
          tags: {
            tags,
            tagCount: tags.length,
            selectedTags,
          },
        },
      },
      attachTo: document.body,
    });

    tagsStore = useTagsStore();
    dialog = new DOMWrapper(document.body);
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders tags button", () => {
      const button = wrapper.find('[data-test="tags-btn"]');
      expect(button.exists()).toBe(true);
    });

    it("shows tag count badge when tags are selected", () => {
      wrapper.unmount();
      mountWrapper({ selectedTags: [mockTag] });
      const badge = wrapper.find(".v-badge__badge");
      expect(badge.exists()).toBe(true);
      expect(badge.text()).toBe("1");
    });

    it("renders tags menu when opened", async () => {
      await openMenu();

      const menu = dialog.find(".v-list");
      expect(menu.exists()).toBe(true);
    });
  });

  describe("Initial data loading", () => {
    it("calls fetchTagList on mount", () => {
      mountWrapper();
      expect(tagsStore.fetchTagList).toHaveBeenCalled();
    });

    it("displays available tags in menu", async () => {
      const multipleTags = [
        { ...mockTag, name: "tag1" },
        { ...mockTag, name: "tag2" },
      ];
      mountWrapper({ tags: multipleTags });

      await openMenu();

      const tagItems = dialog.findAll('[data-test="tag-item"]');
      expect(tagItems.length).toBe(2);
    });
  });

  describe("Tag selection", () => {
    it("calls toggleSelectedTag when clicking on tag", async () => {
      mountWrapper();
      await openMenu();

      const tagItem = dialog.find('[data-test="tag-item"]');
      await tagItem.trigger("click");
      await flushPromises();

      expect(tagsStore.toggleSelectedTag).toHaveBeenCalledWith(mockTag);
    });
  });

  describe("Tag display", () => {
    it("shows checked checkbox for selected tags", async () => {
      mountWrapper();
      await openMenu();

      tagsStore.selectedTags = [mockTag];
      await flushPromises();

      const checkbox = dialog.find('[data-test="tag-checkbox"] input[type="checkbox"]').element as HTMLInputElement;
      expect(checkbox.checked).toBe(true);
    });

    it("displays tag count in badge", () => {
      mountWrapper({ selectedTags: [mockTag, { ...mockTag, name: "tag2" }] });

      const badge = wrapper.find(".v-badge__badge");
      expect(badge.text()).toBe("2");
    });
  });

  describe("Manage tags button", () => {
    it("shows manage tags button in menu", async () => {
      mountWrapper();
      await openMenu();

      const manageBtn = dialog.find('[data-test="manage-tags-btn"]');
      expect(manageBtn.exists()).toBe(true);
      expect(manageBtn.text()).toContain("Manage Tags");
    });
  });
});
