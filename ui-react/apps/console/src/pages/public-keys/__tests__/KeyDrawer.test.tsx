import React from "react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useCreatePublicKey, useUpdatePublicKey } from "../../../hooks/usePublicKeyMutations";
import KeyDrawer from "../KeyDrawer";
import type { PublicKey } from "../../../hooks/usePublicKeys";

vi.mock("../../../hooks/usePublicKeyMutations", () => ({
  useCreatePublicKey: vi.fn(),
  useUpdatePublicKey: vi.fn(),
}));

vi.mock("../../../utils/styles", () => ({
  LABEL: "label-class",
  INPUT: "input-class",
  INPUT_MONO: "input-mono-class",
}));

vi.mock("../../../components/common/Drawer", () => ({
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
    children: React.ReactNode;
    footer?: React.ReactNode;
  }) => {
    if (!open) return null;
    return (
      <div>
        <h2>{title}</h2>
        <button type="button" onClick={onClose}>Cancel</button>
        <div>{children}</div>
        {footer && <div>{footer as React.ReactNode}</div>}
      </div>
    );
  },
}));

vi.mock("../../../components/common/TagsSelector", () => ({
  default: ({
    selected,
    onChange,
    error,
  }: {
    selected: string[];
    onChange: (tags: string[]) => void;
    error?: string;
  }) => (
    <div>
      <button
        type="button"
        data-testid="add-tag-production"
        onClick={() => onChange([...selected, "production"])}
      >
        Add production tag
      </button>
      <button
        type="button"
        data-testid="add-tag-linux"
        onClick={() => onChange([...selected, "linux"])}
      >
        Add linux tag
      </button>
      {selected.map((t) => (
        <span key={t} data-testid={`tag-${t}`}>{t}</span>
      ))}
      {error && <p role="alert">{error}</p>}
    </div>
  ),
}));

vi.mock("../KeyDataInput", () => ({
  default: ({
    value,
    onChange,
    error,
    disabled,
    onFileName: _onFileName,
  }: {
    value: string;
    onChange: (v: string) => void;
    error?: string;
    disabled?: boolean;
    onFileName?: (name: string) => void;
  }) => (
    <div>
      <label htmlFor="key-data">Public key data</label>
      <textarea
        id="key-data"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
      />
      {error && <p role="alert">{error}</p>}
    </div>
  ),
}));

vi.mock("../../../utils/sshKeys", () => ({
  isPublicKeyValid: vi.fn(() => true),
}));

vi.mock("../../../components/icons", () => ({
  DevicesIcon: () => <svg data-testid="devices-icon" />,
}));

const mockCreateMutateAsync = vi.fn();
const mockUpdateMutateAsync = vi.fn();

function makeKey(overrides: Partial<PublicKey> = {}): PublicKey {
  return {
    name: "prod-key",
    fingerprint: "ab:cd:ef",
    created_at: "2024-01-01T00:00:00Z",
    tenant_id: "tenant-1",
    data: btoa("ssh-rsa AAAAB3 test"),
    filter: { hostname: ".*" },
    username: ".*",
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useCreatePublicKey).mockReturnValue({
    mutateAsync: mockCreateMutateAsync,
  } as never);
  vi.mocked(useUpdatePublicKey).mockReturnValue({
    mutateAsync: mockUpdateMutateAsync,
  } as never);
});

afterEach(cleanup);

function renderDrawer(
  overrides: Partial<{ open: boolean; editKey: PublicKey | null; onClose: () => void }> = {},
) {
  const defaults = { open: true, editKey: null, onClose: vi.fn() };
  const props = { ...defaults, ...overrides };
  return { onClose: props.onClose, ...render(<KeyDrawer {...props} />) };
}

const VALID_KEY = "ssh-rsa AAAAB3NzaC1yc2E test@host";

async function fillName(name: string) {
  await userEvent.type(screen.getByPlaceholderText(/name used to identify/i), name);
}

async function fillKeyData(key: string) {
  await userEvent.type(screen.getByLabelText(/public key data/i), key);
}

describe("KeyDrawer", () => {
  describe("rendering — add mode", () => {
    it("renders the 'New Public Key' title", () => {
      renderDrawer();
      expect(screen.getByText("New Public Key")).toBeInTheDocument();
    });

    it("renders 'Create Key' submit button", () => {
      renderDrawer();
      expect(screen.getByRole("button", { name: /create key/i })).toBeInTheDocument();
    });

    it("submit button is disabled when form is empty", () => {
      renderDrawer();
      expect(screen.getByRole("button", { name: /create key/i })).toBeDisabled();
    });

    it("does not render when open is false", () => {
      renderDrawer({ open: false });
      expect(screen.queryByText("New Public Key")).not.toBeInTheDocument();
    });
  });

  describe("rendering — edit mode", () => {
    it("renders the 'Edit Public Key' title", () => {
      renderDrawer({ editKey: makeKey() });
      expect(screen.getByText("Edit Public Key")).toBeInTheDocument();
    });

    it("renders 'Save Changes' submit button", () => {
      renderDrawer({ editKey: makeKey() });
      expect(screen.getByRole("button", { name: /save changes/i })).toBeInTheDocument();
    });

    it("pre-fills the name field with the key name", () => {
      renderDrawer({ editKey: makeKey({ name: "my-server-key" }) });
      expect(screen.getByPlaceholderText(/name used to identify/i)).toHaveValue("my-server-key");
    });

    it("key data input is disabled in edit mode", () => {
      renderDrawer({ editKey: makeKey() });
      expect(screen.getByLabelText(/public key data/i)).toBeDisabled();
    });
  });

  describe("filter initialization from editKey", () => {
    it("selects 'all devices' when hostname is '.*'", () => {
      renderDrawer({ editKey: makeKey({ filter: { hostname: ".*" } }) });
      // hostname input should not be visible — "all" is selected
      expect(screen.queryByPlaceholderText(/e\.g\. \.\*/i)).not.toBeInTheDocument();
    });

    it("selects hostname filter when editKey has a non-wildcard hostname", () => {
      renderDrawer({ editKey: makeKey({ filter: { hostname: "^prod-.*" } }) });
      expect(screen.getByPlaceholderText(/e\.g\. \.\*/i)).toHaveValue("^prod-.*");
    });

    it("selects tags filter and pre-populates tags when editKey has tags", () => {
      renderDrawer({
        editKey: makeKey({ filter: { tags: ["production", "linux"] } }),
      });
      expect(screen.getByTestId("tag-production")).toBeInTheDocument();
      expect(screen.getByTestId("tag-linux")).toBeInTheDocument();
    });

    it("selects 'all' when editKey has no tags and hostname is '.*'", () => {
      renderDrawer({ editKey: makeKey({ filter: { hostname: ".*" } }) });
      // Neither hostname input nor tags should be visible
      expect(screen.queryByPlaceholderText(/e\.g\. \.\*/i)).not.toBeInTheDocument();
      expect(screen.queryByTestId("tag-production")).not.toBeInTheDocument();
    });
  });

  describe("buildFilter — create flow", () => {
    it("sends { hostname: '.*' } when 'All devices' is selected", async () => {
      mockCreateMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await fillName("test-key");
      await fillKeyData(VALID_KEY);
      // "All devices" is selected by default
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        expect(mockCreateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ filter: { hostname: ".*" } }),
          }),
        );
      });
    });

    it("sends { hostname } when 'Filter by hostname' is selected", async () => {
      mockCreateMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await fillName("test-key");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /filter by hostname/i }));
      await userEvent.type(screen.getByPlaceholderText(/e\.g\. \.\*/i), "^prod-.*");
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        expect(mockCreateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ filter: { hostname: "^prod-.*" } }),
          }),
        );
      });
    });

    it("sends { tags: string[] } when 'Filter by tags' is selected", async () => {
      mockCreateMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await fillName("test-key");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /filter by tags/i }));
      await userEvent.click(screen.getByTestId("add-tag-production"));
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        expect(mockCreateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ filter: { tags: ["production"] } }),
          }),
        );
      });
    });

    it("sends tags as plain strings, not objects", async () => {
      mockCreateMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await fillName("test-key");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /filter by tags/i }));
      await userEvent.click(screen.getByTestId("add-tag-production"));
      await userEvent.click(screen.getByTestId("add-tag-linux"));
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        const arg = mockCreateMutateAsync.mock.calls[0][0] as {
          body: { filter: { tags: string[] } };
        };
        const tags = arg.body.filter.tags;
        expect(tags).toEqual(["production", "linux"]);
        // Each element must be a plain string, not an object
        tags.forEach((t) => expect(typeof t).toBe("string"));
      });
    });
  });

  describe("buildFilter — update flow", () => {
    it("sends { hostname } when updating with hostname filter", async () => {
      mockUpdateMutateAsync.mockResolvedValue(undefined);
      renderDrawer({
        editKey: makeKey({ filter: { hostname: "old-host" } }),
      });

      const hostnameInput = screen.getByPlaceholderText(/e\.g\. \.\*/i);
      await userEvent.clear(hostnameInput);
      await userEvent.type(hostnameInput, "new-host");
      await userEvent.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() => {
        expect(mockUpdateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ filter: { hostname: "new-host" } }),
          }),
        );
      });
    });

    it("sends { tags: string[] } when updating with tags filter", async () => {
      mockUpdateMutateAsync.mockResolvedValue(undefined);
      renderDrawer({
        editKey: makeKey({ filter: { tags: ["production"] } }),
      });

      await userEvent.click(screen.getByTestId("add-tag-linux"));
      await userEvent.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() => {
        expect(mockUpdateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              filter: { tags: ["production", "linux"] },
            }),
          }),
        );
      });
    });
  });

  describe("create flow", () => {
    it("calls createKey with base64-encoded key data", async () => {
      mockCreateMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await fillName("my-key");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        expect(mockCreateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ data: btoa(VALID_KEY) }),
          }),
        );
      });
    });

    it("calls createKey with trimmed name", async () => {
      mockCreateMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await fillName("  my-key  ");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        expect(mockCreateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ name: "my-key" }),
          }),
        );
      });
    });

    it("calls onClose after successful create", async () => {
      mockCreateMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDrawer();

      await fillName("my-key");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("shows key error on 409 conflict", async () => {
      mockCreateMutateAsync.mockRejectedValue({ status: 409 });
      renderDrawer();

      await fillName("my-key");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /this public key already exists/i,
        );
      });
    });

    it("shows generic error on unexpected failure", async () => {
      mockCreateMutateAsync.mockRejectedValue(new Error("Server error"));
      renderDrawer();

      await fillName("my-key");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        expect(screen.getByText(/server error/i)).toBeInTheDocument();
      });
    });
  });

  describe("update flow", () => {
    it("calls updateKey with fingerprint and updated name", async () => {
      mockUpdateMutateAsync.mockResolvedValue(undefined);
      renderDrawer({
        editKey: makeKey({ fingerprint: "ab:cd:ef", name: "old-name" }),
      });

      const nameInput = screen.getByPlaceholderText(/name used to identify/i);
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "new-name");
      await userEvent.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() => {
        expect(mockUpdateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { fingerprint: "ab:cd:ef" },
            body: expect.objectContaining({ name: "new-name" }),
          }),
        );
      });
    });

    it("calls onClose after successful update", async () => {
      mockUpdateMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDrawer({ editKey: makeKey() });

      await userEvent.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });
  });

  describe("username filter", () => {
    it("sends username '.*' when 'Allow any user' is selected", async () => {
      mockCreateMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await fillName("my-key");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        expect(mockCreateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ username: ".*" }),
          }),
        );
      });
    });

    it("sends specific username when 'Restrict by username' is selected", async () => {
      mockCreateMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await fillName("my-key");
      await fillKeyData(VALID_KEY);
      await userEvent.click(screen.getByRole("button", { name: /restrict by username/i }));
      await userEvent.type(screen.getByPlaceholderText(/e\.g\. root/i), "ubuntu");
      await userEvent.click(screen.getByRole("button", { name: /create key/i }));

      await waitFor(() => {
        expect(mockCreateMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ username: "ubuntu" }),
          }),
        );
      });
    });
  });

  describe("form reset on open", () => {
    it("clears fields when closed and reopened in add mode", () => {
      const { rerender } = renderDrawer({ editKey: makeKey({ name: "old" }) });

      rerender(<KeyDrawer open={false} editKey={null} onClose={vi.fn()} />);
      rerender(<KeyDrawer open editKey={null} onClose={vi.fn()} />);

      expect(screen.getByPlaceholderText(/name used to identify/i)).toHaveValue("");
    });
  });
});
