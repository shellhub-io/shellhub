import { expect } from "@playwright/test";
import { composeLogs } from "./seed";

const findLink = (emailLogLine: string, path: string) => {
  const url = emailLogLine
    .match(/https?:\/\/[^\s"\\)]+/g)
    ?.map((link) => new URL(link))
    .find((candidate) => candidate.pathname === path);
  return url && { path: url.pathname + url.search, params: url.searchParams };
};

const findLatestEmail = (to: string, path: string) => {
  const lines = composeLogs("server").split("\n").reverse();
  for (const line of lines) {
    if (!line.includes("[mail/dummy]") || !line.includes(`to=${to}`)) continue;
    const link = findLink(line, path);
    if (link) return { line, link };
  }
};

export async function readLatestEmail(to: string, path: string) {
  let email: ReturnType<typeof findLatestEmail>;
  await expect
    .poll(
      () => {
        email = findLatestEmail(to, path);
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
