import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import Tags from "@/views/Tags.vue";
import { mockTags } from "@tests/views/mocks";

describe("Tags View", () => {
  let wrapper: VueWrapper<InstanceType<typeof Tags>>;

  const mountWrapper = (hasTags = true) => {
    const initialState = {
      tags: {
        showTags: hasTags,
        tags: hasTags ? mockTags : [],
      },
    };

    wrapper = mountComponent(Tags, { piniaOptions: { initialState } });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("when tags exist", () => {
    beforeEach(() => mountWrapper());

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="tags-settings-card"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Tags");
      expect(pageHeader.text()).toContain("Organization");
    });

    it("displays the search field", () => {
      const searchField = wrapper.find('[data-test="search-text"]');
      expect(searchField.exists()).toBe(true);
    });

    it("displays the create tag button in header", () => {
      const createButton = wrapper.find('[data-test="tag-create-button"]');
      expect(createButton.exists()).toBe(true);
      expect(createButton.text()).toContain("Create Tag");
    });

    it("displays the tags list", () => {
      const tagList = wrapper.findComponent({ name: "TagList" });
      expect(tagList.exists()).toBe(true);
    });

    it("does not show the no items message", () => {
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
    });

    it("opens create tag dialog when button is clicked", async () => {
      const createButton = wrapper.find('[data-test="tag-create-button"]');
      await createButton.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "TagCreate" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("allows searching for tags", async () => {
      const searchField = wrapper.find('[data-test="search-text"]').find("input");
      await searchField.setValue("test-tag");
      await flushPromises();

      expect(searchField.element.value).toBe("test-tag");
    });
  });

  describe("when no tags exist", () => {
    beforeEach(() => mountWrapper(false));

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="tags-settings-card"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Tags");
    });

    it("does not display the search field", () => {
      expect(wrapper.find('[data-test="search-text"]').exists()).toBe(false);
    });

    it("does not display the create tag button in header", () => {
      expect(wrapper.find('[data-test="tag-create-button"]').exists()).toBe(false);
    });

    it("does not display the tags list", () => {
      const tagList = wrapper.findComponent({ name: "TagList" });
      expect(tagList.exists()).toBe(false);
    });

    it("shows the no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      expect(noItemsMessage.exists()).toBe(true);
      expect(noItemsMessage.text()).toContain("Tags");
      expect(noItemsMessage.text()).toContain("organize your resources using Tags");
    });

    it("displays create tag button in no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      const createButton = noItemsMessage.find("button");
      expect(createButton.exists()).toBe(true);
      expect(createButton.text()).toContain("Create Tag");
    });

    it("opens create tag dialog when no items button is clicked", async () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      const createButton = noItemsMessage.find("button");
      await createButton.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "TagCreate" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("modelValue")).toBe(true);
    });
  });
});
