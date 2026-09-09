import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { Device } from "@/client/model";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockDevice as mockDeviceFactory } from "@/tests/factories";

vi.mock("@/components/common/BaseDialog", async () => ({
  default: (await import("@/tests/mocks")).MockBaseDialog,
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

function setupHandlers({
  suggested = SUGGESTED_DEVICES,
  allDevices = ALL_DEVICES,
  totalCount = ALL_DEVICES.length,
}: {
  suggested?: Device[];
  allDevices?: Device[];
  totalCount?: number;
} = {}) {
  server.use(
    http.get("*/api/billing/devices-most-used", () =>
      HttpResponse.json(suggested),
    ),
    http.get("*/api/devices", () => jsonWithTotal(allDevices, totalCount)),
    http.post(
      "*/api/billing/device-choice",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
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
  setupHandlers();
});

describe("DeviceChooserDialog", () => {
  describe("rendering", () => {
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
    it("Suggested tab is selected by default when suggested devices are non-empty", async () => {
      renderDialog();
      await screen.findByText("hostname-1");
      expect(screen.getByRole("tab", { name: "Suggested" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
  });

  describe("Suggested tab", () => {
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
  });

  describe("when suggested list is empty", () => {
    beforeEach(() => {
      setupHandlers({ suggested: [] });
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
      server.use(
        http.get("*/api/billing/devices-most-used", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
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
    it("Cancel is disabled while mutation is in flight", async () => {
      server.use(
        http.post("*/api/billing/device-choice", () => new Promise(() => {})),
      );
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

    it("Subscribe is disabled while mutation is in flight", async () => {
      server.use(
        http.post("*/api/billing/device-choice", () => new Promise(() => {})),
      );
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
    it("calls onClose after a successful Accept on suggested tab", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() => expect(onClose).toHaveBeenCalledOnce());
    });

    it("calls onClose after a successful Accept on All tab", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("tab", { name: "All" }));
      await user.click(
        await screen.findByRole("checkbox", { name: /select hostname-10/i }),
      );
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() => expect(onClose).toHaveBeenCalledOnce());
    });

    it("Accept is disabled while mutation is in flight", async () => {
      server.use(
        http.post("*/api/billing/device-choice", () => new Promise(() => {})),
      );
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
      server.use(
        http.post("*/api/billing/device-choice", () => new Promise(() => {})),
      );
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
      server.use(
        http.post("*/api/billing/device-choice", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
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
      server.use(
        http.post("*/api/billing/device-choice", () =>
          HttpResponse.json({}, { status: 403 }),
        ),
      );
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

    it("clears the error when switching tabs", async () => {
      server.use(
        http.post("*/api/billing/device-choice", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = userEvent.setup();
      renderDialog();
      await screen.findByText("hostname-1");
      await user.click(screen.getByRole("button", { name: /accept/i }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toBeInTheDocument(),
      );
      server.use(
        http.post(
          "*/api/billing/device-choice",
          () => new HttpResponse(null, { status: 204 }),
        ),
      );
      await user.click(screen.getByRole("tab", { name: "All" }));
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });
});
