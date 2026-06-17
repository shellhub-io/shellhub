import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { defaultConfig } from "@/env";
import PendingDevices from "../PendingDevices";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const mockGetConfig = vi.fn();

vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: (): unknown => mockGetConfig() };
});

const mockAcceptMutateAsync = vi.fn();
const mockRejectMutateAsync = vi.fn();

vi.mock("@/hooks/useDeviceMutations", () => ({
  useAcceptDevice: () => ({
    mutateAsync: mockAcceptMutateAsync,
    isPending: false,
  }),
  useRejectDevice: () => ({
    mutateAsync: mockRejectMutateAsync,
    isPending: false,
  }),
}));

const mockUseDevices = vi.fn<
  () => {
    devices: Array<{
      uid: string;
      name: string;
      identity?: { mac?: string };
      info?: { id?: string; pretty_name?: string };
      tags: string[];
    }>;
    totalCount: number;
    isLoading: boolean;
  }
>();

vi.mock("@/hooks/useDevices", () => ({
  useDevices: () => mockUseDevices(),
}));

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function makeDevice(uid = "device-1", name = "my-device") {
  return {
    uid,
    name,
    identity: { mac: "aa:bb:cc:dd:ee:ff" },
    info: { id: "id-1", pretty_name: "Ubuntu 22.04" },
    tags: [],
  };
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </MemoryRouter>
  );
}

function renderPendingDevices() {
  return render(<PendingDevices />, { wrapper: createWrapper() });
}

async function openDropdown() {
  const trigger = screen.getByRole("button", { name: /pending devices/i });
  await userEvent.click(trigger);
}

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({
    ...defaultConfig,
    enterprise: true,
    cloud: false,
  });
  mockUseDevices.mockReturnValue({
    devices: [makeDevice()],
    totalCount: 1,
    isLoading: false,
  });
  mockAcceptMutateAsync.mockResolvedValue(undefined);
  mockRejectMutateAsync.mockResolvedValue(undefined);
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("PendingDevices", () => {
  describe("inline error state — accept failure", () => {
    it("shows inline license-limit message when accept returns 402 on enterprise", async () => {
      mockAcceptMutateAsync.mockRejectedValue({ status: 402 });

      renderPendingDevices();
      await openDropdown();

      const acceptBtn = screen.getByRole("button", { name: /accept/i });
      await userEvent.click(acceptBtn);

      await waitFor(() => {
        const alert = screen.getByRole("alert");
        expect(alert).toBeInTheDocument();
        expect(alert.textContent).toMatch(/license/i);
      });
    });

    it("shows billing message when accept returns 402 on cloud", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: true,
        cloud: true,
      });
      mockAcceptMutateAsync.mockRejectedValue({ status: 402 });

      renderPendingDevices();
      await openDropdown();

      const acceptBtn = screen.getByRole("button", { name: /accept/i });
      await userEvent.click(acceptBtn);

      await waitFor(() => {
        const alert = screen.getByRole("alert");
        expect(alert).toBeInTheDocument();
        expect(alert.textContent).toMatch(/billing|subscription|plan/i);
      });
    });
  });

  describe("inline error state — reject failure", () => {
    it("shows inline 'Failed to reject device.' message when reject fails", async () => {
      mockRejectMutateAsync.mockRejectedValue(new Error("server error"));

      renderPendingDevices();
      await openDropdown();

      const rejectBtn = screen.getByRole("button", { name: /reject/i });
      await userEvent.click(rejectBtn);

      await waitFor(() => {
        const alert = screen.getByRole("alert");
        expect(alert).toBeInTheDocument();
        expect(alert.textContent).toMatch(/failed to reject device/i);
      });
    });
  });

  describe("success flash", () => {
    it("shows the success flash after a successful accept", async () => {
      renderPendingDevices();
      await openDropdown();

      await userEvent.click(screen.getByRole("button", { name: /accept/i }));

      await waitFor(() =>
        expect(screen.getByText("Accepted")).toBeInTheDocument(),
      );
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("error clears on dropdown close", () => {
    it("clears the error when the dropdown is closed and reopened", async () => {
      mockAcceptMutateAsync.mockRejectedValue({ status: 402 });

      renderPendingDevices();
      await openDropdown();

      const acceptBtn = screen.getByRole("button", { name: /accept/i });
      await userEvent.click(acceptBtn);

      await waitFor(() =>
        expect(screen.getByRole("alert")).toBeInTheDocument(),
      );

      // Close the dropdown by clicking the X button
      const closeBtn = screen.getByRole("button", { name: /close/i });
      await userEvent.click(closeBtn);

      await waitFor(() =>
        expect(screen.queryByRole("alert")).not.toBeInTheDocument(),
      );

      // Reopen and confirm the error is gone
      await openDropdown();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });
});
