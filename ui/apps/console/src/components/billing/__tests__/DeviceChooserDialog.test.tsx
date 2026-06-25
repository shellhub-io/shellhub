import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  render,
  screen,
  fireEvent,
  waitFor,
  cleanup,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

// ── Stub BaseDialog ───────────────────────────────────────────────────────────
// BaseDialog calls showModal() which jsdom does not implement. Replace it with
// a plain div that honours the `open` prop and wires up a close button via the
// native `cancel` event that the real component relies on.
vi.mock("@/components/common/BaseDialog", () => ({
  default: ({
    open,
    canClose,
    children,
    "aria-labelledby": ariaLabelledBy,
    "aria-describedby": ariaDescribedBy,
  }: {
    open: boolean;
    onClose: () => void;
    canClose?: () => boolean;
    children: React.ReactNode;
    "aria-labelledby"?: string;
    "aria-describedby"?: string;
  }) => {
    if (!open) return null;
    return React.createElement(
      "div",
      {
        role: "dialog",
        "aria-labelledby": ariaLabelledBy,
        "aria-describedby": ariaDescribedBy,
        "data-can-close": String(canClose ? canClose() : true),
      },
      children,
    );
  },
}));

vi.mock("@/hooks/useFocusTrap", () => ({ useFocusTrap: vi.fn() }));

// ── Hook mocks ────────────────────────────────────────────────────────────────

const mockMutateAsync = vi.fn();
const mockIsPending = { value: false };

vi.mock("@/hooks/useDeviceChooser", () => ({
  useSuggestedDevices: vi.fn(),
  useChoiceDevices: vi.fn(),
}));

vi.mock("@/hooks/useDevices", () => ({
  useDevices: vi.fn(),
}));

const mockIsSdkError = vi.fn();
vi.mock("@/api/errors", () => ({
  isSdkError: (err: unknown): boolean => mockIsSdkError(err) as boolean,
}));

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async () => {
  const actual =
    await vi.importActual<typeof import("react-router-dom")>(
      "react-router-dom",
    );
  return { ...actual, useNavigate: () => mockNavigate };
});

// ── Post-mock imports ─────────────────────────────────────────────────────────

import {
  useSuggestedDevices,
  useChoiceDevices,
} from "@/hooks/useDeviceChooser";
import { useDevices } from "@/hooks/useDevices";
import type { NormalizedDevice } from "@/hooks/useDevices";
import DeviceChooserDialog from "../DeviceChooserDialog";

const mockUseSuggestedDevices = vi.mocked(useSuggestedDevices);
const mockUseChoiceDevices = vi.mocked(useChoiceDevices);
const mockUseDevices = vi.mocked(useDevices);

// ── Fixtures ──────────────────────────────────────────────────────────────────

function makeDevice(n: number): NormalizedDevice {
  return {
    uid: `uid-${n}`,
    name: `hostname-${n}`,
    tags: [],
    info: { pretty_name: `Ubuntu ${n}` },
  } as unknown as NormalizedDevice;
}

const SUGGESTED_DEVICES = [makeDevice(1), makeDevice(2), makeDevice(3)];
const ALL_DEVICES = [
  makeDevice(10),
  makeDevice(11),
  makeDevice(12),
  makeDevice(13),
  makeDevice(14),
];

// ── Setup helpers ─────────────────────────────────────────────────────────────

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(
      QueryClientProvider,
      { client: queryClient },
      React.createElement(MemoryRouter, null, children),
    );
}

function setupHooks({
  suggested = SUGGESTED_DEVICES,
  suggestedLoading = false,
  allDevices = ALL_DEVICES,
  totalCount = ALL_DEVICES.length,
  allLoading = false,
  isPending = false,
}: {
  suggested?: NormalizedDevice[];
  suggestedLoading?: boolean;
  allDevices?: NormalizedDevice[];
  totalCount?: number;
  allLoading?: boolean;
  isPending?: boolean;
} = {}) {
  mockIsPending.value = isPending;

  mockUseSuggestedDevices.mockReturnValue({
    devices: suggested,
    isLoading: suggestedLoading,
    error: null,
    refetch: vi.fn(),
  });

  mockUseChoiceDevices.mockReturnValue({
    mutateAsync: mockMutateAsync,
    isPending,
    isError: false,
    error: null,
  } as never);

  mockUseDevices.mockReturnValue({
    devices: allDevices,
    totalCount,
    isLoading: allLoading,
    error: null,
    refetch: vi.fn(),
  });
}

function renderDialog(props: { open?: boolean; onClose?: () => void } = {}) {
  const onClose = props.onClose ?? vi.fn();
  return {
    onClose,
    ...render(
      <DeviceChooserDialog open={props.open ?? true} onClose={onClose} />,
      { wrapper: createWrapper() },
    ),
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  mockIsSdkError.mockReturnValue(false);
  setupHooks();
});

afterEach(cleanup);

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("DeviceChooserDialog", () => {
  // ── Rendering ───────────────────────────────────────────────────────────────

  describe("rendering", () => {
    it("renders nothing when open=false", () => {
      renderDialog({ open: false });
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("renders the dialog title when open", () => {
      renderDialog();
      expect(
        screen.getByText(/update account or select three devices/i),
      ).toBeInTheDocument();
    });

    it("renders the description when open", () => {
      renderDialog();
      expect(
        screen.getByText(/subscribe to shellhub cloud/i),
      ).toBeInTheDocument();
    });

    it("dialog has aria-labelledby pointing to the title", () => {
      renderDialog();
      const dialog = screen.getByRole("dialog");
      const titleId = dialog.getAttribute("aria-labelledby");
      expect(titleId).toBeTruthy();
      const titleEl = document.getElementById(titleId!);
      expect(titleEl).not.toBeNull();
      expect(titleEl!.textContent).toMatch(
        /update account or select three devices/i,
      );
    });

    it("dialog has aria-describedby pointing to the description", () => {
      renderDialog();
      const dialog = screen.getByRole("dialog");
      const descId = dialog.getAttribute("aria-describedby");
      expect(descId).toBeTruthy();
      const descEl = document.getElementById(descId!);
      expect(descEl).not.toBeNull();
      expect(descEl!.textContent).toMatch(/subscribe to shellhub cloud/i);
    });
  });

  // ── Tab structure ────────────────────────────────────────────────────────────

  describe("tab structure", () => {
    it("renders a tablist with role=tablist", () => {
      renderDialog();
      expect(screen.getByRole("tablist")).toBeInTheDocument();
    });

    it("renders Suggested and All tabs with role=tab", () => {
      renderDialog();
      const tabs = screen.getAllByRole("tab");
      const labels = tabs.map((t) => t.textContent);
      expect(labels).toContain("Suggested");
      expect(labels).toContain("All");
    });

    it("Suggested tab is selected by default when suggested devices are non-empty", () => {
      renderDialog();
      expect(screen.getByRole("tab", { name: "Suggested" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("renders a tabpanel for the active tab", () => {
      renderDialog();
      expect(screen.getByRole("tabpanel")).toBeInTheDocument();
    });
  });

  // ── Suggested tab ────────────────────────────────────────────────────────────

  describe("Suggested tab", () => {
    it("shows suggested device hostnames", () => {
      renderDialog();
      expect(screen.getByText("hostname-1")).toBeInTheDocument();
      expect(screen.getByText("hostname-2")).toBeInTheDocument();
      expect(screen.getByText("hostname-3")).toBeInTheDocument();
    });

    it("does not render checkboxes on suggested tab (non-editable)", () => {
      renderDialog();
      expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    });

    it("Accept button is enabled when suggested devices are present", () => {
      renderDialog();
      expect(
        screen.getByRole("button", { name: /accept/i }),
      ).not.toBeDisabled();
    });
  });

  // ── Auto-switch to All tab when suggested is empty ───────────────────────────

  describe("when suggested list is empty", () => {
    beforeEach(() => {
      setupHooks({ suggested: [], suggestedLoading: false });
    });

    it("switches to the All tab automatically", async () => {
      renderDialog();
      await waitFor(() =>
        expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
          "aria-selected",
          "true",
        ),
      );
    });

    it("Suggested tab is disabled", async () => {
      renderDialog();
      await waitFor(() =>
        expect(screen.getByRole("tab", { name: "Suggested" })).toBeDisabled(),
      );
    });
  });

  // ── Suggested query error ───────────────────────────────────────────────────

  describe("when the suggested query errors", () => {
    beforeEach(() => {
      mockUseSuggestedDevices.mockReturnValue({
        devices: [],
        isLoading: false,
        error: new Error("network failure"),
        refetch: vi.fn(),
      });
    });

    it("keeps the Suggested tab selected and surfaces the error banner", () => {
      renderDialog();
      expect(screen.getByRole("tab", { name: "Suggested" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
      expect(screen.getByRole("alert")).toHaveTextContent(
        /couldn't load the suggested devices/i,
      );
    });

    it("does not disable the Suggested tab", () => {
      renderDialog();
      expect(screen.getByRole("tab", { name: "Suggested" })).not.toBeDisabled();
    });
  });

  // ── Refetch flips suggested-empty mid-session ───────────────────────────────

  describe("when suggested becomes empty after a refetch", () => {
    it("forces tab to All even if user previously picked Suggested", async () => {
      setupHooks({ suggested: SUGGESTED_DEVICES });
      const { rerender } = render(
        <DeviceChooserDialog open onClose={vi.fn()} />,
        { wrapper: createWrapper() },
      );
      // User sees Suggested tab selected initially.
      expect(screen.getByRole("tab", { name: "Suggested" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
      // Now the refetch returns []; suggested becomes empty.
      setupHooks({ suggested: [] });
      rerender(<DeviceChooserDialog open onClose={vi.fn()} />);
      await waitFor(() =>
        expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
          "aria-selected",
          "true",
        ),
      );
      // Accept is disabled because no device is selected on the All tab.
      expect(screen.getByRole("button", { name: /accept/i })).toBeDisabled();
    });
  });

  // ── Tab-switch selection persistence ────────────────────────────────────────

  describe("tab-switch selection persistence", () => {
    it("keeps the user on All after picking a device, even if Suggested refetches non-empty", async () => {
      setupHooks({ suggested: [] });
      const user = userEvent.setup();
      const { rerender } = render(
        <DeviceChooserDialog open onClose={vi.fn()} />,
        { wrapper: createWrapper() },
      );
      // Forced to All because suggested is empty.
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
      // User picks a device on All — this commits intent.
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      );
      // Suggested refetches with results.
      setupHooks({ suggested: SUGGESTED_DEVICES });
      rerender(<DeviceChooserDialog open onClose={vi.fn()} />);
      // Tab must stay on All so the user's selection is not silently discarded.
      await waitFor(() =>
        expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
          "aria-selected",
          "true",
        ),
      );
      // The selected device is still in scope.
      expect(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      ).toBeChecked();
    });

    it("clears All-tab selections when the user explicitly switches to Suggested", async () => {
      const user = userEvent.setup();
      renderDialog();
      // Switch to All and pick a device.
      await user.click(screen.getByRole("tab", { name: "All" }));
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      );
      // Switch back to Suggested — selections should be cleared so Accept
      // submits the suggested list, not the now-stale All-tab picks.
      await user.click(screen.getByRole("tab", { name: "Suggested" }));
      await user.click(screen.getByRole("tab", { name: "All" }));
      expect(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      ).not.toBeChecked();
    });
  });

  // ── All tab ──────────────────────────────────────────────────────────────────

  describe("All tab", () => {
    async function switchToAll(user: ReturnType<typeof userEvent.setup>) {
      await user.click(screen.getByRole("tab", { name: "All" }));
    }

    it("switches to All tab on click", async () => {
      const user = userEvent.setup();
      renderDialog();
      await switchToAll(user);
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("shows device checkboxes in the All tab", async () => {
      const user = userEvent.setup();
      renderDialog();
      await switchToAll(user);
      const checkboxes = screen.getAllByRole("checkbox");
      expect(checkboxes.length).toBe(ALL_DEVICES.length);
    });

    it("Accept button is disabled when no devices are selected in All tab", async () => {
      const user = userEvent.setup();
      renderDialog();
      await switchToAll(user);
      expect(screen.getByRole("button", { name: /accept/i })).toBeDisabled();
    });

    it("Accept button is enabled after selecting 1 device", async () => {
      const user = userEvent.setup();
      renderDialog();
      await switchToAll(user);
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      );
      expect(
        screen.getByRole("button", { name: /accept/i }),
      ).not.toBeDisabled();
    });

    it("can select up to 3 devices", async () => {
      const user = userEvent.setup();
      renderDialog();
      await switchToAll(user);
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      );
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-11/i }),
      );
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-12/i }),
      );
      const checked = screen
        .getAllByRole("checkbox")
        .filter((cb) => (cb as HTMLInputElement).checked);
      expect(checked).toHaveLength(3);
    });

    it("4th checkbox is disabled when 3 are already selected", async () => {
      const user = userEvent.setup();
      renderDialog();
      await switchToAll(user);
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      );
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-11/i }),
      );
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-12/i }),
      );
      // The 4th unselected checkbox should now be disabled
      const uncheckedDisabled = screen
        .getAllByRole("checkbox")
        .filter(
          (cb) =>
            !(cb as HTMLInputElement).checked &&
            (cb as HTMLInputElement).disabled,
        );
      expect(uncheckedDisabled.length).toBeGreaterThan(0);
    });

    it("deselecting a device removes it from the selection", async () => {
      const user = userEvent.setup();
      renderDialog();
      await switchToAll(user);
      const checkbox = screen.getByRole("checkbox", {
        name: /select hostname-10/i,
      });
      await user.click(checkbox);
      await user.click(checkbox);
      expect((checkbox as HTMLInputElement).checked).toBe(false);
    });

    it("shows the selection counter with aria-live", async () => {
      const user = userEvent.setup();
      renderDialog();
      await switchToAll(user);
      const status = screen.getByRole("status");
      expect(status).toHaveAttribute("aria-live", "polite");
      expect(status.textContent).toMatch(/0 of 3/);
    });

    it("selection counter updates after selecting a device", async () => {
      const user = userEvent.setup();
      renderDialog();
      await switchToAll(user);
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      );
      expect(screen.getByRole("status").textContent).toMatch(/1 of 3/);
    });

    it("typing in search calls useDevices with the new search term", async () => {
      const user = userEvent.setup();
      renderDialog();
      await user.click(screen.getByRole("tab", { name: "All" }));
      const searchInput = screen.getByRole("searchbox");
      await user.type(searchInput, "prod");
      await waitFor(() => {
        const lastCall =
          mockUseDevices.mock.calls[mockUseDevices.mock.calls.length - 1];
        expect(lastCall[0]).toMatchObject({ search: "prod" });
      });
    });

    it("perPage is 5 when calling useDevices", () => {
      renderDialog();
      expect(mockUseDevices).toHaveBeenCalledWith(
        expect.objectContaining({ perPage: 5 }),
      );
    });

    it("requests last_seen/desc sort by default", () => {
      renderDialog();
      expect(mockUseDevices).toHaveBeenCalledWith(
        expect.objectContaining({ sortBy: "last_seen", orderBy: "desc" }),
      );
    });

    it("toggles sort to name/asc when the Hostname header is clicked", async () => {
      const user = userEvent.setup();
      renderDialog();
      await user.click(screen.getByRole("tab", { name: "All" }));
      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      const last =
        mockUseDevices.mock.calls[mockUseDevices.mock.calls.length - 1][0];
      expect(last).toMatchObject({ sortBy: "name", orderBy: "asc" });
    });
  });

  // ── Tab keyboard navigation ──────────────────────────────────────────────────

  describe("tab keyboard navigation", () => {
    it("ArrowRight moves focus from Suggested to All", () => {
      renderDialog();
      const suggested = screen.getByRole("tab", { name: "Suggested" });
      fireEvent.keyDown(suggested, { key: "ArrowRight" });
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("ArrowLeft wraps from Suggested to All (only two tabs)", () => {
      renderDialog();
      const suggested = screen.getByRole("tab", { name: "Suggested" });
      fireEvent.keyDown(suggested, { key: "ArrowLeft" });
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("Home key moves to the first enabled tab", () => {
      render(<DeviceChooserDialog open onClose={vi.fn()} />, {
        wrapper: createWrapper(),
      });
      const allTab = screen.getByRole("tab", { name: "All" });
      fireEvent.click(allTab);
      expect(allTab).toHaveAttribute("aria-selected", "true");
      fireEvent.keyDown(allTab, { key: "Home" });
      expect(screen.getByRole("tab", { name: "Suggested" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("End key moves to the last tab", () => {
      renderDialog();
      const suggested = screen.getByRole("tab", { name: "Suggested" });
      fireEvent.keyDown(suggested, { key: "End" });
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
  });

  // ── Footer buttons ───────────────────────────────────────────────────────────

  describe("Cancel button", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await user.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledOnce();
    });

    it("Cancel is disabled while mutation is in flight", () => {
      setupHooks({ isPending: true });
      renderDialog();
      expect(screen.getByRole("button", { name: /cancel/i })).toBeDisabled();
    });
  });

  describe("Subscribe button", () => {
    it("navigates to /settings#billing when Subscribe is clicked", async () => {
      const user = userEvent.setup();
      renderDialog();
      await user.click(screen.getByRole("button", { name: /subscribe/i }));
      expect(mockNavigate).toHaveBeenCalledWith("/settings#billing");
    });

    it("calls onClose when Subscribe is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await user.click(screen.getByRole("button", { name: /subscribe/i }));
      expect(onClose).toHaveBeenCalledOnce();
    });

    it("Subscribe is disabled while mutation is in flight", () => {
      setupHooks({ isPending: true });
      renderDialog();
      expect(screen.getByRole("button", { name: /subscribe/i })).toBeDisabled();
    });
  });

  describe("Accept button", () => {
    it("calls mutateAsync with the suggested UIDs when on suggested tab", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const user = userEvent.setup();
      renderDialog();
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(mockMutateAsync).toHaveBeenCalledWith({
          body: { choices: ["uid-1", "uid-2", "uid-3"] },
        }),
      );
    });

    it("calls onClose after a successful Accept", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() => expect(onClose).toHaveBeenCalledOnce());
    });

    it("calls mutateAsync with selected UIDs when on All tab", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const user = userEvent.setup();
      renderDialog();
      await user.click(screen.getByRole("tab", { name: "All" }));
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      );
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(mockMutateAsync).toHaveBeenCalledWith({
          body: { choices: ["uid-10"] },
        }),
      );
    });

    it("shows spinner and 'Saving…' text while mutation is pending", () => {
      setupHooks({ isPending: true });
      renderDialog();
      expect(
        screen.getByRole("button", { name: /saving/i }),
      ).toBeInTheDocument();
    });

    it("Accept is disabled while mutation is in flight", () => {
      setupHooks({ isPending: true });
      renderDialog();
      // The button text changes to "Saving…" but remains disabled
      const btn = screen.getByRole("button", { name: /saving/i });
      expect(btn).toBeDisabled();
    });

    it("blocks close (canClose=false) while mutation is in flight", () => {
      setupHooks({ isPending: true });
      renderDialog();
      const dialog = screen.getByRole("dialog");
      expect(dialog.getAttribute("data-can-close")).toBe("false");
    });

    it("allows close (canClose=true) when mutation is idle", () => {
      setupHooks({ isPending: false });
      renderDialog();
      const dialog = screen.getByRole("dialog");
      expect(dialog.getAttribute("data-can-close")).toBe("true");
    });
  });

  // ── Error handling ───────────────────────────────────────────────────────────

  describe("error handling", () => {
    it("shows generic error when Accept fails with a non-403 error", async () => {
      mockMutateAsync.mockRejectedValue(new Error("network error"));
      mockIsSdkError.mockReturnValue(false);
      const user = userEvent.setup();
      renderDialog();
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(
          /couldn't save your selection/i,
        ),
      );
    });

    it("shows permission error when Accept fails with a 403 SDK error", async () => {
      const err = { status: 403 };
      mockMutateAsync.mockRejectedValue(err);
      mockIsSdkError.mockImplementation((e: unknown) => e === err);
      const user = userEvent.setup();
      renderDialog();
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(
          /don't have permission/i,
        ),
      );
    });

    it("error alert is rendered above the footer", async () => {
      mockMutateAsync.mockRejectedValue(new Error("fail"));
      mockIsSdkError.mockReturnValue(false);
      const user = userEvent.setup();
      renderDialog();
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toBeInTheDocument(),
      );
      // Error appears before footer buttons in the DOM
      const alert = screen.getByRole("alert");
      const cancel = screen.getByRole("button", { name: /cancel/i });
      expect(
        alert.compareDocumentPosition(cancel) &
          Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy();
    });

    it("clears the error when switching tabs", async () => {
      mockMutateAsync.mockRejectedValue(new Error("fail"));
      mockIsSdkError.mockReturnValue(false);
      const user = userEvent.setup();
      renderDialog();
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toBeInTheDocument(),
      );
      await user.click(screen.getByRole("tab", { name: "All" }));
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });
});
