import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
vi.unmock("@/hooks/useFocusTrap");
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockDevice, mockNamespace } from "@/tests/factories";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import ConnectModal from "@/components/ConnectModal";

vi.mock("@/utils/browserKey", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/utils/browserKey")>()),
  storedBrowserKeyFingerprint: () => Promise.resolve(null),
}));

vi.mock("@/utils/recordings", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/utils/recordings")>()),
  isRecordingSupported: () => true,
}));

function serveConnectApi(
  accessMode: "legacy" | "identity",
  sessionRecord = false,
) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(
        mockNamespace({
          settings: {
            session_record: sessionRecord,
            connection_announcement: "",
            ssh_access_mode: accessMode,
            ssh_legacy_allowed: true,
          },
        }),
      ),
    ),
    http.get("*/api/devices/:uid", () =>
      HttpResponse.json(mockDevice({ uid: "dev-1", name: "rustagent" })),
    ),
    http.get("*/api/ssh-identities", () => HttpResponse.json([])),
  );
}

function renderModal(sshid = "dev.rustagent@localhost") {
  return render(
    <ClipboardProvider>
      <ConnectModal
        open
        onClose={vi.fn()}
        deviceUid="dev-1"
        deviceName="rustagent"
        sshid={sshid}
      />
    </ClipboardProvider>,
    { wrapper: createTestWrapper() },
  );
}

const externalLink = () =>
  screen.queryByRole("link", { name: /open in external terminal/i });

const recSwitch = () =>
  screen.findByRole("switch", { name: /record this session in this browser/i });

describe("ConnectModal", () => {
  beforeEach(() => {
    seedAuthStore({ name: "gustavo", email: "gustavo@shellhub.io" });
    serveConnectApi("legacy");
    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 0;
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("opens with focus on the login field", () => {
    renderModal();

    expect(screen.getByLabelText("Login")).toHaveFocus();
  });

  describe("external terminal", () => {
    it("links the SSHID as root while the login is empty", () => {
      renderModal();

      expect(externalLink()).toHaveAttribute(
        "href",
        "ssh://root%40dev.rustagent@localhost",
      );
    });

    it("links the SSHID as the login typed", async () => {
      const user = userEvent.setup();
      renderModal();

      await user.type(screen.getByLabelText("Login"), "deploy");

      expect(externalLink()).toHaveAttribute(
        "href",
        "ssh://deploy%40dev.rustagent@localhost",
      );
    });

    it("offers no link before the SSHID is known", () => {
      renderModal("3f1c9a0e7b2d");

      expect(externalLink()).not.toBeInTheDocument();
    });
  });

  describe("recording", () => {
    it("records in this browser by default", async () => {
      renderModal();

      expect(await recSwitch()).toHaveAttribute("aria-checked", "true");
    });

    it("stops recording in this browser when REC is switched off", async () => {
      const user = userEvent.setup();
      renderModal();
      const rec = await recSwitch();

      await user.click(rec);

      expect(rec).toHaveAttribute("aria-checked", "false");
      expect(screen.getByText(/won't be recorded/i)).toBeInTheDocument();
    });

    it("states the server recording and offers no switch when the namespace records", async () => {
      serveConnectApi("legacy", true);
      renderModal();

      expect(
        await screen.findByText(/recorded and stored on the server/i),
      ).toBeInTheDocument();
      expect(screen.queryByRole("switch")).not.toBeInTheDocument();
    });
  });

  it("lets the shell open in the browser with the login empty, in an identity namespace", async () => {
    serveConnectApi("identity");
    renderModal();

    await screen.findByRole("group", { name: "Identity" });

    expect(
      screen.getByRole("button", { name: "Open in browser" }),
    ).toBeEnabled();
  });

  it("says a new browser key will be added when this browser has none", async () => {
    serveConnectApi("identity");
    renderModal();

    expect(
      await screen.findByRole("group", { name: "Identity" }),
    ).toHaveAccessibleDescription(/adds it to your SSH identities/i);
  });
});
