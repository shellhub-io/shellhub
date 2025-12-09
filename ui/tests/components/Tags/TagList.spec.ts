import { flushPromises, mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import TagList from "@/components/Tags/TagList.vue";
import { router } from "@/router";
import { tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const tags = [
  {
    tenant_id: "fake-tenant-data",
    name: "123x",
    created_at: "2025-01-01T10:00:00Z",
    updated_at: "2025-01-01T10:00:00Z",
  },
  {
    tenant_id: "fake-tenant-data",
    name: "newtag",
    created_at: "2025-01-02T12:00:00Z",
    updated_at: "2025-01-02T12:00:00Z",
  },
];

describe("Tag List", () => {
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
