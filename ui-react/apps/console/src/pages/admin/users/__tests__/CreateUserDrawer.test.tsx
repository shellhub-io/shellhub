import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import CreateUserDrawer from "../CreateUserDrawer";
import { useCreateUser } from "../../../../hooks/useAdminUserMutations";
vi.mock("../../../../hooks/useAdminUserMutations", () => ({
  useCreateUser: vi.fn(),
}));

vi.mock("../../../../utils/styles", () => ({
  LABEL: "label",
  INPUT: "input",
}));

vi.mock("../../../../components/common/Drawer", async () => ({
  default: (await import("./mocks")).MockDrawer,
}));

const mockMutateAsync = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useCreateUser).mockReturnValue({
    mutateAsync: mockMutateAsync,
  } as never);
});

function renderDrawer(
  overrides: Partial<{ open: boolean; onClose: () => void }> = {},
) {
  const defaults = { open: true, onClose: vi.fn() };
  const props = { ...defaults, ...overrides };
  return { onClose: props.onClose, ...render(<CreateUserDrawer {...props} />) };
}

async function fillForm({
  name = "Alice",
  username = "alice",
  email = "alice@example.com",
  password = "pass123",
}: Partial<{
  name: string;
  username: string;
  email: string;
  password: string;
}> = {}) {
  if (name) await userEvent.type(screen.getByLabelText(/^name$/i), name);
  if (username)
    await userEvent.type(screen.getByLabelText(/^username$/i), username);
  if (email) await userEvent.type(screen.getByLabelText(/^email$/i), email);
  if (password)
    await userEvent.type(screen.getByLabelText(/^password$/i), password);
}

describe("CreateUserDrawer", () => {
  describe("rendering — closed", () => {
    it("renders nothing when open is false", () => {
      renderDrawer({ open: false });
      expect(screen.queryByText("Create User")).not.toBeInTheDocument();
    });
  });

  describe("rendering — open", () => {
    it("renders the 'Create User' title", () => {
      renderDrawer();
      expect(
        screen.getByRole("heading", { name: "Create User" }),
      ).toBeInTheDocument();
    });

    it("renders the Name input field", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^name$/i)).toBeInTheDocument();
    });

    it("renders the Username input field", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^username$/i)).toBeInTheDocument();
    });

    it("renders the Email input field", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument();
    });

    it("renders the Password input field", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
    });

    it("renders the 'Create User' submit button", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /create user/i }),
      ).toBeInTheDocument();
    });

    it("renders the Cancel button", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /cancel/i }),
      ).toBeInTheDocument();
    });

    it("submit button is disabled when form is empty", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /create user/i }),
      ).toBeDisabled();
    });

    it("password field is of type password by default", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^password$/i)).toHaveAttribute(
        "type",
        "password",
      );
    });
  });

  describe("form enabling", () => {
    it("enables submit button when all required fields are filled", async () => {
      renderDrawer();
      await fillForm();
      expect(
        screen.getByRole("button", { name: /create user/i }),
      ).not.toBeDisabled();
    });

    it("keeps submit disabled when name is missing", async () => {
      renderDrawer();
      await fillForm({ name: "" });
      expect(
        screen.getByRole("button", { name: /create user/i }),
      ).toBeDisabled();
    });

    it("keeps submit disabled when username is missing", async () => {
      renderDrawer();
      await fillForm({ username: "" });
      expect(
        screen.getByRole("button", { name: /create user/i }),
      ).toBeDisabled();
    });

    it("keeps submit disabled when email is missing", async () => {
      renderDrawer();
      await fillForm({ email: "" });
      expect(
        screen.getByRole("button", { name: /create user/i }),
      ).toBeDisabled();
    });

    it("keeps submit disabled when password is missing", async () => {
      renderDrawer();
      await fillForm({ password: "" });
      expect(
        screen.getByRole("button", { name: /create user/i }),
      ).toBeDisabled();
    });
  });

  describe("password visibility toggle", () => {
    it("shows password in plaintext when Show password button is clicked", async () => {
      renderDrawer();
      await userEvent.click(
        screen.getByRole("button", { name: /show password/i }),
      );
      expect(screen.getByLabelText(/^password$/i)).toHaveAttribute(
        "type",
        "text",
      );
    });

    it("hides password again when Hide password button is clicked", async () => {
      renderDrawer();
      await userEvent.click(
        screen.getByRole("button", { name: /show password/i }),
      );
      await userEvent.click(
        screen.getByRole("button", { name: /hide password/i }),
      );
      expect(screen.getByLabelText(/^password$/i)).toHaveAttribute(
        "type",
        "password",
      );
    });
  });

  describe("namespace limit controls", () => {
    it("does not show namespace limit sub-options by default", () => {
      renderDrawer();
      expect(
        screen.queryByLabelText(/disable namespace creation/i),
      ).not.toBeInTheDocument();
    });

    it("shows sub-options when 'Set namespace creation limit' is checked", async () => {
      renderDrawer();
      await userEvent.click(
        screen.getByLabelText(/set namespace creation limit/i),
      );
      expect(
        screen.getByLabelText(/disable namespace creation/i),
      ).toBeInTheDocument();
    });

    it("shows max namespaces input when limit is enabled but disable is unchecked", async () => {
      renderDrawer();
      await userEvent.click(
        screen.getByLabelText(/set namespace creation limit/i),
      );
      expect(screen.getByLabelText(/max namespaces/i)).toBeInTheDocument();
    });

    it("hides max namespaces input when 'Disable namespace creation' is checked", async () => {
      renderDrawer();
      await userEvent.click(
        screen.getByLabelText(/set namespace creation limit/i),
      );
      await userEvent.click(
        screen.getByLabelText(/disable namespace creation/i),
      );
      expect(
        screen.queryByLabelText(/max namespaces/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("admin checkbox", () => {
    it("renders 'Admin user' checkbox unchecked by default", () => {
      renderDrawer();
      expect(screen.getByLabelText(/admin user/i)).not.toBeChecked();
    });
  });

  describe("submit — success", () => {
    it("calls mutateAsync with the correct payload", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          body: expect.objectContaining({
            name: "Alice",
            username: "alice",
            email: "alice@example.com",
            password: "pass123",
            admin: false,
          }),
        });
      });
    });

    it("calls onClose after successful creation", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("sends max_namespaces as undefined when limit is not enabled", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          body: expect.objectContaining({ max_namespaces: undefined }),
        });
      });
    });

    it("sends max_namespaces as 0 when namespace creation is disabled", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByLabelText(/set namespace creation limit/i),
      );
      await userEvent.click(
        screen.getByLabelText(/disable namespace creation/i),
      );
      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          body: expect.objectContaining({ max_namespaces: 0 }),
        });
      });
    });
  });

  describe("submit — error handling", () => {
    it("shows conflict error message for 409 responses", async () => {
      mockMutateAsync.mockRejectedValue({ status: 409 });
      renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );

      await waitFor(() => {
        expect(screen.getByText(/already exists/i)).toBeInTheDocument();
      });
    });

    it("shows generic error for 400 responses", async () => {
      mockMutateAsync.mockRejectedValue({ status: 400 });
      renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );

      await waitFor(() => {
        expect(screen.getByText(/failed to create user/i)).toBeInTheDocument();
      });
    });

    it("shows generic error for unexpected failures", async () => {
      mockMutateAsync.mockRejectedValue(new Error("network error"));
      renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );

      await waitFor(() => {
        expect(screen.getByText(/failed to create user/i)).toBeInTheDocument();
      });
    });

    it("renders error with role='alert'", async () => {
      mockMutateAsync.mockRejectedValue(new Error("network error"));
      renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });

    it("does not call onClose when creation fails", async () => {
      mockMutateAsync.mockRejectedValue(new Error("network error"));
      const { onClose } = renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );

      await waitFor(() => screen.getByRole("alert"));
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const { onClose } = renderDrawer();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("does not call mutateAsync when Cancel is clicked", async () => {
      renderDrawer();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });
  });

  describe("state reset on reopen", () => {
    it("clears the name field when closed then reopened", async () => {
      const { rerender } = renderDrawer();
      await userEvent.type(screen.getByLabelText(/^name$/i), "Alice");

      rerender(<CreateUserDrawer open={false} onClose={vi.fn()} />);
      rerender(<CreateUserDrawer open={true} onClose={vi.fn()} />);

      expect(screen.getByLabelText(/^name$/i)).toHaveValue("");
    });

    it("clears any error when closed then reopened", async () => {
      mockMutateAsync.mockRejectedValue(new Error("fail"));
      const { rerender } = renderDrawer();
      await fillForm();
      await userEvent.click(
        screen.getByRole("button", { name: /create user/i }),
      );
      await waitFor(() => screen.getByRole("alert"));

      rerender(<CreateUserDrawer open={false} onClose={vi.fn()} />);
      rerender(<CreateUserDrawer open={true} onClose={vi.fn()} />);

      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });
});
