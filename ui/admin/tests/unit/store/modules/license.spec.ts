import { describe, expect, it } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useLicenseStore from "@admin/store/modules/license";

const mockLicense = {
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

describe("License Pinia Store", () => {
  setActivePinia(createPinia());
  const licenseStore = useLicenseStore();

  it("returns default license state", () => {
    expect(licenseStore.license).toEqual({});
    expect(licenseStore.isExpired).toBe(true);
  });

  it("updates license state correctly", () => {
    licenseStore.license = mockLicense;

    expect(licenseStore.license).toEqual(mockLicense);
    expect(licenseStore.isExpired).toBe(false);
  });
});
