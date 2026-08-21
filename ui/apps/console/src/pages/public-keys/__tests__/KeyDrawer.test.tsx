import { useController, type Control, type Path } from "react-hook-form";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import KeyDrawer from "../KeyDrawer";
import type { PublicKeyResponse } from "@/client";
import type { KeyFormValues } from "../keySchema";
import { mockSdkResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { mockPublicKey, mockTag } from "@/tests/factories";
import { mockTags } from "@/tests/mockTags";

const mockCreatePublicKey = vi.hoisted(() => vi.fn());
const mockUpdatePublicKey = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    createPublicKey: mockCreatePublicKey,
    updatePublicKey: mockUpdatePublicKey,
    getTags: vi.fn(),
  };
});

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

vi.mock("../KeyDataInput", () => ({
  default: function MockKeyDataInput({
    name,
    control,
    disabled,
    onFileName,
  }: {
    name: Path<KeyFormValues>;
    control: Control<KeyFormValues>;
    disabled?: boolean;
    onFileName?: (name: string) => void;
  }) {
    const {
      field,
      fieldState: { error },
    } = useController({ name, control });
    return (
      <div>
        <label htmlFor="key-data">Public key data</label>
        <textarea
          id="key-data"
          value={String(field.value ?? "")}
          onChange={(e) => {
            field.onChange(e.target.value);
          }}
          disabled={disabled}
        />
        {error?.message && <p role="alert">{error.message}</p>}
        {onFileName && (
          <button
            type="button"
            data-testid="trigger-filename"
            onClick={() => onFileName("my-key-file")}
          >
            Trigger filename
          </button>
        )}
      </div>
    );
  },
}));

const Wrapper = createTestWrapper();

function renderDrawer(
  props: Partial<{
    open: boolean;
    editKey: PublicKeyResponse | null;
    onClose: () => void;
  }> = {},
) {
  const merged = { open: true, editKey: null, onClose: vi.fn(), ...props };
  return render(
    <Wrapper>
      <KeyDrawer {...merged} />
    </Wrapper>,
  );
}

function getSubmitButton() {
  return screen.getByRole("button", { name: /create key|save changes/i });
}

async function fillName(
  user: ReturnType<typeof userEvent.setup>,
  name: string,
) {
  const input = screen.getByPlaceholderText(/name used to identify/i);
  await user.clear(input);
  if (name) await user.type(input, name);
}

async function fillKeyData(
  user: ReturnType<typeof userEvent.setup>,
  key: string,
) {
  const ta = screen.getByLabelText(/public key data/i);
  await user.clear(ta);
  if (key) await user.type(ta, key);
}

const VALID_KEY = "ssh-rsa AAAAB3NzaC1yc2E test@host";

describe("KeyDrawer", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreatePublicKey.mockResolvedValue(mockSdkResponse(undefined));
    mockUpdatePublicKey.mockResolvedValue(mockSdkResponse(undefined));
    mockTags(["production", "linux", "staging"]);
  });

  describe("add mode UI", () => {
    it("shows 'New Public Key' title", () => {
      renderDrawer();
      expect(
        screen.getByRole("heading", { name: /new public key/i }),
      ).toBeInTheDocument();
    });

    it("shows 'Create Key' on the submit button", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /create key/i }),
      ).toBeInTheDocument();
    });

    it("submit is disabled when form is empty", () => {
      renderDrawer();
      expect(getSubmitButton()).toBeDisabled();
    });

    it("does not render when open is false", () => {
      renderDrawer({ open: false });
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("edit mode UI", () => {
    it("shows 'Edit Public Key' title", () => {
      renderDrawer({ editKey: mockPublicKey() });
      expect(
        screen.getByRole("heading", { name: /edit public key/i }),
      ).toBeInTheDocument();
    });

    it("shows 'Save Changes' on the submit button", () => {
      renderDrawer({ editKey: mockPublicKey() });
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeInTheDocument();
    });

    it("pre-fills the name field from editKey", () => {
      renderDrawer({ editKey: mockPublicKey({ name: "my-server-key" }) });
      expect(screen.getByPlaceholderText(/name used to identify/i)).toHaveValue(
        "my-server-key",
      );
    });

    it("key data textarea is disabled in edit mode", () => {
      renderDrawer({ editKey: mockPublicKey() });
      expect(screen.getByLabelText(/public key data/i)).toBeDisabled();
    });
  });

  describe("filter pre-population from editKey", () => {
    it("shows no hostname input when editKey filter is all (hostname '.*')", () => {
      renderDrawer({
        editKey: mockPublicKey({ filter: { hostname: ".*", tags: [] } }),
      });
      expect(
        screen.queryByPlaceholderText(/e\.g\. \.\*/i),
      ).not.toBeInTheDocument();
    });

    it("pre-populates hostname when editKey has a non-wildcard hostname", () => {
      renderDrawer({
        editKey: mockPublicKey({ filter: { hostname: "^prod-.*", tags: [] } }),
      });
      expect(screen.getByPlaceholderText(/e\.g\. \.\*/i)).toHaveValue(
        "^prod-.*",
      );
    });

    it("pre-populates tags when editKey has tags", () => {
      renderDrawer({
        editKey: mockPublicKey({
          filter: {
            tags: [mockTag({ name: "production" }), mockTag({ name: "linux" })],
          },
        }),
      });
      expect(screen.getByText("production")).toBeInTheDocument();
      expect(screen.getByText("linux")).toBeInTheDocument();
    });
  });

  describe("create happy-path — all devices", () => {
    it("sends { hostname: '.*' } and base64-encoded key data", async () => {
      const user = userEvent.setup();
      renderDrawer();

      await fillName(user, "test-key");
      await fillKeyData(user, VALID_KEY);
      await user.click(getSubmitButton());

      await waitFor(() =>
        expect(mockCreatePublicKey).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              data: btoa(VALID_KEY),
              filter: { hostname: ".*" },
            }),
          }),
        ),
      );
    });

    it("calls onClose after successful create", async () => {
      const user = userEvent.setup();
      const onClose = vi.fn();
      renderDrawer({ onClose });

      await fillName(user, "test-key");
      await fillKeyData(user, VALID_KEY);
      await user.click(getSubmitButton());

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });
  });

  describe("create happy-path — hostname filter", () => {
    it("sends { hostname } when hostname filter is selected", async () => {
      const user = userEvent.setup();
      renderDrawer();

      await fillName(user, "test-key");
      await fillKeyData(user, VALID_KEY);
      await user.click(
        screen.getByRole("radio", { name: /filter by hostname/i }),
      );
      await user.type(screen.getByPlaceholderText(/e\.g\. \.\*/i), "^prod-.*");
      await user.click(getSubmitButton());

      await waitFor(() =>
        expect(mockCreatePublicKey).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ filter: { hostname: "^prod-.*" } }),
          }),
        ),
      );
    });
  });

  describe("create happy-path — tags filter", () => {
    it("sends { tags: string[] } when tags filter is selected and tag chosen", async () => {
      const user = userEvent.setup();
      renderDrawer();

      await fillName(user, "test-key");
      await fillKeyData(user, VALID_KEY);
      await user.click(screen.getByRole("radio", { name: /filter by tags/i }));

      const tagInput = screen.getByPlaceholderText("Search tags...");
      await user.click(tagInput);
      await user.click(
        await screen.findByRole("option", { name: "production" }),
      );

      await user.click(getSubmitButton());

      await waitFor(() =>
        expect(mockCreatePublicKey).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ filter: { tags: ["production"] } }),
          }),
        ),
      );
    });
  });

  describe("409 error handling", () => {
    it("shows a 409 conflict alert on the key data field", async () => {
      const user = userEvent.setup();
      mockCreatePublicKey.mockRejectedValue({ status: 409 });
      renderDrawer();

      await fillName(user, "test-key");
      await fillKeyData(user, VALID_KEY);
      await user.click(getSubmitButton());

      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(
          /this public key already exists/i,
        ),
      );
    });
  });

  describe("generic root error", () => {
    it("shows the error message in a root error paragraph", async () => {
      const user = userEvent.setup();
      mockCreatePublicKey.mockRejectedValue(new Error("Server meltdown"));
      renderDrawer();

      await fillName(user, "test-key");
      await fillKeyData(user, VALID_KEY);
      await user.click(getSubmitButton());

      await waitFor(() =>
        expect(screen.getByText(/server meltdown/i)).toBeInTheDocument(),
      );
    });
  });

  describe("update happy-path", () => {
    it("calls updatePublicKey with fingerprint path param and updated name, no data field", async () => {
      const user = userEvent.setup();
      renderDrawer({
        editKey: mockPublicKey({ fingerprint: "ab:cd:ef", name: "old-name" }),
      });

      const nameInput = screen.getByPlaceholderText(/name used to identify/i);
      await user.clear(nameInput);
      await user.type(nameInput, "new-name");
      await user.click(getSubmitButton());

      await waitFor(() =>
        expect(mockUpdatePublicKey).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { fingerprint: "ab:cd:ef" },
            body: expect.objectContaining({ name: "new-name" }),
          }),
        ),
      );

      const callArg = mockUpdatePublicKey.mock.calls[0][0] as {
        body: Record<string, unknown>;
      };
      expect(callArg.body).not.toHaveProperty("data");
    });

    it("calls onClose after successful update", async () => {
      const user = userEvent.setup();
      const onClose = vi.fn();
      renderDrawer({ editKey: mockPublicKey(), onClose });

      await user.click(getSubmitButton());

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });
  });

  describe("reset on reopen", () => {
    it("clears fields when reopened in add mode after edit", () => {
      const { rerender } = renderDrawer({
        editKey: mockPublicKey({ name: "old" }),
      });

      rerender(
        <Wrapper>
          <KeyDrawer open={false} editKey={null} onClose={vi.fn()} />
        </Wrapper>,
      );
      rerender(
        <Wrapper>
          <KeyDrawer open editKey={null} onClose={vi.fn()} />
        </Wrapper>,
      );

      expect(screen.getByPlaceholderText(/name used to identify/i)).toHaveValue(
        "",
      );
    });

    it("updates pre-filled values when editKey changes on reopen", () => {
      const key1 = mockPublicKey({ name: "key-one" });
      const key2 = mockPublicKey({ name: "key-two" });
      const { rerender } = renderDrawer({ editKey: key1 });

      expect(screen.getByPlaceholderText(/name used to identify/i)).toHaveValue(
        "key-one",
      );

      rerender(
        <Wrapper>
          <KeyDrawer open={false} editKey={key2} onClose={vi.fn()} />
        </Wrapper>,
      );
      rerender(
        <Wrapper>
          <KeyDrawer open editKey={key2} onClose={vi.fn()} />
        </Wrapper>,
      );

      expect(screen.getByPlaceholderText(/name used to identify/i)).toHaveValue(
        "key-two",
      );
    });
  });

  describe("onFileName auto-fill", () => {
    it("auto-fills the name field when it is empty and a filename is provided", async () => {
      const user = userEvent.setup();
      renderDrawer();

      await user.click(screen.getByTestId("trigger-filename"));

      expect(screen.getByPlaceholderText(/name used to identify/i)).toHaveValue(
        "my-key-file",
      );
    });

    it("does not overwrite a name already typed by the user", async () => {
      const user = userEvent.setup();
      renderDrawer();

      await fillName(user, "existing-name");
      await user.click(screen.getByTestId("trigger-filename"));

      expect(screen.getByPlaceholderText(/name used to identify/i)).toHaveValue(
        "existing-name",
      );
    });
  });

  describe("name length validation", () => {
    it("keeps submit disabled when name exceeds 64 characters", async () => {
      const user = userEvent.setup();
      renderDrawer();

      const longName = "a".repeat(65);
      await fillName(user, longName);
      await fillKeyData(user, VALID_KEY);

      await waitFor(() => expect(getSubmitButton()).toBeDisabled());
    });

    it("enables submit when name is exactly 64 characters", async () => {
      const user = userEvent.setup();
      renderDrawer();

      const exactName = "a".repeat(64);
      await fillName(user, exactName);
      await fillKeyData(user, VALID_KEY);

      await waitFor(() => expect(getSubmitButton()).not.toBeDisabled());
    });
  });
});
