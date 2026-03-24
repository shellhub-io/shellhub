import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import AdminLicense from "../License";
import { useAdminLicense } from "../../../hooks/useAdminLicense";
import { useUploadLicense } from "../../../hooks/useUploadLicense";

vi.mock("../../../hooks/useAdminLicense", () => ({
  useAdminLicense: vi.fn(),
}));
vi.mock("../../../hooks/useUploadLicense", () => ({
  useUploadLicense: vi.fn(),
}));

// navigator.clipboard is not available in jsdom
Object.assign(navigator, {
  clipboard: { writeText: vi.fn().mockResolvedValue(undefined) },
});

const validLicense = {
  id: "xxx-xxx",
  expired: false,
  about_to_expire: false,
  grace_period: false,
  issued_at: 1704110400,
  starts_at: 1704110400,
  expires_at: 1735732800,
  allowed_regions: [] as string[],
  customer: { id: "cust-xxx", name: "Test Customer", email: "test@example.com", company: "Test Co" },
  features: { devices: -1, session_recording: true, firewall_rules: true, reports: false, login_link: false, billing: false },
};
const expiredLicense = { ...validLicense, expired: true, grace_period: false };
const aboutToExpireLicense = { ...validLicense, about_to_expire: true };
const gracePeriodLicense = { ...validLicense, expired: true, grace_period: true };
const regionalLicense = { ...validLicense, allowed_regions: ["BR", "US"] };

const mockMutateAsync = vi.fn();

function setupHooks({
  data,
  isLoading = false,
  isError = false,
  error,
  isPending = false,
}: {
  data?: object;
  isLoading?: boolean;
  isError?: boolean;
  error?: object;
  isPending?: boolean;
}) {
  vi.mocked(useAdminLicense).mockReturnValue({ data, isLoading, isError, error } as never);
  vi.mocked(useUploadLicense).mockReturnValue({ mutateAsync: mockMutateAsync, isPending } as never);
}

function renderPage() {
  const result = render(
    <MemoryRouter>
      <AdminLicense />
    </MemoryRouter>,
  );
  const fileInput = () => result.container.querySelector<HTMLInputElement>("#license-file")!;
  return { ...result, fileInput };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("AdminLicense", () => {
  describe("loading state", () => {
    it("renders spinner with role='status'", () => {
      setupHooks({ isLoading: true });
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders error message with role='alert' for non-400 errors", () => {
      setupHooks({ isError: true, error: { status: 500 } });
      renderPage();
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("Failed to load license information")).toBeInTheDocument();
    });

    it("shows no-license info alert and upload section when API returns 400", () => {
      setupHooks({ isError: true, error: { status: 400 } });
      renderPage();
      expect(screen.getByText("You do not have an installed license")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /choose a \.dat file/i })).toBeInTheDocument();
      expect(screen.queryByText("Failed to load license information")).not.toBeInTheDocument();
    });
  });

  describe("no data (query disabled)", () => {
    it("renders upload section but no license details", () => {
      setupHooks({ data: undefined });
      renderPage();
      expect(screen.getByRole("button", { name: /choose a \.dat file/i })).toBeInTheDocument();
      expect(screen.queryByText("License Information")).not.toBeInTheDocument();
    });
  });

  describe("status alerts", () => {
    it("shows info alert when no license (data is empty object)", () => {
      setupHooks({ data: {} });
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("You do not have an installed license")).toBeInTheDocument();
    });

    it("shows info alert when about_to_expire", () => {
      setupHooks({ data: aboutToExpireLicense });
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Your license is about to expire!")).toBeInTheDocument();
    });

    it("shows warning when expired + grace period", () => {
      setupHooks({ data: gracePeriodLicense });
      renderPage();
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText(/grace period/i)).toBeInTheDocument();
    });

    it("shows error when expired without grace period", () => {
      setupHooks({ data: expiredLicense });
      renderPage();
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("Your license has expired!")).toBeInTheDocument();
    });

    it("shows no alert when license is valid", () => {
      setupHooks({ data: validLicense });
      renderPage();
      expect(screen.queryByText(/license has expired/i)).not.toBeInTheDocument();
      expect(screen.queryByText(/about to expire/i)).not.toBeInTheDocument();
      expect(screen.queryByText(/do not have an installed license/i)).not.toBeInTheDocument();
    });
  });

  describe("license details", () => {
    it("does not render license details section when no license", () => {
      setupHooks({ data: {} });
      renderPage();
      expect(screen.queryByText("License Information")).not.toBeInTheDocument();
    });

    it("renders dates formatted correctly", () => {
      setupHooks({ data: validLicense });
      renderPage();
      // issued_at / starts_at: 1704110400 → Jan 1, 2024 (appears twice); expires_at: 1735732800 → Jan 1, 2025
      const jan2024 = screen.getAllByText("Jan 1, 2024");
      expect(jan2024.length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText("Jan 1, 2025")).toBeInTheDocument();
    });

    it("shows 'Now' for -1 timestamps", () => {
      const licenseWithNow = { ...validLicense, issued_at: -1, starts_at: -1 };
      setupHooks({ data: licenseWithNow });
      renderPage();
      const nowElements = screen.getAllByText("Now");
      expect(nowElements.length).toBeGreaterThanOrEqual(2);
    });

    it("shows 'Global' when allowed_regions is empty", () => {
      setupHooks({ data: validLicense });
      renderPage();
      expect(screen.getByText("Global")).toBeInTheDocument();
    });

    it("shows region list when regions are non-empty", () => {
      setupHooks({ data: regionalLicense });
      renderPage();
      expect(screen.getByText("BR, US")).toBeInTheDocument();
    });
  });

  describe("license owner", () => {
    it("displays customer fields", () => {
      setupHooks({ data: validLicense });
      renderPage();
      expect(screen.getByText("cust-xxx")).toBeInTheDocument();
      expect(screen.getByText("Test Customer")).toBeInTheDocument();
      expect(screen.getByText("test@example.com")).toBeInTheDocument();
      expect(screen.getByText("Test Co")).toBeInTheDocument();
    });

    it("renders copy button for customer ID", () => {
      setupHooks({ data: validLicense });
      renderPage();
      expect(screen.getByRole("button", { name: "Copy" })).toBeInTheDocument();
    });
  });

  describe("license features", () => {
    it("shows 'Unlimited' for devices = -1", () => {
      setupHooks({ data: validLicense });
      renderPage();
      expect(screen.getByText("Unlimited")).toBeInTheDocument();
    });

    it("renders check icon for enabled boolean features", () => {
      setupHooks({ data: validLicense });
      renderPage();
      // session_recording: true → "Included"; firewall_rules: true → "Included"
      const included = screen.getAllByLabelText("Included");
      expect(included.length).toBeGreaterThanOrEqual(2);
    });

    it("renders cross icon for disabled boolean features", () => {
      setupHooks({ data: validLicense });
      renderPage();
      // billing: false → "Not included"
      const notIncluded = screen.getAllByLabelText("Not included");
      expect(notIncluded.length).toBeGreaterThanOrEqual(1);
    });

    it("does not render login_link or reports features", () => {
      setupHooks({ data: validLicense });
      renderPage();
      expect(screen.queryByText("Login link")).not.toBeInTheDocument();
      expect(screen.queryByText("Reports")).not.toBeInTheDocument();
    });
  });

  describe("license upload", () => {
    it("renders drop zone and hidden file input", () => {
      setupHooks({ data: {} });
      const { fileInput } = renderPage();
      expect(screen.getByRole("button", { name: /choose a \.dat file/i })).toBeInTheDocument();
      expect(fileInput()).toBeInTheDocument();
    });

    it("upload button is disabled by default (no file selected)", () => {
      setupHooks({ data: {} });
      renderPage();
      expect(screen.getByRole("button", { name: /upload/i })).toBeDisabled();
    });

    it("shows validation error for wrong file extension", () => {
      setupHooks({ data: {} });
      const { fileInput } = renderPage();
      const badFile = new File(["content"], "license.txt", { type: "text/plain" });
      fireEvent.change(fileInput(), { target: { files: [badFile] } });
      expect(screen.getByText("Only .dat files are allowed")).toBeInTheDocument();
    });

    it("shows remove button and clears file on click", async () => {
      setupHooks({ data: {} });
      const { fileInput } = renderPage();
      const validFile = new File(["content"], "license.dat", { type: "application/octet-stream" });
      await userEvent.upload(fileInput(), validFile);
      const removeBtn = screen.getByRole("button", { name: /remove selected file/i });
      expect(removeBtn).toBeInTheDocument();
      await userEvent.click(removeBtn);
      expect(screen.getByRole("button", { name: /upload/i })).toBeDisabled();
    });

    it("calls mutateAsync when upload button is clicked with valid file", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      setupHooks({ data: {} });
      const { fileInput } = renderPage();
      const validFile = new File(["license-content"], "license.dat", { type: "application/octet-stream" });
      await userEvent.upload(fileInput(), validFile);
      const uploadBtn = screen.getByRole("button", { name: /upload/i });
      expect(uploadBtn).not.toBeDisabled();
      await userEvent.click(uploadBtn);
      expect(mockMutateAsync).toHaveBeenCalledWith({ body: { file: validFile } });
    });

    it("shows success message after upload", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      setupHooks({ data: {} });
      const { fileInput } = renderPage();
      const validFile = new File(["license-content"], "license.dat", { type: "application/octet-stream" });
      await userEvent.upload(fileInput(), validFile);
      await userEvent.click(screen.getByRole("button", { name: /upload/i }));
      await waitFor(() =>
        expect(screen.getByText("License uploaded successfully.")).toBeInTheDocument(),
      );
    });

    it("shows error message on failed upload", async () => {
      mockMutateAsync.mockRejectedValue(new Error("upload failed"));
      setupHooks({ data: {} });
      const { fileInput } = renderPage();
      const validFile = new File(["license-content"], "license.dat", { type: "application/octet-stream" });
      await userEvent.upload(fileInput(), validFile);
      await userEvent.click(screen.getByRole("button", { name: /upload/i }));
      await waitFor(() =>
        expect(screen.getByText("Failed to upload the license.")).toBeInTheDocument(),
      );
    });
  });
});
