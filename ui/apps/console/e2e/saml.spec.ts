import { type Locator, type Page, expect, test } from "@playwright/test";
import { SignedXml } from "xml-crypto";
import {
  createNamespace,
  getAuthenticationSettings,
  getInfo,
  getSamlAuthUrl,
  getUser,
  getUserInfo,
} from "@/client";
import { buildRequestContext, loginAs } from "./api";
import { isEnterprise, requireEnv } from "./env";
import { createTeam, required, signInAndOpen, signOut } from "./helpers";
import {
  type SignOnRequest,
  adminContext,
  answerSignOn,
  buildSamlUser,
  disableSaml,
  enableSaml,
  identityProvider,
  idpCertificate,
  samlReason,
  signInThroughApi,
  signInWithSso,
  signOnURLs,
  waitForSessionToken,
} from "./saml";
import { buildShortId, composeExec, password } from "./seed";

test.skip(!isEnterprise, samlReason);
test.afterEach(disableSaml);

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

const bindingNames = {
  post: "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
  redirect: "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
};

type Binding = keyof typeof bindingNames;

function buildIdpMetadata(
  entityId: string,
  bindings: Binding[] = ["post", "redirect"],
) {
  return [
    `<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" entityID="${entityId}">`,
    '<md:IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">',
    '<md:KeyDescriptor use="signing"><ds:KeyInfo><ds:X509Data>',
    `<ds:X509Certificate>${stripCertificate(idpCertificate)}</ds:X509Certificate>`,
    "</ds:X509Data></ds:KeyInfo></md:KeyDescriptor>",
    ...bindings.map(
      (binding) =>
        `<md:SingleSignOnService Binding="${bindingNames[binding]}" Location="${signOnURLs[binding]}"/>`,
    ),
    "</md:IDPSSODescriptor>",
    "</md:EntityDescriptor>",
  ].join("");
}

async function publishIdpMetadata(entityId: string, bindings?: Binding[]) {
  const file = `e2e-idp-${buildShortId()}.xml`;
  const metadata = buildIdpMetadata(entityId, bindings);
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

async function saveMetadataUrl(page: Page, metadataUrl: string) {
  const dialog = await openSamlDialog(page);
  await checkBox(dialog, "Use Metadata URL");
  await dialog.getByLabel("IdP Metadata URL").fill(metadataUrl);
  await saveSamlDialog(dialog);
}

async function openEditDialog(page: Page) {
  await page.getByRole("button", { name: "Edit Configuration" }).click();
  return samlDialog(page);
}

test.describe("Admin configuration", () => {
  test("SAML takes the identity provider from its metadata URL", async ({
    page,
  }) => {
    const entityId = buildEntityId();
    const metadataUrl = await publishIdpMetadata(entityId);
    await openAuthenticationSettings(page);

    await saveMetadataUrl(page, metadataUrl);

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

    const dialog = await openEditDialog(page);
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

test.describe("Sign-in", () => {
  test.beforeEach(() => enableSaml());

  test("a returning SAML user signs in to their namespace", async ({
    page,
  }) => {
    const user = buildSamlUser();
    const first = await signInThroughApi(user);
    const namespace = `e2e-saml-${buildShortId()}`;
    await createNamespace({
      ...buildRequestContext({ token: first }),
      body: { name: namespace },
    });
    await answerSignOn(page.context(), user);

    const token = await signInWithSso(page);

    await expect(page).toHaveURL(/\/dashboard$/);
    await expect(
      page
        .getByRole("listitem")
        .filter({ has: page.getByRole("heading", { name: "Namespace" }) })
        .getByText(namespace, { exact: true }),
    ).toBeVisible();
    expect((await readSessionUser(token)).id).toBe(
      (await readSessionUser(first)).id,
    );
  });

  test("a first SAML sign-in creates the account", async ({ page }) => {
    const user = buildSamlUser();
    await answerSignOn(page.context(), user);

    const token = await signInWithSso(page);

    const account = await readSessionUser(token);
    expect(account).toMatchObject({
      email: user.email,
      name: user.name,
      origin: "saml",
      auth_methods: ["saml"],
    });
    const { data } = await getUser({
      ...(await adminContext()),
      path: { id: account.id },
    });
    expect(data).toMatchObject({ email: user.email, username: "" });
  });

  test("a SAML sign-in with a local account's email signs in to that account", async ({
    page,
  }) => {
    const { owner } = await createTeam();
    const { id } = await loginAs(owner.username, password);
    await answerSignOn(page.context(), {
      email: owner.email,
      name: owner.username,
    });

    const token = await signInWithSso(page);

    await expect(page).toHaveURL(/\/dashboard$/);
    expect(await readSessionUser(token)).toMatchObject({
      id,
      user: owner.username,
      auth_methods: ["local", "saml"],
    });
    expect((await loginAs(owner.username, password)).id).toBe(id);
  });
});

test.describe("Bindings", () => {
  async function signInThroughEndpoint(page: Page, endpoint: string) {
    const user = buildSamlUser();
    const requests = await answerSignOn(page.context(), user);

    const token = await signInWithSso(page);

    expect(requests).toHaveLength(1);
    expect(endpointOf(requests[0].url)).toBe(endpoint);
    expect(requests[0].document.getAttribute("Destination")).toBe(endpoint);
    expect((await readSessionUser(token)).email).toBe(user.email);
  }

  test("the redirect binding sends the request to the redirect endpoint", async ({
    page,
  }) => {
    await enableSaml({
      binding: {
        post: signOnURLs.post,
        redirect: signOnURLs.redirect,
        preferred: "redirect",
      },
    });

    await signInThroughEndpoint(page, signOnURLs.redirect);
  });

  test("the POST binding sends the request to the POST endpoint", async ({
    page,
  }) => {
    await enableSaml({
      binding: {
        post: signOnURLs.post,
        redirect: signOnURLs.redirect,
        preferred: "post",
      },
    });

    await signInThroughEndpoint(page, signOnURLs.post);
  });

  test("with both endpoints and no preference, the POST binding is used", async ({
    page,
  }) => {
    await enableSaml({
      binding: { post: signOnURLs.post, redirect: signOnURLs.redirect },
    });

    await signInThroughEndpoint(page, signOnURLs.post);
  });

  test("metadata offering only the POST binding signs in through it", async ({
    page,
  }) => {
    const metadataUrl = await publishIdpMetadata(identityProvider, ["post"]);
    const admin = await openAuthenticationSettings(page);

    await saveMetadataUrl(page, metadataUrl);

    const { idp } = await readSamlSettings();
    expect(idp).toMatchObject({
      entity_id: identityProvider,
      binding: { post: signOnURLs.post, redirect: "" },
    });
    await signOut(page, admin.username);
    await signInThroughEndpoint(page, signOnURLs.post);
  });

  test("editing the configuration keeps the preferred binding", async ({
    page,
  }) => {
    await enableSaml({
      binding: {
        post: signOnURLs.post,
        redirect: signOnURLs.redirect,
        preferred: "redirect",
      },
    });
    await openAuthenticationSettings(page);

    const dialog = await openEditDialog(page);
    const entityId = buildEntityId();
    await dialog.getByLabel("Entity ID").fill(entityId);
    await saveSamlDialog(dialog);
    await expect(page.getByText(entityId)).toBeVisible();

    expect((await readSamlSettings()).idp?.binding?.preferred).toBe("redirect");
    expect(await readSignOnTarget()).toBe(signOnURLs.redirect);
  });
});
