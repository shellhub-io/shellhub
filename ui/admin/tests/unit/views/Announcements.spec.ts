import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import Announcements from "@admin/views/Announcements.vue";
import { Router } from "vue-router";

describe("Announcements", () => {
  let wrapper: VueWrapper<InstanceType<typeof Announcements>>;
  let router: Router;

  beforeEach(async () => {
    router = createCleanAdminRouter();
    await router.push({ name: "announcements" });
    await router.isReady();

    wrapper = mountComponent(Announcements, { global: { plugins: [router] } });
  });

  afterEach(() => { wrapper?.unmount(); });

  it("displays the page header with correct title", () => {
    const header = wrapper.find('[data-test="announcement-title"]');
    expect(header.exists()).toBe(true);
    expect(wrapper.text()).toContain("Announcements");
  });

  it("displays the page header with correct overline", () => {
    expect(wrapper.text()).toContain("Platform Messaging");
  });

  it("displays the page header description", () => {
    expect(wrapper.text()).toContain("Share important system broadcasts with every namespace administrator.");
  });

  it("displays the new announcement button", () => {
    const newBtn = wrapper.find('[data-test="new-announcement-btn"]');
    expect(newBtn.exists()).toBe(true);
    expect(newBtn.text()).toBe("New");
  });

  it("displays the announcement list component", () => {
    const list = wrapper.find('[data-test="announcement-list"]');
    const emptyState = wrapper.find('[data-test="announcements-empty-state"]');
    expect(list.exists() || emptyState.exists()).toBe(true);
  });

  it("navigates to new announcement page when button is clicked", async () => {
    const pushSpy = vi.spyOn(router, "push");
    const newBtn = wrapper.find('[data-test="new-announcement-btn"]');

    await newBtn.trigger("click");

    expect(pushSpy).toHaveBeenCalledWith({ name: "new-announcement" });
  });
});
