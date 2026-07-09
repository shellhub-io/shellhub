import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import SamlConfigDrawer from "../SamlConfigDrawer";

vi.mock("@/client", () => ({
  configureSamlAuthentication: vi.fn(),
}));

import { configureSamlAuthentication as configureSamlAuthenticationSdk } from "@/client";

const mockedConfigureSaml = vi.mocked(configureSamlAuthenticationSdk);

const VALID_URL = "https://idp.example.com/sso";
const VALID_METADATA_URL = "https://idp.example.com/metadata.xml";
const VALID_ENTITY_ID = "https://idp.example.com/entity";
const VALID_CERT =
  "-----BEGIN CERTIFICATE-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\n-----END CERTIFICATE-----";

type SdkResponse<T = unknown> = {
  data: T;
  request: Request;
  response: Response;
};

function mockSdkResponse<T>(data: T): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(),
  };
}

const defaultProps = {
  open: true,
  onClose: vi.fn(),
  onSaved: vi.fn(),
  existingConfig: null,
};

function renderDrawer(props: Partial<typeof defaultProps> = {}) {
  return render(<SamlConfigDrawer {...defaultProps} {...props} />);
}

async function toggleMetadataMode(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByLabelText(/use metadata url/i));
}

function getSubmitButton() {
  return screen.getByRole("button", { name: /save configuration/i });
}

describe("SamlConfigDrawer", () => {
  beforeEach(() => {
    mockedConfigureSaml.mockReset();
    defaultProps.onClose.mockReset();
    defaultProps.onSaved.mockReset();
  });

  describe("mode toggle", () => {
    it("switches to metadata URL field when 'Use Metadata URL' is checked", async () => {
      const user = userEvent.setup();
      renderDrawer();

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
      renderDrawer();

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
      renderDrawer();

      await toggleMetadataMode(user);
      await user.type(screen.getByLabelText(/idp metadata url/i), "not-a-url");

      expect(getSubmitButton()).toBeDisabled();
    });

    it("enables submit when metadataUrl is a valid URL", async () => {
      const user = userEvent.setup();
      renderDrawer();

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
      renderDrawer();

      await user.type(screen.getByLabelText(/sso post url/i), VALID_URL);
      await user.type(screen.getByLabelText(/x\.509 certificate/i), VALID_CERT);

      expect(getSubmitButton()).toBeDisabled();
    });

    it("keeps submit disabled when no SSO URL is provided", async () => {
      const user = userEvent.setup();
      renderDrawer();

      await user.type(screen.getByLabelText(/entity id/i), VALID_ENTITY_ID);
      await user.type(screen.getByLabelText(/x\.509 certificate/i), VALID_CERT);

      expect(getSubmitButton()).toBeDisabled();
    });
  });

  describe("successful submission", () => {
    it("calls the API with correct metadata-mode body and closes the drawer", async () => {
      mockedConfigureSaml.mockResolvedValue(mockSdkResponse({}));
      const user = userEvent.setup();
      renderDrawer();

      await toggleMetadataMode(user);
      await user.type(
        screen.getByLabelText(/idp metadata url/i),
        VALID_METADATA_URL,
      );
      await user.click(getSubmitButton());

      await waitFor(() => expect(mockedConfigureSaml).toHaveBeenCalledTimes(1));
      expect(mockedConfigureSaml).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            enable: true,
            idp: { metadata_url: VALID_METADATA_URL },
            sp: { sign_requests: false },
          }),
          throwOnError: true,
        }),
      );
      expect(defaultProps.onSaved).toHaveBeenCalledTimes(1);
      expect(defaultProps.onClose).toHaveBeenCalledTimes(1);
    });

    it("calls the API with correct manual-mode body", async () => {
      mockedConfigureSaml.mockResolvedValue(mockSdkResponse({}));
      const user = userEvent.setup();
      renderDrawer();

      await user.type(screen.getByLabelText(/sso post url/i), VALID_URL);
      await user.type(screen.getByLabelText(/entity id/i), VALID_ENTITY_ID);
      await user.type(screen.getByLabelText(/x\.509 certificate/i), VALID_CERT);
      await user.click(getSubmitButton());

      await waitFor(() => expect(mockedConfigureSaml).toHaveBeenCalledTimes(1));
      expect(mockedConfigureSaml).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            enable: true,
            idp: expect.objectContaining({
              entity_id: VALID_ENTITY_ID,
              binding: { post: VALID_URL },
            }),
            sp: { sign_requests: false },
          }),
          throwOnError: true,
        }),
      );
    });
  });

  describe("save failure", () => {
    it("displays an error alert when the API call fails", async () => {
      mockedConfigureSaml.mockRejectedValue(new Error("network error"));
      const user = userEvent.setup();
      renderDrawer();

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
