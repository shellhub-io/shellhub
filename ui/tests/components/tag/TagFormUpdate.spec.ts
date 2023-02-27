import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import TagFormUpdate from "../../../src/components/Tags/TagFormUpdate.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("TagFormUpdate", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagFormUpdate>>;
  const vuetify = createVuetify();

  const tests = [
    {
      description: "Dialog closed with add",
      props: {
        deviceUid: "",
        tagsList: [],
      },
      data: {
        showDialog: false,
        listTagLocal: [],
        errorMsg: "",
      },
      computed: {
        hasTag: false,
      },
      template: {
        "edit-icon": true,
        "edit-title": true,
        "tagForm-card": false,
      },
      templateText: {
        "edit-title": "Add tags",
      },
    },
    {
      description: "Dialog closed with edit",
      props: {
        deviceUid: "xxxxxxx",
        tagsList: ["ShellHub"],
      },
      data: {
        showDialog: false,
        listTagLocal: ["ShellHub"],
        errorMsg: "",
      },
      computed: {
        hasTag: true,
      },
      template: {
        "edit-icon": true,
        "edit-title": true,
        "tagForm-card": false,
      },
      templateText: {
        "edit-title": "Edit tags",
      },
    },
    {
      description: "Dialog opened",
      props: {
        deviceUid: "xxxxxxx",
        tagsList: ["ShellHub"],
      },
      data: {
        showDialog: true,
        listTagLocal: ["ShellHub"],
        errorMsg: "",
      },
      computed: {
        hasTag: true,
      },
      template: {
        "edit-icon": true,
        "edit-title": true,
        "tagForm-card": true,
      },
      templateText: {
        "edit-title": "Edit tags",
      },
    },
  ];

  const store = createStore({
    state: {},
    getters: {},
    actions: {
      "devices/updateDeviceTag": vi.fn(),
      "snackbar/showSnackbarSuccessAction": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(TagFormUpdate, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            deviceUid: test.props.deviceUid,
            tagsList: test.props.tagsList,
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
        expect(wrapper.vm.deviceUid).toBe(test.props.deviceUid);
        expect(wrapper.vm.tagsList).toStrictEqual(test.props.tagsList);
      });
      it("Compare data with default value", () => {
        expect(wrapper.vm.showDialog).toBe(test.data.showDialog);
        expect(wrapper.vm.inputTags).toStrictEqual(test.data.listTagLocal);
        expect(wrapper.vm.tagsError).toBe(test.data.errorMsg);
      });
      it("Process data in the computed", () => {
        expect(wrapper.vm.hasTags).toBe(test.computed.hasTag);
      });
    });
  });
});
