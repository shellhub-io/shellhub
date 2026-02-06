import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { mockTag } from "@tests/mocks/tag";
import * as hasPermissionModule from "@/utils/permission";
import TagList from "@/components/Tags/TagList.vue";
import useTagsStore from "@/store/modules/tags";

describe("TagList", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagList>>;
  let tagsStore: ReturnType<typeof useTagsStore>;

  const mountWrapper = ({
    tags = [mockTag],
    canEditTag = true,
    canRemoveTag = true,
  } = {}) => {
    vi.spyOn(hasPermissionModule, "default")
      .mockImplementation((permission: string) => {
        if (permission === "tag:edit") return canEditTag;
        if (permission === "tag:remove") return canRemoveTag;
        return false;
      });

    wrapper = mountComponent(TagList, {
      piniaOptions: {
        initialState: {
          tags: { tags, tagCount: tags.length },
        },
      },
    });

    tagsStore = useTagsStore();
  };

  afterEach(() => {
    wrapper?.unmount();
  });

  describe("Component rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders data table", () => {
      const table = wrapper.find('[data-test="tag-list"]');
      expect(table.exists()).toBe(true);
    });

    it("displays tag name", () => {
      const tagName = wrapper.find('[data-test="tag-name"]');
      expect(tagName.exists()).toBe(true);
      expect(tagName.text()).toBe(mockTag.name);
    });

    it("displays tag created date", () => {
      const createdAt = wrapper.find('[data-test="tag-created-at"]');
      expect(createdAt.exists()).toBe(true);
    });

    it("displays actions button for each tag", () => {
      const actionsBtn = wrapper.find('[data-test="tag-list-actions"]');
      expect(actionsBtn.exists()).toBe(true);
    });
  });

  describe("Permissions", () => {
    it("enables edit option when user has edit permission", () => {
      mountWrapper({ canEditTag: true });
      const editComponent = wrapper.findComponent({ name: "TagEdit" });
      expect(editComponent.find('[data-test="open-tag-edit"]').classes()).not.toContain("v-list-item--disabled");
    });

    it("disables edit option when user lacks edit permission", () => {
      mountWrapper({ canEditTag: false });
      const editComponent = wrapper.findComponent({ name: "TagEdit" });
      expect(editComponent.find('[data-test="open-tag-edit"]').classes()).toContain("v-list-item--disabled");
    });

    it("enables remove option when user has remove permission", () => {
      mountWrapper({ canRemoveTag: true });
      const removeComponent = wrapper.findComponent({ name: "TagRemove" });
      expect(removeComponent.find('[data-test="open-tag-remove"]').classes()).not.toContain("v-list-item--disabled");
    });

    it("disables remove option when user lacks remove permission", () => {
      mountWrapper({ canRemoveTag: false });
      const removeComponent = wrapper.findComponent({ name: "TagRemove" });
      expect(removeComponent.find('[data-test="open-tag-remove"]').classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Data loading", () => {
    it("calls fetchTagList on mount", () => {
      mountWrapper();
      expect(tagsStore.fetchTagList).toHaveBeenCalled();
    });

    it("displays multiple tags when available", () => {
      const multipleTags = [
        { ...mockTag, name: "tag1" },
        { ...mockTag, name: "tag2" },
        { ...mockTag, name: "tag3" },
      ];
      mountWrapper({ tags: multipleTags });

      const tagNames = wrapper.findAll('[data-test="tag-name"]');
      expect(tagNames).toHaveLength(3);
    });

    it("shows empty state when no tags exist", () => {
      mountWrapper({ tags: [] });
      expect(wrapper.text()).toContain("No data available");
    });
  });

  describe("Pagination", () => {
    it("refetches tags when page changes", async () => {
      mountWrapper({ tags: Array(15).fill(mockTag) });

      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(tagsStore.fetchTagList).toHaveBeenCalled();
    });

    it("refetches tags when items per page changes", async () => {
      mountWrapper({ tags: Array(20).fill(mockTag) });

      const ippCombo = wrapper.find('[data-test="ipp-combo"] input');
      await ippCombo.setValue(20);
      await flushPromises();

      expect(tagsStore.fetchTagList).toHaveBeenCalled();
    });
  });
});
