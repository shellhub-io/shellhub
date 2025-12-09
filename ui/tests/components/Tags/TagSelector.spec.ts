import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, afterEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import TagSelector from "@/components/Tags/TagSelector.vue";
import { router } from "@/router";
import { tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const tags = [
  { name: "tag1" },
  { name: "tag2" },
  { name: "tag3" },
  { name: "tag4" },
  { name: "tag5" },
  { name: "tag6" },
  { name: "tag7" },
  { name: "tag8" },
  { name: "tag9" },
  { name: "tag10" },
  { name: "tag11" },
];

describe("Tag Selector", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagSelector>>;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  localStorage.setItem("tenant", "fake-tenant-data");
  mockTagsApi
    .onGet("http://localhost:3000/api/tags?filter=&page=1&per_page=10")
    .reply(200, tags);

  beforeEach(() => {
    wrapper = mount(TagSelector, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: {
        variant: "device",
      },
    });
  });

  afterEach(async () => {
    await flushPromises();
    wrapper.unmount();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", () => {
    expect(wrapper.findComponent('[data-test="tags-btn"]').exists()).toBe(true);
  });

  it("Successfully loads tags", async () => {
    await wrapper.findComponent('[data-test="tags-btn"]').trigger("click");
    await flushPromises();
    expect(wrapper.vm.fetchedTags).toEqual(tags);
  });
});
