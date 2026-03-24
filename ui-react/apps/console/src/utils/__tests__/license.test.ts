import { describe, it, expect } from "vitest";
import {
  formatLicenseTimestamp,
  formatDeviceCount,
  formatRegions,
  getDisplayFeatures,
  validateLicenseFile,
  getLicenseAlertConfig,
} from "../license";

describe("formatLicenseTimestamp", () => {
  it('returns "Now" for -1', () => {
    expect(formatLicenseTimestamp(-1)).toBe("Now");
  });

  it("returns a formatted date string for a valid Unix timestamp", () => {
    // 2024-01-15T12:00:00Z — mid-day to avoid timezone edge cases
    expect(formatLicenseTimestamp(1705320000)).toBe("Jan 15, 2024");
  });

  it("handles timestamps at the start of the epoch", () => {
    // Mid-day to avoid timezone edge cases
    expect(formatLicenseTimestamp(43200)).toBe("Jan 1, 1970");
  });
});

describe("formatDeviceCount", () => {
  it('returns "Unlimited" for -1', () => {
    expect(formatDeviceCount(-1)).toBe("Unlimited");
  });

  it("returns the number as a string for positive values", () => {
    expect(formatDeviceCount(50)).toBe("50");
  });

  it("returns '0' for zero", () => {
    expect(formatDeviceCount(0)).toBe("0");
  });
});

describe("formatRegions", () => {
  it('returns "Global" for an empty array', () => {
    expect(formatRegions([])).toBe("Global");
  });

  it("returns a single region as-is", () => {
    expect(formatRegions(["us-east-1"])).toBe("us-east-1");
  });

  it("returns comma-separated regions for multiple entries", () => {
    expect(formatRegions(["us-east-1", "eu-west-1", "ap-southeast-1"])).toBe(
      "us-east-1, eu-west-1, ap-southeast-1",
    );
  });
});

describe("getDisplayFeatures", () => {
  const features = {
    devices: 100,
    session_recording: true,
    firewall_rules: false,
    billing: true,
    login_link: true,
    reports: true,
  };

  it("returns exactly 4 display features", () => {
    expect(getDisplayFeatures(features)).toHaveLength(4);
  });

  it("includes devices, session_recording, firewall_rules, and billing", () => {
    const names = getDisplayFeatures(features).map((f) => f.name);
    expect(names).toEqual(["devices", "session_recording", "firewall_rules", "billing"]);
  });

  it("excludes login_link and reports", () => {
    const names = getDisplayFeatures(features).map((f) => f.name);
    expect(names).not.toContain("login_link");
    expect(names).not.toContain("reports");
  });

  it("maps values from the features object", () => {
    const result = getDisplayFeatures(features);
    expect(result.find((f) => f.name === "devices")?.value).toBe(100);
    expect(result.find((f) => f.name === "session_recording")?.value).toBe(true);
    expect(result.find((f) => f.name === "firewall_rules")?.value).toBe(false);
    expect(result.find((f) => f.name === "billing")?.value).toBe(true);
  });

  it("sets type to 'number' for devices and 'boolean' for the rest", () => {
    const result = getDisplayFeatures(features);
    expect(result.find((f) => f.name === "devices")?.type).toBe("number");
    expect(result.find((f) => f.name === "session_recording")?.type).toBe("boolean");
    expect(result.find((f) => f.name === "firewall_rules")?.type).toBe("boolean");
    expect(result.find((f) => f.name === "billing")?.type).toBe("boolean");
  });
});

describe("validateLicenseFile", () => {
  it("returns null for a valid .dat file under 32 KB", () => {
    const file = new File(["content"], "license.dat");
    expect(validateLicenseFile(file)).toBeNull();
  });

  it("returns an error message for a non-.dat extension", () => {
    const file = new File(["content"], "license.txt");
    expect(validateLicenseFile(file)).toBe("Only .dat files are allowed");
  });

  it("accepts uppercase .DAT extension", () => {
    const file = new File(["content"], "LICENSE.DAT");
    expect(validateLicenseFile(file)).toBeNull();
  });

  it("returns an error message for a file without an extension", () => {
    const file = new File(["content"], "license");
    expect(validateLicenseFile(file)).toBe("Only .dat files are allowed");
  });

  it("returns an error message for a file exactly 32 KB in size", () => {
    const content = "x".repeat(32 * 1024);
    const file = new File([content], "license.dat");
    expect(validateLicenseFile(file)).toBe("File must be smaller than 32 KB");
  });

  it("returns an error message for a file larger than 32 KB", () => {
    const content = "x".repeat(32 * 1024 + 1);
    const file = new File([content], "license.dat");
    expect(validateLicenseFile(file)).toBe("File must be smaller than 32 KB");
  });

  it("accepts a .dat file that is 1 byte under the 32 KB limit", () => {
    const content = "x".repeat(32 * 1024 - 1);
    const file = new File([content], "license.dat");
    expect(validateLicenseFile(file)).toBeNull();
  });
});

describe("getLicenseAlertConfig", () => {
  it("returns an info alert when license is null", () => {
    expect(getLicenseAlertConfig(null)).toEqual({
      variant: "info",
      message: "You do not have an installed license",
    });
  });

  it("returns an info alert when license is about to expire", () => {
    expect(
      getLicenseAlertConfig({ expired: false, about_to_expire: true, grace_period: false }),
    ).toEqual({ variant: "info", message: "Your license is about to expire!" });
  });

  it("returns a warning alert when license is expired but within grace period", () => {
    expect(
      getLicenseAlertConfig({ expired: true, about_to_expire: false, grace_period: true }),
    ).toEqual({
      variant: "warning",
      message: "Your license has expired, but you are still within the grace period.",
    });
  });

  it("returns an error alert when license is expired and outside grace period", () => {
    expect(
      getLicenseAlertConfig({ expired: true, about_to_expire: false, grace_period: false }),
    ).toEqual({ variant: "error", message: "Your license has expired!" });
  });

  it("returns null for a valid, non-expiring license", () => {
    expect(
      getLicenseAlertConfig({ expired: false, about_to_expire: false, grace_period: false }),
    ).toBeNull();
  });

  it("prioritizes expired over about_to_expire when both flags are true", () => {
    const result = getLicenseAlertConfig({
      expired: true,
      about_to_expire: true,
      grace_period: false,
    });
    expect(result?.variant).toBe("error");
    expect(result?.message).toBe("Your license has expired!");
  });
});
