import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import TagList from "../../../src/components/Tags/TagList.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

interface TagObject {
  name: string;
}

describe("TagList", () => {
  let wrapper: VueWrapper<InstanceType<typeof TagList>>;
  const vuetify = createVuetify();

  const numberTagsGlobal = 3;
  const tagsGlobal = ["tag1", "tag2", "tag3"];
  const tagsObject = tagsGlobal.map((str) => ({ name: str }));

  const headers = [
    {
      text: "Name",
      value: "name",
      align: "center",
      sortable: false,
    },
    {
      text: "Actions",
      value: "actions",
      align: "center",
      sortable: false,
    },
  ];

  const tests = [
    {
      description: "List data when user has owner role",
      role: {
        type: "owner",
        permission: true,
      },
      variables: {
        tagsObject,
        numberTagsGlobal,
      },
      data: {
        headers,
        hasAuthorizationRemove: true,
        hasAuthorizationEdit: true,
      },
      computed: {
        tags: tagsObject,
      },
    },
    {
      description: "List data when user has operator role",
      role: {
        type: "operator",
        permission: false,
      },
      variables: {
        tagsObject,
        numberTagsGlobal,
      },
      data: {
        headers,
        hasAuthorizationRemove: false,
        hasAuthorizationEdit: false,
      },
      computed: {
        tags: tagsObject,
      },
    },
  ];

  const store = (tags: Array<TagObject>, numberTags: number, hasAuthorizationRemove: string) => createStore({
    state: {
      tags,
      numberTags,
      hasAuthorizationRemove,
    },
    getters: {
      "tags/list": (state) => state.tags,
      "tags/getNumberTags": (state) => state.numberTags,
      "tags/hasAuthorizationRemove": (state) => state.hasAuthorizationRemove,
    },
    actions: {
      "tags/fetch": () => vi.fn(),
      "snackbar/showSnackbarErrorLoading": () => vi.fn(),
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(TagList, {
          global: {
            plugins: [
              [
                store(
                  test.variables.tagsObject,
                  test.variables.numberTagsGlobal,
                  test.role.type,
                ),
                key,
              ],
              routes,
              vuetify,
            ],
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
      it("Compare data with default value", () => {
        expect(wrapper.vm.headers).toEqual(test.data.headers);
      });
      it("Check computed properties", () => {
        expect(wrapper.vm.tags).toEqual(test.computed.tags);
      });

      //////
      // HTML validation
      //////

      it("Renders the template with data", () => {
        expect(
          wrapper.find('[data-test="tagListList-dataTable"]').exists(),
        ).toBeTruthy();
      });
    });
  });
});
