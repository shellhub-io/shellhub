import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import PublicKeyList from "../../../src/components/PublicKeys/PublicKeysList.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const numberPublicKeysGlobal = 2;

const publicKeysGlobal = [
  {
    data: "BBGVvbmF",
    fingerprint: "00:00:00",
    created_at: "2020-11-23T20:59:13.323Z",
    tenant_id: "xxxxxxxx",
    name: "shellhub",
  },
  {
    data: "AbGVvbmF",
    fingerprint: "00:00:00",
    created_at: "2020-11-23T20:59:13.323Z",
    tenant_id: "xxxxxxxx",
    name: "shellhub",
  },
];

const headers = [
  {
    text: "Name",
    value: "name",
  },
  {
    text: "Fingerprint",
    value: "fingerprint",
  },
  {
    text: "Filter",
    value: "filter",
  },
  {
    text: "Username",
    value: "username",
  },
  {
    text: "Created At",
    value: "created_at",
  },
  {
    text: "Actions",
    value: "actions",
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
      publicKeysGlobal,
      numberPublicKeysGlobal,
    },
    data: {
      pagination: {},
      loading: false,
      itemsPerPage: 10,
      page: 1,
      publicKeyFormDialogShow: [],
      publicKeyDeleteShow: [],
      editAction: "edit",
      headers,
    },
    computed: {
      publicKeys: publicKeysGlobal,
      getNumberPublicKeys: numberPublicKeysGlobal,
      hasAuthorizationFormDialogEdit: true,
      hasAuthorizationFormDialogRemove: true,
    },
  },
  {
    description: "List data when user has operator role",
    role: {
      type: "operator",
      permission: false,
    },
    variables: {
      publicKeysGlobal,
      numberPublicKeysGlobal,
    },
    data: {
      pagination: {},
      loading: false,
      itemsPerPage: 10,
      page: 1,
      publicKeyFormDialogShow: [],
      publicKeyDeleteShow: [],
      editAction: "edit",
      headers,
    },
    computed: {
      publicKeys: publicKeysGlobal,
      getNumberPublicKeys: numberPublicKeysGlobal,
      hasAuthorizationFormDialogEdit: false,
      hasAuthorizationFormDialogRemove: false,
    },
  },
];

const store = (
  publicKeys: typeof publicKeysGlobal,
  numberPublicKeys: number,
  status: string,
  currentRole: boolean,
) => createStore({
  state: {
    publicKeys,
    numberPublicKeys,
    status,
    currentRole,
  },
  getters: {
    "publicKeys/list": (state) => state.publicKeys,
    "publicKeys/getNumberPublicKeys": (state) => state.numberPublicKeys,
    "box/getStatus": (state) => state.status,
    "auth/role": (state) => state.currentRole,
  },
  actions: {
    "publicKeys/fetch": vi.fn(),
    "snackbar/showSnackbarErrorLoading": vi.fn(),
    "box/setStatus": vi.fn(),
  },
});

describe("PublicKeyList", () => {
  let wrapper: VueWrapper<InstanceType<typeof PublicKeyList>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(PublicKeyList, {
          global: {
            plugins: [
              [
                store(
                  test.variables.publicKeysGlobal,
                  test.variables.numberPublicKeysGlobal,
                  test.role.type,
                  test.role.permission,
                ),
                key,
              ],
              routes,
              vuetify,
            ],
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
      // Data and Props checking
      //////

      it("Data is defined", () => {
        expect(wrapper.vm.$data).toBeDefined();
      });

      it("Compare data with default value", () => {
        expect(wrapper.vm.headers).toStrictEqual(headers);
        expect(wrapper.vm.loading).toStrictEqual(test.data.loading);
        expect(wrapper.vm.page).toStrictEqual(test.data.page);
        expect(wrapper.vm.itemsPerPage).toStrictEqual(test.data.itemsPerPage);
      });

      it("Compare the computed with the default value", () => {
        expect(wrapper.vm.publicKeys).toStrictEqual(test.computed.publicKeys);
        expect(wrapper.vm.getNumberPublicKeys).toStrictEqual(
          test.computed.getNumberPublicKeys,
        );
      });

      //////
      // HTML validation
      //////

      it("Renders the template with data", () => {
        const dt = wrapper.find('[data-test="publicKeys-list"]');
        expect(dt.exists()).toBeTruthy();
      });
    });
  });
});
