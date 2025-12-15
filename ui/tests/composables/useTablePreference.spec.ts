import { describe, it, expect, beforeEach, vi } from "vitest";
import { useTablePreference } from "@/composables/useTablePreference";

describe("useTablePreference", () => {
  const STORAGE_KEY = "tablePreferences";

  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  describe("getItemsPerPage", () => {
    it("should return default value 10 when localStorage is empty", () => {
      const { getItemsPerPage } = useTablePreference();

      expect(getItemsPerPage("sessions")).toBe(10);
    });

    it("should return default value 10 for missing table key", () => {
      const { getItemsPerPage } = useTablePreference();
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ devices: 20 }));

      expect(getItemsPerPage("sessions")).toBe(10);
    });

    it("should return preferencesObject value for existing table", () => {
      const { getItemsPerPage } = useTablePreference();
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ sessions: 50 }));

      expect(getItemsPerPage("sessions")).toBe(50);
    });

    it("should handle JSON parse errors gracefully", () => {
      const { getItemsPerPage } = useTablePreference();
      localStorage.setItem(STORAGE_KEY, "invalid json{");

      expect(getItemsPerPage("sessions")).toBe(10);
    });

    it("should handle localStorage access errors gracefully", () => {
      const { getItemsPerPage } = useTablePreference();
      vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
        throw new Error("localStorage is disabled");
      });

      expect(getItemsPerPage("sessions")).toBe(10);
    });
  });

  describe("setItemsPerPage", () => {
    it("should create localStorage entry with single table preference", () => {
      const { setItemsPerPage } = useTablePreference();

      setItemsPerPage("sessions", 20);

      const preferencesObject = JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}");
      expect(preferencesObject).toEqual({ sessions: 20 });
    });

    it("should update existing table preference without affecting others", () => {
      const { setItemsPerPage } = useTablePreference();
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ devices: 30, sessions: 10 }));

      setItemsPerPage("sessions", 50);

      const preferencesObject = JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}");
      expect(preferencesObject).toEqual({ devices: 30, sessions: 50 });
    });

    it("should add new table preference to existing storage", () => {
      const { setItemsPerPage } = useTablePreference();
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ sessions: 20 }));

      setItemsPerPage("devices", 100);

      const preferencesObject = JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}");
      expect(preferencesObject).toEqual({ sessions: 20, devices: 100 });
    });

    it("should handle localStorage quota exceeded errors silently", () => {
      const { setItemsPerPage } = useTablePreference();
      vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
        throw new DOMException("QuotaExceededError");
      });

      expect(() => setItemsPerPage("sessions", 20)).not.toThrow();
    });

    it("should handle JSON parse errors when reading existing data", () => {
      const { setItemsPerPage } = useTablePreference();
      localStorage.setItem(STORAGE_KEY, "invalid json{");

      // Should not throw when encountering invalid JSON
      expect(() => setItemsPerPage("sessions", 20)).not.toThrow();
    });
  });

  describe("round-trip persistence", () => {
    it("should successfully save and retrieve value", () => {
      const { getItemsPerPage, setItemsPerPage } = useTablePreference();

      setItemsPerPage("sessions", 75);
      expect(getItemsPerPage("sessions")).toBe(75);
    });

    it("should maintain independent values for multiple tables", () => {
      const { getItemsPerPage, setItemsPerPage } = useTablePreference();

      setItemsPerPage("sessions", 20);
      setItemsPerPage("devices", 50);
      setItemsPerPage("firewallRules", 100);

      expect(getItemsPerPage("sessions")).toBe(20);
      expect(getItemsPerPage("devices")).toBe(50);
      expect(getItemsPerPage("firewallRules")).toBe(100);
    });

    it("should handle all defined table names", () => {
      const { getItemsPerPage, setItemsPerPage } = useTablePreference();
      const tableNames = [
        "sessions",
        "devices",
        "containers",
        "firewallRules",
        "publicKeys",
        "apiKeys",
        "invitations",
        "tags",
        "connectors",
        "webEndpoints",
        "adminSessions",
        "adminDevices",
        "adminNamespaces",
        "adminUsers",
        "adminFirewallRules",
        "adminAnnouncements",
      ] as const;

      tableNames.forEach((tableName, index) => {
        const value = (index + 1) * 5;
        setItemsPerPage(tableName, value);
      });

      tableNames.forEach((tableName, index) => {
        const expectedValue = (index + 1) * 5;
        expect(getItemsPerPage(tableName)).toBe(expectedValue);
      });
    });
  });
});
