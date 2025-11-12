import { createVuetify } from "vuetify";
import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useAuthStore from "@admin/store/modules/auth";
import Unauthorized from "@admin/views/Unauthorized.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Unauthorized", () => {
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  authStore.logout = vi.fn();

  const wrapper = mount(Unauthorized, { global: { plugins: [createVuetify(), SnackbarPlugin] } });

  it("Renders the correct heading", () => {
    expect(wrapper.find("h1").text()).toBe("Admin Access Required");
  });

  it("Renders the correct subheading", () => {
    expect(wrapper.find("h2").text()).toBe("What you can do:");
  });

  it("Renders all action items", () => {
    const actionItems = [
      "Return to the main ShellHub application",
      "Contact your system administrator for admin access",
      "Manage your devices, sessions, and namespaces in the main app",
    ];

    const listItems = wrapper.findAll(".v-list-item-title");
    expect(listItems).toHaveLength(actionItems.length);

    actionItems.forEach((item, index) => {
      expect(listItems[index].text()).toBe(item);
    });
  });

  it("Renders the info alert", () => {
    const alert = wrapper.find(".v-alert");
    expect(alert.exists()).toBe(true);
    expect(alert.text()).toContain("If you believe you should have admin access");
  });

  it("Renders both action buttons", () => {
    const buttons = wrapper.findAll(".v-btn");
    expect(buttons).toHaveLength(2);

    expect(buttons[0].text()).toContain("Logout");
    expect(buttons[1].text()).toContain("Go to ShellHub");
  });

  it("Calls logout when logout button is clicked", async () => {
    const authStore = useAuthStore();
    const logoutButton = wrapper.findAll(".v-btn")[0];

    await logoutButton.trigger("click");

    expect(authStore.logout).toHaveBeenCalled();
    expect(window.location.href).toBe("/login");
  });

  it("Redirects to main app when 'Go to ShellHub' button is clicked", async () => {
    const goToMainAppButton = wrapper.findAll(".v-btn")[1];

    await goToMainAppButton.trigger("click");

    expect(window.location.href).toBe("/");
  });
});
