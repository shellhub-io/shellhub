import { createVuetify } from "vuetify";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import SnackbarComponent from "../../../../src/components/Snackbar/Snackbar.vue";
import routes from "../../../../src/router";
import AppLayout from "../../../../src/layouts/AppLayout.vue";

type AppLayoutWrapper = VueWrapper<InstanceType<typeof AppLayout>>;

vi.mock("@admin/store/modules/auth", () => ({
  default: vi.fn(() => ({
    isLoggedIn: true,
    currentUser: "ossystem",
    logout: vi.fn(),
  })),
}));

vi.mock("@admin/store/modules/layout", () => ({
  default: vi.fn(() => ({
    getLayout: "AppLayout",
    getStatusDarkMode: "dark",
    setLayout: vi.fn(),
    setStatusDarkMode: vi.fn(),
  })),
}));

vi.mock("@admin/store/modules/license", () => ({
  default: vi.fn(() => ({
    isExpired: false,
    license: {
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
    },
    get: vi.fn(),
  })),
}));

vi.mock("@admin/store/modules/spinner", () => ({
  default: vi.fn(() => ({
    getStatus: false,
  })),
}));

describe("AppLayout", () => {
  let wrapper: AppLayoutWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    wrapper = shallowMount(AppLayout, {
      global: {
        plugins: [vuetify, routes],
        components: { SnackbarComponent },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
