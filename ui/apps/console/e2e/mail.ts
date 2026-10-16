import { expect } from "@playwright/test";
import { JSDOM } from "jsdom";
import { requireEnv } from "./env";

type MailpitSummary = { ID: string; To: { Address: string }[] };
type MailpitSearch = { messages: MailpitSummary[] };
type MailpitMessage = { HTML: string };

const readMailpit = async <T>(path: string) => {
  const response = await fetch(new URL(path, requireEnv("E2E_MAILPIT_URL")));
  if (!response.ok) {
    throw new Error(
      `expected Mailpit to answer ${path}, got ${response.status}`,
    );
  }
  return (await response.json()) as T;
};

const withoutStyles = (html: string) =>
  html.replace(/<style\b[^>]*>[\s\S]*?<\/style>/gi, "");

const findLink = (body: HTMLElement, path: string) => {
  const url = Array.from(
    body.querySelectorAll("a[href]"),
    (anchor) => anchor.getAttribute("href") ?? "",
  )
    .filter((href) => URL.canParse(href))
    .map((href) => new URL(href))
    .find((candidate) => candidate.pathname === path);
  return url && { path: url.pathname + url.search, params: url.searchParams };
};

const findLatestEmail = async (to: string, path: string) => {
  const query = new URLSearchParams({ query: `to:"${to}"` });
  const { messages } = await readMailpit<MailpitSearch>(
    `/api/v1/search?${query}`,
  );
  const sentTo = ({ To }: MailpitSummary) =>
    To.some(({ Address }) => Address.toLowerCase() === to.toLowerCase());
  for (const { ID } of messages.filter(sentTo)) {
    const { HTML } = await readMailpit<MailpitMessage>(
      `/api/v1/message/${encodeURIComponent(ID)}`,
    );
    const { body } = new JSDOM(withoutStyles(HTML)).window.document;
    const link = findLink(body, path);
    if (link) return { body, link };
  }
};

export async function readLatestEmail(to: string, path: string) {
  let email: Awaited<ReturnType<typeof findLatestEmail>>;
  await expect
    .poll(
      async () => {
        email = await findLatestEmail(to, path);
        return email;
      },
      { message: `an email to ${to} with a ${path} link` },
    )
    .toBeTruthy();
  if (!email) throw new Error(`expected an email to ${to} with a ${path} link`);
  return email;
}

export async function readEmailLink(to: string, path: string) {
  return (await readLatestEmail(to, path)).link.path;
}
