import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { vi, expect, describe, it, beforeEach } from "vitest";
import NotFound from "@/views/NotFound.vue";

describe("NotFound.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof NotFound>>;
  const vuetify = createVuetify();
  const $router = { push: vi.fn() };

  beforeEach(async () => {
    wrapper = mount(NotFound, {
      global: {
        plugins: [vuetify],
        mocks: {
          $router,
        },
      },
    });
  });

  it("renders the correct text", () => {
    expect(wrapper.find("h1.text-h3.font-weight-bold.mt-6.mb-4").text()).toMatch("Whoops!");
    expect(wrapper.find("h1.text-h1.font-weight-bold.mt-4.mb-2.text-primary").text()).toMatch("404");
    expect(wrapper.find("p.font-weight-bold.text-h3").text()).toMatch("Page not found");
    // eslint-disable-next-line vue/max-len
    expect(wrapper.find("p.font-weight-bold.text-h6").text()).toMatch("The requested URL was not found on the server. You can go back to the dashboard by clicking the button below.");
    expect(wrapper.find('[data-test="dashboard-btn"]').exists()).toBe(true);
  });

  it("navigates to dashboard on button click", async () => {
    await wrapper.find('[data-test="dashboard-btn"]').trigger("click");
    expect($router.push).toHaveBeenCalledWith({ name: "Dashboard" });
  });
});
