import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import BillingDialog from "@/components/Billing/BillingDialog.vue";
import { billingApi, namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type BillingDialogWrapper = VueWrapper<InstanceType<typeof BillingDialog>>;

describe("Billing Dialog", () => {
  let wrapper: BillingDialogWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockBilling: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const customerData = {
    id: "cus_test",
    name: "test",
    email: "test@test.com",
    payment_methods: [
      {
        id: "test_id",
        number: "xxxxxxxxxxxx4242",
        brand: "visa",
        exp_month: 3,
        exp_year: 2029,
        cvc: "",
        default: true,
      },
    ],
  };

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockBilling = new MockAdapter(billingApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockBilling.onPost("http://localhost:3000/api/billing/customer").reply(200);
    mockBilling.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(BillingDialog, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders dialog text", async () => {
    wrapper.vm.showCheckoutDialog = true;
    await flushPromises();

    expect(wrapper.find('[data-test="card-first-page"]').exists());
    expect(wrapper.find('[data-test="card-second-page"]').exists());
    expect(wrapper.find('[data-test="card-third-page"]').exists());
    expect(wrapper.find('[data-test="card-fourth-page"]').exists());
  });

  it("Render pagination", async () => {
    wrapper.vm.showCheckoutDialog = true;
    await flushPromises();

    expect(wrapper.findComponent('[data-test="payment-letter-next-button"]').exists());
    expect(wrapper.findComponent('[data-test="payment-letter-close-button"]').exists());
    expect(wrapper.findComponent('[data-test="payment-details-back-button"]').exists());
    expect(wrapper.findComponent('[data-test="payment-details-next-button"]').exists());
    expect(wrapper.findComponent('[data-test="checkout-back-button"]').exists());
    expect(wrapper.findComponent('[data-test="checkout-button"]').exists());
    expect(wrapper.findComponent('[data-test="successful-close-button"]').exists());
  });

  it("Pagination logic test", async () => {
    wrapper.vm.showCheckoutDialog = true;
    await flushPromises();

    expect(wrapper.vm.el).toEqual(1);
    await wrapper.findComponent('[data-test="payment-letter-next-button"]').trigger("click");
    expect(wrapper.vm.el).toEqual(2);
    await wrapper.findComponent('[data-test="payment-details-back-button"]').trigger("click");
    expect(wrapper.vm.el).toEqual(1);
  });

  it("Subscribe to Premium", async () => {
    wrapper.vm.showCheckoutDialog = true;
    wrapper.vm.el = 3;
    await flushPromises();

    mockBilling.onPost("http://localhost:3000/api/billing/subscription").reply(200);

    const subscribeSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="checkout-button"]').trigger("click");

    vi.runOnlyPendingTimers();

    expect(subscribeSpy).toHaveBeenCalledWith("customer/createSubscription");
  });
});
