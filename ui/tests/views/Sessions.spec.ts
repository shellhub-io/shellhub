import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import Sessions from "../../src/views/Sessions.vue";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("Sessions", () => {
  let wrapper: VueWrapper<InstanceType<typeof Sessions>>;
  const vuetify = createVuetify();

  const numberSessionsEqualZero = 0;
  const numberSessionsGreaterThanZero = 1;

  const storeWithoutSessions = createStore({
    state: {
      numberSessions: numberSessionsEqualZero,
    },
    getters: {
      "sessions/getNumberSessions": (state) => state.numberSessions,
    },
    actions: {
      "sessions/refresh": vi.fn(),
      "box/setStatus": vi.fn(),
      "sessions/resetPagePerpage": vi.fn(),
      "snackbar/showSnackbarErrorLoading": vi.fn(),
      "sessions/fetch": vi.fn(),
    },
  });

  const storeWithSessions = createStore({
    state: {
      numberSessions: numberSessionsGreaterThanZero,
    },
    getters: {
      "sessions/getNumberSessions": (state) => state.numberSessions,
    },
    actions: {
      "sessions/refresh": vi.fn(),
      "box/setStatus": vi.fn(),
      "sessions/resetPagePerpage": vi.fn(),
      "snackbar/showSnackbarErrorLoading": vi.fn(),
      "sessions/fetch": vi.fn(),
    },
  });

  ///////
  // In this case, the rendering of the component that shows the
  // message when it does not have access to the device is tested.
  ///////

  describe("Without sessions", () => {
    beforeEach(async () => {
      wrapper = mount(Sessions, {
        global: {
          plugins: [[storeWithoutSessions, key], vuetify, routes],
        },
      });
    });

    ///////
    // Component Rendering
    //////

    it("Is a Vue instance", () => {
      expect(wrapper).toBeTruthy();
    });
    it("Renders the component", () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data checking
    //////
    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });
    it("Compare data with the default and defined value", () => {
      expect(wrapper.vm.show).toEqual(true);
    });
    it("Process data in the computed", () => {
      expect(wrapper.vm.hasSession).toEqual(false);
      expect(wrapper.vm.showBoxMessage).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it("Renders the template with components", () => {
      expect(
        wrapper.find('[data-test="BoxMessageSession-component"]').exists(),
      ).toBe(true);
    });
  });

  ///////
  // In this case, it is tested when it has already accessed a
  // device.
  ///////

  describe("With sessions", () => {
    beforeEach(async () => {
      wrapper = mount(Sessions, {
        global: {
          plugins: [[storeWithSessions, key], vuetify, routes],
        },
      });
    });

    ///////
    // Component Rendering
    //////

    it("Is a Vue instance", () => {
      expect(wrapper).toBeTruthy();
    });
    it("Renders the component", () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data checking
    //////

    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });
    it("Compare data with the default and defined value", () => {
      expect(wrapper.vm.show).toEqual(true);
    });
    it("Process data in the computed", () => {
      expect(wrapper.vm.hasSession).toEqual(true);
      expect(wrapper.vm.showBoxMessage).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it("Renders the template with components", () => {
      expect(wrapper.find('[data-test="BoxMessageSession-component"]').exists()).toBe(false);
    });
  });
});
