import { type ReactNode } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { FirewallRulesResponse } from "@/client";

const mockCreateMutateAsync = vi.fn();
const mockUpdateMutateAsync = vi.fn();

vi.mock("@/hooks/useFirewallRuleMutations", () => ({
  useCreateFirewallRule: () => ({
    mutateAsync: mockCreateMutateAsync,
    isPending: false,
  }),
  useUpdateFirewallRule: () => ({
    mutateAsync: mockUpdateMutateAsync,
    isPending: false,
  }),
}));

vi.mock("@/hooks/useTags", () => ({
  useTags: vi.fn(),
}));

vi.mock("@/components/common/Drawer", () => ({
  default: ({
    open,
    onClose,
    title,
    children,
    footer,
  }: {
    open: boolean;
    onClose: () => void;
    title: string;
    children: ReactNode;
    footer?: ReactNode;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog" aria-label={title}>
        <button type="button" onClick={onClose}>
          Close
        </button>
        {children}
        {footer}
      </div>
    );
  },
}));

import { useTags } from "@/hooks/useTags";
import RuleDrawer from "../RuleDrawer";

function makeRule(
  overrides: Partial<FirewallRulesResponse> = {},
): FirewallRulesResponse {
  return {
    id: "rule-1",
    tenant_id: "tenant-abc",
    priority: 100,
    action: "allow",
    active: true,
    source_ip: ".*",
    username: ".*",
    filter: { hostname: ".*", tags: [] },
    ...overrides,
  };
}

function renderDrawer(
  props: Partial<{
    open: boolean;
    editRule: FirewallRulesResponse | null;
    onClose: () => void;
  }> = {},
) {
  const merged = {
    open: true,
    editRule: null,
    onClose: vi.fn(),
    ...props,
  };
  return render(<RuleDrawer {...merged} />);
}

async function typePriority(
  user: ReturnType<typeof userEvent.setup>,
  value: string,
) {
  const input = screen.getByLabelText(/priority/i);
  await user.clear(input);
  if (value) await user.type(input, value);
}

function getConfirmButton() {
  return screen.getByRole("button", { name: /create rule|save changes/i });
}

beforeEach(() => {
  vi.clearAllMocks();
  mockCreateMutateAsync.mockResolvedValue(undefined);
  mockUpdateMutateAsync.mockResolvedValue(undefined);
  vi.mocked(useTags).mockReturnValue({
    tags: [{ name: "production" }, { name: "staging" }, { name: "dev" }],
    totalCount: 3,
    isLoading: false,
    error: null,
  } as never);
});

describe("RuleDrawer — create mode", () => {
  it("submit is disabled until a valid priority is entered", async () => {
    const user = userEvent.setup();
    renderDrawer();

    expect(getConfirmButton()).toBeDisabled();

    await typePriority(user, "100");

    await waitFor(() => expect(getConfirmButton()).not.toBeDisabled());
  });

  it("keeps submit disabled when source-IP regexp is invalid", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await typePriority(user, "50");

    const sourceIpGroup = screen.getByRole("radiogroup", {
      name: /source ip/i,
    });
    await user.click(
      within(sourceIpGroup).getByRole("radio", {
        name: /restrict with regexp/i,
      }),
    );

    const ipInput = await screen.findByPlaceholderText(/192/i);
    // fireEvent.change instead of user.type: "[" in user.type is parsed as a key descriptor
    fireEvent.change(ipInput, { target: { value: "[invalid(" } });

    await waitFor(() => expect(getConfirmButton()).toBeDisabled());
  });

  it("keeps submit disabled when tags mode has no tag selected", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await typePriority(user, "50");
    await user.click(screen.getByRole("radio", { name: /filter by tags/i }));

    await waitFor(() => expect(getConfirmButton()).toBeDisabled());
  });

  it("caps tag selection at 3, ignoring a 4th tag and keeping submit enabled", async () => {
    const user = userEvent.setup();
    vi.mocked(useTags).mockReturnValue({
      tags: [{ name: "a" }, { name: "b" }, { name: "c" }, { name: "d" }],
      totalCount: 4,
      isLoading: false,
      error: null,
    } as never);

    renderDrawer();

    await typePriority(user, "50");
    await user.click(screen.getByRole("radio", { name: /filter by tags/i }));

    const tagInput = screen.getByPlaceholderText("Search tags...");
    await user.click(tagInput);
    await user.click(screen.getByRole("button", { name: "a" }));
    await user.click(screen.getByRole("button", { name: "b" }));
    await user.click(screen.getByRole("button", { name: "c" }));
    // The selector hard-caps selection at 3, so this 4th click is ignored —
    // the drawer can never submit more than 3 tags (the schema's >3 rule is a
    // defensive backstop that the UI cannot reach).
    await user.click(screen.getByRole("button", { name: "d" }));

    expect(screen.getAllByRole("button", { name: /remove tag/i })).toHaveLength(
      3,
    );
    await waitFor(() => expect(getConfirmButton()).not.toBeDisabled());
  });

  it("calls createRule.mutateAsync with the correct body and calls onClose on success", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    renderDrawer({ onClose });

    await typePriority(user, "42");
    await user.click(getConfirmButton());

    await waitFor(() =>
      expect(mockCreateMutateAsync).toHaveBeenCalledWith({
        body: {
          priority: 42,
          action: "allow",
          active: true,
          source_ip: ".*",
          username: ".*",
          filter: { hostname: ".*" },
        },
      }),
    );
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });
});

describe("RuleDrawer — edit mode", () => {
  it("prefills all fields from editRule", async () => {
    const rule = makeRule({
      priority: 99,
      action: "deny",
      active: false,
      source_ip: "10\\.0\\..*",
      username: "admin",
      filter: { hostname: "web-.*", tags: [] },
    });
    renderDrawer({ editRule: rule });

    expect(screen.getByLabelText(/priority/i)).toHaveValue("99");
    expect(screen.getByRole("radio", { name: /deny/i })).toBeChecked();
    expect(screen.getByRole("switch", { name: /status/i })).toHaveAttribute(
      "aria-checked",
      "false",
    );
    expect(await screen.findByPlaceholderText(/192/i)).toHaveValue(
      "10\\.0\\..*",
    );
    expect(screen.getByPlaceholderText(/e\.g\. root/i)).toHaveValue("admin");
    expect(screen.getByPlaceholderText(/e\.g\. web-/i)).toHaveValue("web-.*");
  });

  it("calls updateRule.mutateAsync with the correct body on save", async () => {
    const user = userEvent.setup();
    const rule = makeRule({ priority: 10, action: "deny" });
    const onClose = vi.fn();
    renderDrawer({ editRule: rule, onClose });

    await user.click(getConfirmButton());

    await waitFor(() =>
      expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
        path: { id: "rule-1" },
        body: {
          priority: 10,
          action: "deny",
          active: true,
          source_ip: ".*",
          username: ".*",
          filter: { hostname: ".*" },
        },
      }),
    );
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });

  it("prefills a legacy priority '0' rule but blocks submit until it is positive", async () => {
    const user = userEvent.setup();
    const rule = makeRule({ priority: 0 });
    const onClose = vi.fn();
    renderDrawer({ editRule: rule, onClose });

    expect(screen.getByLabelText(/priority/i)).toHaveValue("0");
    // 0 is not a positive integer, so the drawer keeps submit disabled.
    await waitFor(() => expect(getConfirmButton()).toBeDisabled());

    await typePriority(user, "5");
    await user.click(getConfirmButton());

    await waitFor(() =>
      expect(mockUpdateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({ priority: 5 }),
        }),
      ),
    );
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });
});

describe("RuleDrawer — API rejection", () => {
  it("shows root error alert and does not call onClose when API rejects", async () => {
    const user = userEvent.setup();
    mockCreateMutateAsync.mockRejectedValue(new Error("Server error"));
    const onClose = vi.fn();
    renderDrawer({ onClose });

    await typePriority(user, "1");
    await user.click(getConfirmButton());

    await waitFor(() =>
      expect(screen.getByText(/server error/i)).toBeInTheDocument(),
    );
    expect(onClose).not.toHaveBeenCalled();
  });
});
