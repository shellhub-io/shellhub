import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import DeviceRename from "../../../src/components/Devices/DeviceRename.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const name = "39-5e-2a";
const uid = "a582b47a42d";

const tests = [
  {
    description: "Dialog closed",
    props: {
      name,
      uid,
    },
    data: {
      showDialog: false,
      invalid: false,
      editName: name,
      messages: "Examples: (foobar, foo-bar-ba-z-qux, foo-example, 127-0-0-1)",
    },
    computed: {
      device() {
        return {
          name,
          uid,
        };
      },
    },
    template: {
      "rename-icon": true,
      "rename-title": true,
      "deviceRename-card": false,
    },
    templateText: {
      "rename-title": "Rename",
    },
  },
  {
    description: "Dialog opened",
    props: {
      name,
      uid,
    },
    data: {
      showDialog: true,
      invalid: false,
      editName: name,
      messages: "Examples: (foobar, foo-bar-ba-z-qux, foo-example, 127-0-0-1)",
      editNameError: false,
      setEditNameError: vi.fn(),
    },
    computed: {
      device() {
        return {
          name,
          uid,
        };
      },
    },
    template: {
      "rename-icon": true,
      "rename-title": true,
      "deviceRename-card": true,
      "text-title": true,
      "hostname-field": true,
      "close-btn": true,
      "rename-btn": true,
    },
    templateText: {
      "rename-title": "Rename",
      "text-title": "Rename Device",
      "close-btn": "Close",
      "rename-btn": "Rename",
    },
  },
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "devices/rename": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

tests.forEach((test) => {
  describe(`${test.description}`, () => {
    let wrapper: VueWrapper<InstanceType<typeof DeviceRename>>;

    beforeEach(() => {
      const vuetify = createVuetify();
      const wrapper = mount(DeviceRename, {
        global: {
          plugins: [[store, key], routes, vuetify],
          stubs: ["router-link", "router-view"],
        },
        props: {
          name: test.props.name,
          uid: test.props.uid,
        },
        setup() {
          return {
            ...test.data,
            ...test.computed,
          };
        },
        shallow: true,
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

    it("Data is correct in props", () => {
      expect(wrapper.vm.name).toBe(test.props.name);
      expect(wrapper.vm.uid).toBe(test.props.uid);
    });

    it("Data is correct in data", () => {
      expect(wrapper.vm.showDialog).toBe(test.data.showDialog);
      expect(wrapper.vm.invalid).toBe(test.data.invalid);
      expect(wrapper.vm.editName).toBe(test.data.editName);
      expect(wrapper.vm.messages).toBe(test.data.messages);
    });

    it("Data is correct in computed", () => {
      expect(wrapper.vm.device).toStrictEqual(test.computed.device);
    });
  });
});
