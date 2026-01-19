import { IAdminLicense } from "@admin/interfaces/ILicense";

export const mockLicense: IAdminLicense = {
  id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  expired: false,
  about_to_expire: false,
  grace_period: false,
  issued_at: 1704067200, // Jan 1, 2024
  starts_at: 1704067200,
  expires_at: 1735689600, // Jan 1, 2025
  allowed_regions: [],
  customer: {
    id: "customer-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    name: "Test Customer",
    email: "test@example.com",
    company: "Test Company Inc.",
  },
  features: {
    devices: -1,
    session_recording: true,
    firewall_rules: true,
    billing: false,
    reports: false,
    login_link: false,
  },
};

export const mockLicenseExpired: IAdminLicense = {
  ...mockLicense,
  expired: true,
  grace_period: false,
  expires_at: 1672531200, // Jan 1, 2023 (past date)
};

export const mockLicenseAboutToExpire: IAdminLicense = {
  ...mockLicense,
  about_to_expire: true,
};

export const mockLicenseGracePeriod: IAdminLicense = {
  ...mockLicense,
  expired: true,
  grace_period: true,
};

export const mockLicenseRegional: IAdminLicense = {
  ...mockLicense,
  allowed_regions: ["US", "EU"] as IAdminLicense["allowed_regions"],
};

export const mockNoLicense: Partial<IAdminLicense> = {};
