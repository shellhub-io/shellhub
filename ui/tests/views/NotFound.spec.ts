import { vi, expect, describe, it } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import NotFound from "@/views/NotFound.vue";
import createCleanRouter from "@tests/utils/router";

describe("Not Found Page", () => {
  const router = createCleanRouter();
  const wrapper = mountComponent(NotFound, { global: { plugins: [router] } });

  it("renders all page elements correctly", () => {
    expect(wrapper.find('[data-test="whoops-heading"]').text()).toBe("Whoops!");
    expect(wrapper.find('[data-test="404-heading"]').text()).toBe("404");
    expect(wrapper.find('[data-test="not-found-message"]').text()).toBe("Page not found");
    expect(wrapper.find('[data-test="help-text"]').text()).toContain("The requested URL was not found on the server");
    expect(wrapper.find('[data-test="error-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="home-btn"]').exists()).toBe(true);
  });

  it("navigates to home when button is clicked", async () => {
    const pushSpy = vi.spyOn(router, "push");
    await wrapper.find('[data-test="home-btn"]').trigger("click");
    expect(pushSpy).toHaveBeenCalledWith({ name: "Home" });
  });
});
