import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useSessionsStore } from "@admin/store/modules/sessions";
import { useSnackbarStore } from "@admin/store/modules/snackbar";
import routes from "../../../../src/router";
import Sessions from "../../../../src/views/Sessions.vue";

type SessionsWrapper = VueWrapper<InstanceType<typeof Sessions>>;

describe("Sessions", () => {
  let wrapper: SessionsWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const sessionsStore = useSessionsStore();
    sessionsStore.fetch = vi.fn();

    const snackbarStore = useSnackbarStore();
    snackbarStore.showSnackbarErrorAction = vi.fn();

    const vuetify = createVuetify();

    wrapper = mount(Sessions, {
      global: {
        plugins: [pinia, vuetify, routes],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Should render all the components in the screen", () => {
    expect(wrapper.find("h1").text()).toContain("Sessions");
    expect(wrapper.find("[data-test='session-list']").exists()).toBe(true);
  });
});
