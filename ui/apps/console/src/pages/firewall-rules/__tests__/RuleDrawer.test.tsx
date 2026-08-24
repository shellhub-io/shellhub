import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import RuleDrawer from "../RuleDrawer";
import type { FirewallRulesResponse } from "@/client";
import { mockSdkResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { mockFirewallRule } from "@/tests/factories";
import { mockTags } from "@/tests/mockTags";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    createFirewallRule: vi.fn(),
    updateFirewallRule: vi.fn(),
    getTags: vi.fn(),
  }),
);

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

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
  return render(<RuleDrawer {...merged} />, {
    wrapper: createTestWrapper(),
  });
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
  sdk.createFirewallRule.mockResolvedValue(mockSdkResponse(undefined));
  sdk.updateFirewallRule.mockResolvedValue(mockSdkResponse(undefined));
  mockTags(["production", "staging", "dev"]);
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
    mockTags(["a", "b", "c", "d"]);

    renderDrawer();

    await typePriority(user, "50");
    await user.click(screen.getByRole("radio", { name: /filter by tags/i }));

    const tagInput = screen.getByPlaceholderText("Search tags...");
    await user.click(tagInput);
    await user.click(await screen.findByRole("option", { name: "a" }));
    await user.click(screen.getByRole("option", { name: "b" }));
    await user.click(screen.getByRole("option", { name: "c" }));
    await user.click(screen.getByRole("option", { name: "d" }));

    expect(screen.getAllByRole("button", { name: /remove tag/i })).toHaveLength(
      3,
    );
    await waitFor(() => expect(getConfirmButton()).not.toBeDisabled());
  });

  it("calls createFirewallRule with the correct body and calls onClose on success", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    renderDrawer({ onClose });

    await typePriority(user, "42");
    await user.click(getConfirmButton());

    await waitFor(() =>
      expect(sdk.createFirewallRule).toHaveBeenCalledWith(
        expect.objectContaining({
          body: {
            priority: 42,
            action: "allow",
            active: true,
            source_ip: ".*",
            username: ".*",
            filter: { hostname: ".*" },
          },
        }),
      ),
    );
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });
});

describe("RuleDrawer — edit mode", () => {
  it("prefills all fields from editRule", async () => {
    const rule = mockFirewallRule({
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

  it("calls updateFirewallRule with the correct body on save", async () => {
    const user = userEvent.setup();
    const rule = mockFirewallRule({ priority: 10, action: "deny" });
    const onClose = vi.fn();
    renderDrawer({ editRule: rule, onClose });

    await user.click(getConfirmButton());

    await waitFor(() =>
      expect(sdk.updateFirewallRule).toHaveBeenCalledWith(
        expect.objectContaining({
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
      ),
    );
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });

  it("prefills a legacy priority '0' rule but blocks submit until it is positive", async () => {
    const user = userEvent.setup();
    const rule = mockFirewallRule({ priority: 0 });
    const onClose = vi.fn();
    renderDrawer({ editRule: rule, onClose });

    expect(screen.getByLabelText(/priority/i)).toHaveValue("0");
    await waitFor(() => expect(getConfirmButton()).toBeDisabled());

    await typePriority(user, "5");
    await user.click(getConfirmButton());

    await waitFor(() =>
      expect(sdk.updateFirewallRule).toHaveBeenCalledWith(
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
    sdk.createFirewallRule.mockRejectedValue(new Error("Server error"));
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
