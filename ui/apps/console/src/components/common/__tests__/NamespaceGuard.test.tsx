import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useConnectivityStore } from "@/stores/connectivityStore";
import { createTestWrapper } from "@/tests/wrapper";
import { paginatedResponse, mockSdkResponse } from "@/tests/sdk";
import { mockNamespace } from "@/tests/factories";
import NamespaceGuard from "../NamespaceGuard";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getNamespaces: vi.fn(),
    getNamespaceToken: vi.fn(),
  }),
);

vi.mock("../CreateNamespace", () => ({
  default: () => <div data-testid="create-namespace" />,
}));

vi.mock("@/components/layout/UserMenu", () => ({
  default: () => <div data-testid="user-menu" />,
}));

beforeEach(() => {
  vi.clearAllMocks();
  sdk.getNamespaces.mockResolvedValue(paginatedResponse([]));
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token" }),
  );
  useConnectivityStore.getState().markUp();
});

function renderGuard(initialPath = "/dashboard") {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route element={<NamespaceGuard />}>
          <Route path="/dashboard" element={<div>dashboard content</div>} />
          <Route path="/profile" element={<div>profile content</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("NamespaceGuard", () => {
  describe("loading state", () => {
    it("shows a loading spinner while namespaces are not yet loaded", () => {
      sdk.getNamespaces.mockReturnValue(new Promise(() => {}));
      renderGuard();
      expect(screen.getByText(/loading/i)).toBeInTheDocument();
    });

    it("does not render the outlet while loading", () => {
      sdk.getNamespaces.mockReturnValue(new Promise(() => {}));
      renderGuard();
      expect(screen.queryByText("dashboard content")).not.toBeInTheDocument();
    });
  });

  describe("with namespaces", () => {
    it("renders the outlet when namespaces exist", async () => {
      sdk.getNamespaces.mockResolvedValue(
        paginatedResponse([mockNamespace({ tenant_id: "t1", name: "ns1" })]),
      );
      renderGuard();
      expect(await screen.findByText("dashboard content")).toBeInTheDocument();
    });

    it("does not show the create-namespace screen when namespaces exist", async () => {
      sdk.getNamespaces.mockResolvedValue(
        paginatedResponse([mockNamespace({ tenant_id: "t1", name: "ns1" })]),
      );
      renderGuard();
      expect(await screen.findByText("dashboard content")).toBeInTheDocument();
      expect(screen.queryByTestId("create-namespace")).not.toBeInTheDocument();
    });
  });

  describe("without namespaces — non-profile route", () => {
    it("shows the create-namespace screen", async () => {
      renderGuard("/dashboard");
      expect(await screen.findByTestId("create-namespace")).toBeInTheDocument();
    });

    it("does not render the outlet", async () => {
      renderGuard("/dashboard");
      expect(await screen.findByTestId("create-namespace")).toBeInTheDocument();
      expect(screen.queryByText("dashboard content")).not.toBeInTheDocument();
    });

    it("renders UserMenu in the minimal header", async () => {
      renderGuard("/dashboard");
      expect(await screen.findByTestId("user-menu")).toBeInTheDocument();
    });
  });

  describe("without namespaces — /profile route", () => {
    it("renders the outlet instead of the create-namespace screen", async () => {
      renderGuard("/profile");
      expect(await screen.findByText("profile content")).toBeInTheDocument();
    });

    it("does not show the create-namespace screen", async () => {
      renderGuard("/profile");
      expect(await screen.findByText("profile content")).toBeInTheDocument();
      expect(screen.queryByTestId("create-namespace")).not.toBeInTheDocument();
    });

    it("does not show the minimal header", async () => {
      renderGuard("/profile");
      expect(await screen.findByText("profile content")).toBeInTheDocument();
      expect(screen.queryByTestId("user-menu")).not.toBeInTheDocument();
    });
  });
});
