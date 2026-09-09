import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import DataTable, { type Column } from "../DataTable";

type Row = { id: string; name: string };

const COLUMNS: Column<Row>[] = [
  { key: "id", header: "ID", render: (row) => row.id },
  { key: "name", header: "Name", render: (row) => row.name },
];

const SORTABLE_COLUMNS: Column<Row>[] = [
  { key: "id", header: "ID", sortable: true, render: (row) => row.id },
  { key: "name", header: "Name", sortable: false, render: (row) => row.name },
];

const ROWS: Row[] = [
  { id: "1", name: "Alice" },
  { id: "2", name: "Bob" },
];

interface RenderOptions {
  columns?: Column<Row>[];
  data?: Row[];
  page?: number;
  totalPages?: number;
  totalCount?: number;
  itemLabel?: string;
  onPageChange?: (page: number) => void;
  sortField?: string;
  sortOrder?: "asc" | "desc";
  onSort?: (field: string) => void;
  onRowClick?: (row: Row) => void;
  rowClassName?: (row: Row) => string | undefined;
  isLoading?: boolean;
  loadingMessage?: string;
  emptyState?: React.ReactNode;
  emptyMessage?: string;
  noWrapper?: boolean;
  label?: string;
}

function renderTable(options: RenderOptions = {}) {
  const { columns = COLUMNS, data = ROWS, ...rest } = options;

  return render(
    <DataTable<Row>
      columns={columns}
      data={data}
      rowKey={(row) => row.id}
      {...rest}
    />,
  );
}

function dataRows() {
  return screen.getAllByRole("row").slice(1);
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("DataTable", () => {
  it("renders a header per column and a row per data item", () => {
    renderTable();
    expect(screen.getByRole("columnheader", { name: "ID" })).toBeInTheDocument();
    expect(
      screen.getByRole("columnheader", { name: "Name" }),
    ).toBeInTheDocument();
    expect(screen.getAllByRole("row")).toHaveLength(ROWS.length + 1);
    expect(screen.getByText("Alice")).toBeInTheDocument();
    expect(screen.getByText("Bob")).toBeInTheDocument();
  });

  describe("loading state", () => {
    it.each([
      [undefined, "Loading..."],
      ["Fetching data…", "Fetching data…"],
    ])("loadingMessage=%s shows '%s'", (loadingMessage, expected) => {
      renderTable({ isLoading: true, data: [], loadingMessage });
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(expected)).toBeInTheDocument();
    });

    it("does NOT show spinner when isLoading is true but data is non-empty", () => {
      renderTable({ isLoading: true, data: ROWS });
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it.each([
      [undefined, "No data available"],
      ["Nothing here yet.", "Nothing here yet."],
    ])("emptyMessage=%s shows '%s'", (emptyMessage, expected) => {
      renderTable({ data: [], emptyMessage });
      expect(screen.getByText(expected)).toBeInTheDocument();
    });

    it("uses emptyState over emptyMessage when both are provided", () => {
      renderTable({
        data: [],
        emptyMessage: "Should not appear",
        emptyState: <span data-testid="node-wins">Node wins</span>,
      });
      expect(screen.getByTestId("node-wins")).toBeInTheDocument();
      expect(screen.queryByText("Should not appear")).not.toBeInTheDocument();
    });

    it("uses colSpan equal to the number of columns on the empty cell", () => {
      renderTable({ data: [] });
      const td = screen.getByText("No data available").closest("td");
      expect(td).toHaveAttribute("colspan", String(COLUMNS.length));
    });
  });

  describe("row click", () => {
    it("calls onRowClick with the correct row when clicked", async () => {
      const onRowClick = vi.fn();
      renderTable({ onRowClick });
      await userEvent.click(screen.getByText("Alice").closest("tr")!);
      expect(onRowClick).toHaveBeenCalledWith(ROWS[0]);
    });

    it.each([
      ["provided", vi.fn(), "0"],
      ["undefined", undefined, null],
    ])(
      "onRowClick %s sets tabindex to %s",
      (_label, onRowClick, expected) => {
        renderTable({ onRowClick });
        dataRows().forEach((row) => {
          if (expected === null) expect(row).not.toHaveAttribute("tabindex");
          else expect(row).toHaveAttribute("tabindex", expected);
        });
      },
    );

    it.each([
      ["Enter", true],
      [" ", true],
      ["Tab", false],
    ] as const)("pressing %s on a row activates it: %s", (key, activates) => {
      const onRowClick = vi.fn();
      renderTable({ onRowClick });
      fireEvent.keyDown(dataRows()[0], { key });
      if (activates) expect(onRowClick).toHaveBeenCalledWith(ROWS[0]);
      else expect(onRowClick).not.toHaveBeenCalled();
    });
  });

  it("applies the class returned by rowClassName to each row", () => {
    renderTable({
      rowClassName: (row) => (row.id === "1" ? "row-highlight" : undefined),
    });
    expect(dataRows()[0].className).toContain("row-highlight");
    expect(dataRows()[1].className).not.toContain("row-highlight");
  });

  describe("pagination", () => {
    it.each([
      ["page, totalPages and onPageChange are all undefined", {}],
      ["totalPages <= 1", { page: 1, totalPages: 1, onPageChange: vi.fn() }],
    ])("does not render pagination when %s", (_label, options) => {
      renderTable(options);
      expect(screen.queryByText("Prev")).not.toBeInTheDocument();
      expect(screen.queryByText("Next")).not.toBeInTheDocument();
    });

    it.each([
      ["Next", 1, 2],
      ["Prev", 2, 1],
    ])("clicking %s from page %i calls onPageChange with %i", async (
      button,
      page,
      expected,
    ) => {
      const onPageChange = vi.fn();
      renderTable({ page, totalPages: 3, totalCount: 25, onPageChange });
      await userEvent.click(screen.getByText(button));
      expect(onPageChange).toHaveBeenCalledWith(expected);
    });

    it.each([
      [1, "Prev", "Next"],
      [3, "Next", "Prev"],
    ])("on page %i of 3, %s is disabled and %s is not", (
      page,
      disabled,
      enabled,
    ) => {
      renderTable({ page, totalPages: 3, totalCount: 25, onPageChange: vi.fn() });
      expect(screen.getByText(disabled)).toBeDisabled();
      expect(screen.getByText(enabled)).not.toBeDisabled();
    });
  });

  describe("sorting", () => {
    it("renders a sort button only for sortable columns", () => {
      renderTable({ columns: SORTABLE_COLUMNS, onSort: vi.fn() });
      expect(
        screen.getByRole("button", { name: /sort by id/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /sort by name/i }),
      ).not.toBeInTheDocument();
    });

    it("does not render sort buttons when onSort is not provided", () => {
      renderTable({ columns: SORTABLE_COLUMNS });
      expect(
        screen.queryByRole("button", { name: /sort by/i }),
      ).not.toBeInTheDocument();
    });

    it("calls onSort with the column key when sort button is clicked", async () => {
      const onSort = vi.fn();
      renderTable({ columns: SORTABLE_COLUMNS, onSort });
      await userEvent.click(screen.getByRole("button", { name: /sort by id/i }));
      expect(onSort).toHaveBeenCalledWith("id");
    });

    it.each([
      ["asc", "ascending"],
      ["desc", "descending"],
      [undefined, "none"],
    ] as const)("sortOrder=%s sets aria-sort='%s' on the sorted column", (
      sortOrder,
      expected,
    ) => {
      renderTable({
        columns: SORTABLE_COLUMNS,
        onSort: vi.fn(),
        sortField: "id",
        sortOrder,
      });
      expect(screen.getByRole("columnheader", { name: /id/i })).toHaveAttribute(
        "aria-sort",
        expected,
      );
    });

    it("sets aria-sort='none' on a sortable column that is not the active sort field", () => {
      renderTable({
        columns: SORTABLE_COLUMNS.map((column) => ({
          ...column,
          sortable: true,
        })),
        onSort: vi.fn(),
        sortField: "name",
        sortOrder: "asc",
      });
      const idSortBtn = screen.getByRole("button", { name: /sort by id/i });
      expect(idSortBtn.closest("th")).toHaveAttribute("aria-sort", "none");
    });

    it("does not set aria-sort on non-sortable columns", () => {
      renderTable({
        columns: SORTABLE_COLUMNS,
        onSort: vi.fn(),
        sortField: "id",
        sortOrder: "asc",
      });
      expect(
        screen.getByRole("columnheader", { name: "Name" }),
      ).not.toHaveAttribute("aria-sort");
    });
  });

  it.each([
    ["Users", "Users"],
    [undefined, null],
  ])("label=%s sets the table's accessible name to %s", (label, expected) => {
    renderTable({ label });
    if (expected === null) {
      expect(screen.getByRole("table")).not.toHaveAttribute("aria-label");
    } else {
      expect(screen.getByRole("table", { name: expected })).toBeInTheDocument();
    }
  });
});
