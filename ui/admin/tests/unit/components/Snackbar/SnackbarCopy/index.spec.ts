import { createVuetify } from "vuetify";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useSnackbarStore from "@admin/store/modules/snackbar";

import SnackbarCopy from "../../../../../src/components/Snackbar/SnackbarCopy.vue";
import routes from "../../../../../src/router";

type SnackbarCopyWrapper = VueWrapper<InstanceType<typeof SnackbarCopy>>;

const mainContent = "Command";
const snackbarCopy = true;
const message = `${mainContent} copied to clipboard.`;

describe("SnackbarCopy", () => {
  let wrapper: SnackbarCopyWrapper;
  let snackbarStore: ReturnType<typeof useSnackbarStore>;

  beforeEach(() => {
    vi.useFakeTimers();
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    snackbarStore = useSnackbarStore();
    snackbarStore.snackbarCopy = snackbarCopy;
    vi.spyOn(snackbarStore, "unsetShowStatusSnackbarCopy");

    wrapper = shallowMount(SnackbarCopy, {
      global: {
        plugins: [vuetify, routes],
      },
      props: {
        mainContent,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Receives data in props", () => {
    expect(wrapper.vm.mainContent).toEqual(mainContent);
  });

  it("Processes data in computed properties", () => {
    expect(wrapper.vm.snackbar).toBe(true);
    expect(wrapper.vm.message).toEqual(message);
  });

  it("Calls unsetShowStatusSnackbarCopy after timeout", async () => {
    wrapper.vm.snackbar = false;

    vi.runAllTimers();

    expect(snackbarStore.unsetShowStatusSnackbarCopy).toHaveBeenCalled();
  });
});
