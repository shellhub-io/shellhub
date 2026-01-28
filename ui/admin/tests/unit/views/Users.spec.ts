import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import Users from "@admin/views/Users.vue";

describe("Users", () => {
  let wrapper: VueWrapper<InstanceType<typeof Users>>;
  let router: ReturnType<typeof createCleanAdminRouter>;

  beforeEach(async () => {
    router = createCleanAdminRouter();
    await router.push({ name: "users" });
    await router.isReady();

    wrapper = mountComponent(Users, { global: { plugins: [router] } });
  });

  afterEach(() => {
    wrapper?.unmount();
  });

  it("displays the page header with correct title", () => {
    expect(wrapper.text()).toContain("Users");
    expect(wrapper.text()).toContain("Account Management");
  });

  it("displays the search input field", () => {
    const searchInput = wrapper.find('input[type="text"]');
    expect(searchInput.exists()).toBe(true);
  });

  it("displays the export users button", () => {
    expect(wrapper.find('[data-test="users-export-btn"]').exists()).toBe(true);
  });

  it("displays the add user button", () => {
    expect(wrapper.find('[data-test="user-add-btn"]').exists()).toBe(true);
  });

  it("displays the users list component", () => {
    const list = wrapper.find('[data-test="users-list"]');
    const emptyState = wrapper.find('[data-test="users-empty-state"]');
    expect(list.exists() || emptyState.exists()).toBe(true);
  });
});
