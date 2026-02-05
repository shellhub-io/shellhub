import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import DataTable from "@/components/Tables/DataTable.vue";
import * as useTablePreferenceModule from "@/composables/useTablePreference";

describe("DataTable", () => {
  let wrapper: VueWrapper<InstanceType<typeof DataTable>>;

  const defaultHeaders = [
    { text: "Name", value: "name", sortable: true },
    { text: "Owner", value: "owner", sortable: false },
    { text: "Created", value: "createdAt", sortable: true },
  ];

  const defaultItems = [
    { id: 1, name: "Item 1" },
    { id: 2, name: "Item 2" },
  ];

  const defaultRowsSlot = `
    <tr data-test="row">
      <td>Alpha</td>
      <td>Test</td>
      <td>2025-10-21</td>
    </tr>
  `;

  const createTablePreferenceSpy = (getItemsPerPageValue = 10) => {
    const getItemsPerPage = vi.fn(() => getItemsPerPageValue);
    const setItemsPerPage = vi.fn();

    vi.spyOn(useTablePreferenceModule, "useTablePreference").mockReturnValue({
      getItemsPerPage,
      setItemsPerPage,
    });

    return { getItemsPerPage, setItemsPerPage };
  };

  const mountWrapper = ({
    headers = defaultHeaders,
    items = defaultItems,
    totalCount = 42,
    loading = false,
    itemsPerPageOptions = [5, 10, 25],
    page = 2,
    itemsPerPage = 10,
    tableName = undefined as useTablePreferenceModule.TableName | undefined,
    rowsSlot = defaultRowsSlot,
  } = {}) => {
    wrapper = mountComponent(DataTable, {
      props: {
        headers,
        items,
        totalCount,
        loading,
        itemsPerPageOptions,
        page,
        itemsPerPage,
        tableName,
      },
      slots: {
        rows: rowsSlot,
      },
    });
  };

  beforeEach(() => {
    createTablePreferenceSpy();
    mountWrapper();
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Table rendering", () => {
    it("Renders v-table component", () => {
      const table = wrapper.find('[data-test="data-table"]');
      expect(table.exists()).toBe(true);
    });

    it("Renders all table headers", () => {
      const headers = wrapper.findAll('[data-test$="-header"]');
      expect(headers).toHaveLength(3);
      expect(headers[0].text()).toContain("Name");
      expect(headers[1].text()).toContain("Owner");
      expect(headers[2].text()).toContain("Created");
    });

    it("Shows rows when items are present", () => {
      const tbody = wrapper.find('[data-test="tbody-has-items"]');
      expect(tbody.exists()).toBe(true);

      const row = wrapper.find('[data-test="row"]');
      expect(row.exists()).toBe(true);
      expect(row.text()).toContain("Alpha");
    });

    it("Shows empty state when there are no items", () => {
      wrapper.unmount();
      mountWrapper({ items: [] });

      const emptyTbody = wrapper.find('[data-test="tbody-empty"]');
      expect(emptyTbody.exists()).toBe(true);

      const emptyState = wrapper.find('[data-test="empty-state"]');
      expect(emptyState.exists()).toBe(true);
      expect(emptyState.text()).toBe("No data available");
    });
  });

  describe("Sortable headers", () => {
    it("Renders sortable header as clickable", () => {
      const sortableHeader = wrapper.find('[data-test="sort-name"]');
      expect(sortableHeader.exists()).toBe(true);
      expect(sortableHeader.classes()).toContain("cursor-pointer");
    });

    it("Renders non-sortable header as plain text", () => {
      const nonSortableHeader = wrapper.find('[data-test="owner-header"]');
      const label = nonSortableHeader.find('[data-test="header-label"]');
      expect(label.exists()).toBe(true);
      expect(label.text()).toBe("Owner");
    });

    it("Emits update:sort when clicking sortable header", async () => {
      const sortableHeader = wrapper.find('[data-test="sort-name"]');
      await sortableHeader.trigger("click");

      expect(wrapper.emitted("update:sort")).toBeTruthy();
      expect(wrapper.emitted("update:sort")?.[0]).toEqual(["name"]);
    });

    it("Emits update:sort when pressing Enter on sortable header", async () => {
      const sortableHeader = wrapper.find('[data-test="sort-createdAt"]');
      await sortableHeader.trigger("keypress.enter");

      expect(wrapper.emitted("update:sort")).toBeTruthy();
      expect(wrapper.emitted("update:sort")?.[0]).toEqual(["createdAt"]);
    });
  });

  describe("Loading state", () => {
    it("Shows loading progress when loading is true", () => {
      wrapper.unmount();
      mountWrapper({ loading: true });

      const loadingBar = wrapper.find('[data-test="loading"]');
      expect(loadingBar.exists()).toBe(true);
    });

    it("Hides loading progress when loading is false", () => {
      const loadingBar = wrapper.find('[data-test="loading"]');
      expect(loadingBar.exists()).toBe(false);
    });
  });

  describe("Pagination", () => {
    it("Renders pagination when itemsPerPageOptions are provided", () => {
      const pager = wrapper.find('[data-test="pager"]');
      expect(pager.exists()).toBe(true);
    });

    it("Does not render pagination when itemsPerPageOptions is empty", () => {
      wrapper.unmount();
      mountWrapper({ itemsPerPageOptions: [] });

      const pager = wrapper.find('[data-test="pager"]');
      expect(pager.exists()).toBe(false);
    });

    it("Shows correct page text", () => {
      const pagerText = wrapper.find('[data-test="pager-text"]');
      expect(pagerText.text()).toBe("2 of 5");
    });

    it("Calculates page quantity correctly", () => {
      wrapper.unmount();
      mountWrapper({ totalCount: 100, itemsPerPage: 10 });

      const pagerText = wrapper.find('[data-test="pager-text"]');
      expect(pagerText.text()).toBe("2 of 10");
    });

    it("Shows at least 1 page when totalCount is 0", () => {
      wrapper.unmount();
      mountWrapper({ totalCount: 0, itemsPerPage: 10, page: 1 });

      const pagerText = wrapper.find('[data-test="pager-text"]');
      expect(pagerText.text()).toBe("1 of 1");
    });
  });

  describe("Page navigation", () => {
    it("Decrements page when clicking previous button", async () => {
      const prevBtn = wrapper.find('[data-test="pager-prev"]');
      await prevBtn.trigger("click");

      expect(wrapper.emitted("update:page")).toBeTruthy();
      expect(wrapper.emitted("update:page")?.[0]).toEqual([1]);
    });

    it("Increments page when clicking next button", async () => {
      const nextBtn = wrapper.find('[data-test="pager-next"]');
      await nextBtn.trigger("click");

      expect(wrapper.emitted("update:page")).toBeTruthy();
      expect(wrapper.emitted("update:page")?.[0]).toEqual([3]);
    });

    it("Disables previous button on first page", () => {
      wrapper.unmount();
      mountWrapper({ page: 1 });

      const prevBtn = wrapper.find('[data-test="pager-prev"]');
      expect(prevBtn.attributes("disabled")).toBeDefined();
    });

    it("Disables next button on last page", () => {
      wrapper.unmount();
      mountWrapper({ page: 5, totalCount: 42, itemsPerPage: 10 });

      const nextBtn = wrapper.find('[data-test="pager-next"]');
      expect(nextBtn.attributes("disabled")).toBeDefined();
    });

    it("Disables next button when only one page", () => {
      wrapper.unmount();
      mountWrapper({ page: 1, totalCount: 5, itemsPerPage: 10 });

      const nextBtn = wrapper.find('[data-test="pager-next"]');
      expect(nextBtn.attributes("disabled")).toBeDefined();
    });
  });

  describe("Items per page", () => {
    it("Shows items per page label", () => {
      const label = wrapper.find('[data-test="ipp-label"]');
      expect(label.text()).toBe("Items per page:");
    });

    it("Renders items per page combobox", () => {
      const combo = wrapper.find('[data-test="ipp-combo"]');
      expect(combo.exists()).toBe(true);
    });

    it("Updates itemsPerPage when selecting valid value", async () => {
      const combo = wrapper.findComponent({ name: "v-combobox" });
      await combo.setValue(25);

      expect(wrapper.emitted("update:itemsPerPage")).toBeTruthy();
      expect(wrapper.emitted("update:itemsPerPage")?.[0]).toEqual([25]);
    });

    it("Resets page to 1 when changing items per page", async () => {
      const combo = wrapper.findComponent({ name: "v-combobox" });
      await combo.setValue(25);

      expect(wrapper.emitted("update:page")).toBeTruthy();
      expect(wrapper.emitted("update:page")?.[0]).toEqual([1]);
    });

    it("Shows error when value is less than 1", async () => {
      const combo = wrapper.findComponent({ name: "v-combobox" });
      await combo.setValue(0);

      expect(combo.props("errorMessages")).toBe("Minimum is 1");
    });

    it("Shows error when value is greater than 100", async () => {
      const combo = wrapper.findComponent({ name: "v-combobox" });
      await combo.setValue(101);

      expect(combo.props("errorMessages")).toBe("Maximum is 100");
    });

    it("Clamps value to 1 on blur when invalid", async () => {
      const combo = wrapper.find("[data-test=\"ipp-combo\"] input");
      await combo.setValue(0);
      await combo.trigger("blur");

      expect(wrapper.emitted("update:itemsPerPage")).toBeTruthy();
      const events = wrapper.emitted("update:itemsPerPage") as number[][];
      expect(events[events.length - 1]).toEqual([1]);
    });

    it("Clamps value to 100 on blur when too high", async () => {
      const combo = wrapper.find("[data-test=\"ipp-combo\"] input");
      await combo.setValue(200);
      await combo.trigger("blur");

      expect(wrapper.emitted("update:itemsPerPage")).toBeTruthy();
      const events = wrapper.emitted("update:itemsPerPage") as number[][];
      expect(events[events.length - 1]).toEqual([100]);
    });

    it("Clamps value on Enter key press", async () => {
      const combo = wrapper.find("[data-test=\"ipp-combo\"] input");
      await combo.setValue(0);
      await combo.trigger("keydown.enter");

      expect(wrapper.emitted("update:itemsPerPage")).toBeTruthy();
      const events = wrapper.emitted("update:itemsPerPage") as number[][];
      expect(events[events.length - 1]).toEqual([1]);
    });
  });

  describe("Table preference persistence", () => {
    it("Calls getItemsPerPage on mount when tableName is provided", () => {
      wrapper.unmount();
      const { getItemsPerPage } = createTablePreferenceSpy(25);
      mountWrapper({ tableName: "devices" });

      expect(getItemsPerPage).toHaveBeenCalledWith("devices");
    });

    it("Updates itemsPerPage from stored preference on mount", () => {
      wrapper.unmount();
      createTablePreferenceSpy(25);
      mountWrapper({ tableName: "devices", itemsPerPage: 10 });

      expect(wrapper.emitted("update:itemsPerPage")).toBeTruthy();
      expect(wrapper.emitted("update:itemsPerPage")?.[0]).toEqual([25]);
    });

    it("Calls setItemsPerPage when itemsPerPage changes", async () => {
      wrapper.unmount();
      const { setItemsPerPage } = createTablePreferenceSpy();
      mountWrapper({ tableName: "devices", itemsPerPage: 10 });

      const combo = wrapper.findComponent({ name: "v-combobox" });
      await combo.setValue(25);

      expect(setItemsPerPage).toHaveBeenCalledWith("devices", 25);
    });

    it("Does not call setItemsPerPage when tableName is not provided", async () => {
      wrapper.unmount();
      const { setItemsPerPage } = createTablePreferenceSpy();
      mountWrapper({ tableName: undefined });

      const combo = wrapper.findComponent({ name: "v-combobox" });
      await combo.setValue(25);

      expect(setItemsPerPage).not.toHaveBeenCalled();
    });
  });
});
