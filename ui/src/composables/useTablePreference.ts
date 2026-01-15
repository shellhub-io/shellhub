import handleError from "@/utils/handleError";

// eslint-disable-next-line vue/max-len
export type TableName = "sessions" | "devices" | "containers" | "firewallRules" | "publicKeys" | "apiKeys" | "invitations" | "tags" | "connectors" | "webEndpoints" | "adminSessions" | "adminDevices" | "adminNamespaces" | "adminUsers" | "adminFirewallRules" | "adminAnnouncements";

const STORAGE_KEY = "tablePreferences";
const DEFAULT_ITEMS_PER_PAGE = 10; // Default fallback value

// Custom default items per page for specific tables
const CUSTOM_DEFAULT_ITEMS_PER_PAGE_MAP: Partial<Record<TableName, number>> = {
  tags: 100,
};

export function useTablePreference() {
  const getItemsPerPage = (tableName: TableName): number => {
    const defaultItemsPerPage = CUSTOM_DEFAULT_ITEMS_PER_PAGE_MAP[tableName] ?? DEFAULT_ITEMS_PER_PAGE;
    try {
      const preferencesItem = localStorage.getItem(STORAGE_KEY);
      if (!preferencesItem) return defaultItemsPerPage;

      const preferencesObject = JSON.parse(preferencesItem) as Record<TableName, number>;
      return preferencesObject[tableName] ?? defaultItemsPerPage;
    } catch {
      return defaultItemsPerPage;
    }
  };

  const setItemsPerPage = (tableName: TableName, value: number): void => {
    try {
      const preferencesItem = localStorage.getItem(STORAGE_KEY);
      const preferencesObject = preferencesItem ? (JSON.parse(preferencesItem) as Record<string, number>) : {};

      preferencesObject[tableName] = value;
      localStorage.setItem(STORAGE_KEY, JSON.stringify(preferencesObject));
    } catch (error) {
      handleError(error);
    }
  };

  return {
    getItemsPerPage,
    setItemsPerPage,
  };
}
