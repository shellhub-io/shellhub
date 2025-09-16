import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useSessionsStore from "@admin/store/modules/sessions";
import routes from "@admin/router";
import Sessions from "@admin/views/Sessions.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SessionsWrapper = VueWrapper<InstanceType<typeof Sessions>>;

describe("Sessions", () => {
  let wrapper: SessionsWrapper;
  const pinia = createPinia();
  setActivePinia(pinia);
  const sessionsStore = useSessionsStore();
  sessionsStore.fetchSessionList = vi.fn();

  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(Sessions, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
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
