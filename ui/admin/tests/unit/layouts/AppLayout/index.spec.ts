import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import SnackbarComponent from "../../../../src/components/Snackbar/Snackbar.vue";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import AppLayout from "../../../../src/layouts/AppLayout.vue";

const layout = "AppLayout";
type AppLayoutWrapper = VueWrapper<InstanceType<typeof AppLayout>>;

const license = {
  id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  expired: false,
  aboutToExpire: false,
  gracePeriod: false,
  issuedAt: -1,
  startsAt: -1,
  expiresAt: -1,
  allowedRegions: [],
  customer: {
    id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    name: "ShellHub",
    email: "contato@ossystems.com.br",
    company: "O.S. Systems",
  },
  features: {
    devices: -1,
    sessionRecording: true,
    firewallRules: true,
    billing: false,
  },
};

const store = createStore({
  state: {
    layout,
    license,
  },
  getters: {
    "layout/getLayout": (state) => state.layout,
    "license/license": (state) => state.license,
    "auth/isLoggedIn": () => true,
    "spinner/status": () => false,
    "auth/currentUser": () => "ossystem",
    "layout/getStatusDarkMode": () => "dark",
  },
  actions: {
    "layout/setLayout": vi.fn(),
    "layout/setStatusDarkMode": vi.fn(),
    "auth/logout": vi.fn(),
    "license/get": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("AppLayout", () => {
  let wrapper: AppLayoutWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = shallowMount(AppLayout, {
      global: {
        plugins: [[store, key], vuetify, routes],
        components: { SnackbarComponent },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
