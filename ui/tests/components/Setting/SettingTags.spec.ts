import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import SettingTags from "@/components/Setting/SettingTags.vue";
import { tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { envVariables } from "@/envVariables";

type SettingTagsWrapper = VueWrapper<InstanceType<typeof SettingTags>>;

describe("Setting Tags", () => {
  let wrapper: SettingTagsWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockTags = new MockAdapter(tagsApi.getAxios());

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockTags
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, [{ name: "1" }, { name: "2" }]);

    wrapper = mount(SettingTags, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="tagList-component"]').exists());
  });
});
