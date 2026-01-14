import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useStatsStore from "@admin/store/modules/stats";
import Dashboard from "@admin/views/Dashboard.vue";
import { mockStats } from "../mocks";
import getStats from "@admin/store/api/stats";

vi.mock("@admin/store/api/stats");

describe("Dashboard", () => {
  let wrapper: VueWrapper<InstanceType<typeof Dashboard>>;

  const mountWrapper = async (mockError?: Error) => {
    const statsStore = useStatsStore();
    if (mockError) {
      vi.mocked(statsStore.getStats).mockRejectedValueOnce(mockError);
      vi.mocked(getStats).mockRejectedValueOnce(mockError);
    }

    wrapper = mountComponent(Dashboard, {
      piniaOptions: {
        initialState: { adminStats: mockError ? {} : { stats: mockStats } },
        stubActions: !mockError,
      },
    });

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when stats load successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays the page header with correct title", () => {
      expect(wrapper.text()).toContain("System Overview");
      expect(wrapper.text()).toContain("Admin Dashboard");
    });

    it("displays the stats section heading", () => {
      expect(wrapper.text()).toContain("Stats");
    });

    it("displays all six stat cards with correct values", () => {
      const statsCards = [
        { title: "Registered Users", value: mockStats.registered_users },
        { title: "Registered Devices", value: mockStats.registered_devices },
        { title: "Online Devices", value: mockStats.online_devices },
        { title: "Pending Devices", value: mockStats.pending_devices },
        { title: "Rejected Devices", value: mockStats.rejected_devices },
        { title: "Active Sessions", value: mockStats.active_sessions },
      ];

      statsCards.forEach(({ title, value }) => {
        expect(wrapper.text()).toContain(title);
        expect(wrapper.text()).toContain(String(value));
      });
    });

    it("does not show the error message", () => {
      expect(wrapper.find('[data-test="dashboard-failed"]').exists()).toBe(false);
    });
  });

  describe("when stats fail to load", () => {
    it("displays error message when loading fails", async () => {
      await mountWrapper(createAxiosError(500, "Internal Server Error"));

      expect(wrapper.find('[data-test="dashboard-failed"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="dashboard-failed"]').text()).toContain("Something is wrong, try again!");
    });

    it("shows error snackbar for general errors", async () => {
      await mountWrapper(createAxiosError(500, "Internal Server Error"));

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load the dashboard stats. Please try again.");
    });

    it("shows specific error snackbar for license errors", async () => {
      await mountWrapper(createAxiosError(402, "Payment Required"));

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load the dashboard stats. Check your license and try again.");
    });
  });
});
