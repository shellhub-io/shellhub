import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse, type JsonBodyType } from "msw";
import { server } from "@/tests/msw";
import AdminLicense from "../License";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";

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

function setLicense(data: JsonBodyType) {
  server.use(http.get("*/admin/api/license", () => HttpResponse.json(data)));
}

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

function datFile(content = "license-content") {
  return new File([content], "license.dat", {
    type: "application/octet-stream",
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true });
  server.use(
    http.get("*/admin/api/license", () => HttpResponse.json({})),
    http.post(
      "*/admin/api/license",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

describe("AdminLicense", () => {
  it("renders spinner with role='status' while loading", () => {
    server.use(http.get("*/admin/api/license", () => new Promise(() => {})));
    renderPage();
    expect(screen.getByRole("status")).toBeInTheDocument();
  });

  it("renders an error alert for non-400 errors", async () => {
    server.use(
      http.get("*/admin/api/license", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    renderPage();
    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(
      screen.getByText("Failed to load license information"),
    ).toBeInTheDocument();
  });

  it("shows no-license info alert and upload section when 400 (no license stored)", async () => {
    server.use(
      http.get("*/admin/api/license", () =>
        HttpResponse.json({}, { status: 400 }),
      ),
    );
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

  it("renders no license details when the query is disabled for non-admins", () => {
    useAuthStore.setState({ isAdmin: false });
    renderPage();
    expect(screen.queryByText("License Information")).not.toBeInTheDocument();
  });

  describe("status alerts", () => {
    it.each([
      ["no license", {}, "You do not have an installed license"],
      [
        "about to expire",
        aboutToExpireLicense,
        "Your license is about to expire!",
      ],
      ["expired within the grace period", gracePeriodLicense, /grace period/i],
      ["expired", expiredLicense, "Your license has expired!"],
    ])("a %s license reports '%s'", async (_label, license, message) => {
      setLicense(license);
      renderPage();
      expect(await screen.findByText(message)).toBeInTheDocument();
    });

    it("shows no alert when license is valid", async () => {
      setLicense(validLicense);
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
    it("renders dates formatted correctly", async () => {
      setLicense(validLicense);
      renderPage();
      await screen.findByText("License Information");
      expect(screen.getAllByText("Jan 1, 2024").length).toBeGreaterThanOrEqual(
        1,
      );
      expect(screen.getByText("Jan 1, 2025")).toBeInTheDocument();
    });

    it("shows 'Now' for -1 timestamps", async () => {
      setLicense({ ...validLicense, issued_at: -1, starts_at: -1 });
      renderPage();
      await screen.findByText("License Information");
      expect(screen.getAllByText("Now").length).toBeGreaterThanOrEqual(2);
    });

    it.each([
      [[], "Global"],
      [["BR", "US"], "BR, US"],
    ])("allowed_regions=%j renders as '%s'", async (regions, expected) => {
      setLicense({ ...validLicense, allowed_regions: regions });
      renderPage();
      expect(await screen.findByText(expected)).toBeInTheDocument();
    });

    it("displays customer fields", async () => {
      setLicense(validLicense);
      renderPage();
      await screen.findByText("License Information");
      expect(screen.getByText("cust-xxx")).toBeInTheDocument();
      expect(screen.getByText("Test Customer")).toBeInTheDocument();
      expect(screen.getByText("test@example.com")).toBeInTheDocument();
      expect(screen.getByText("Test Co")).toBeInTheDocument();
    });
  });

  describe("license features", () => {
    it("shows 'Unlimited' for devices = -1", async () => {
      setLicense(validLicense);
      renderPage();
      expect(await screen.findByText("Unlimited")).toBeInTheDocument();
    });

    it("marks enabled features as included and disabled ones as not included", async () => {
      setLicense(validLicense);
      renderPage();
      await screen.findByText("License Information");
      expect(
        screen.getAllByLabelText("Included").length,
      ).toBeGreaterThanOrEqual(2);
      expect(
        screen.getAllByLabelText("Not included").length,
      ).toBeGreaterThanOrEqual(1);
    });

    it("does not render login_link or reports features", async () => {
      setLicense(validLicense);
      renderPage();
      await screen.findByText("License Information");
      expect(screen.queryByText("Login link")).not.toBeInTheDocument();
      expect(screen.queryByText("Reports")).not.toBeInTheDocument();
    });
  });

  describe("license upload", () => {
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

    it("re-disables upload after the selected file is removed", async () => {
      const { fileInput } = renderPage();
      await screen.findByText("You do not have an installed license");
      await userEvent.upload(fileInput(), datFile());
      await userEvent.click(screen.getByRole("button", { name: /remove file/i }));
      expect(
        screen.getByRole("button", { name: /upload license/i }),
      ).toBeDisabled();
    });

    it("uploads license when upload button is clicked with valid file", async () => {
      const { fileInput } = renderPage();
      await screen.findByText("You do not have an installed license");
      await userEvent.upload(fileInput(), datFile());
      const uploadBtn = screen.getByRole("button", { name: /upload license/i });
      expect(uploadBtn).not.toBeDisabled();
      await userEvent.click(uploadBtn);
      await waitFor(() =>
        expect(
          screen.getByText("License uploaded successfully."),
        ).toBeInTheDocument(),
      );
    });

    it("shows error message on failed upload", async () => {
      server.use(
        http.post("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const { fileInput } = renderPage();
      await screen.findByText("You do not have an installed license");
      await userEvent.upload(fileInput(), datFile());
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
