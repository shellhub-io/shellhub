import { flushPromises, mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import TagList from "@/components/Tags/TagList.vue";
import { router } from "@/router";
import { tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const tags = [{ name: "123x" }, { name: "newtag" }];

describe("Tag List", async () => {
  const vuetify = createVuetify();
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  setActivePinia(createPinia());
  localStorage.setItem("tenant", "fake-tenant-data");

  mockTagsApi
    .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
    .reply(200, tags);

  const wrapper = mount(TagList, { global: { plugins: [vuetify, router, SnackbarPlugin] } });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component table", async () => {
    await flushPromises();
    expect(wrapper.find('[data-test="tag-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="tag-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="tag-list-actions"]').exists()).toBe(true);
  });
});
