import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import SessionDelete from "../../../src/components/Sessions/SessionDelete.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const tests = [
  {
    description: "Dialog closed",
    props: {
      uid: "8c354a00",
      notHasAuthorization: false,
    },
    data: {
      showDialog: false,
    },
    template: {
      "removeRecord-icon": true,
      "removeRecord-title": true,
      "sessionDeleteRecord-card": false,
    },
    templateText: {
      "removeRecord-title": "Delete Session Record",
    },
  },
  {
    description: "Dialog opened",
    props: {
      uid: "8c354a00",
      notHasAuthorization: false,
    },
    data: {
      showDialog: true,
    },
    template: {
      "removeRecord-icon": true,
      "removeRecord-title": true,
      "sessionDeleteRecord-card": true,
      "text-title": true,
      "text-text": true,
      "cancel-btn": true,
      "delete-btn": true,
    },
    templateText: {
      "removeRecord-title": "Delete Session Record",
      "text-title": "Are you sure?",
      "text-text":
        "You are going to delete the logs recorded for this session.",
      "cancel-btn": "Cancel",
      "delete-btn": "Delete",
    },
  },
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "sessions/deleteSessionLogs": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("SessionDeleteRecord", () => {
  let wrapper: VueWrapper<InstanceType<typeof SessionDelete>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(SessionDelete, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            uid: test.props.uid,
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
      it("Receive data in props", () => {
        expect(wrapper.vm.uid).toEqual(test.props.uid);
        expect(wrapper.vm.notHasAuthorization).toEqual(
          test.props.notHasAuthorization,
        );
      });
      it("Compare data with default value", () => {
        expect(wrapper.vm.showDialog).toEqual(test.data.showDialog);
      });

    // todo html checking
    });
  });
});
