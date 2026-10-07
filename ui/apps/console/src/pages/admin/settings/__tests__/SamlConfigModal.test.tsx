import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import SamlConfigModal from "../SamlConfigModal";
import type { SamlSettings } from "../samlSchema";

const VALID_URL = "https://idp.example.com/sso";
const VALID_METADATA_URL = "https://idp.example.com/metadata.xml";
const VALID_ENTITY_ID = "https://idp.example.com/entity";
const VALID_CERT =
  "-----BEGIN CERTIFICATE-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\n-----END CERTIFICATE-----";

const samlSpy = vi.fn();

const defaultProps = {
  open: true,
  onClose: vi.fn(),
  onSaved: vi.fn(),
  existingConfig: null as SamlSettings | null,
};

function renderModal(props: Partial<typeof defaultProps> = {}) {
  return render(<SamlConfigModal {...defaultProps} {...props} />);
}

async function toggleMetadataMode(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByLabelText(/use metadata url/i));
}

function getSubmitButton() {
  return screen.getByRole("button", { name: /save configuration/i });
}

describe("SamlConfigModal", () => {
  beforeEach(() => {
    samlSpy.mockReset();
    defaultProps.onClose.mockReset();
    defaultProps.onSaved.mockReset();
    server.use(
      http.put("*/admin/api/authentication/saml", async ({ request }) => {
        samlSpy({ body: await request.json() });
        return HttpResponse.json({});
      }),
    );
  });

  describe("mode toggle", () => {
    it("switches to metadata URL field when 'Use Metadata URL' is checked", async () => {
      const user = userEvent.setup();
      renderModal();

      expect(screen.getByLabelText(/entity id/i)).toBeInTheDocument();
      expect(
        screen.queryByLabelText(/idp metadata url/i),
      ).not.toBeInTheDocument();

      await toggleMetadataMode(user);

      expect(screen.getByLabelText(/idp metadata url/i)).toBeInTheDocument();
      expect(screen.queryByLabelText(/entity id/i)).not.toBeInTheDocument();
    });

    it("switches back to manual fields when 'Use Metadata URL' is unchecked", async () => {
      const user = userEvent.setup();
      renderModal();

      await toggleMetadataMode(user);
      await toggleMetadataMode(user);

      expect(screen.getByLabelText(/entity id/i)).toBeInTheDocument();
      expect(
        screen.queryByLabelText(/idp metadata url/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("metadata mode validation", () => {
    it("keeps submit disabled when metadataUrl is invalid", async () => {
      const user = userEvent.setup();
      renderModal();

      await toggleMetadataMode(user);
      await user.type(screen.getByLabelText(/idp metadata url/i), "not-a-url");

      expect(getSubmitButton()).toBeDisabled();
    });

    it("enables submit when metadataUrl is a valid URL", async () => {
      const user = userEvent.setup();
      renderModal();

      await toggleMetadataMode(user);
      await user.type(
        screen.getByLabelText(/idp metadata url/i),
        VALID_METADATA_URL,
      );

      expect(getSubmitButton()).toBeEnabled();
    });
  });

  describe("manual mode validation", () => {
    it("keeps submit disabled when entityId is missing", async () => {
      const user = userEvent.setup();
      renderModal();

      await user.type(screen.getByLabelText(/sso post url/i), VALID_URL);
      await user.type(screen.getByLabelText(/x\.509 certificate/i), VALID_CERT);

      expect(getSubmitButton()).toBeDisabled();
    });

    it("keeps submit disabled when no SSO URL is provided", async () => {
      const user = userEvent.setup();
      renderModal();

      await user.type(screen.getByLabelText(/entity id/i), VALID_ENTITY_ID);
      await user.type(screen.getByLabelText(/x\.509 certificate/i), VALID_CERT);

      expect(getSubmitButton()).toBeDisabled();
    });
  });

  describe("successful submission", () => {
    it("calls the API with correct metadata-mode body and closes the modal", async () => {
      const user = userEvent.setup();
      renderModal();

      await toggleMetadataMode(user);
      await user.type(
        screen.getByLabelText(/idp metadata url/i),
        VALID_METADATA_URL,
      );
      await user.click(getSubmitButton());

      await waitFor(() => expect(samlSpy).toHaveBeenCalledTimes(1));
      expect(samlSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            enable: true,
            idp: { metadata_url: VALID_METADATA_URL },
            sp: { sign_requests: false },
          }),
        }),
      );
      expect(defaultProps.onSaved).toHaveBeenCalledTimes(1);
      expect(defaultProps.onClose).toHaveBeenCalledTimes(1);
    });

    it("calls the API with correct manual-mode body", async () => {
      const user = userEvent.setup();
      renderModal();

      await user.type(screen.getByLabelText(/sso post url/i), VALID_URL);
      await user.type(screen.getByLabelText(/entity id/i), VALID_ENTITY_ID);
      await user.type(screen.getByLabelText(/x\.509 certificate/i), VALID_CERT);
      await user.click(getSubmitButton());

      await waitFor(() => expect(samlSpy).toHaveBeenCalledTimes(1));
      expect(samlSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            enable: true,
            idp: expect.objectContaining({
              entity_id: VALID_ENTITY_ID,
              binding: { post: VALID_URL },
            }),
            sp: { sign_requests: false },
          }),
        }),
      );
    });
  });

  describe("editing a stored configuration", () => {
    const REDIRECT_URL = "https://idp.example.com/sso/redirect";
    const storedConfig = (preferred: "post" | "redirect"): SamlSettings => ({
      enabled: true,
      idp: {
        entity_id: VALID_ENTITY_ID,
        certificates: [VALID_CERT],
        binding: { post: VALID_URL, redirect: REDIRECT_URL, preferred },
      },
      sp: { sign_auth_requests: false },
    });

    it.each(["post", "redirect"] as const)(
      "keeps a preferred %s binding the form does not show",
      async (preferred) => {
        const user = userEvent.setup();
        renderModal({ existingConfig: storedConfig(preferred) });

        await user.clear(screen.getByLabelText(/entity id/i));
        await user.type(screen.getByLabelText(/entity id/i), "https://idp.example.com/other");
        await user.click(getSubmitButton());

        await waitFor(() => expect(samlSpy).toHaveBeenCalledTimes(1));
        expect(samlSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              idp: expect.objectContaining({
                binding: { post: VALID_URL, redirect: REDIRECT_URL, preferred },
              }),
            }),
          }),
        );
      },
    );

    it.each([
      { preferred: "post", cleared: /sso post url/i, binding: { redirect: REDIRECT_URL } },
      { preferred: "redirect", cleared: /sso redirect url/i, binding: { post: VALID_URL } },
    ] as const)(
      "drops a preferred $preferred binding once its URL is cleared",
      async ({ preferred, cleared, binding }) => {
        const user = userEvent.setup();
        renderModal({ existingConfig: storedConfig(preferred) });

        await user.clear(screen.getByLabelText(cleared));
        await user.click(getSubmitButton());

        await waitFor(() => expect(samlSpy).toHaveBeenCalledTimes(1));
        expect(samlSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              idp: expect.objectContaining({ binding }),
            }),
          }),
        );
      },
    );
  });

  describe("save failure", () => {
    it("displays an error alert when the API call fails", async () => {
      server.use(
        http.put("*/admin/api/authentication/saml", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = userEvent.setup();
      renderModal();

      await toggleMetadataMode(user);
      await user.type(
        screen.getByLabelText(/idp metadata url/i),
        VALID_METADATA_URL,
      );
      await user.click(getSubmitButton());

      expect(await screen.findByRole("alert")).toBeInTheDocument();
      expect(screen.getByRole("alert")).toHaveTextContent(
        /failed to save saml configuration/i,
      );
      expect(defaultProps.onSaved).not.toHaveBeenCalled();
      expect(defaultProps.onClose).not.toHaveBeenCalled();
    });
  });
});
