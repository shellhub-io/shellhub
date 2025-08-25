import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import SettingTags from "@/components/Setting/SettingTags.vue";
import { tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SettingTagsWrapper = VueWrapper<InstanceType<typeof SettingTags>>;

describe("Setting Tags", () => {
  let wrapper: SettingTagsWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());

  beforeEach(async () => {
    mockTagsApi.onGet("http://localhost:3000/api/tags").reply(200, ["1", "2"]);

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
