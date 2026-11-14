import type { BrowserContext, Page, Request } from "@playwright/test";
import { randomUUID } from "node:crypto";
import { inflateRawSync } from "node:zlib";
import sshpk from "sshpk";
import { DOMParser, type Element } from "@xmldom/xmldom";
import { SignedXml } from "xml-crypto";
import { configureSamlAuthentication, getSamlAuthUrl } from "@/client";
import { buildRequestContext, loginAs } from "./api";
import { adminUser, requireEnv } from "./env";
import { required } from "./helpers";
import { buildRandomEmail } from "./seed";
import { buildPrivateKey } from "./vault";

export const samlReason = "only enterprise signs users in with SAML";

export const identityProvider = "http://idp.e2e.test";

export const signOnURLs = {
  post: `${identityProvider}/sso/post`,
  redirect: `${identityProvider}/sso/redirect`,
};

export type AttributeNames = { email: string; name: string };

const defaultAttributes: AttributeNames = {
  email: "emailAddress",
  name: "displayName",
};

export type SamlUser = { email: string; name: string };

export function buildSamlUser(): SamlUser {
  const email = buildRandomEmail("saml");
  return { email, name: `E2E ${email}` };
}

export type SignOnRequest = {
  url: URL;
  document: Element;
  relayState: string;
};

const buildSamlId = () => `_${randomUUID()}`;

function buildSigningKey() {
  const privateKey = buildPrivateKey();
  const certificate = sshpk.createSelfSignedCertificate(
    sshpk.identityForHost(new URL(identityProvider).hostname),
    sshpk.parsePrivateKey(privateKey),
    {
      validFrom: new Date(Date.now() - 60_000),
      validUntil: new Date(Date.now() + 86_400_000),
    },
  );
  return { privateKey, certificate: certificate.toString("pem") };
}

const signingKey = buildSigningKey();

export const idpCertificate = signingKey.certificate;

export async function adminContext() {
  const { token } = await loginAs(adminUser.username, adminUser.password);
  return buildRequestContext({ token });
}

export type SamlOptions = {
  binding?: {
    post?: string;
    redirect?: string;
    preferred?: "post" | "redirect";
  };
};

export async function enableSaml({
  binding = { redirect: signOnURLs.redirect },
}: SamlOptions = {}) {
  await configureSamlAuthentication({
    ...(await adminContext()),
    body: {
      enable: true,
      idp: {
        entity_id: identityProvider,
        certificate: idpCertificate,
        binding,
      },
      sp: { sign_requests: false },
    },
  });
}

export async function disableSaml() {
  await configureSamlAuthentication({
    ...(await adminContext()),
    body: { enable: false, idp: {}, sp: {} },
  });
}

const requireAttribute = (request: Element, name: string) =>
  required(request.getAttribute(name), `${name} in the SAML request`);

type AssertionFields = {
  user: SamlUser;
  attributes: AttributeNames;
  requestId?: string;
  consumer: string;
  audience: string;
  issued: string;
  expires: string;
};

const inResponseTo = (requestId?: string) =>
  requestId ? ` InResponseTo="${requestId}"` : "";

function buildAssertion(fields: AssertionFields) {
  const { user, attributes, requestId, consumer, audience, issued, expires } =
    fields;
  return [
    `<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="${buildSamlId()}" Version="2.0" IssueInstant="${issued}">`,
    `<saml:Issuer>${identityProvider}</saml:Issuer>`,
    "<saml:Subject>",
    `<saml:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress">${user.email}</saml:NameID>`,
    '<saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">',
    `<saml:SubjectConfirmationData${inResponseTo(requestId)} NotOnOrAfter="${expires}" Recipient="${consumer}"/>`,
    "</saml:SubjectConfirmation>",
    "</saml:Subject>",
    `<saml:Conditions NotBefore="${issued}" NotOnOrAfter="${expires}">`,
    `<saml:AudienceRestriction><saml:Audience>${audience}</saml:Audience></saml:AudienceRestriction>`,
    "</saml:Conditions>",
    `<saml:AuthnStatement AuthnInstant="${issued}" SessionIndex="${buildSamlId()}">`,
    "<saml:AuthnContext><saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:Password</saml:AuthnContextClassRef></saml:AuthnContext>",
    "</saml:AuthnStatement>",
    "<saml:AttributeStatement>",
    `<saml:Attribute Name="${attributes.email}"><saml:AttributeValue>${user.email}</saml:AttributeValue></saml:Attribute>`,
    `<saml:Attribute Name="${attributes.name}"><saml:AttributeValue>${user.name}</saml:AttributeValue></saml:Attribute>`,
    "</saml:AttributeStatement>",
    "</saml:Assertion>",
  ].join("");
}

function signAssertion(response: string) {
  const signature = new SignedXml({
    privateKey: signingKey.privateKey,
    publicCert: signingKey.certificate,
    signatureAlgorithm: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
    canonicalizationAlgorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
  });
  signature.addReference({
    xpath: "//*[local-name(.)='Assertion']",
    transforms: [
      "http://www.w3.org/2000/09/xmldsig#enveloped-signature",
      "http://www.w3.org/2001/10/xml-exc-c14n#",
    ],
    digestAlgorithm: "http://www.w3.org/2001/04/xmlenc#sha256",
  });
  signature.computeSignature(response, {
    prefix: "ds",
    location: {
      reference: "//*[local-name(.)='Assertion']/*[local-name(.)='Issuer']",
      action: "after",
    },
  });
  return signature.getSignedXml();
}

function buildResponse(
  request: Element,
  user: SamlUser,
  attributes: AttributeNames,
  unsolicited: boolean,
) {
  const now = new Date();
  const fields = {
    user,
    attributes,
    requestId: unsolicited ? undefined : requireAttribute(request, "ID"),
    consumer: requireAttribute(request, "AssertionConsumerServiceURL"),
    audience: required(
      request.getElementsByTagNameNS(
        "urn:oasis:names:tc:SAML:2.0:assertion",
        "Issuer",
      )[0]?.textContent,
      "an Issuer in the SAML request",
    ),
    issued: now.toISOString(),
    expires: new Date(now.getTime() + 5 * 60_000).toISOString(),
  };
  const response = [
    `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="${buildSamlId()}" Version="2.0" IssueInstant="${fields.issued}" Destination="${fields.consumer}"${inResponseTo(fields.requestId)}>`,
    `<saml:Issuer>${identityProvider}</saml:Issuer>`,
    '<samlp:Status><samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/></samlp:Status>',
    buildAssertion(fields),
    "</samlp:Response>",
  ].join("");
  return { consumer: fields.consumer, xml: signAssertion(response) };
}

function buildSignOnRequest(
  url: URL,
  xml: string,
  relayState: string | null,
): SignOnRequest {
  const document = new DOMParser().parseFromString(xml, "text/xml").documentElement;
  if (!document) throw new Error("expected a root element in the SAML request");
  return {
    url,
    document,
    relayState: relayState ?? "",
  };
}

function parseRedirectRequest(url: URL): SignOnRequest {
  const encoded = required(
    url.searchParams.get("SAMLRequest"),
    `a SAMLRequest in ${url}`,
  );
  return buildSignOnRequest(
    url,
    inflateRawSync(Buffer.from(encoded, "base64")).toString(),
    url.searchParams.get("RelayState"),
  );
}

function parsePostRequest(url: URL, body: string): SignOnRequest {
  const form = new URLSearchParams(body);
  const encoded = required(
    form.get("SAMLRequest"),
    `a SAMLRequest in the form posted to ${url}`,
  );
  return buildSignOnRequest(
    url,
    Buffer.from(encoded, "base64").toString(),
    form.get("RelayState"),
  );
}

function readSignOnRequest(request: Request) {
  const url = new URL(request.url());
  const endpoint = `${url.origin}${url.pathname}`;
  if (endpoint === signOnURLs.post && request.method() === "POST") {
    return parsePostRequest(url, request.postData() ?? "");
  }
  if (endpoint === signOnURLs.redirect && request.method() === "GET") {
    return parseRedirectRequest(url);
  }
  return undefined;
}

const consumerURL = (consumer: string) =>
  new URL(new URL(consumer).pathname, requireEnv("E2E_BASE_URL"));

export function answerSignOnRequest(
  request: SignOnRequest,
  user: SamlUser,
  attributes = defaultAttributes,
  unsolicited = false,
) {
  const { consumer, xml } = buildResponse(
    request.document,
    user,
    attributes,
    unsolicited,
  );
  return {
    action: consumerURL(consumer),
    samlResponse: Buffer.from(xml).toString("base64"),
  };
}

export type AnswerOptions = {
  attributes?: AttributeNames;
  replaceRelayState?: (relayState: string) => string;
};

export async function answerSignOn(
  context: BrowserContext,
  user: SamlUser,
  { attributes, replaceRelayState }: AnswerOptions = {},
) {
  const requests: SignOnRequest[] = [];
  await context.route(`${identityProvider}/sso/**`, async (route) => {
    const request = readSignOnRequest(route.request());
    if (!request) {
      await route.fulfill({
        status: 405,
        body: `${route.request().method()} is not the binding of ${route.request().url()}`,
      });
      return;
    }
    const { action, samlResponse } = answerSignOnRequest(
      request,
      user,
      attributes,
    );
    requests.push(request);
    const relayState =
      replaceRelayState?.(request.relayState) ?? request.relayState;
    await route.fulfill({
      contentType: "text/html",
      body: autoSubmitSignOnResponse(action, {
        SAMLResponse: samlResponse,
        RelayState: relayState,
      }),
    });
  });
  return requests;
}

export function autoSubmitSignOnResponse(
  action: URL | string,
  fields: Record<string, string>,
) {
  const inputs = Object.entries(fields)
    .map(
      ([name, value]) =>
        `<input type="hidden" name="${name}" value="${value}">`,
    )
    .join("");
  return `<form method="post" action="${action}">${inputs}</form><script>document.forms[0].submit()</script>`;
}

export async function buildSignOnResponse(
  user: SamlUser,
  { unsolicited = false } = {},
) {
  const { data } = await getSamlAuthUrl(buildRequestContext());
  return answerSignOnRequest(
    parseRedirectRequest(new URL(data.url)),
    user,
    defaultAttributes,
    unsolicited,
  );
}

export function postSignOnResponse(action: URL, samlResponse: string) {
  return fetch(action, {
    method: "POST",
    redirect: "manual",
    body: new URLSearchParams({ SAMLResponse: samlResponse }),
  });
}

export async function signInThroughApi(user: SamlUser) {
  const { action, samlResponse } = await buildSignOnResponse(user);
  const response = await postSignOnResponse(action, samlResponse);
  const location = required(
    response.headers.get("location"),
    `a redirect from the assertion consumer, got ${response.status}`,
  );
  return required(
    new URL(location, action).searchParams.get("token"),
    `a session token in ${location}`,
  );
}

export async function waitForSessionToken(context: BrowserContext) {
  const landing = await context.waitForEvent("request", (request) => {
    const url = new URL(request.url());
    return (
      url.pathname === "/login" &&
      (url.searchParams.has("token") ||
        url.searchParams.has("missing_assertions"))
    );
  });
  const url = new URL(landing.url());
  return required(
    url.searchParams.get("token"),
    `a session token from the assertion consumer, got missing assertions "${url.searchParams.get("missing_assertions")}"`,
  );
}

export async function signInWithSso(page: Page) {
  await page.goto("/login");
  const token = waitForSessionToken(page.context());
  await page.getByRole("button", { name: "Login with SSO" }).click();
  return token;
}
