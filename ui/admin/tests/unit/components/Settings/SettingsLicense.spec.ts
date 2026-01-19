import { describe, expect, it, vi, afterEach, beforeEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useLicenseStore from "@admin/store/modules/license";
import SettingsLicense from "@admin/components/Settings/SettingsLicense.vue";
import {
  mockLicense,
  mockLicenseExpired,
  mockLicenseAboutToExpire,
  mockLicenseGracePeriod,
  mockLicenseRegional,
  mockNoLicense,
} from "../../mocks";
import { IAdminLicense } from "@admin/interfaces/ILicense";
import * as licenseApi from "@admin/store/api/license";

vi.mock("@admin/store/api/license");

describe("SettingsLicense", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingsLicense>>;
  let licenseStore: ReturnType<typeof useLicenseStore>;

  const mountWrapper = async (license: Partial<IAdminLicense> = mockLicense) => {
    wrapper = mountComponent(SettingsLicense, { piniaOptions: { initialState: { adminLicense: { license } } } });

    licenseStore = useLicenseStore();

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders page header", () => {
      const header = wrapper.find('[title-test-id="license-header"]');
      expect(header.exists()).toBe(true);
      expect(header.text()).toContain("License Details");
    });

    it("renders license card", () => {
      const card = wrapper.find('[data-test="license-card"]');
      expect(card.exists()).toBe(true);
    });

    it("displays issued at date", () => {
      const issuedAt = wrapper.find('[data-test="issued-at-field"]');
      expect(issuedAt.exists()).toBe(true);
    });

    it("displays starts at date", () => {
      const startsAt = wrapper.find('[data-test="starts-at-field"]');
      expect(startsAt.exists()).toBe(true);
    });

    it("displays expires at date", () => {
      const expiresAt = wrapper.find('[data-test="expires-at-field"]');
      expect(expiresAt.exists()).toBe(true);
    });

    it("displays global license badge when no regions specified", () => {
      const globalBadge = wrapper.text();
      expect(globalBadge).toContain("Global");
    });

    it("displays regional license badge with regions", async () => {
      wrapper.unmount();
      await mountWrapper(mockLicenseRegional);

      const badge = wrapper.text();
      expect(badge).toContain("Limited");
      expect(badge).toContain("US, EU");
    });

    it("displays customer id", () => {
      const customerId = wrapper.find('[data-test="id"]');
      expect(customerId.exists()).toBe(true);
    });

    it("displays customer name", () => {
      const customerName = wrapper.find('[data-test="name"]');
      expect(customerName.exists()).toBe(true);
    });

    it("displays customer email", () => {
      const customerEmail = wrapper.find('[data-test="email"]');
      expect(customerEmail.exists()).toBe(true);
    });

    it("displays customer company", () => {
      const customerCompany = wrapper.find('[data-test="company"]');
      expect(customerCompany.exists()).toBe(true);
    });

    it("displays devices feature", () => {
      const devices = wrapper.find('[data-test="devices"]');
      expect(devices.exists()).toBe(true);
    });

    it("displays session_recording feature", () => {
      const sessionRecording = wrapper.find('[data-test="session_recording"]');
      expect(sessionRecording.exists()).toBe(true);
    });

    it("displays firewall_rules feature", () => {
      const firewallRules = wrapper.find('[data-test="firewall_rules"]');
      expect(firewallRules.exists()).toBe(true);
    });

    it("shows included icon for enabled boolean features", () => {
      const includedIcons = wrapper.findAll('[data-test="included-icon"]');
      expect(includedIcons.length).toBeGreaterThan(0);
    });

    it("shows not included icon for disabled boolean features", () => {
      const notIncludedIcons = wrapper.findAll('[data-test="not-included-icon"]');
      expect(notIncludedIcons.length).toBeGreaterThan(0);
    });
  });

  describe("license alerts", () => {
    it("shows no license alert when license not installed", async () => {
      await mountWrapper(mockNoLicense);

      const alert = wrapper.find('[data-test="license-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("You do not have an installed license");
    });

    it("shows about to expire alert", async () => {
      await mountWrapper(mockLicenseAboutToExpire);

      const alert = wrapper.find('[data-test="license-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("Your license is about to expire");
    });

    it("shows grace period alert when expired in grace period", async () => {
      await mountWrapper(mockLicenseGracePeriod);

      const alert = wrapper.find('[data-test="license-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("expired, but you are still within the grace period");
    });

    it("shows expired alert when license expired", async () => {
      await mountWrapper(mockLicenseExpired);

      const alert = wrapper.find('[data-test="license-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("Your license has expired!");
    });

    it("shows no alert when license is valid", async () => {
      await mountWrapper(mockLicense);

      const alert = wrapper.find('[data-test="license-alert"]');
      expect(alert.exists()).toBe(false);
    });
  });

  describe("initial data loading", () => {
    it("fetches license on mount", async () => {
      await mountWrapper();
      expect(licenseStore.getLicense).toHaveBeenCalled();
    });

    it("shows error when license fetch fails", async () => {
      vi.mocked(licenseApi.getLicense).mockRejectedValue(
        createAxiosError(500, "Internal Server Error"),
      );
      mountComponent(SettingsLicense, { piniaOptions: { stubActions: false } });
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Error loading license.");
    });
  });

  describe("file upload", () => {
    beforeEach(() => mountWrapper());

    it("renders file input", () => {
      const fileInput = wrapper.find('input[type="file"]');
      expect(fileInput.exists()).toBe(true);
    });

    it("accepts only .dat files", () => {
      const fileInput = wrapper.find('input[type="file"]');
      expect(fileInput.attributes("accept")).toBe(".dat");
    });

    it("has upload button disabled by default", () => {
      const uploadBtn = wrapper.find('[data-test="upload-license-btn"]');
      expect(uploadBtn.attributes("disabled")).toBeDefined();
    });

    it("uploads license file successfully", async () => {
      const file = new File(["license content"], "license.dat", { type: "application/octet-stream" });
      const fileInput = wrapper.find('input[type="file"]');

      Object.defineProperty(fileInput.element, "files", {
        value: [file],
        writable: false,
      });
      await fileInput.trigger("change");
      await flushPromises();

      const uploadBtn = wrapper.find('[data-test="upload-license-btn"]');
      await uploadBtn.trigger("click");
      await flushPromises();

      expect(licenseStore.uploadLicense).toHaveBeenCalledWith(file);
      expect(licenseStore.getLicense).toHaveBeenCalled();
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("License uploaded successfully.");
    });

    it("shows error when upload fails", async () => {
      vi.mocked(licenseStore.uploadLicense).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const file = new File(["license content"], "license.dat", { type: "application/octet-stream" });
      const fileInput = wrapper.find('input[type="file"]');

      Object.defineProperty(fileInput.element, "files", {
        value: [file],
        writable: false,
      });
      await fileInput.trigger("change");
      await flushPromises();

      const uploadBtn = wrapper.find('[data-test="upload-license-btn"]');
      await uploadBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to upload the license.");
    });
  });

  describe("conditional rendering", () => {
    it("hides license details when no license installed", async () => {
      await mountWrapper(mockNoLicense);

      const issuedAt = wrapper.find('[data-test="issued-at-field"]');
      expect(issuedAt.exists()).toBe(false);
    });

    it("shows license details when license installed", async () => {
      await mountWrapper(mockLicense);

      const issuedAt = wrapper.find('[data-test="issued-at-field"]');
      expect(issuedAt.exists()).toBe(true);
    });
  });
});
