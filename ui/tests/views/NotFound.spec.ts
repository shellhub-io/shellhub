import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { vi, expect, describe, it, beforeEach } from "vitest";
import NotFound from "@/views/NotFound.vue";
import { key, store } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Not Found Page", () => {
  let wrapper: VueWrapper<InstanceType<typeof NotFound>>;
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(NotFound, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("renders the correct text and elements", () => {
    expect(wrapper.find("h1.text-h3.font-weight-bold.mt-6.mb-4").text()).toMatch("Whoops!");
    expect(wrapper.find("h1.text-h1.font-weight-bold.mt-4.mb-2.text-primary").text()).toMatch("404");
    expect(wrapper.find("p.font-weight-bold.text-h3").text()).toMatch("Page not found");
    expect(wrapper.find("p.font-weight-bold.text-h6").text())
      .toMatch("The requested URL was not found on the server. You can go back to the dashboard by clicking the button below.");
    expect(wrapper.find('[data-test="dashboard-btn"]').exists()).toBe(true);
  });

  it("navigates to dashboard on button click", async () => {
    const pushSpy = vi.spyOn(router, "push");
    await wrapper.findComponent('[data-test="dashboard-btn"]').trigger("click");
    expect(pushSpy).toHaveBeenCalledWith({ name: "Dashboard" });
  });

  it("navigates to dashboard on button click", async () => {
    const pushSpy = vi.spyOn(router, "push");
    await wrapper.vm.goToDashboard();
    expect(pushSpy).toHaveBeenCalledWith({ name: "Dashboard" });
  });

  it("navigates to the correct route", () => {
    const pushSpy = vi.spyOn(router, "push");
    wrapper.vm.goToDashboard();
    expect(pushSpy).toHaveBeenCalledWith({ name: "Dashboard" });
    pushSpy.mockRestore();
  });
});
