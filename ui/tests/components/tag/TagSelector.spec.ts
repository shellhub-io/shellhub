import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import TagSelector from "../../../src/components/Tags/TagSelector.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("TagSelector", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const tagsGlobal = ["ShellHub", "Shell", "Hub"];

  const tests = [
    {
      description: "Without tags",
      variables: {
        tags: [],
        selectedTags: [],
      },
      data: {
        prevSelectedLength: 0,
      },
      computed: {
        getListTags: [],
      },
      template: {
        "tags-btn": true,
      },
    },
    {
      description: "With tags",
      variables: {
        tags: tagsGlobal,
        selectedTags: [],
      },
      data: {
        prevSelectedLength: 0,
      },
      computed: {
        getListTags: tagsGlobal,
      },
    },
  ];

  const store = (tags: any, selectedTags?: any) => {
    return createStore({
      state: {
        tags,
        selectedTags,
      },
      getters: {
        "tags/list": (state) => state.tags,
        "tags/selected": (state) => state.selectedTags,
      },
      actions: {
        "tags/setSelected": vi.fn(),
        "tags/fetch": vi.fn(),
        "devices/setFilter": vi.fn(),
        "devices/refresh": vi.fn(),
        "snackbar/showSnackbarErrorAssociation": vi.fn(),
        "snackbar/showSnackbarErrorDefault": vi.fn(),
      },
    });
  };

  tests.forEach((test, index) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        if (index === 0) {
          wrapper = mount(TagSelector, {
            global: {
              plugins: [
                [store(test.variables.tags, test.variables.selectedTags), key],
                routes,
                vuetify,
              ],
            },
          });
        } else {
          wrapper = mount(TagSelector, {
            global: {
              plugins: [
                [store(test.variables.tags, test.variables.selectedTags), key],
                routes,
                vuetify,
              ],
            },
            computed: {
              getListTags: () => test.computed.getListTags,
            },
          });
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
      it("Compare data with default value", () => {
        expect(wrapper.vm.prevSelectedLength).toBe(
          test.data.prevSelectedLength
        );
      });

      //////
      // HTML validation
      //////

      it("Renders the template with data", () => {
        if (index === 0) {
          // @ts-ignore
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(
              // @ts-ignore
              test.template[item]
            );
          });
        }
      });
    });
  });
});
