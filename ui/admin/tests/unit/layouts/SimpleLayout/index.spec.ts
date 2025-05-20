import { beforeEach, describe, expect, it } from "vitest";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { createPinia, setActivePinia } from "pinia";
import { SnackbarPlugin } from "@/plugins/snackbar";
import SimpleLayout from "../../../../src/layouts/SimpleLayout.vue";
import routes from "../../../../src/router";

type SimpleLayoutWrapper = VueWrapper<InstanceType<typeof SimpleLayout>>;

describe("SimpleLayout", () => {
  let wrapper: SimpleLayoutWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());

    const vuetify = createVuetify();

    wrapper = shallowMount(SimpleLayout, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
    }) as SimpleLayoutWrapper;
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
