import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import BillingWarning from "@/components/Billing/BillingWarning.vue";
import { key } from "@/store";
import { router } from "@/router";
import useAuthStore from "@/store/modules/auth";
import useBillingStore from "@/store/modules/billing";

setActivePinia(createPinia());
const authStore = useAuthStore();
const billingStore = useBillingStore();
const statusUpdateAccountDialog = true;
const statusUpdateAccountDialogByDeviceAction = false;

const stats = {
  registered_devices: 0,
  online_devices: 0,
  active_sessions: 0,
  pending_devices: 1,
  rejected_devices: 0,
};

const store = (statsData: typeof stats, billingEnabled: boolean, role: string) => {
  authStore.role = role;
  billingStore.billing.active = billingEnabled;
  return createStore({
    state: {
      stats: statsData,
      statusUpdateAccountDialog,
      statusUpdateAccountDialogByDeviceAction,
    },
    getters: {
      "stats/stats": (state) => state.stats,
      "users/statusUpdateAccountDialog": (state) => state.statusUpdateAccountDialog,
      "users/statusUpdateAccountDialogByDeviceAction": (state) => state.statusUpdateAccountDialogByDeviceAction,
    },
    actions: {
      "users/setStatusUpdateAccountDialog": vi.fn(),
    },
  });
};

const tests = [
  {
    ///////
    // In this case, uhe test dialog is closes when the user has less
    // than 3 devices and has no subscription.
    ///////

    description: "Less than 3 devices, no subscription",
    storeData: {
      stats,
      enable: false,
      permission: "owner",
    },
    computed: {
      showWarningDialog: false,
    },
    template: {
      '[data-test="billing-warning-dialog"]': false,
      '[data-test="close-btn"]': false,
      '[data-test="goToBilling-btn"]': false,
    },
  },
  {
    ///////
    // In this case, the test dialog is closes when the user has 3
    // devices and has subscription.
    ///////

    description: "3 devices, with subscription",
    storeData: {
      stats: { ...stats, registered_devices: 3 },
      enable: true,
      permission: "owner",
    },
    computed: {
      showWarningDialog: false,
    },
    template: {
      '[data-test="billing-warning-dialog"]': false,
      '[data-test="close-btn"]': false,
      '[data-test="goToBilling-btn"]': false,
    },
  },
  {
    ///////
    // In this case, the test dialog is closes when the user has 3
    // devices and has no subscription.
    ///////

    description: "3 devices, no subscription",
    storeData: {
      stats: { ...stats, registered_devices: 3 },
      enable: false,
      permission: "owner",
    },
    computed: {
      showWarningDialog: true,
      hasAuthorization: true,
    },
    template: {
      '[data-test="billing-warning-dialog"]': true,
      '[data-test="close-btn"]': true,
      '[data-test="goToBilling-btn"]': true,
    },
  },
  {
    ///////
    // In this case, the test dialog does not open for user not owner
    ///////

    description: "Not the owner",
    storeData: {
      stats: { ...stats, registered_devices: 3 },
      enable: false,
      permission: "operator",
    },
    computed: {
      showWarningDialog: true,
      hasAuthorization: false,
    },
    template: {
      '[data-test="billing-warning-dialog"]': false,
      '[data-test="close-btn"]': false,
      '[data-test="goToBilling-btn"]': false,
    },
  },
];

describe("BillingWarning", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingWarning>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(test.description, () => {
      beforeEach(() => {
        wrapper = mount(BillingWarning, {
          global: {
            plugins: [
              // eslint-disable-next-line @typescript-eslint/ban-ts-comment
              // @ts-ignore
              [store(...Object.values(test.storeData)), key],
              router,
              vuetify,
            ],
          },

          shallow: true,
        });
      });

      it("Is a Vue instance", () => {
        expect(wrapper).toBeTruthy();
      });
      it("Renders the component", () => {
        expect(wrapper.html()).toMatchSnapshot();
      });

      it("Process data in the computed", () => {
        Reflect.ownKeys(test.computed).forEach((c) => {
          expect(wrapper.vm[c]).toEqual(test.computed[c]);
        });
      });
    });
  });
});
