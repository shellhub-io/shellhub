import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useLicenseStore from "@admin/store/modules/license";
import { GetLicense200Response } from "@admin/api/client/api";

describe("License Pinia Store", () => {
  let licenseStore: ReturnType<typeof useLicenseStore>;

  const mockLicense: GetLicense200Response = {
    expired: false,
    about_to_expire: false,
    grace_period: false,
    id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    issued_at: -1,
    starts_at: -1,
    expires_at: -1,
    allowed_regions: [],
    customer: {
      id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      name: "ShellHub",
      email: "contato@ossystems.com.br",
      company: "O.S. Systems",
    },
    features: {
      devices: -1,
      session_recording: true,
      firewall_rules: true,
      reports: false,
      login_link: false,
      billing: false,
    },
  };

  beforeEach(() => {
    setActivePinia(createPinia());
    licenseStore = useLicenseStore();
  });

  it("returns default license state", () => {
    expect(licenseStore.getLicense).toEqual({});
    expect(licenseStore.isExpired).toBe(true);
  });

  it("updates license state correctly", () => {
    licenseStore.license = mockLicense;

    expect(licenseStore.getLicense).toEqual(mockLicense);
    expect(licenseStore.isExpired).toBe(false);
  });
});
