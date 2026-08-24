import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import AdminLicense from "../License";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getLicense: vi.fn(),
    sendLicense: vi.fn(),
  }),
);

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
  customer: {
    id: "cust-xxx",
    name: "Test Customer",
    email: "test@example.com",
    company: "Test Co",
  },
  features: {
    devices: -1,
    session_recording: true,
    firewall_rules: true,
    reports: false,
    login_link: false,
    billing: false,
  },
};
const expiredLicense = { ...validLicense, expired: true, grace_period: false };
const aboutToExpireLicense = { ...validLicense, about_to_expire: true };
const gracePeriodLicense = {
  ...validLicense,
  expired: true,
  grace_period: true,
};
const regionalLicense = { ...validLicense, allowed_regions: ["BR", "US"] };

function renderPage() {
  const result = render(
    <ClipboardProvider>
      <MemoryRouter>
        <AdminLicense />
      </MemoryRouter>
    </ClipboardProvider>,
    { wrapper: createTestWrapper() },
  );
  const fileInput = () =>
    result.container.querySelector<HTMLInputElement>("#license-file")!;
  return { ...result, fileInput };
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true });
  sdk.getLicense.mockResolvedValue(mockSdkResponse({}));
  sdk.sendLicense.mockResolvedValue(mockSdkResponse(undefined));
});

describe("AdminLicense", () => {
  describe("loading state", () => {
    it("renders spinner with role='status'", () => {
      sdk.getLicense.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders error message with role='alert' for non-400 errors", async () => {
      sdk.getLicense.mockRejectedValue(makeSdkError(500));
      renderPage();
      expect(await screen.findByRole("alert")).toBeInTheDocument();
      expect(
        screen.getByText("Failed to load license information"),
      ).toBeInTheDocument();
    });

    it("shows no-license info alert and upload section when 400 (no license stored)", async () => {
      sdk.getLicense.mockRejectedValue(makeSdkError(400));
      renderPage();
      expect(
        await screen.findByText("You do not have an installed license"),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /choose a \.dat file/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByText("Failed to load license information"),
      ).not.toBeInTheDocument();
    });
  });

  describe("no data (query disabled)", () => {
    it("renders upload section but no license details", () => {
      useAuthStore.setState({ isAdmin: false });
      renderPage();
      expect(screen.queryByText("License Information")).not.toBeInTheDocument();
    });
  });

  describe("status alerts", () => {
    it("shows info alert when no license (data is empty object)", async () => {
      renderPage();
      expect(
        await screen.findByText("You do not have an installed license"),
      ).toBeInTheDocument();
    });

    it("shows info alert when about_to_expire", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(aboutToExpireLicense));
      renderPage();
      expect(
        await screen.findByText("Your license is about to expire!"),
      ).toBeInTheDocument();
    });

    it("shows warning when expired + grace period", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(gracePeriodLicense));
      renderPage();
      expect(await screen.findByRole("alert")).toBeInTheDocument();
      expect(screen.getByText(/grace period/i)).toBeInTheDocument();
    });

    it("shows error when expired without grace period", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(expiredLicense));
      renderPage();
      expect(
        await screen.findByText("Your license has expired!"),
      ).toBeInTheDocument();
    });

    it("shows no alert when license is valid", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(validLicense));
      renderPage();
      await screen.findByText("License Information");
      expect(
        screen.queryByText(/license has expired/i),
      ).not.toBeInTheDocument();
      expect(screen.queryByText(/about to expire/i)).not.toBeInTheDocument();
      expect(
        screen.queryByText(/do not have an installed license/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("license details", () => {
    it("does not render license details section when no license", async () => {
      renderPage();
      await screen.findByText("You do not have an installed license");
      expect(screen.queryByText("License Information")).not.toBeInTheDocument();
    });

    it("renders dates formatted correctly", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(validLicense));
      renderPage();
      await screen.findByText("License Information");
      const jan2024 = screen.getAllByText("Jan 1, 2024");
      expect(jan2024.length).toBeGreaterThanOrEqual(1);
      expect(screen.getByText("Jan 1, 2025")).toBeInTheDocument();
    });

    it("shows 'Now' for -1 timestamps", async () => {
      const licenseWithNow = { ...validLicense, issued_at: -1, starts_at: -1 };
      sdk.getLicense.mockResolvedValue(mockSdkResponse(licenseWithNow));
      renderPage();
      await screen.findByText("License Information");
      const nowElements = screen.getAllByText("Now");
      expect(nowElements.length).toBeGreaterThanOrEqual(2);
    });

    it("shows 'Global' when allowed_regions is empty", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(validLicense));
      renderPage();
      expect(await screen.findByText("Global")).toBeInTheDocument();
    });

    it("shows region list when regions are non-empty", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(regionalLicense));
      renderPage();
      expect(await screen.findByText("BR, US")).toBeInTheDocument();
    });
  });

  describe("license owner", () => {
    it("displays customer fields", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(validLicense));
      renderPage();
      await screen.findByText("License Information");
      expect(screen.getByText("cust-xxx")).toBeInTheDocument();
      expect(screen.getByText("Test Customer")).toBeInTheDocument();
      expect(screen.getByText("test@example.com")).toBeInTheDocument();
      expect(screen.getByText("Test Co")).toBeInTheDocument();
    });

    it("renders copy button for customer ID", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(validLicense));
      renderPage();
      await screen.findByText("License Information");
      expect(screen.getByRole("button", { name: "Copy" })).toBeInTheDocument();
    });
  });

  describe("license features", () => {
    it("shows 'Unlimited' for devices = -1", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(validLicense));
      renderPage();
      expect(await screen.findByText("Unlimited")).toBeInTheDocument();
    });

    it("renders check icon for enabled boolean features", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(validLicense));
      renderPage();
      await screen.findByText("License Information");
      const included = screen.getAllByLabelText("Included");
      expect(included.length).toBeGreaterThanOrEqual(2);
    });

    it("renders cross icon for disabled boolean features", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(validLicense));
      renderPage();
      await screen.findByText("License Information");
      const notIncluded = screen.getAllByLabelText("Not included");
      expect(notIncluded.length).toBeGreaterThanOrEqual(1);
    });

    it("does not render login_link or reports features", async () => {
      sdk.getLicense.mockResolvedValue(mockSdkResponse(validLicense));
      renderPage();
      await screen.findByText("License Information");
      expect(screen.queryByText("Login link")).not.toBeInTheDocument();
      expect(screen.queryByText("Reports")).not.toBeInTheDocument();
    });
  });

  describe("license upload", () => {
    it("renders drop zone and hidden file input", async () => {
      const { fileInput } = renderPage();
      await screen.findByText("You do not have an installed license");
      expect(
        screen.getByRole("button", { name: /choose a \.dat file/i }),
      ).toBeInTheDocument();
      expect(fileInput()).toBeInTheDocument();
    });

    it("upload button is disabled by default (no file selected)", async () => {
      renderPage();
      await screen.findByText("You do not have an installed license");
      expect(
        screen.getByRole("button", { name: /upload license/i }),
      ).toBeDisabled();
    });

    it("shows validation error for wrong file extension", async () => {
      const { fileInput } = renderPage();
      await screen.findByText("You do not have an installed license");
      const badFile = new File(["content"], "license.txt", {
        type: "text/plain",
      });
      fireEvent.change(fileInput(), { target: { files: [badFile] } });
      expect(
        screen.getByText("Only .dat files are allowed"),
      ).toBeInTheDocument();
    });

    it("shows remove button and clears file on click", async () => {
      const { fileInput } = renderPage();
      await screen.findByText("You do not have an installed license");
      const validFile = new File(["content"], "license.dat", {
        type: "application/octet-stream",
      });
      await userEvent.upload(fileInput(), validFile);
      const removeBtn = screen.getByRole("button", { name: /remove file/i });
      expect(removeBtn).toBeInTheDocument();
      await userEvent.click(removeBtn);
      expect(
        screen.getByRole("button", { name: /upload license/i }),
      ).toBeDisabled();
    });

    it("calls sendLicense when upload button is clicked with valid file", async () => {
      sdk.sendLicense.mockResolvedValue(mockSdkResponse(undefined));
      const { fileInput } = renderPage();
      await screen.findByText("You do not have an installed license");
      const validFile = new File(["license-content"], "license.dat", {
        type: "application/octet-stream",
      });
      await userEvent.upload(fileInput(), validFile);
      const uploadBtn = screen.getByRole("button", {
        name: /upload license/i,
      });
      expect(uploadBtn).not.toBeDisabled();
      await userEvent.click(uploadBtn);
      await waitFor(() => {
        expect(sdk.sendLicense).toHaveBeenCalledWith(
          expect.objectContaining({
            body: { file: validFile },
          }),
        );
      });
    });

    it("shows success message after upload", async () => {
      sdk.sendLicense.mockResolvedValue(mockSdkResponse(undefined));
      const { fileInput } = renderPage();
      await screen.findByText("You do not have an installed license");
      const validFile = new File(["license-content"], "license.dat", {
        type: "application/octet-stream",
      });
      await userEvent.upload(fileInput(), validFile);
      await userEvent.click(
        screen.getByRole("button", { name: /upload license/i }),
      );
      await waitFor(() =>
        expect(
          screen.getByText("License uploaded successfully."),
        ).toBeInTheDocument(),
      );
    });

    it("shows error message on failed upload", async () => {
      sdk.sendLicense.mockRejectedValue(new Error("upload failed"));
      const { fileInput } = renderPage();
      await screen.findByText("You do not have an installed license");
      const validFile = new File(["license-content"], "license.dat", {
        type: "application/octet-stream",
      });
      await userEvent.upload(fileInput(), validFile);
      await userEvent.click(
        screen.getByRole("button", { name: /upload license/i }),
      );
      await waitFor(() =>
        expect(
          screen.getByText("Failed to upload the license."),
        ).toBeInTheDocument(),
      );
    });
  });
});
