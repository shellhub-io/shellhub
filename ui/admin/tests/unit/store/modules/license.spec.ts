import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import axios from "axios";
import useLicenseStore from "@admin/store/modules/license";
import { IAdminLicense } from "@admin/interfaces/ILicense";

const mockLicense: IAdminLicense = {
  expired: false,
  about_to_expire: false,
  grace_period: false,
  id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  issued_at: 1704067200,
  starts_at: 1704067200,
  expires_at: 1735689600,
  allowed_regions: [],
  customer: {
    id: "customer-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    name: "ShellHub",
    email: "contact@shellhub.io",
    company: "ShellHub Inc",
  },
  features: {
    devices: 100,
    session_recording: true,
    firewall_rules: true,
    reports: true,
    login_link: true,
    billing: true,
  },
};

const mockExpiredLicense: IAdminLicense = {
  ...mockLicense,
  expired: true,
};

describe("Admin License Store", () => {
  let licenseStore: ReturnType<typeof useLicenseStore>;
  let mockAdminApi: MockAdapter;
  let mockAxios: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    licenseStore = useLicenseStore();
    mockAdminApi = new MockAdapter(adminApi.getAxios());
    mockAxios = new MockAdapter(axios);
    localStorage.clear();
  });

  afterEach(() => {
    mockAdminApi.reset();
    mockAxios.reset();
    localStorage.clear();
  });

  describe("Initial State", () => {
    it("should have empty license object", () => {
      expect(licenseStore.license).toEqual({});
    });

    it("should have isExpired as true when license is empty", () => {
      expect(licenseStore.isExpired).toBe(true);
    });
  });

  describe("Computed Properties", () => {
    it("should compute isExpired as false when license is not expired", () => {
      licenseStore.license = mockLicense;

      expect(licenseStore.isExpired).toBe(false);
    });

    it("should compute isExpired as true when license is expired", () => {
      licenseStore.license = mockExpiredLicense;

      expect(licenseStore.isExpired).toBe(true);
    });

    it("should compute isExpired as true when expired field is undefined", () => {
      const licenseWithoutExpired = { ...mockLicense };
      delete (licenseWithoutExpired as Partial<IAdminLicense>).expired;
      licenseStore.license = licenseWithoutExpired as IAdminLicense;

      expect(licenseStore.isExpired).toBe(true);
    });
  });

  describe("getLicense", () => {
    const baseUrl = "http://localhost:3000/admin/api/license";

    it("should fetch license successfully and update state", async () => {
      mockAdminApi.onGet(baseUrl).reply(200, mockLicense);

      await licenseStore.getLicense();

      expect(licenseStore.license).toEqual(mockLicense);
      expect(licenseStore.isExpired).toBe(false);
    });

    it("should throw on not found error when fetching license", async () => {
      mockAdminApi.onGet(baseUrl).reply(404);

      await expect(licenseStore.getLicense()).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when fetching license", async () => {
      mockAdminApi.onGet(baseUrl).networkError();

      await expect(licenseStore.getLicense()).rejects.toThrow("Network Error");
    });
  });

  describe("uploadLicense", () => {
    const uploadUrl = `${window.location.origin}/admin/api/license`;

    it("should upload license file successfully with FormData", async () => {
      const mockFile = new File(["license content"], "license.txt", { type: "text/plain" });
      const token = "test-bearer-token";
      localStorage.setItem("token", token);

      mockAxios
        .onPost(uploadUrl)
        .reply((config) => {
          expect(config.data).toBeInstanceOf(FormData);
          expect(config.headers?.Authorization).toBe(`Bearer ${token}`);
          expect(config.headers?.["Content-Type"]).toBe("multipart/form-data");
          return [200, { success: true }];
        });

      await licenseStore.uploadLicense(mockFile);
    });

    it("should throw on server error when uploading license", async () => {
      const mockFile = new File(["license content"], "license.txt", { type: "text/plain" });
      localStorage.setItem("token", "test-token");

      mockAxios.onPost(uploadUrl).reply(500);

      await expect(licenseStore.uploadLicense(mockFile)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when uploading license", async () => {
      const mockFile = new File(["license content"], "license.txt", { type: "text/plain" });
      localStorage.setItem("token", "test-token");

      mockAxios.onPost(uploadUrl).networkError();

      await expect(licenseStore.uploadLicense(mockFile)).rejects.toThrow("Network Error");
    });
  });
});
