import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import DataTable, { type Column } from "../DataTable";

// ---------------------------------------------------------------------------
// Test type and fixtures
// ---------------------------------------------------------------------------

type Row = { id: string; name: string };

const COLUMNS: Column<Row>[] = [
  { key: "id", header: "ID", render: (row) => row.id },
  { key: "name", header: "Name", render: (row) => row.name },
];

const ROWS: Row[] = [
  { id: "1", name: "Alice" },
  { id: "2", name: "Bob" },
];

// ---------------------------------------------------------------------------
// Render helper
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

beforeEach(() => {
  vi.clearAllMocks();
});

describe("DataTable", () => {
  // -------------------------------------------------------------------------
  // Column headers
  // -------------------------------------------------------------------------
  describe("column headers", () => {
    it("renders a header cell for each column", () => {
      renderTable();
      expect(
        screen.getByRole("columnheader", { name: "ID" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("columnheader", { name: "Name" }),
      ).toBeInTheDocument();
    });

    it("renders the correct number of header cells", () => {
      renderTable();
      expect(screen.getAllByRole("columnheader")).toHaveLength(COLUMNS.length);
    });
  });

  // -------------------------------------------------------------------------
  // Data rows
  // -------------------------------------------------------------------------
  describe("data rows", () => {
    it("renders a row for each data item", () => {
      renderTable();
      // thead row + 2 data rows
      expect(screen.getAllByRole("row")).toHaveLength(ROWS.length + 1);
    });

    it("renders cell content via column render functions", () => {
      renderTable();
      expect(screen.getByText("Alice")).toBeInTheDocument();
      expect(screen.getByText("Bob")).toBeInTheDocument();
      expect(screen.getByText("1")).toBeInTheDocument();
      expect(screen.getByText("2")).toBeInTheDocument();
    });

    it("renders custom ReactNode output from render functions", () => {
      const columns: Column<Row>[] = [
        {
          key: "name",
          header: "Name",
          render: (row) => <strong data-testid="bold-name">{row.name}</strong>,
        },
      ];
      renderTable({ columns });
      expect(screen.getAllByTestId("bold-name")).toHaveLength(ROWS.length);
    });
  });

  // -------------------------------------------------------------------------
  // Loading state
  // -------------------------------------------------------------------------
  describe("loading state", () => {
    it("shows spinner with role='status' when isLoading and data is empty", () => {
      renderTable({ isLoading: true, data: [] });
      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("shows the default loading message when loadingMessage is omitted", () => {
      renderTable({ isLoading: true, data: [] });
      expect(screen.getByText("Loading...")).toBeInTheDocument();
    });

    it("shows a custom loadingMessage when provided", () => {
      renderTable({
        isLoading: true,
        data: [],
        loadingMessage: "Fetching data\u2026",
      });
      expect(screen.getByText("Fetching data\u2026")).toBeInTheDocument();
    });

    it("does NOT show spinner when isLoading is true but data is non-empty", () => {
      renderTable({ isLoading: true, data: ROWS });
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    it("uses colSpan equal to the number of columns on the loading cell", () => {
      renderTable({ isLoading: true, data: [] });
      const td = screen.getByRole("status").closest("td");
      expect(td).toHaveAttribute("colspan", String(COLUMNS.length));
    });
  });

  // -------------------------------------------------------------------------
  // Empty state
  // -------------------------------------------------------------------------
  describe("empty state", () => {
    it("shows 'No data available' by default when data is empty and not loading", () => {
      renderTable({ data: [] });
      expect(screen.getByText("No data available")).toBeInTheDocument();
    });

    it("shows a custom emptyMessage when provided", () => {
      renderTable({ data: [], emptyMessage: "Nothing here yet." });
      expect(screen.getByText("Nothing here yet.")).toBeInTheDocument();
    });

    it("shows a custom emptyState ReactNode when provided", () => {
      renderTable({
        data: [],
        emptyState: <span data-testid="custom-empty">No rows found</span>,
      });
      expect(screen.getByTestId("custom-empty")).toBeInTheDocument();
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
      const emptyText = screen.getByText("No data available");
      const td = emptyText.closest("td");
      expect(td).toHaveAttribute("colspan", String(COLUMNS.length));
    });

    it("does not render data rows when data is empty", () => {
      renderTable({ data: [] });
      // thead row + one empty-state row = 2 total rows
      expect(screen.getAllByRole("row")).toHaveLength(2);
    });
  });

  // -------------------------------------------------------------------------
  // Row click
  // -------------------------------------------------------------------------
  describe("row click", () => {
    it("calls onRowClick with the correct row when clicked", async () => {
      const onRowClick = vi.fn();
      renderTable({ onRowClick });
      await userEvent.click(screen.getByText("Alice").closest("tr")!);
      expect(onRowClick).toHaveBeenCalledWith(ROWS[0]);
    });

    it("calls onRowClick for second row with correct data", async () => {
      const onRowClick = vi.fn();
      renderTable({ onRowClick });
      await userEvent.click(screen.getByText("Bob").closest("tr")!);
      expect(onRowClick).toHaveBeenCalledWith(ROWS[1]);
    });

    it("adds cursor-pointer class to rows when onRowClick is provided", () => {
      const onRowClick = vi.fn();
      renderTable({ onRowClick });
      const dataRows = screen.getAllByRole("row").slice(1);
      dataRows.forEach((row) => {
        expect(row.className).toContain("cursor-pointer");
      });
    });

    it("sets tabIndex=0 on rows when onRowClick is provided", () => {
      renderTable({ onRowClick: vi.fn() });
      const dataRows = screen.getAllByRole("row").slice(1);
      dataRows.forEach((row) => {
        expect(row).toHaveAttribute("tabindex", "0");
      });
    });
  });

  // -------------------------------------------------------------------------
  // No row click
  // -------------------------------------------------------------------------
  describe("no row click handler", () => {
    it("does not add cursor-pointer class when onRowClick is undefined", () => {
      renderTable();
      const dataRows = screen.getAllByRole("row").slice(1);
      dataRows.forEach((row) => {
        expect(row.className).not.toContain("cursor-pointer");
      });
    });

    it("does not set tabIndex when onRowClick is undefined", () => {
      renderTable();
      const dataRows = screen.getAllByRole("row").slice(1);
      dataRows.forEach((row) => {
        expect(row).not.toHaveAttribute("tabindex");
      });
    });
  });

  // -------------------------------------------------------------------------
  // Keyboard navigation
  // -------------------------------------------------------------------------
  describe("keyboard navigation on clickable rows", () => {
    it("calls onRowClick when Enter is pressed on a row", () => {
      const onRowClick = vi.fn();
      renderTable({ onRowClick });
      const firstDataRow = screen.getAllByRole("row")[1];
      fireEvent.keyDown(firstDataRow, { key: "Enter" });
      expect(onRowClick).toHaveBeenCalledWith(ROWS[0]);
    });

    it("calls onRowClick when Space is pressed on a row", () => {
      const onRowClick = vi.fn();
      renderTable({ onRowClick });
      const firstDataRow = screen.getAllByRole("row")[1];
      fireEvent.keyDown(firstDataRow, { key: " " });
      expect(onRowClick).toHaveBeenCalledWith(ROWS[0]);
    });

    it("does NOT call onRowClick for other keys", () => {
      const onRowClick = vi.fn();
      renderTable({ onRowClick });
      const firstDataRow = screen.getAllByRole("row")[1];
      fireEvent.keyDown(firstDataRow, { key: "Tab" });
      expect(onRowClick).not.toHaveBeenCalled();
    });

    it("does not attach a keyDown handler when onRowClick is undefined", () => {
      renderTable();
      const firstDataRow = screen.getAllByRole("row")[1];
      expect(() =>
        fireEvent.keyDown(firstDataRow, { key: "Enter" }),
      ).not.toThrow();
    });
  });

  // -------------------------------------------------------------------------
  // Focus-visible classes
  // -------------------------------------------------------------------------
  describe("focus-visible outline classes", () => {
    it("applies focus-visible outline classes to clickable rows", () => {
      renderTable({ onRowClick: vi.fn() });
      const dataRows = screen.getAllByRole("row").slice(1);
      dataRows.forEach((row) => {
        expect(row.className).toContain("focus-visible:outline");
        expect(row.className).toContain("focus-visible:outline-primary/50");
      });
    });

    it("does not apply focus-visible outline classes when not clickable", () => {
      renderTable();
      const dataRows = screen.getAllByRole("row").slice(1);
      dataRows.forEach((row) => {
        expect(row.className).not.toContain("focus-visible:outline");
      });
    });
  });

  // -------------------------------------------------------------------------
  // rowClassName
  // -------------------------------------------------------------------------
  describe("rowClassName", () => {
    it("applies the class returned by rowClassName to each row", () => {
      renderTable({
        rowClassName: (row) => (row.id === "1" ? "row-highlight" : undefined),
      });
      const dataRows = screen.getAllByRole("row").slice(1);
      expect(dataRows[0].className).toContain("row-highlight");
      expect(dataRows[1].className).not.toContain("row-highlight");
    });

    it("does not break when rowClassName returns undefined for all rows", () => {
      expect(() =>
        renderTable({ rowClassName: () => undefined }),
      ).not.toThrow();
    });

    it("combines rowClassName output with base row classes", () => {
      renderTable({ rowClassName: () => "extra-class" });
      const dataRows = screen.getAllByRole("row").slice(1);
      dataRows.forEach((row) => {
        expect(row.className).toContain("extra-class");
        expect(row.className).toContain("transition-colors");
      });
    });
  });

  // -------------------------------------------------------------------------
  // Pagination
  // -------------------------------------------------------------------------
  describe("pagination", () => {
    it("does not render pagination when page, totalPages and onPageChange are all undefined", () => {
      renderTable();
      expect(screen.queryByText("Prev")).not.toBeInTheDocument();
      expect(screen.queryByText("Next")).not.toBeInTheDocument();
    });

    it("does not render pagination when only page is provided", () => {
      renderTable({ page: 1, totalPages: 3 });
      expect(screen.queryByText("Prev")).not.toBeInTheDocument();
    });

    it("does not render pagination when only onPageChange is provided", () => {
      renderTable({ onPageChange: vi.fn() });
      expect(screen.queryByText("Prev")).not.toBeInTheDocument();
    });

    it("does not render pagination when totalPages is missing", () => {
      renderTable({ page: 1, onPageChange: vi.fn() });
      expect(screen.queryByText("Prev")).not.toBeInTheDocument();
    });

    it("does not render pagination when totalPages <= 1", () => {
      renderTable({ page: 1, totalPages: 1, onPageChange: vi.fn() });
      expect(screen.queryByText("Prev")).not.toBeInTheDocument();
    });

    it("renders Prev and Next buttons when there are multiple pages", () => {
      renderTable({
        page: 1,
        totalPages: 3,
        totalCount: 25,
        onPageChange: vi.fn(),
      });
      expect(screen.getByText("Prev")).toBeInTheDocument();
      expect(screen.getByText("Next")).toBeInTheDocument();
    });

    it("calls onPageChange with next page number when Next is clicked", async () => {
      const onPageChange = vi.fn();
      renderTable({ page: 1, totalPages: 3, totalCount: 25, onPageChange });
      await userEvent.click(screen.getByText("Next"));
      expect(onPageChange).toHaveBeenCalledWith(2);
    });

    it("calls onPageChange with previous page number when Prev is clicked", async () => {
      const onPageChange = vi.fn();
      renderTable({ page: 2, totalPages: 3, totalCount: 25, onPageChange });
      await userEvent.click(screen.getByText("Prev"));
      expect(onPageChange).toHaveBeenCalledWith(1);
    });

    it("disables Prev button on the first page", () => {
      renderTable({
        page: 1,
        totalPages: 3,
        totalCount: 25,
        onPageChange: vi.fn(),
      });
      expect(screen.getByText("Prev")).toBeDisabled();
    });

    it("disables Next button on the last page", () => {
      renderTable({
        page: 3,
        totalPages: 3,
        totalCount: 25,
        onPageChange: vi.fn(),
      });
      expect(screen.getByText("Next")).toBeDisabled();
    });
  });

  // -------------------------------------------------------------------------
  // Sorting
  // -------------------------------------------------------------------------
  describe("sorting", () => {
    const sortableColumns: Column<Row>[] = [
      { key: "id", header: "ID", sortable: true, render: (row) => row.id },
      {
        key: "name",
        header: "Name",
        sortable: false,
        render: (row) => row.name,
      },
    ];

    it("renders a sort button only for sortable columns", () => {
      renderTable({ columns: sortableColumns, onSort: vi.fn() });
      expect(
        screen.getByRole("button", { name: /sort by id/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /sort by name/i }),
      ).not.toBeInTheDocument();
    });

    it("does not render sort buttons when onSort is not provided", () => {
      renderTable({ columns: sortableColumns });
      expect(
        screen.queryByRole("button", { name: /sort by/i }),
      ).not.toBeInTheDocument();
    });

    it("calls onSort with the column key when sort button is clicked", async () => {
      const onSort = vi.fn();
      renderTable({ columns: sortableColumns, onSort });
      await userEvent.click(
        screen.getByRole("button", { name: /sort by id/i }),
      );
      expect(onSort).toHaveBeenCalledWith("id");
    });

    it("sets aria-sort='ascending' on the active sorted column when order is asc", () => {
      renderTable({
        columns: sortableColumns,
        onSort: vi.fn(),
        sortField: "id",
        sortOrder: "asc",
      });
      expect(screen.getByRole("columnheader", { name: /id/i })).toHaveAttribute(
        "aria-sort",
        "ascending",
      );
    });

    it("sets aria-sort='descending' on the active sorted column when order is desc", () => {
      renderTable({
        columns: sortableColumns,
        onSort: vi.fn(),
        sortField: "id",
        sortOrder: "desc",
      });
      expect(screen.getByRole("columnheader", { name: /id/i })).toHaveAttribute(
        "aria-sort",
        "descending",
      );
    });

    it("sets aria-sort='none' on a sortable column that is not the active sort field", () => {
      const threeColumns: Column<Row>[] = [
        { key: "id", header: "ID", sortable: true, render: (row) => row.id },
        {
          key: "name",
          header: "Name",
          sortable: true,
          render: (row) => row.name,
        },
      ];
      renderTable({
        columns: threeColumns,
        onSort: vi.fn(),
        sortField: "name",
        sortOrder: "asc",
      });
      const idSortBtn = screen.getByRole("button", { name: /sort by id/i });
      expect(idSortBtn.closest("th")).toHaveAttribute("aria-sort", "none");
    });

    it("sets aria-sort='none' when sortField matches but sortOrder is undefined", () => {
      renderTable({
        columns: sortableColumns,
        onSort: vi.fn(),
        sortField: "id",
      });
      expect(screen.getByRole("columnheader", { name: /id/i })).toHaveAttribute(
        "aria-sort",
        "none",
      );
    });

    it("does not set aria-sort on non-sortable columns", () => {
      renderTable({
        columns: sortableColumns,
        onSort: vi.fn(),
        sortField: "id",
        sortOrder: "asc",
      });
      expect(
        screen.getByRole("columnheader", { name: "Name" }),
      ).not.toHaveAttribute("aria-sort");
    });

    it("sort button has aria-label matching Sort by <header>", () => {
      renderTable({ columns: sortableColumns, onSort: vi.fn() });
      const btn = screen.getByRole("button", { name: /sort by id/i });
      expect(btn).toHaveAttribute("aria-label", "Sort by ID");
    });
  });

  // -------------------------------------------------------------------------
  // noWrapper prop
  // -------------------------------------------------------------------------
  describe("noWrapper prop", () => {
    it("renders the default card wrapper when noWrapper is omitted", () => {
      const { container } = renderTable();
      const wrapper = container.firstChild as HTMLElement;
      expect(wrapper.className).toContain("bg-card");
      expect(wrapper.className).toContain("border");
      expect(wrapper.className).toContain("rounded-xl");
    });

    it("skips the default card wrapper when noWrapper is true", () => {
      const { container } = renderTable({ noWrapper: true });
      const firstChild = container.firstChild as HTMLElement;
      // With noWrapper, the first child is the overflow-x-auto scroll container
      // that holds the <table>, not a styled card wrapper.
      expect(firstChild.className).not.toContain("bg-card");
      expect(firstChild.className).not.toContain("rounded-xl");
      expect(firstChild.className).toContain("overflow-x-auto");
    });
  });

  // -------------------------------------------------------------------------
  // label prop (accessible name for the table)
  // -------------------------------------------------------------------------
  describe("label prop", () => {
    it("sets an aria-label on the table when label is provided", () => {
      renderTable({ label: "Users" });
      expect(
        screen.getByRole("table", { name: "Users" }),
      ).toBeInTheDocument();
    });

    it("does not set an aria-label when label is omitted", () => {
      renderTable();
      expect(screen.getByRole("table")).not.toHaveAttribute("aria-label");
    });
  });

  // -------------------------------------------------------------------------
  // Table structure
  // -------------------------------------------------------------------------
  describe("table structure", () => {
    it("renders a <table> element", () => {
      renderTable();
      expect(screen.getByRole("table")).toBeInTheDocument();
    });

    it("renders a thead and tbody as row groups", () => {
      renderTable();
      expect(screen.getAllByRole("rowgroup")).toHaveLength(2);
    });
  });
});
