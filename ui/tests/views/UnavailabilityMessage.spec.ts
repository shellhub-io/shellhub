import { flushPromises, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Router } from "vue-router";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import UnavailabilityMessage from "@/views/UnavailabilityMessage.vue";
import { envVariables } from "@/envVariables";
import useUsersStore from "@/store/modules/users";
import { createAxiosError } from "@tests/utils/axiosError";
import { routes } from "@/router";

vi.mock("@/envVariables", () => ({
  envVariables: {
    isCloud: false,
    isEnterprise: false,
  },
}));

const mockRoutes = [
  ...routes,
  // Use System Unavailable without route guard
  { name: "System Unavailable", path: "/system-unavailable", component: UnavailabilityMessage },
];

describe("UnavailabilityMessage View", () => {
  let wrapper: VueWrapper<InstanceType<typeof UnavailabilityMessage>>;
  let router: Router;
  let usersStore: ReturnType<typeof useUsersStore>;

  const mountWrapper = async (redirectPath = "/devices") => {
    router = createCleanRouter(mockRoutes);
    await router.push({ name: "System Unavailable", query: { redirect: redirectPath } });
    await router.isReady();

    wrapper = mountComponent(UnavailabilityMessage, { global: { plugins: [router] } });

    usersStore = useUsersStore();
    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("community instance", () => {
    beforeEach(async () => {
      envVariables.isCloud = false;
      envVariables.isEnterprise = false;
      await mountWrapper();
    });

    it("renders the title for a connection failure", () => {
      expect(wrapper.find('[data-test="unavailability-title"]').text()).toBe("Unable to Connect to API");
    });

    it("renders the description for a connection failure", () => {
      expect(wrapper.find('[data-test="unavailability-description"]').text()).toContain("trouble connecting to your ShellHub server");
    });

    it("renders the retry button with correct label", () => {
      expect(wrapper.find('[data-test="retry-button"]').text()).toContain("Retry Connection");
    });

    it("renders the troubleshooting support link", () => {
      const link = wrapper.find('[data-test="support-link"]');
      expect(link.exists()).toBe(true);
      expect(link.text()).toContain("Troubleshooting Guide");
    });
  });

  describe("cloud instance", () => {
    beforeEach(async () => {
      envVariables.isCloud = true;
      envVariables.isEnterprise = false;
      await mountWrapper();
    });

    it("renders the title for a cloud outage", () => {
      expect(wrapper.find('[data-test="unavailability-title"]').text()).toBe("System Temporarily Unavailable");
    });

    it("renders the description for a cloud outage", () => {
      expect(wrapper.find('[data-test="unavailability-description"]').text()).toContain("temporarily down");
    });

    it("renders the retry button with correct label", () => {
      expect(wrapper.find('[data-test="retry-button"]').text()).toContain("Retry Connection");
    });

    it("renders the status page support link", () => {
      const link = wrapper.find('[data-test="support-link"]');
      expect(link.exists()).toBe(true);
      expect(link.text()).toContain("View Status Page");
    });
  });

  describe("enterprise instance", () => {
    beforeEach(async () => {
      envVariables.isCloud = false;
      envVariables.isEnterprise = true;
      await mountWrapper();
    });

    it("renders the title for an enterprise outage", () => {
      expect(wrapper.find('[data-test="unavailability-title"]').text()).toBe("System Temporarily Unavailable");
    });

    it("renders the description for an enterprise outage", () => {
      expect(wrapper.find('[data-test="unavailability-description"]').text()).toContain("currently unavailable");
    });

    it("renders the retry button with correct label", () => {
      expect(wrapper.find('[data-test="retry-button"]').text()).toContain("Try Again");
    });

    it("renders the report issue support link", () => {
      const link = wrapper.find('[data-test="support-link"]');
      expect(link.exists()).toBe(true);
      expect(link.text()).toContain("Report Issue");
    });
  });

  describe("retry behavior", () => {
    beforeEach(async () => {
      envVariables.isCloud = false;
      envVariables.isEnterprise = false;
      await mountWrapper();
    });

    it("navigates to / when retry succeeds and no redirect query param is set", async () => {
      wrapper.unmount();
      await mountWrapper("");
      const routerPushSpy = vi.spyOn(router, "push").mockImplementation(() => Promise.resolve());

      await wrapper.find('[data-test="retry-button"]').trigger("click");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith("/");
    });

    it("navigates to the redirect path when retry succeeds and redirect query param is set", async () => {
      const routerPushSpy = vi.spyOn(router, "push").mockImplementation(() => Promise.resolve());

      await wrapper.find('[data-test="retry-button"]').trigger("click");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith("/devices");
    });

    it("shows an error snackbar when retry fails", async () => {
      vi.spyOn(usersStore, "checkHealth").mockRejectedValueOnce(createAxiosError(503, "Service Unavailable"));

      await wrapper.find('[data-test="retry-button"]').trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "Still unable to connect to the API. Please try again later.",
      );
    });

    it("disables the retry button while retrying", async () => {
      vi.spyOn(usersStore, "checkHealth").mockImplementationOnce(
        () => new Promise((resolve) => setTimeout(resolve, 100)),
      );

      await wrapper.find('[data-test="retry-button"]').trigger("click");

      const retryButton = wrapper.find('[data-test="retry-button"]');
      expect(retryButton.attributes("disabled")).toBeDefined();
    });
  });
});
