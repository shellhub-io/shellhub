import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import Sessions from "@admin/views/Sessions.vue";

describe("Sessions", () => {
  let wrapper: VueWrapper<InstanceType<typeof Sessions>>;
  let router: ReturnType<typeof createCleanAdminRouter>;

  beforeEach(async () => {
    router = createCleanAdminRouter();
    await router.push({ name: "sessions" });
    await router.isReady();

    wrapper = mountComponent(Sessions, { global: { plugins: [router] } });
  });

  afterEach(() => { wrapper?.unmount(); });

  it("displays the page header with correct title", () => {
    expect(wrapper.text()).toContain("Sessions");
    expect(wrapper.text()).toContain("Activity Monitoring");
  });

  it("renders the session list component", () => {
    expect(wrapper.find('[data-test="session-list"]').exists()).toBe(true);
  });
});
