import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// Stub the community-polling useEffect so it does not leak timers
vi.mock("@/client", () => ({
  getNamespaces: vi.fn().mockResolvedValue({ data: [] }),
}));
vi.mock("@/hooks/useNamespaceMutations", () => ({
  useCreateNamespace: vi.fn(),
  useSwitchNamespace: vi.fn(),
}));

import { getConfig, defaultConfig } from "@/env";
import {
  useCreateNamespace,
  useSwitchNamespace,
} from "@/hooks/useNamespaceMutations";
import CreateNamespace from "../CreateNamespace";

const mockGetConfig = vi.mocked(getConfig);
const mockUseCreateNamespace = vi.mocked(useCreateNamespace);
const mockUseSwitchNamespace = vi.mocked(useSwitchNamespace);

function makeCreateNs(
  overrides?: Partial<ReturnType<typeof useCreateNamespace>>,
) {
  return {
    mutateAsync: vi.fn<() => Promise<void>>().mockResolvedValue(undefined),
    isPending: false,
    error: null,
    reset: vi.fn(),
    ...overrides,
  } as unknown as ReturnType<typeof useCreateNamespace>;
}

beforeEach(() => {
  // Default: cloud mode → renders CloudForm
  mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
  mockUseCreateNamespace.mockReturnValue(makeCreateNs());
  mockUseSwitchNamespace.mockReturnValue({
    mutateAsync: vi.fn(),
    isPending: false,
  } as unknown as ReturnType<typeof useSwitchNamespace>);
});

afterEach(cleanup);

function renderComponent() {
  return render(<CreateNamespace />);
}

describe("CreateNamespace — CloudForm", () => {
  it("calls mutateAsync with the typed name on valid submit (no-op regression guard)", async () => {
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockResolvedValue(undefined);
    mockUseCreateNamespace.mockReturnValue(makeCreateNs({ mutateAsync }));

    const user = userEvent.setup();
    renderComponent();

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith("my-ns"));
  });

  it("shows 'A namespace with this name already exists.' on 409", async () => {
    const sdkError = { status: 409 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
    mockUseCreateNamespace.mockReturnValue(makeCreateNs({ mutateAsync }));

    const user = userEvent.setup();
    renderComponent();

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() =>
      expect(
        screen.getByText("A namespace with this name already exists."),
      ).toBeInTheDocument(),
    );
  });

  it("shows the limit/permission message on 403", async () => {
    const sdkError = { status: 403 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
    mockUseCreateNamespace.mockReturnValue(makeCreateNs({ mutateAsync }));

    const user = userEvent.setup();
    renderComponent();

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() =>
      expect(
        screen.getByText(
          "You have reached the namespace limit or do not have permission.",
        ),
      ).toBeInTheDocument(),
    );
  });

  it("shows the invalid-name message on 400", async () => {
    const sdkError = { status: 400 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
    mockUseCreateNamespace.mockReturnValue(makeCreateNs({ mutateAsync }));

    const user = userEvent.setup();
    renderComponent();

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() =>
      expect(
        screen.getByText("The namespace name is invalid."),
      ).toBeInTheDocument(),
    );
  });

  it("shows the generic fallback message on 500", async () => {
    const sdkError = { status: 500 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
    mockUseCreateNamespace.mockReturnValue(makeCreateNs({ mutateAsync }));

    const user = userEvent.setup();
    renderComponent();

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() =>
      expect(
        screen.getByText("An unexpected error occurred. Please try again."),
      ).toBeInTheDocument(),
    );
  });

  it("clears the error text from the DOM when the user types after a failed submission", async () => {
    const sdkError = { status: 409 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
    mockUseCreateNamespace.mockReturnValue(makeCreateNs({ mutateAsync }));

    const user = userEvent.setup();
    renderComponent();

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() =>
      expect(
        screen.getByText("A namespace with this name already exists."),
      ).toBeInTheDocument(),
    );

    // Now type more — error must disappear
    await user.type(screen.getByPlaceholderText("my-namespace"), "x");

    expect(
      screen.queryByText("A namespace with this name already exists."),
    ).not.toBeInTheDocument();
  });
});
