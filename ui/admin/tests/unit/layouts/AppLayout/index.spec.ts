import { createVuetify } from "vuetify";
import { describe, expect, it, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { SnackbarPlugin } from "@/plugins/snackbar";
import routes from "@admin/router";
import AppLayout from "@admin/layouts/AppLayout.vue";

vi.mock("@admin/store/modules/auth", () => ({
  default: vi.fn(() => ({
    isLoggedIn: true,
  })),
}));

vi.mock("@admin/store/modules/license", () => ({
  default: vi.fn(() => ({
    isExpired: false,
    license: {},
  })),
}));

vi.mock("@admin/store/modules/spinner", () => ({
  default: vi.fn(() => ({
    status: false,
  })),
}));

describe("AppLayout", () => {
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const wrapper = shallowMount(AppLayout, { global: { plugins: [vuetify, routes, SnackbarPlugin] } });
  it("Renders the component", () => { expect(wrapper.html()).toMatchSnapshot(); });
  it("Passes AppBarContent flags", () => {
    const appBarContent = wrapper.findComponent({ name: "AppBarContent" });
    expect(appBarContent.exists()).toBe(true);
    expect(appBarContent.props("showMenuToggle")).toBe(true);
    expect(appBarContent.props("showSupport")).toBe(true);
  });
});
