import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SessionClose from "../../../src/components/Sessions/SessionClose.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const session = {
  uid: "8c354a00",
  device_uid: "a582b47a",
};

const tests = [
  {
    description: "Dialog closed",
    variables: {
      session,
    },
    props: {
      uid: session.uid,
      device: session,
      notHasAuthorization: false,
    },
    data: {
      session,
      showDialog: false,
    },
  },
  {
    description: "Dialog opened",
    variables: {
      session,
    },
    props: {
      uid: session.uid,
      device: session,
      notHasAuthorization: false,
    },
    data: {
      session,
      showDialog: true,
    },
  },
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "sessions/close": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("SessionClose", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(SessionClose, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            uid: test.props.uid,
            device: test.props.device,
            notHasAuthorization: test.props.notHasAuthorization,
          },
          shallow: true,
        });

        if (test.data.showDialog) wrapper.vm.showDialog = true;
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
      it('Receive data in props', () => {
        expect(wrapper.vm.uid).toEqual(test.props.uid);
        expect(wrapper.vm.device).toEqual(test.props.device);
        expect(wrapper.vm.notHasAuthorization).toEqual(test.props.notHasAuthorization);
      });
      it('Compare data with default value', () => {
        expect(wrapper.vm.showDialog).toEqual(test.data.showDialog);
      });

      // todo
    });
  });
});
