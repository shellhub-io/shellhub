import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { getUserInfoOptions } from "@/client";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import { mockNamespace, mockUserAuth } from "@/tests/factories";
import { setEdition } from "@/tests/edition";
import { seedAuthStore, VALID_JWT } from "@/tests/seedAuthStore";
import { useAuthStore } from "@/stores/authStore";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import {
  namespaceTab,
  useWorkspaceTabsStore,
} from "@/stores/workspaceTabsStore";
import NewTabPicker from "@/components/commandPalette/NewTabPicker";

const namespaces = [
  mockNamespace({ tenant_id: "t-alpha", name: "alpha" }),
  mockNamespace({ tenant_id: "t-zeta", name: "zeta" }),
  mockNamespace({ tenant_id: "tenant-456", name: "dev" }),
  mockNamespace({ tenant_id: "t-beta", name: "beta" }),
];

function renderPicker({
  path = "/dashboard",
  queryClient = createTestQueryClient(),
} = {}) {
  return render(<NewTabPicker />, {
    wrapper: createTestWrapper({ initialEntries: [path], queryClient }),
  });
}

async function openPicker() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: "New tab" }));
  const picker = await screen.findByRole("dialog", { name: "New tab" });
  await within(picker).findByText("alpha");
  return { user, picker };
}

function expectOptions(picker: HTMLElement, names: string[]) {
  const options = within(picker).getAllByRole("option");
  expect(options).toHaveLength(names.length);
  names.forEach((name, i) =>
    expect(within(options[i]).getByText(name)).toBeInTheDocument(),
  );
  return options;
}

describe("NewTabPicker", () => {
  beforeEach(() => {
    setEdition("community");
    seedAuthStore({ isAdmin: false });
    useCommandPaletteStore.setState({ view: null });
    useWorkspaceTabsStore.setState({ tabs: [], failures: {} });
    server.use(
      http.get("*/api/namespaces", () => jsonWithTotal(namespaces)),
      http.get("*/api/auth/user", () => HttpResponse.json(mockUserAuth())),
    );
  });

  it("lists the current namespace, then open tabs, then the rest A to Z", async () => {
    useWorkspaceTabsStore.setState({
      tabs: [namespaceTab("t-zeta", "zeta")],
    });
    renderPicker();

    const { picker } = await openPicker();

    const options = expectOptions(picker, [
      "dev",
      "zeta",
      "alpha",
      "beta",
      "Connect to a device",
    ]);
    expect(within(options[0]).getByText("Current")).toBeInTheDocument();
  });

  it("filters by the typed name", async () => {
    renderPicker();
    const { user, picker } = await openPicker();

    await user.type(within(picker).getByRole("combobox"), "bet");

    expectOptions(picker, ["beta"]);
  });

  it("says nothing matches when the filter hides every row", async () => {
    renderPicker();
    const { user, picker } = await openPicker();

    await user.type(within(picker).getByRole("combobox"), "zzz");

    expect(within(picker).getByRole("status")).toHaveTextContent(
      "Nothing matches",
    );
    expect(within(picker).queryByRole("option")).not.toBeInTheDocument();
  });

  it("opens the chosen namespace's tab and closes", async () => {
    server.use(
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({ token: VALID_JWT, role: "owner" }),
      ),
    );
    renderPicker();
    const { user, picker } = await openPicker();

    await user.click(within(picker).getByText("alpha"));

    expect(useCommandPaletteStore.getState().view).toBeNull();
    expect(useWorkspaceTabsStore.getState().tabs).toContainEqual(
      expect.objectContaining({ id: "ns:t-alpha" }),
    );
  });

  it("hands over to the device palette from Connect to a device", async () => {
    renderPicker();
    const { user, picker } = await openPicker();

    await user.click(within(picker).getByText("Connect to a device"));

    expect(useCommandPaletteStore.getState().view).toBe("connect");
  });

  it("runs the highlighted row on Enter", async () => {
    renderPicker();
    const { user, picker } = await openPicker();

    await user.type(within(picker).getByRole("combobox"), "connect{Enter}");

    expect(useCommandPaletteStore.getState().view).toBe("connect");
  });

  it("leaves out Connect to a device in the admin console", async () => {
    renderPicker({ path: "/admin/instance" });
    const { picker } = await openPicker();

    expect(
      within(picker).queryByText("Connect to a device"),
    ).not.toBeInTheDocument();
  });

  it("closes on Escape", async () => {
    renderPicker();
    const { user } = await openPicker();

    await user.keyboard("{Escape}");

    expect(useCommandPaletteStore.getState().view).toBeNull();
    expect(
      screen.queryByRole("dialog", { name: "New tab" }),
    ).not.toBeInTheDocument();
  });

  it("offers Create namespace on cloud to an account with an allowance", async () => {
    setEdition("cloud");
    renderPicker();
    const { picker } = await openPicker();

    expect(
      await within(picker).findByText("Create namespace"),
    ).toBeInTheDocument();
  });

  it("hides Create namespace on cloud from an account with no allowance", async () => {
    setEdition("cloud");
    server.use(
      http.get("*/api/auth/user", () =>
        HttpResponse.json(mockUserAuth({ max_namespaces: 0 })),
      ),
    );
    const queryClient = createTestQueryClient();
    renderPicker({ queryClient });
    const { picker } = await openPicker();

    await waitFor(() =>
      expect(
        queryClient.getQueryState(getUserInfoOptions().queryKey)?.status,
      ).toBe("success"),
    );
    expect(
      within(picker).queryByText("Create namespace"),
    ).not.toBeInTheDocument();
  });

  it("offers Admin Console to an instance admin on enterprise", async () => {
    setEdition("enterprise");
    useAuthStore.setState({ isAdmin: true });
    renderPicker();
    const { picker } = await openPicker();

    expect(within(picker).getByText("Admin Console")).toBeInTheDocument();
  });

  it.each([
    { edition: "enterprise", isAdmin: false },
    { edition: "community", isAdmin: true },
  ] as const)(
    "hides Admin Console on $edition when isAdmin is $isAdmin",
    async ({ edition, isAdmin }) => {
      setEdition(edition);
      useAuthStore.setState({ isAdmin });
      renderPicker();
      const { picker } = await openPicker();

      expect(
        within(picker).queryByText("Admin Console"),
      ).not.toBeInTheDocument();
    },
  );
});
