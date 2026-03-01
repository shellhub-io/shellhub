import React from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useVaultStore, DuplicateKeyError } from "../../../stores/vaultStore";
import KeyDrawer from "../KeyDrawer";
import type { VaultKeyEntry } from "../../../types/vault";

vi.mock("../../../stores/vaultStore", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../../stores/vaultStore")>();
  return {
    ...actual,
    useVaultStore: vi.fn(),
  };
});

vi.mock("../../../utils/ssh-keys", () => ({
  validatePrivateKey: vi.fn(),
  getFingerprint: vi.fn(),
}));

// KeyFileInput internally uses a hook + file input. To keep tests focused on
// KeyDrawer logic, we replace it with a simple textarea that forwards the same
// onChange/value interface that KeyDrawer expects.
vi.mock("../../../utils/styles", () => ({
  LABEL: "label-class",
  INPUT: "input-class",
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
        <button onClick={onClose}>Close Drawer</button>
        <div>{children as React.ReactNode}</div>
        {footer && <div>{footer as React.ReactNode}</div>}
      </div>
    );
  },
}));

vi.mock("../../../components/common/KeyFileInput", () => ({
  default: ({
    label,
    id,
    value,
    onChange,
    disabled,
    error,
  }: {
    label: string;
    id: string;
    value: string;
    onChange: (v: string) => void;
    disabled?: boolean;
    error?: string | null;
  }) => (
    <div>
      <label htmlFor={id}>{label}</label>
      <textarea
        id={id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        aria-invalid={!!error}
      />
      {error && <p role="alert">{error}</p>}
    </div>
  ),
}));

import { validatePrivateKey, getFingerprint } from "../../../utils/ssh-keys";

const mockAddKey = vi.fn();
const mockUpdateKey = vi.fn();

const mockEntry: VaultKeyEntry = {
  id: "key-1",
  name: "Production Server",
  data: "-----BEGIN OPENSSH PRIVATE KEY-----\nexisting\n-----END OPENSSH PRIVATE KEY-----",
  hasPassphrase: false,
  fingerprint: "aa:bb:cc:dd",
  createdAt: "2024-01-01T00:00:00Z",
  updatedAt: "2024-01-01T00:00:00Z",
};

const VALID_KEY =
  "-----BEGIN OPENSSH PRIVATE KEY-----\nvalid\n-----END OPENSSH PRIVATE KEY-----";

function setupStore() {
  vi.mocked(useVaultStore).mockReturnValue({
    addKey: mockAddKey,
    updateKey: mockUpdateKey,
  } as never);
}

beforeEach(() => {
  vi.clearAllMocks();
  setupStore();

  // Default: key is valid, unencrypted
  vi.mocked(validatePrivateKey).mockReturnValue({ valid: true, encrypted: false });
  vi.mocked(getFingerprint).mockReturnValue("aa:bb:cc:dd");
});

function renderDrawer(
  overrides: Partial<{ open: boolean; editKey: VaultKeyEntry | null; onClose: () => void }> = {},
) {
  const defaults = { open: true, editKey: null, onClose: vi.fn() };
  const props = { ...defaults, ...overrides };
  return { onClose: props.onClose, ...render(<KeyDrawer {...props} />) };
}

async function fillName(name: string) {
  await userEvent.type(screen.getByLabelText(/^name$/i), name);
}

async function fillKey(pem: string) {
  await userEvent.type(screen.getByLabelText(/private key/i), pem);
}

describe("KeyDrawer", () => {
  describe("rendering — add mode", () => {
    it("renders the 'Add Private Key' title when editKey is null", () => {
      renderDrawer();
      expect(screen.getByText("Add Private Key")).toBeInTheDocument();
    });

    it("renders empty name field initially", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^name$/i)).toHaveValue("");
    });

    it("renders 'Add Key' submit button", () => {
      renderDrawer();
      expect(screen.getByRole("button", { name: /add key/i })).toBeInTheDocument();
    });

    it("submit button is disabled when form is empty", () => {
      renderDrawer();
      expect(screen.getByRole("button", { name: /add key/i })).toBeDisabled();
    });

    it("does not render passphrase field initially", () => {
      renderDrawer();
      expect(screen.queryByLabelText(/passphrase/i)).not.toBeInTheDocument();
    });
  });

  describe("rendering — edit mode", () => {
    it("renders the 'Edit Private Key' title", () => {
      renderDrawer({ editKey: mockEntry });
      expect(screen.getByText("Edit Private Key")).toBeInTheDocument();
    });

    it("pre-fills the name field with the entry name", () => {
      renderDrawer({ editKey: mockEntry });
      expect(screen.getByLabelText(/^name$/i)).toHaveValue("Production Server");
    });

    it("renders 'Save Changes' submit button", () => {
      renderDrawer({ editKey: mockEntry });
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeInTheDocument();
    });

    it("key field is disabled in edit mode", () => {
      renderDrawer({ editKey: mockEntry });
      expect(screen.getByLabelText(/private key/i)).toBeDisabled();
    });
  });

  describe("form state reset on open", () => {
    it("clears fields when closed then reopened without editKey", async () => {
      const { rerender } = renderDrawer({ editKey: mockEntry });

      // Reopen without editKey
      rerender(<KeyDrawer open={false} editKey={null} onClose={vi.fn()} />);
      rerender(<KeyDrawer open editKey={null} onClose={vi.fn()} />);

      expect(screen.getByLabelText(/^name$/i)).toHaveValue("");
    });
  });

  describe("name field validation", () => {
    it("enables submit when both name and key are valid", async () => {
      renderDrawer();

      await fillKey(VALID_KEY);
      await fillName("My Key");

      expect(screen.getByRole("button", { name: /add key/i })).not.toBeDisabled();
    });

    it("clears name error when the user starts typing again", async () => {
      mockAddKey.mockRejectedValue(new DuplicateKeyError("name"));
      renderDrawer();

      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(screen.getByText(/name is already used/i)).toBeInTheDocument();
      });

      await userEvent.type(screen.getByLabelText(/^name$/i), "X");
      expect(screen.queryByText(/name is already used/i)).not.toBeInTheDocument();
    });
  });

  describe("key validation", () => {
    it("shows key error when key is invalid", async () => {
      vi.mocked(validatePrivateKey).mockReturnValue({
        valid: false,
        error: "Invalid private key format.",
      });

      renderDrawer();
      await fillKey("not-a-key");

      expect(screen.getByText(/invalid private key format/i)).toBeInTheDocument();
    });

    it("shows passphrase field when key is encrypted", async () => {
      vi.mocked(validatePrivateKey).mockReturnValue({ valid: true, encrypted: true });

      renderDrawer();
      await fillKey(VALID_KEY);

      expect(screen.getByLabelText(/passphrase/i)).toBeInTheDocument();
    });

    it("requires passphrase to be filled before enabling submit", async () => {
      vi.mocked(validatePrivateKey).mockReturnValue({ valid: true, encrypted: true });

      renderDrawer();
      await fillKey(VALID_KEY);
      await fillName("My Key");

      // Passphrase field visible but empty — submit must be disabled
      expect(screen.getByRole("button", { name: /add key/i })).toBeDisabled();
    });

    it("enables submit after providing passphrase for encrypted key", async () => {
      vi.mocked(validatePrivateKey).mockReturnValue({ valid: true, encrypted: true });

      renderDrawer();
      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.type(screen.getByLabelText(/passphrase/i), "secret");

      expect(screen.getByRole("button", { name: /add key/i })).not.toBeDisabled();
    });
  });

  describe("add flow", () => {
    it("calls addKey with correct payload on submit", async () => {
      mockAddKey.mockResolvedValue(undefined);
      renderDrawer();

      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(mockAddKey).toHaveBeenCalledWith({
          name: "My Key",
          data: VALID_KEY,
          hasPassphrase: false,
          fingerprint: "aa:bb:cc:dd",
        });
      });
    });

    it("calls onClose after successful add", async () => {
      mockAddKey.mockResolvedValue(undefined);
      const { onClose } = renderDrawer();

      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(onClose).toHaveBeenCalledTimes(1);
      });
    });

    it("calls addKey with passphrase fingerprint for encrypted key", async () => {
      vi.mocked(validatePrivateKey).mockReturnValue({ valid: true, encrypted: true });
      vi.mocked(getFingerprint).mockReturnValue("ee:ff:00:11");
      mockAddKey.mockResolvedValue(undefined);
      renderDrawer();

      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.type(screen.getByLabelText(/passphrase/i), "secret");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(getFingerprint).toHaveBeenCalledWith(VALID_KEY, "secret");
        expect(mockAddKey).toHaveBeenCalledWith(
          expect.objectContaining({ fingerprint: "ee:ff:00:11", hasPassphrase: true }),
        );
      });
    });
  });

  describe("edit flow", () => {
    it("calls updateKey with the entry id and updated name on submit", async () => {
      mockUpdateKey.mockResolvedValue(undefined);
      renderDrawer({ editKey: mockEntry });

      // Clear current name and type new one
      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "Updated Name");

      await userEvent.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() => {
        expect(mockUpdateKey).toHaveBeenCalledWith(
          "key-1",
          expect.objectContaining({ name: "Updated Name" }),
        );
      });
    });

    it("calls onClose after successful update", async () => {
      mockUpdateKey.mockResolvedValue(undefined);
      const { onClose } = renderDrawer({ editKey: mockEntry });

      await userEvent.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() => {
        expect(onClose).toHaveBeenCalledTimes(1);
      });
    });
  });

  describe("error states — duplicate key", () => {
    it("shows name error when DuplicateKeyError with field 'name'", async () => {
      mockAddKey.mockRejectedValue(new DuplicateKeyError("name"));
      renderDrawer();

      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(screen.getByText(/name is already used/i)).toBeInTheDocument();
      });
    });

    it("shows key error when DuplicateKeyError with field 'private_key'", async () => {
      mockAddKey.mockRejectedValue(new DuplicateKeyError("private_key"));
      renderDrawer();

      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(screen.getByText(/private key is already stored/i)).toBeInTheDocument();
      });
    });

    it("shows both name and key errors when DuplicateKeyError with field 'both'", async () => {
      mockAddKey.mockRejectedValue(new DuplicateKeyError("both"));
      renderDrawer();

      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(screen.getByText(/name is already used/i)).toBeInTheDocument();
        expect(screen.getByText(/private key is already stored/i)).toBeInTheDocument();
      });
    });
  });

  describe("error states — generic error", () => {
    it("shows a generic error message on unexpected failure", async () => {
      mockAddKey.mockRejectedValue(new Error("Network error"));
      renderDrawer();

      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument();
      });
    });

    it("shows passphrase error when getFingerprint throws KeyParseError for encrypted key", async () => {
      vi.mocked(validatePrivateKey).mockReturnValue({ valid: true, encrypted: true });
      const err = new Error("Bad key");
      (err as { name?: string }).name = "KeyParseError";
      vi.mocked(getFingerprint).mockImplementation(() => { throw err; });

      renderDrawer();
      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.type(screen.getByLabelText(/passphrase/i), "wrongpass");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(screen.getByText(/incorrect passphrase/i)).toBeInTheDocument();
      });
    });

    it("shows passphrase error when getFingerprint throws generic error for encrypted key", async () => {
      vi.mocked(validatePrivateKey).mockReturnValue({ valid: true, encrypted: true });
      vi.mocked(getFingerprint).mockImplementation(() => { throw new Error("Decryption failed"); });

      renderDrawer();
      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.type(screen.getByLabelText(/passphrase/i), "wrongpass");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(screen.getByText(/could not decrypt key/i)).toBeInTheDocument();
      });
    });

    it("shows key error when getFingerprint throws for unencrypted key", async () => {
      vi.mocked(getFingerprint).mockImplementation(() => { throw new Error("Unreadable"); });

      renderDrawer();
      await fillKey(VALID_KEY);
      await fillName("My Key");
      await userEvent.click(screen.getByRole("button", { name: /add key/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed to read private key/i)).toBeInTheDocument();
      });
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const { onClose } = renderDrawer();

      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });
});
