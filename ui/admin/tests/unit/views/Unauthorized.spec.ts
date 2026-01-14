import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import useAuthStore from "@admin/store/modules/auth";
import Unauthorized from "@admin/views/Unauthorized.vue";

describe("Unauthorized", () => {
  let wrapper: VueWrapper<InstanceType<typeof Unauthorized>>;
  let authStore: ReturnType<typeof useAuthStore>;

  beforeEach(() => {
    wrapper = mountComponent(Unauthorized);
    authStore = useAuthStore();
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  it("displays the access denied heading", () => {
    expect(wrapper.find("h1").text()).toBe("Admin Access Required");
  });

  it("displays the what you can do section", () => {
    expect(wrapper.find("h2").text()).toBe("What you can do:");
  });

  it("displays all action items", () => {
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

  it("displays the info alert", () => {
    const alert = wrapper.find(".v-alert");
    expect(alert.exists()).toBe(true);
    expect(alert.text()).toContain("If you believe you should have admin access");
  });

  it("displays logout and go to main app buttons", () => {
    const buttons = wrapper.findAll(".v-btn");
    expect(buttons).toHaveLength(2);

    expect(buttons[0].text()).toContain("Logout");
    expect(buttons[1].text()).toContain("Go to ShellHub");
  });

  it("calls logout and redirects when logout button is clicked", async () => {
    const logoutButton = wrapper.findAll(".v-btn")[0];
    await logoutButton.trigger("click");

    expect(authStore.logout).toHaveBeenCalled();
    expect(window.location.href).toBe("/login");
  });

  it("redirects to main app when go to shellhub button is clicked", async () => {
    const goToMainAppButton = wrapper.findAll(".v-btn")[1];
    await goToMainAppButton.trigger("click");

    expect(window.location.href).toBe("/");
  });
});
