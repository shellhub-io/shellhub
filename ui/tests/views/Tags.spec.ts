import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import Tags from "@/views/Tags.vue";
import { tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { envVariables } from "@/envVariables";
import TagList from "@/components/Tags/TagList.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import useTagsStore from "@/store/modules/tags";

const mockTags = [
  {
    tenant_id: "fake-tenant-data",
    name: "1",
    created_at: "2025-12-09T10:00:00Z",
    updated_at: "2025-12-09T10:00:00Z",
  },
  {
    tenant_id: "fake-tenant-data",
    name: "2",
    created_at: "2025-12-09T12:00:00Z",
    updated_at: "2025-12-09T12:00:00Z",
  },
];

describe("Tags View", async () => {
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  const tagsStore = useTagsStore();
  tagsStore.showTags = true;

  localStorage.setItem("tenant", "fake-tenant-data");
  envVariables.isCloud = true;

  mockTagsApi
    .onGet("http://localhost:3000/api/tags?filter=&page=1&per_page=10")
    .reply(200, mockTags, { "x-total-count": "2" });

  const wrapper = mount(Tags, { global: { plugins: [vuetify, SnackbarPlugin] } });
  await flushPromises();

  it("Renders the main heading", () => {
    expect(wrapper.find("h1").text()).toBe("Tags");
  });

  it("Renders the TagList component with 2 tag rows", () => {
    const tagList = wrapper.findComponent(TagList);
    expect(tagList.exists()).toBe(true);

    const tagRows = tagList.findAll('[data-test="tag-name"]');
    expect(tagRows).toHaveLength(2);
    expect(tagRows[0].text()).toBe("1");
    expect(tagRows[1].text()).toBe("2");
  });

  it("Renders the search field when tags exist", () => {
    const searchField = wrapper.find('[data-test="search-text"]');
    expect(searchField.exists()).toBe(true);
    expect(searchField.text()).toContain("Search by Tag Name"); // Input label
  });

  it("Renders the create tag button when tags exist", () => {
    const createButton = wrapper.find('[data-test="tag-create-button"]');
    expect(createButton.exists()).toBe(true);
    expect(createButton.text()).toBe("Create Tag");
  });

  it("Renders NoItemsMessage when no tags exist", async () => {
    tagsStore.showTags = false;
    mockTagsApi.reset();
    mockTagsApi
      .onGet("http://localhost:3000/api/tags?filter=&page=1&per_page=10")
      .reply(200, [], { "x-total-count": "0" });

    wrapper.unmount();
    const emptyWrapper = mount(Tags, { global: { plugins: [vuetify, SnackbarPlugin] } });
    await flushPromises();

    expect(emptyWrapper.findComponent(TagList).exists()).toBe(false);
    expect(emptyWrapper.find('[data-test="search-text"]').exists()).toBe(false);

    const noItemsMessage = emptyWrapper.findComponent(NoItemsMessage);
    expect(noItemsMessage.exists()).toBe(true);
    expect(noItemsMessage.props("item")).toBe("Tags");
    expect(noItemsMessage.props("icon")).toBe("mdi-tag-multiple");
  });
});
