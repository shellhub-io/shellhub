import { type Locator, type Page, expect, test } from "@playwright/test";
import { SignedXml } from "xml-crypto";
import {
  getAuthenticationSettings,
  getInfo,
  getSamlAuthUrl,
  getUserInfo,
} from "@/client";
import { buildRequestContext } from "./api";
import { isEnterprise, requireEnv } from "./env";
import { createTeam, signInAndOpen, signOut } from "./helpers";
import {
  type SamlUser,
  type SignOnRequest,
  adminContext,
  answerSignOn,
  disableSaml,
  enableSaml,
  identityProvider,
  idpCertificate,
  required,
  samlReason,
  signInWithSso,
  signOnURLs,
  waitForSessionToken,
} from "./saml";
import { buildRandomEmail, buildShortId, composeExec } from "./seed";

test.skip(!isEnterprise, samlReason);
test.afterEach(disableSaml);

function buildSamlUser(): SamlUser {
  const email = buildRandomEmail("saml");
  return { email, name: `E2E ${email}` };
}

async function readSamlSettings() {
  const { data } = await getAuthenticationSettings(await adminContext());
  if (!data.saml) throw new Error("expected the SAML settings");
  return data.saml;
}

async function readSessionUser(token: string) {
  const { data } = await getUserInfo(buildRequestContext({ token }));
  return data;
}

const endpointOf = (url: URL) => `${url.origin}${url.pathname}`;

async function readSignOnTarget() {
  const { data } = await getSamlAuthUrl(buildRequestContext());
  return endpointOf(new URL(data.url));
}

async function readOfferedSaml() {
  const { data } = await getInfo(buildRequestContext());
  return data.authentication?.saml;
}

const stripCertificate = (pem: string) =>
  pem.replace(/-----(BEGIN|END) CERTIFICATE-----|\s/g, "");

function expectIdpCertificate(certificates: string[] | undefined) {
  expect(certificates?.map(stripCertificate)).toEqual([
    stripCertificate(idpCertificate),
  ]);
}

function buildIdpMetadata(entityId: string) {
  return [
    `<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" entityID="${entityId}">`,
    '<md:IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">',
    '<md:KeyDescriptor use="signing"><ds:KeyInfo><ds:X509Data>',
    `<ds:X509Certificate>${stripCertificate(idpCertificate)}</ds:X509Certificate>`,
    "</ds:X509Data></ds:KeyInfo></md:KeyDescriptor>",
    `<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="${signOnURLs.post}"/>`,
    `<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="${signOnURLs.redirect}"/>`,
    "</md:IDPSSODescriptor>",
    "</md:EntityDescriptor>",
  ].join("");
}

async function publishIdpMetadata(entityId: string) {
  const file = `e2e-idp-${buildShortId()}.xml`;
  const metadata = buildIdpMetadata(entityId);
  composeExec("ui", ["sh", "-c", `cat > /var/www/${file}`], metadata);
  const served = await fetch(new URL(file, requireEnv("E2E_BASE_URL")));
  if ((await served.text()) !== metadata) {
    throw new Error(`expected the ui service to serve ${file}`);
  }
  return `http://ui:8080/${file}`;
}

function isSignedBy(request: SignOnRequest, certificate: string) {
  const signature = request.document.getElementsByTagNameNS(
    "http://www.w3.org/2000/09/xmldsig#",
    "Signature",
  )[0];
  if (!signature) throw new Error("expected a signed SAML request");
  const verifier = new SignedXml({ publicCert: certificate });
  verifier.loadSignature(signature);
  return verifier.checkSignature(request.xml);
}

const buildEntityId = () => `${identityProvider}/${buildShortId()}`;

async function openAuthenticationSettings(page: Page) {
  const { owner: admin } = await createTeam({ admin: true });
  await signInAndOpen(page, admin.username, "/admin/settings/authentication");
  return admin;
}

const samlSwitch = (page: Page) =>
  page.getByRole("switch", { name: "Toggle SAML authentication" });

const samlDialog = (page: Page) =>
  page.getByRole("dialog", { name: "Configure single sign-on" });

async function openSamlDialog(page: Page) {
  await samlSwitch(page).click();
  return samlDialog(page);
}

async function fillManualFields(dialog: Locator, entityId = identityProvider) {
  await dialog.getByLabel("SSO Redirect URL").fill(signOnURLs.redirect);
  await dialog.getByLabel("Entity ID").fill(entityId);
  await dialog.getByLabel("X.509 Certificate").fill(idpCertificate);
}

async function checkBox(dialog: Locator, name: string) {
  const box = dialog.getByRole("checkbox", { name });
  await box.press("Space");
  await expect(box).toBeChecked();
}

async function saveSamlDialog(dialog: Locator) {
  await dialog.getByRole("button", { name: "Save Configuration" }).click();
  await expect(dialog).toBeHidden();
}

test.describe("Admin configuration", () => {
  test("SAML takes the identity provider from its metadata URL", async ({
    page,
  }) => {
    const entityId = buildEntityId();
    const metadataUrl = await publishIdpMetadata(entityId);
    await openAuthenticationSettings(page);

    const dialog = await openSamlDialog(page);
    await checkBox(dialog, "Use Metadata URL");
    await dialog.getByLabel("IdP Metadata URL").fill(metadataUrl);
    await saveSamlDialog(dialog);

    await expect(page.getByText(entityId)).toBeVisible();
    const saml = await readSamlSettings();
    expect(saml).toMatchObject({
      enabled: true,
      idp: {
        entity_id: entityId,
        binding: { post: signOnURLs.post, redirect: signOnURLs.redirect },
      },
    });
    expectIdpCertificate(saml.idp?.certificates);
  });

  test("SAML takes the identity provider from fields entered by hand", async ({
    page,
  }) => {
    const entityId = buildEntityId();
    await openAuthenticationSettings(page);

    const dialog = await openSamlDialog(page);
    await fillManualFields(dialog, entityId);
    await saveSamlDialog(dialog);

    await expect(page.getByText(entityId)).toBeVisible();
    const saml = await readSamlSettings();
    expect(saml).toMatchObject({
      enabled: true,
      idp: { entity_id: entityId, binding: { redirect: signOnURLs.redirect } },
    });
    expectIdpCertificate(saml.idp?.certificates);
    expect(await readSignOnTarget()).toBe(signOnURLs.redirect);
  });

  test("custom attribute mappings name the user's email and name", async ({
    page,
  }) => {
    const admin = await openAuthenticationSettings(page);
    const dialog = await openSamlDialog(page);
    await fillManualFields(dialog);
    await dialog.getByRole("button", { name: "Advanced Settings" }).click();
    await dialog.getByLabel("Email Attribute Mapping").fill("mail");
    await dialog.getByLabel("Name Attribute Mapping").fill("cn");
    await saveSamlDialog(dialog);
    expect((await readSamlSettings()).idp?.mappings).toEqual({
      email: "mail",
      name: "cn",
    });
    await signOut(page, admin.username);

    const user = buildSamlUser();
    await answerSignOn(page.context(), user, {
      attributes: { email: "mail", name: "cn" },
    });
    const token = await signInWithSso(page);

    expect(await readSessionUser(token)).toMatchObject({
      email: user.email,
      name: user.name,
    });
  });

  test("ShellHub signs its authentication requests when asked to", async ({
    page,
  }) => {
    const admin = await openAuthenticationSettings(page);
    const dialog = await openSamlDialog(page);
    await fillManualFields(dialog);
    await dialog.getByRole("button", { name: "Advanced Settings" }).click();
    await checkBox(dialog, "Sign authorization requests");
    await saveSamlDialog(dialog);
    const { sp } = await readSamlSettings();
    expect(sp?.sign_auth_requests).toBe(true);
    const certificate = required(sp?.certificate, "an SP certificate");
    await signOut(page, admin.username);

    const user = buildSamlUser();
    const requests = await answerSignOn(page.context(), user);
    const token = await signInWithSso(page);

    expect(requests).toHaveLength(1);
    expect(isSignedBy(requests[0], certificate)).toBe(true);
    expect((await readSessionUser(token)).email).toBe(user.email);
  });

  test("turning SAML off withdraws the SSO sign-in", async ({ page }) => {
    await enableSaml();
    await openAuthenticationSettings(page);
    await expect(samlSwitch(page)).toBeChecked();
    expect(await readOfferedSaml()).toBe(true);

    await samlSwitch(page).click();

    await expect(samlSwitch(page)).not.toBeChecked();
    expect((await readSamlSettings()).enabled).toBe(false);
    expect(await readOfferedSaml()).toBe(false);
  });

  test("editing the configuration keeps what was not changed", async ({
    page,
  }) => {
    await enableSaml();
    await openAuthenticationSettings(page);
    const entityId = buildEntityId();

    await page.getByRole("button", { name: "Edit Configuration" }).click();
    const dialog = samlDialog(page);
    await expect(dialog.getByLabel("Entity ID")).toHaveValue(identityProvider);
    await dialog.getByLabel("Entity ID").fill(entityId);
    await dialog.getByLabel("SSO POST URL").fill(signOnURLs.post);
    await saveSamlDialog(dialog);

    await expect(page.getByText(entityId)).toBeVisible();
    const saml = await readSamlSettings();
    expect(saml.idp).toMatchObject({
      entity_id: entityId,
      binding: { post: signOnURLs.post, redirect: signOnURLs.redirect },
    });
    expectIdpCertificate(saml.idp?.certificates);
  });

  test("the test link signs a user in through the identity provider", async ({
    page,
  }) => {
    await enableSaml();
    await openAuthenticationSettings(page);
    const user = buildSamlUser();
    const requests = await answerSignOn(page.context(), user);
    const assertionURL = required(
      (await readSamlSettings()).assertion_url,
      "an assertion URL",
    );

    const token = waitForSessionToken(page.context());
    await page.getByRole("link", { name: "Test Auth Integration" }).click();

    expect((await readSessionUser(await token)).email).toBe(user.email);
    expect(requests).toHaveLength(1);
    expect(
      requests[0].document.getAttribute("AssertionConsumerServiceURL"),
    ).toBe(assertionURL);
  });
});
