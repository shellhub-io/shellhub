import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import NamespaceDelete from "../../../src/components/Namespace/NamespaceDelete.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { envVariables } from "../../../src/envVariables";

const infoData = {
  info: {
    periodEnd: "2021-12-24T18:16:21Z",
    description: "Shellhub",
    latestPaymentDue: 0,
    latestPaymentPaid: 0,
    nextPaymentDue: 0,
    nextPaymenPaid: 0,
  },
  card: {
    brand: "visa",
    expYear: 2024,
    default: true,
    expMonth: 4,
    last4: "4042",
    id: "pm_1JzQ80KJsksFHO6pREJA5TrK",
  },
  cards: [
    {
      brand: "visa",
      expYear: 2024,
      default: true,
      expMonth: 4,
      last4: "4042",
      id: "pm_1JzQ80KJsksFHO6pREJA5TrK",
    },
    {
      brand: "visa",
      expYear: 2028,
      default: false,
      expMonth: 4,
      last4: "4042",
      id: "pm_1JzQ80KJsksFHO6pREJA5TrG",
    },
    {
      brand: "visa",
      expYear: 2029,
      default: false,
      expMonth: 4,
      last4: "4042",
      id: "pm_1JzQ80KJsksFHO6pREJA5TrF",
    },
  ],
};

const nsTenant = "xxxxxx";

const role = ["owner", "operator"];

const hasAuthorization = {
  owner: true,
  operator: false,
};

const inactiveBilling = {
  active: false,
  current_period_end: 0,
  customer_id: "",
  subscription_id: "",
  payment_method_id: "",
};

const activeBilling = {
  active: true,
  current_period_end: 12121,
  customer_id: "cus_123",
  subscription_id: "subs_123",
  payment_method_id: "pm_123",
};

const namespaceObject = {
  name: "namespace3",
  owner: "user1",
  member_names: ["user1", "user7", "user8"],
  tenant_id: "xxxxxxxx",
};

const getter = {
  billingActive: [inactiveBilling, activeBilling],
};

const tests = [
  {
    description: "Button",
    props: {
      nsTenant,
    },
    data: {
      name: namespaceObject.name,
      dialog: false,
      action: "remove",
    },
    computed: {
      tenant: nsTenant,
      active: getter.billingActive[0].active,
      billing: getter.billingActive[0],
      billingActive: getter.billingActive[0].active,
    },
    namespace: namespaceObject,
    env: {
      billingEnable: false,
    },
    info: {},
    template: {
      "delete-btn": true,
      "namespaceDelete-dialog": false,
      "contentSubscription-p": false,
      "close-btn": false,
      "remove-btn": false,
    },
  },
  {
    description: "Dialog without subscription",
    props: {
      nsTenant,
    },
    data: {
      name: namespaceObject.name,
      dialog: false,
      action: "remove",
    },
    computed: {
      tenant: nsTenant,
      active: true,
      billing: getter.billingActive[0],
      billingActive: true,
    },
    namespace: namespaceObject,
    env: {
      billingEnable: false,
    },
    info: {},
    template: {
      "delete-btn": true,
      "namespaceDelete-dialog": true,
      "contentSubscription-p": false,
      "close-btn": true,
      "remove-btn": true,
    },
  },
  {
    description: "Dialog with subscription",
    props: {
      nsTenant,
    },
    data: {
      name: namespaceObject.name,
      dialog: false,
      action: "remove",
    },
    computed: {
      tenant: nsTenant,
      active: false,
      billing: getter.billingActive[1],
      billingActive: false,
    },
    namespace: namespaceObject,
    env: {
      billingEnable: true,
    },
    info: infoData,
    template: {
      "delete-btn": true,
      "namespaceDelete-dialog": true,
      "contentSubscription-p": true,
      "close-btn": true,
      "remove-btn": true,
    },
  },
];

const store = (
  active: boolean,
  billing: typeof getter.billingActive[0],
  currentrole: string,
  namespace: typeof namespaceObject,
  info: typeof infoData | {},
) => createStore({
  state: {
    active,
    billing,
    currentrole,
    namespace,
    info,
  },
  getters: {
    "billing/active": (state) => state.active,
    "billing/get": (state) => state.billing,
    "auth/role": (state) => state.currentrole,
    "namespaces/get": (state) => state.namespace,
    "billing/getBillInfoData": (state) => state.info,
  },
  actions: {
    "namespaces/remove": vi.fn(),
    "auth/logout": vi.fn(),
    "billing/getSubscription": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
    "snackbar/showSnackbarErrorDefault": vi.fn(),
  },
});

describe("NamespaceDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceDelete>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(NamespaceDelete, {
            global: {
              plugins: [
                [
                  store(
                    test.computed.active,
                    test.computed.billing,
                    currentrole,
                    test.namespace,
                    test.info,
                  ),
                  key,
                ],
                routes,
                vuetify,
              ],
            },
            props: { nsTenant: test.props.nsTenant },
            mocks: {
              $stripe: {
                elements: () => ({
                  create: () => ({
                    mount: () => null,
                  }),
                }),
              },
            },
            shallow: true,
          });

          envVariables.billingEnable = test.env.billingEnable;
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
          Object.keys(test.props).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.props[item]);
          });
        });

        it("Receive data in data", () => {
          expect(wrapper.vm.dialog).toEqual(test.data.dialog);
          expect(wrapper.vm.name).toEqual(test.data.name);
        });
        it("Process data in the computed", () => {
          expect(wrapper.vm.tenant).toEqual(test.computed.tenant);
          expect(wrapper.vm.billingActive).toEqual(test.computed.billingActive);
          expect(wrapper.vm.billing).toEqual(test.computed.billing);
        });

        //////
        // HTML validation
        //////
        // TODO
      });
    });
  });
});
