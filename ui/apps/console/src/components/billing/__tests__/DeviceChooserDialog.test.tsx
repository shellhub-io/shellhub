import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { Device } from "@/client";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { mockDevice as mockDeviceFactory } from "@/tests/factories";

vi.mock("@/components/common/BaseDialog", async () => ({
  default: (await import("@/tests/mocks")).MockBaseDialog,
}));

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getDevicesMostUsed: vi.fn(),
    getDevices: vi.fn(),
    choiceDevices: vi.fn(),
  }),
);

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

import DeviceChooserDialog from "../DeviceChooserDialog";

function makeDevice(n: number): Device {
  return mockDeviceFactory({
    uid: `uid-${n}`,
    name: `hostname-${n}`,
    tags: [],
    info: {
      pretty_name: `Ubuntu ${n}`,
      id: "ubuntu",
      arch: "x86_64",
      platform: "native",
      version: "0.14.0",
    },
  });
}

const SUGGESTED_DEVICES = [makeDevice(1), makeDevice(2), makeDevice(3)];
const ALL_DEVICES = [
  makeDevice(10),
  makeDevice(11),
  makeDevice(12),
  makeDevice(13),
  makeDevice(14),
];

function setupSdk({
  suggested = SUGGESTED_DEVICES,
  allDevices = ALL_DEVICES,
  totalCount = ALL_DEVICES.length,
}: {
  suggested?: Device[];
  allDevices?: Device[];
  totalCount?: number;
} = {}) {
  sdk.getDevicesMostUsed.mockResolvedValue(mockSdkResponse(suggested));
  sdk.getDevices.mockResolvedValue(paginatedResponse(allDevices, totalCount));
  sdk.choiceDevices.mockResolvedValue(mockSdkResponse(undefined));
}

function renderDialog(props: { open?: boolean; onClose?: () => void } = {}) {
  const onClose = props.onClose ?? vi.fn();
  return {
    onClose,
    ...render(
      <DeviceChooserDialog open={props.open ?? true} onClose={onClose} />,
      { wrapper: createTestWrapper({ initialEntries: ["/"] }) },
    ),
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  mockIsSdkError.mockReturnValue(false);
  setupSdk();
});

describe("DeviceChooserDialog", () => {
  describe("rendering", () => {
    it("renders nothing when open=false", () => {
      renderDialog({ open: false });
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("renders the dialog title when open", async () => {
      renderDialog();
      expect(
        await screen.findByText(/update account or select three devices/i),
      ).toBeInTheDocument();
    });

    it("renders the description when open", async () => {
      renderDialog();
      expect(
        await screen.findByText(/subscribe to shellhub cloud/i),
      ).toBeInTheDocument();
    });

    it("dialog has aria-labelledby pointing to the title", async () => {
      renderDialog();
      await screen.findByText(/update account or select three devices/i);
      const dialog = screen.getByRole("dialog");
      const titleId = dialog.getAttribute("aria-labelledby");
      expect(titleId).toBeTruthy();
      const titleEl = document.getElementById(titleId!);
      expect(titleEl).not.toBeNull();
      expect(titleEl!.textContent).toMatch(
        /update account or select three devices/i,
      );
    });

    it("dialog has aria-describedby pointing to the description", async () => {
      renderDialog();
      await screen.findByText(/subscribe to shellhub cloud/i);
      const dialog = screen.getByRole("dialog");
      const descId = dialog.getAttribute("aria-describedby");
      expect(descId).toBeTruthy();
      const descEl = document.getElementById(descId!);
      expect(descEl).not.toBeNull();
      expect(descEl!.textContent).toMatch(/subscribe to shellhub cloud/i);
    });
  });

  describe("tab structure", () => {
    it("renders a tablist with role=tablist", async () => {
      renderDialog();
      expect(await screen.findByRole("tablist")).toBeInTheDocument();
    });

    it("renders Suggested and All tabs with role=tab", async () => {
      renderDialog();
      await screen.findByRole("tablist");
      const tabs = screen.getAllByRole("tab");
      const labels = tabs.map((t) => t.textContent);
      expect(labels).toContain("Suggested");
      expect(labels).toContain("All");
    });

    it("Suggested tab is selected by default when suggested devices are non-empty", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      expect(screen.getByRole("tab", { name: "Suggested" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("renders a tabpanel for the active tab", async () => {
      renderDialog();
      expect(await screen.findByRole("tabpanel")).toBeInTheDocument();
    });
  });

  describe("Suggested tab", () => {
    it("shows suggested device hostnames", async () => {
      renderDialog();
      expect(await screen.findByText("hostname-1")).toBeInTheDocument();
      expect(screen.getByText("hostname-2")).toBeInTheDocument();
      expect(screen.getByText("hostname-3")).toBeInTheDocument();
    });

    it("does not render checkboxes on suggested tab (non-editable)", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    });

    it("Accept button is enabled when suggested devices are present", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      expect(
        screen.getByRole("button", { name: /accept/i }),
      ).not.toBeDisabled();
    });

    it("renders CheckIcon (heroicons) not an inline custom SVG for the selected-row check", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      const inlinePath = document.querySelector('path[d="M3 8l3.5 3.5L13 5"]');
      expect(inlinePath).toBeNull();
      const checkSpan = document.querySelector('span[aria-hidden="true"] svg');
      expect(checkSpan).not.toBeNull();
    });
  });

  describe("when suggested list is empty", () => {
    beforeEach(() => {
      setupSdk({ suggested: [] });
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

  describe("when the suggested query errors", () => {
    beforeEach(() => {
      sdk.getDevicesMostUsed.mockRejectedValue(new Error("network failure"));
    });

    it("keeps the Suggested tab selected and surfaces the error banner", async () => {
      renderDialog();
      expect(await screen.findByRole("alert")).toHaveTextContent(
        /couldn't load the suggested devices/i,
      );
      expect(screen.getByRole("tab", { name: "Suggested" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("does not disable the Suggested tab", async () => {
      renderDialog();
      await screen.findByRole("alert");
      expect(screen.getByRole("tab", { name: "Suggested" })).not.toBeDisabled();
    });
  });

  describe("when suggested becomes empty after a refetch", () => {
    it("forces tab to All when suggested starts empty", async () => {
      setupSdk({ suggested: [] });
      renderDialog();
      await waitFor(() =>
        expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
          "aria-selected",
          "true",
        ),
      );
    });
  });

  describe("tab-switch selection persistence", () => {
    it("clears All-tab selections when the user explicitly switches to Suggested", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("tab", { name: "All" }));
      await user.click(
        await screen.findByRole("checkbox", { name: /select hostname-10/i }),
      );
      await user.click(screen.getByRole("tab", { name: "Suggested" }));
      await user.click(screen.getByRole("tab", { name: "All" }));
      await waitFor(() =>
        expect(
          screen.getByRole("checkbox", { name: /select hostname-10/i }),
        ).not.toBeChecked(),
      );
    });
  });

  describe("All tab", () => {
    async function switchToAll(user: ReturnType<typeof userEvent.setup>) {
      await user.click(screen.getByRole("tab", { name: "All" }));
    }

    it("switches to All tab on click", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await switchToAll(user);
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("shows device checkboxes in the All tab", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await switchToAll(user);
      await screen.findByRole("checkbox", { name: /select hostname-10/i });
      const checkboxes = screen.getAllByRole("checkbox");
      expect(checkboxes.length).toBe(ALL_DEVICES.length);
    });

    it("Accept button is disabled when no devices are selected in All tab", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await switchToAll(user);
      expect(screen.getByRole("button", { name: /accept/i })).toBeDisabled();
    });

    it("Accept button is enabled after selecting 1 device", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await switchToAll(user);
      await user.click(
        await screen.findByRole("checkbox", { name: /select hostname-10/i }),
      );
      expect(
        screen.getByRole("button", { name: /accept/i }),
      ).not.toBeDisabled();
    });

    it("can select up to 3 devices", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await switchToAll(user);
      await screen.findByRole("checkbox", { name: /select hostname-10/i });
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
      await screen.findByText("hostname-1");
      await switchToAll(user);
      await screen.findByRole("checkbox", { name: /select hostname-10/i });
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-10/i }),
      );
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-11/i }),
      );
      await user.click(
        screen.getByRole("checkbox", { name: /select hostname-12/i }),
      );
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
      await screen.findByText("hostname-1");
      await switchToAll(user);
      const checkbox = await screen.findByRole("checkbox", {
        name: /select hostname-10/i,
      });
      await user.click(checkbox);
      await user.click(checkbox);
      expect((checkbox as HTMLInputElement).checked).toBe(false);
    });

    it("shows the selection counter with aria-live", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await switchToAll(user);
      await screen.findByRole("checkbox", { name: /select hostname-10/i });
      const status = screen.getByRole("status");
      expect(status).toHaveAttribute("aria-live", "polite");
      expect(status.textContent).toMatch(/0 of 3/);
    });

    it("selection counter updates after selecting a device", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await switchToAll(user);
      await user.click(
        await screen.findByRole("checkbox", { name: /select hostname-10/i }),
      );
      expect(screen.getByRole("status").textContent).toMatch(/1 of 3/);
    });

    it("typing in search filters devices via the SDK", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("tab", { name: "All" }));
      await screen.findByRole("searchbox");
      const searchInput = screen.getByRole("searchbox");
      await user.type(searchInput, "prod");
      await waitFor(() => {
        const call = sdk.getDevices.mock.calls.at(-1)?.[0] as {
          query?: { filter?: string };
        };
        const decoded = atob(call?.query?.filter ?? "");
        expect(decoded).toContain("prod");
      });
    });

    it("requests per_page=5 from the SDK", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await switchToAll(user);
      await waitFor(() => {
        expect(sdk.getDevices).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ per_page: 5 }),
          }),
        );
      });
    });

    it("requests last_seen/desc sort by default", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await switchToAll(user);
      await waitFor(() => {
        expect(sdk.getDevices).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({
              sort_by: "last_seen",
              order_by: "desc",
            }),
          }),
        );
      });
    });

    it("toggles sort to name/asc when the Hostname header is clicked", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("tab", { name: "All" }));
      await screen.findByRole("checkbox", { name: /select hostname-10/i });
      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      await waitFor(() => {
        expect(sdk.getDevices).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({
              sort_by: "name",
              order_by: "asc",
            }),
          }),
        );
      });
    });
  });

  describe("tab keyboard navigation", () => {
    it("ArrowRight moves focus from Suggested to All", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      const suggested = screen.getByRole("tab", { name: "Suggested" });
      fireEvent.keyDown(suggested, { key: "ArrowRight" });
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("ArrowLeft wraps from Suggested to All (only two tabs)", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      const suggested = screen.getByRole("tab", { name: "Suggested" });
      fireEvent.keyDown(suggested, { key: "ArrowLeft" });
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("Home key moves to the first enabled tab", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      const allTab = screen.getByRole("tab", { name: "All" });
      fireEvent.click(allTab);
      expect(allTab).toHaveAttribute("aria-selected", "true");
      fireEvent.keyDown(allTab, { key: "Home" });
      expect(screen.getByRole("tab", { name: "Suggested" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("End key moves to the last tab", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      const suggested = screen.getByRole("tab", { name: "Suggested" });
      fireEvent.keyDown(suggested, { key: "End" });
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
  });

  describe("Cancel button", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledOnce();
    });

    it("Cancel is disabled while mutation is in flight", async () => {
      sdk.choiceDevices.mockReturnValue(new Promise(() => {}));
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("button", { name: /cancel/i })).toBeDisabled(),
      );
    });
  });

  describe("Subscribe button", () => {
    it("navigates to /settings#billing when Subscribe is clicked", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /subscribe/i }));
      expect(mockNavigate).toHaveBeenCalledWith("/settings#billing");
    });

    it("calls onClose when Subscribe is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /subscribe/i }));
      expect(onClose).toHaveBeenCalledOnce();
    });

    it("Subscribe is disabled while mutation is in flight", async () => {
      sdk.choiceDevices.mockReturnValue(new Promise(() => {}));
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /subscribe/i }),
        ).toBeDisabled(),
      );
    });
  });

  describe("Accept button", () => {
    it("calls choiceDevices with the suggested UIDs when on suggested tab", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(sdk.choiceDevices).toHaveBeenCalledWith(
          expect.objectContaining({
            body: { choices: ["uid-1", "uid-2", "uid-3"] },
            throwOnError: true,
          }),
        ),
      );
    });

    it("calls onClose after a successful Accept", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() => expect(onClose).toHaveBeenCalledOnce());
    });

    it("calls choiceDevices with selected UIDs when on All tab", async () => {
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("tab", { name: "All" }));
      await user.click(
        await screen.findByRole("checkbox", { name: /select hostname-10/i }),
      );
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(sdk.choiceDevices).toHaveBeenCalledWith(
          expect.objectContaining({
            body: { choices: ["uid-10"] },
            throwOnError: true,
          }),
        ),
      );
    });

    it("shows spinner and 'Saving…' text while mutation is pending", async () => {
      sdk.choiceDevices.mockReturnValue(new Promise(() => {}));
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /saving/i }),
        ).toBeInTheDocument(),
      );
    });

    it("Accept is disabled while mutation is in flight", async () => {
      sdk.choiceDevices.mockReturnValue(new Promise(() => {}));
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() => {
        const btn = screen.getByRole("button", { name: /saving/i });
        expect(btn).toBeDisabled();
      });
    });

    it("blocks close (canClose=false) while mutation is in flight", async () => {
      sdk.choiceDevices.mockReturnValue(new Promise(() => {}));
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() => {
        const dialog = screen.getByRole("dialog");
        expect(dialog.getAttribute("data-can-close")).toBe("false");
      });
    });

    it("allows close (canClose=true) when mutation is idle", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      const dialog = screen.getByRole("dialog");
      expect(dialog.getAttribute("data-can-close")).toBe("true");
    });
  });

  describe("error handling", () => {
    it("shows generic error when Accept fails with a non-403 error", async () => {
      sdk.choiceDevices.mockRejectedValue(new Error("network error"));
      mockIsSdkError.mockReturnValue(false);
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(
          /couldn't save your selection/i,
        ),
      );
    });

    it("shows permission error when Accept fails with a 403 SDK error", async () => {
      const err = { status: 403 };
      sdk.choiceDevices.mockRejectedValue(err);
      mockIsSdkError.mockImplementation((e: unknown) => e === err);
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(
          /don't have permission/i,
        ),
      );
    });

    it("error alert is rendered above the footer", async () => {
      sdk.choiceDevices.mockRejectedValue(new Error("fail"));
      mockIsSdkError.mockReturnValue(false);
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toBeInTheDocument(),
      );
      const alert = screen.getByRole("alert");
      const cancel = screen.getByRole("button", { name: /cancel/i });
      expect(
        alert.compareDocumentPosition(cancel) &
          Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy();
    });

    it("clears the error when switching tabs", async () => {
      sdk.choiceDevices.mockRejectedValue(new Error("fail"));
      mockIsSdkError.mockReturnValue(false);
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toBeInTheDocument(),
      );
      sdk.choiceDevices.mockResolvedValue(mockSdkResponse(undefined));
      await user.click(screen.getByRole("tab", { name: "All" }));
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });
});
