import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import DeviceDelete from "../../../src/components/Devices/DeviceDelete.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const tests = [
  {
    description: "Dialog closed",
    props: {
      uid: "a582b47a42d",
      redirect: false,
    },
    data: {
      showDialog: false,
    },
    template: {
      "remove-icon": true,
      "remove-title": true,
      "deviceDelete-card": false,
    },
    templateText: {
      "remove-title": "Remove",
    },
  },
  {
    description: "Dialog opened without redirect",
    props: {
      uid: "a582b47a42d",
      redirect: false,
    },
    template: {
      "remove-icon": true,
      "remove-title": true,
      "text-title": true,
      "text-text": true,
      "close-btn": true,
      "remove-btn": true,
    },
    data: {
      showDialog: true,
    },
    templateText: {
      "remove-title": "Remove",
      "text-title": "Are you sure?",
      "text-text": "You are about to remove this device.",
      "close-btn": "Close",
      "remove-btn": "Remove",
    },
  },
  {
    description: "Dialog opened with redirect",
    props: {
      uid: "a582b47a42d",
      redirect: true,
    },
    template: {
      "remove-icon": true,
      "remove-title": true,
      "text-title": true,
      "text-text": true,
      "close-btn": true,
      "remove-btn": true,
    },
    data: {
      showDialog: true,
    },
    templateText: {
      "remove-title": "Remove",
      "text-title": "Are you sure?",
      "text-text": "You are about to remove this device.",
      "close-btn": "Close",
      "remove-btn": "Remove",
    },
  },
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "devices/remove": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

tests.forEach((test) => {
  describe(`${test.description}`, () => {
    let wrapper: VueWrapper<any>;

    beforeEach(() => {
      const vuetify = createVuetify();

      wrapper = mount(DeviceDelete, {
        global: {
          plugins: [routes, vuetify, [store, key]],
          stubs: ["router-link"],
        },
        props: {
          uid: test.props.uid,
          redirect: test.props.redirect,
        },
        shallow: true,
        setup() {
          return {
            showDialog: test.data.showDialog,
          };
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

    it("Receive data in props", () => {
      expect(wrapper.vm.uid).toBe(test.props.uid);
      expect(wrapper.vm.redirect).toBe(test.props.redirect);
    });

    //////
    // HTML validation
    //////

    // it("Renders the template with data", () => {
    //   expect(wrapper.find('[data-test="remove-icon"').exists()).toBe(
    //     test.template["remove-icon"],
    //   );
    //   expect(wrapper.find('[data-test="remove-title"').exists()).toBe(
    //     test.template["remove-title"],
    //   );
    // });
  });
});
