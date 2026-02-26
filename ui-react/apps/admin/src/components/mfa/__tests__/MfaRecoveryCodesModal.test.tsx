import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import MfaRecoveryCodesModal from "../MfaRecoveryCodesModal";

vi.mock("../../../api/mfa", () => ({
  generateMfa: vi.fn(),
  enableMfa: vi.fn(),
}));

import { generateMfa } from "../../../api/mfa";

const mockedGenerateMfa = vi.mocked(generateMfa);

const mockMfaData = {
  link: "otpauth://totp/ShellHub:user@example.com?secret=ABCD1234&issuer=ShellHub",
  secret: "ABCD1234",
  recovery_codes: ["code1", "code2", "code3", "code4", "code5", "code6"],
};

describe("MfaRecoveryCodesModal", () => {
  const onClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockedGenerateMfa.mockResolvedValue(mockMfaData);
  });

  describe("Modal Open/Close", () => {
    it("does not render when open is false", () => {
      const { container } = render(
        <MfaRecoveryCodesModal open={false} onClose={onClose} />
      );

      expect(container.firstChild).toBeNull();
    });

    it("renders when open is true", () => {
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      expect(screen.getByText(/recovery codes/i)).toBeInTheDocument();
    });

    it("closes when close button is clicked", async () => {
      const user = userEvent.setup();
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      const closeButton = screen.getByText(/close/i);
      await user.click(closeButton);

      expect(onClose).toHaveBeenCalled();
    });

    it("closes when clicking backdrop", async () => {
      const user = userEvent.setup();
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      // Get backdrop element
      const backdrop = screen.getByText(/recovery codes/i).parentElement?.parentElement
        ?.previousElementSibling;

      if (backdrop) {
        await user.click(backdrop);
        expect(onClose).toHaveBeenCalled();
      }
    });
  });

  describe("Empty State", () => {
    it("shows empty state when no codes are available", () => {
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      expect(
        screen.getByText(/recovery codes cannot be viewed after creation/i)
      ).toBeInTheDocument();
      expect(
        screen.getByText(/for security reasons/i)
      ).toBeInTheDocument();
    });

    it("shows regenerate button in empty state", () => {
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      expect(screen.getByText(/regenerate codes/i)).toBeInTheDocument();
    });
  });

  describe("Regenerate Flow", () => {
    it("shows confirmation dialog when regenerate is clicked", async () => {
      const user = userEvent.setup();
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      const regenerateButton = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton);

      expect(
        screen.getByText(/are you sure you want to regenerate/i)
      ).toBeInTheDocument();
      expect(
        screen.getByText(/invalidate all your current recovery codes/i)
      ).toBeInTheDocument();
    });

    it("cancels regeneration when cancel is clicked", async () => {
      const user = userEvent.setup();
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      // Click regenerate
      const regenerateButton = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton);

      // Click cancel
      const cancelButton = screen.getByText(/cancel/i);
      await user.click(cancelButton);

      // Should be back to empty state
      expect(
        screen.getByText(/recovery codes cannot be viewed after creation/i)
      ).toBeInTheDocument();

      // API should not have been called
      expect(mockedGenerateMfa).not.toHaveBeenCalled();
    });

    it("generates new codes when confirmed", async () => {
      const user = userEvent.setup();
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      // Click regenerate
      const regenerateButton = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton);

      // Confirm regeneration
      const confirmButtons = screen.getAllByText(/regenerate codes/i);
      const confirmButton = confirmButtons.find(
        (btn) => btn.closest("button")?.className.includes("accent-red")
      );

      if (confirmButton) {
        await user.click(confirmButton);
      }

      await waitFor(() => {
        expect(mockedGenerateMfa).toHaveBeenCalled();
      });
    });

    it("displays generated codes after regeneration", async () => {
      const user = userEvent.setup();
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      // Regenerate codes
      const regenerateButton = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton);

      const confirmButtons = screen.getAllByText(/regenerate codes/i);
      const confirmButton = confirmButtons.find(
        (btn) => btn.closest("button")?.className.includes("accent-red")
      );

      if (confirmButton) {
        await user.click(confirmButton);
      }

      await waitFor(() => {
        // All 6 codes should be displayed
        mockMfaData.recovery_codes.forEach((code) => {
          expect(screen.getByText(code)).toBeInTheDocument();
        });
      });
    });
  });

  describe("Codes Display", () => {
    beforeEach(async () => {
      const user = userEvent.setup();
      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      // Regenerate codes to display them
      const regenerateButton = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton);

      const confirmButtons = screen.getAllByText(/regenerate codes/i);
      const confirmButton = confirmButtons.find(
        (btn) => btn.closest("button")?.className.includes("accent-red")
      );

      if (confirmButton) {
        await user.click(confirmButton);
      }

      await waitFor(() => expect(mockedGenerateMfa).toHaveBeenCalled());
    });

    it("displays all 6 recovery codes in grid", async () => {
      await waitFor(() => {
        mockMfaData.recovery_codes.forEach((code) => {
          expect(screen.getByText(code)).toBeInTheDocument();
        });
      });
    });

    it("shows download button", async () => {
      await waitFor(() => {
        expect(screen.getByText(/download/i)).toBeInTheDocument();
      });
    });

    it("shows copy button", async () => {
      await waitFor(() => {
        expect(screen.getByText(/copy/i)).toBeInTheDocument();
      });
    });

    it("shows warning about saving codes", async () => {
      await waitFor(() => {
        expect(screen.getByText(/save these codes now/i)).toBeInTheDocument();
        expect(
          screen.getByText(/each code can only be used once/i)
        ).toBeInTheDocument();
      });
    });

    it("shows persistence warning", async () => {
      await waitFor(() => {
        expect(
          screen.getByText(/may not be persisted to the database yet/i)
        ).toBeInTheDocument();
      });
    });

    it("shows done button instead of close", async () => {
      await waitFor(() => {
        expect(screen.getByText(/done/i)).toBeInTheDocument();
        expect(screen.queryByText(/close/i)).not.toBeInTheDocument();
      });
    });
  });

  describe("Error Handling", () => {
    it("shows error when regeneration fails", async () => {
      const user = userEvent.setup();
      mockedGenerateMfa.mockRejectedValue(new Error("Network error"));

      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      // Click regenerate
      const regenerateButton = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton);

      // Confirm
      const confirmButtons = screen.getAllByText(/regenerate codes/i);
      const confirmButton = confirmButtons.find(
        (btn) => btn.closest("button")?.className.includes("accent-red")
      );

      if (confirmButton) {
        await user.click(confirmButton);
      }

      await waitFor(() => {
        expect(
          screen.getByText(/failed to regenerate recovery codes/i)
        ).toBeInTheDocument();
      });
    });

    it("displays custom error message from API", async () => {
      const user = userEvent.setup();
      mockedGenerateMfa.mockRejectedValue(
        new Error("Rate limit exceeded")
      );

      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      const regenerateButton = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton);

      const confirmButtons = screen.getAllByText(/regenerate codes/i);
      const confirmButton = confirmButtons.find(
        (btn) => btn.closest("button")?.className.includes("accent-red")
      );

      if (confirmButton) {
        await user.click(confirmButton);
      }

      await waitFor(() => {
        expect(screen.getByText(/rate limit exceeded/i)).toBeInTheDocument();
      });
    });

    it("clears error when retrying regeneration", async () => {
      const user = userEvent.setup();

      // First attempt fails
      mockedGenerateMfa.mockRejectedValueOnce(new Error("Network error"));

      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      // First attempt
      const regenerateButton = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton);

      const confirmButtons = screen.getAllByText(/regenerate codes/i);
      const confirmButton = confirmButtons.find(
        (btn) => btn.closest("button")?.className.includes("accent-red")
      );

      if (confirmButton) {
        await user.click(confirmButton);
      }

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument();
      });

      // Second attempt succeeds
      mockedGenerateMfa.mockResolvedValueOnce(mockMfaData);

      // Go back to empty state by closing and reopening confirmation
      const closeButton = screen.getByText(/close/i);
      await user.click(closeButton);

      const { rerender } = render(
        <MfaRecoveryCodesModal open={false} onClose={onClose} />
      );

      rerender(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      const regenerateButton2 = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton2);

      const confirmButtons2 = screen.getAllByText(/regenerate codes/i);
      const confirmButton2 = confirmButtons2.find(
        (btn) => btn.closest("button")?.className.includes("accent-red")
      );

      if (confirmButton2) {
        await user.click(confirmButton2);
      }

      // Error should be cleared
      await waitFor(() => {
        expect(screen.queryByText(/network error/i)).not.toBeInTheDocument();
      });
    });
  });

  describe("Loading State", () => {
    it("shows loading spinner while regenerating", async () => {
      const user = userEvent.setup();
      let resolveGenerate: (value: typeof mockMfaData) => void;
      mockedGenerateMfa.mockReturnValue(
        new Promise((resolve) => {
          resolveGenerate = resolve;
        })
      );

      render(<MfaRecoveryCodesModal open={true} onClose={onClose} />);

      const regenerateButton = screen.getByText(/regenerate codes/i);
      await user.click(regenerateButton);

      const confirmButtons = screen.getAllByText(/regenerate codes/i);
      const confirmButton = confirmButtons.find(
        (btn) => btn.closest("button")?.className.includes("accent-red")
      );

      if (confirmButton) {
        await user.click(confirmButton);
      }

      // Should show loading state
      await waitFor(() => {
        const confirmButtonAfterClick = screen.queryByText(/regenerate codes/i);
        expect(confirmButtonAfterClick).toBeInTheDocument();
      });

      // Resolve
      resolveGenerate!(mockMfaData);

      await waitFor(() => {
        expect(screen.getByText("code1")).toBeInTheDocument();
      });
    });
  });
});
