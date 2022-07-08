import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import TagEdit from "../../../src/components/Tags/TagEdit.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("TagEdit", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const invalidName = ["xxx/", "xxx@", "xxx&", "xxx:"];
  const invalidMinAndMaxCharacters = ["x", "xx"];
  const tests = [
    {
      description: "Dialog closed",
      props: {
        tagName: "ShellHub",
        notHasAuthorization: false,
      },
      data: {
        tagLocal: "",
        showDialog: false,
      },
      template: {
        "edit-icon": true,
        "edit-title": true,
        "edit-itemtagForm-card": false,
      },
      templateText: {
        "edit-title": "Edit",
      },
    },
    {
      description: "Dialog Opened",
      props: {
        tagName: "ShellHub",
        notHasAuthorization: false,
      },
      data: {
        showDialog: true,
        tagLocal: "",
      },
      template: {
        "edit-icon": true,
        "edit-title": true,
        "tagForm-card": true,
        "text-title": true,
        "name-field": true,
        "cancel-btn": true,
        "edit-btn": true,
      },
      templateText: {
        "edit-title": "Edit",
        "text-title": "Edit tag",
        "cancel-btn": "Cancel",
        "edit-btn": "Edit",
      },
    },
  ];

  const store = createStore({
    state: {},
    getters: {},
    actions: {
      "tags/edit": vi.fn(),
      "snackbar/showSnackbarSuccessAction": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(TagEdit, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            tag: test.props.tagName,
            notHasAuthorization: test.props.notHasAuthorization,
          },
          shallow: true,
        });

        if (test.data.showDialog) {
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

      ///////
      // Data checking
      //////
      it("Data is defined", () => {
        expect(wrapper.vm.$data).toBeDefined();
      });
      it("Receive data in props", () => {
        expect(wrapper.vm.tag).toEqual(test.props.tagName);
        expect(wrapper.vm.notHasAuthorization).toEqual(
          test.props.notHasAuthorization
        );
      });
      it("Compare data with default value", () => {
        expect(wrapper.vm.tagLocal).toEqual(test.props.tagName);
        expect(wrapper.vm.showDialog).toEqual(test.data.showDialog);
      });
    });
  });
});
