import { createVuetify } from "vuetify";
import { mount, VueWrapper, flushPromises } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import DataTable from "@/components/Tables/DataTable.vue";

type DataTableWrapper = VueWrapper<InstanceType<typeof DataTable>>;

describe("DataTable", () => {
  const vuetify = createVuetify();
  let wrapper: DataTableWrapper;

  const defaultHeaders = [
    { text: "Name", value: "name", sortable: true },
    { text: "Owner", value: "owner", sortable: false },
    { text: "Created", value: "createdAt", sortable: true },
  ] as const;

  const mountWrapper = (
    props: Partial<InstanceType<typeof DataTable>["$props"]> = {},
    slots: Record<string, string> = {},
  ) => mount(DataTable, {
    global: { plugins: [vuetify] },
    attachTo: document.body,
    props: {
      headers: defaultHeaders as unknown as Array<{ text: string; value: string; sortable: boolean }>,
      items: [{ id: 1 }],
      totalCount: 42,
      loading: false,
      itemsPerPageOptions: [5, 10, 25],
      page: 2,
      itemsPerPage: 10,
      ...props,
    },
    slots: {
      rows: `
          <tr data-test="row">
            <td>Alpha</td>
            <td>Test</td>
            <td>2025-10-21</td>
          </tr>
        `,
      ...slots,
    },
  });

  const uiTick = async () => {
    await flushPromises();
    await Promise.resolve();
  };

  beforeEach(async () => {
    document.body.innerHTML = "";
    wrapper = mountWrapper();
    await uiTick();
  });

  afterEach(() => {
    vi.clearAllMocks();
    wrapper.unmount();
    document.body.innerHTML = "";
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders table headers", () => {
    const ths = wrapper.findAll('th[data-test^="th-"]');
    expect(ths.map((th) => th.text().trim())).toEqual(["Name", "Owner", "Created"]);
  });

  it("renders rows slot when items are present", () => {
    const row = wrapper.find('[data-test="row"]');
    expect(row.exists()).toBe(true);
    expect(row.text()).toContain("Alpha");
  });

  it("shows empty state when there are no items", async () => {
    await wrapper.setProps({ items: [] });
    await uiTick();

    const emptyCell = wrapper.find('[data-test="empty-state"]');
    expect(emptyCell.exists()).toBe(true);
    expect(emptyCell.text()).toContain("No data available");
  });

  it("emits sort when clicking a sortable header", async () => {
    const sortable = wrapper.find('[data-test="sort-name"]');
    expect(sortable.exists()).toBe(true);

    await sortable.trigger("click");
    const sortEvents = wrapper.emitted("update:sort");
    expect(sortEvents && sortEvents[0]).toEqual(["name"]);
  });

  it("emits sort when pressing Enter on a sortable header", async () => {
    const sortable = wrapper.find('[data-test="sort-createdAt"]');
    expect(sortable.exists()).toBe(true);

    await sortable.trigger("keypress", { key: "Enter" });
    const sortEvents = wrapper.emitted("update:sort");
    expect(sortEvents && sortEvents.at(-1)).toEqual(["createdAt"]);
  });

  it("shows linear progress when loading", async () => {
    await wrapper.setProps({ loading: true });
    await uiTick();

    const progress = wrapper.find('[data-test="loading"]');
    expect(progress.exists()).toBe(true);
    expect(progress.attributes("alt")).toBe("Data table loading");
  });

  describe("pagination controls", () => {
    const getPagerRoot = () => wrapper.find('[data-test="pager"]');

    it("renders items-per-page combobox and pager when options are provided", () => {
      const pager = getPagerRoot();
      expect(pager.exists()).toBe(true);
      expect(pager.find('[data-test="pager-text"]').text()).toContain("2 of");
    });

    it("hides pagination controls when itemsPerPageOptions is empty/undefined", async () => {
      await wrapper.setProps({ itemsPerPageOptions: [] as number[] });
      await uiTick();

      expect(getPagerRoot().exists()).toBe(false);
    });

    it("resets to first page when itemsPerPage changes via combobox", async () => {
      await wrapper.setProps({ page: 3 });
      await uiTick();

      const combo = wrapper.findComponent({ name: "VCombobox" });
      expect(combo.exists()).toBe(true);

      combo.vm.$emit("update:modelValue", 25);
      await uiTick();

      const pageEvents = wrapper.emitted("update:page");
      expect(pageEvents).toBeTruthy();
      expect(pageEvents && pageEvents.some((e) => e[0] === 1)).toBe(true);

      const ippEvents = wrapper.emitted("update:itemsPerPage");
      expect(ippEvents).toBeTruthy();
      expect(ippEvents && ippEvents.at(-1)).toEqual([25]);
    });

    it("increments/decrements page via chevrons and respects disabled states", async () => {
      const left = wrapper.find('[data-test="pager-prev"]');
      const right = wrapper.find('[data-test="pager-next"]');

      expect(left.attributes("disabled")).toBeUndefined();

      await left.trigger("click");
      await uiTick();
      let pageEvents = wrapper.emitted("update:page");
      expect(pageEvents).toBeTruthy();
      expect(pageEvents && pageEvents.at(-1)).toEqual([1]);

      await wrapper.setProps({ page: 5 });
      await uiTick();

      expect(right.attributes("disabled")).toBeDefined();

      await wrapper.setProps({ page: 4 });
      await uiTick();

      expect(right.attributes("disabled")).toBeUndefined();

      await right.trigger("click");
      await uiTick();
      pageEvents = wrapper.emitted("update:page");
      expect(pageEvents && pageEvents.at(-1)).toEqual([5]);
    });

    it("shows 'page of pageQuantity' text based on totalCount and itemsPerPage", async () => {
      let text = wrapper.find('[data-test="pager-text"]').text();
      expect(text).toContain("2 of 5");

      await wrapper.setProps({ totalCount: 7, itemsPerPage: 5, page: 1 });
      await uiTick();

      text = wrapper.find('[data-test="pager-text"]').text();
      expect(text).toContain("1 of 2");
    });
  });
});
