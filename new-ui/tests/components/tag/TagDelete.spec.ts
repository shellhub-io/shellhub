import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import TagRemove from "../../../src/components/Tags/TagRemove.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("TagDelete", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const tests = [
    {
      description: "Dialog Closed",
      props: {
        tagName: "tag",
        notHasAuthorization: false,
      },
      data: {
        showDialog: false,
      },
      template: {
        "remove-icon": true,
        "remove-title": true,
        "tagDelete-card": false,
      },
      templateText: {
        "remove-title": "Remove",
      },
    },
    {
      description: "Dialog opened",
      props: {
        tagName: "tag",
        notHasAuthorization: false,
      },
      data: {
        showDialog: true,
      },
      template: {
        "remove-icon": true,
        "remove-title": true,
        "tagDelete-card": true,
        "text-title": true,
        "text-text": true,
        "close-btn": true,
        "remove-btn": true,
      },
      templateText: {
        "remove-title": "Remove",
        "text-title": "Are you sure?",
        "text-text": "You are about to remove this tag.",
        "close-btn": "Close",
        "remove-btn": "Remove",
      },
    },
  ];

  const store = createStore({
    state: {},
    getters: {},
    actions: {
      "tags/remove": vi.fn(),
      "snackbar/showSnackbarSuccessAction": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(TagRemove, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            tagName: test.props.tagName,
            notHasAuthorization: test.props.notHasAuthorization,
          },
          shallow: true,
        });

        if(test.data.showDialog) {
          wrapper.vm.showDialog = true;
        }
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

      ///////s
      // Data checking
      //////
      it("Data is defined", () => {
        expect(wrapper.vm.$data).toBeDefined();
      });
      it('Receive data in props', () => {
        expect(wrapper.vm.tagName).toBe(test.props.tagName);
        expect(wrapper.vm.notHasAuthorization).toBe(test.props.notHasAuthorization);
      });
      it('Receive data', () => {
        expect(wrapper.vm.showDialog).toBe(test.data.showDialog);
      });

    });
  });
});
