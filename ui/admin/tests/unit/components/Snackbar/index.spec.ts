import { createVuetify } from "vuetify";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useSnackbarStore from "@admin/store/modules/snackbar";
import Snackbar from "../../../../src/components/Snackbar/Snackbar.vue";
import router from "../../../../src/router";

type SnackbarWrapper = VueWrapper<InstanceType<typeof Snackbar>>;

describe("Snackbar.vue", () => {
  let wrapper: SnackbarWrapper;
  const vuetify = createVuetify();

  beforeEach(() => {
    setActivePinia(createPinia());

    const snackbarStore = useSnackbarStore();
    snackbarStore.snackbarMessageAndContentType = {
      typeMessage: "",
      typeContent: "",
    };

    wrapper = shallowMount(Snackbar, {
      global: {
        plugins: [vuetify, router],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Processes data from the computed property", () => {
    const store = useSnackbarStore();
    expect(wrapper.vm.message).toEqual(store.getSnackbarMessageAndContentType);
  });
});
