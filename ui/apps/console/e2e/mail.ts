import { expect } from "@playwright/test";
import { composeLogs } from "./seed";

const findLatestEmail = (to: string) =>
  composeLogs("server")
    .split("\n")
    .reverse()
    .find((l) => l.includes("[mail/dummy]") && l.includes(`to=${to}`));

export async function readLatestEmail(to: string) {
  await expect
    .poll(() => findLatestEmail(to), { message: `an email to ${to}` })
    .toBeTruthy();
  return findLatestEmail(to) ?? "";
}

export function findEmailLink(emailLogLine: string, path: string) {
  const url = emailLogLine
    .match(/https?:\/\/[^\s"\\)]+/g)
    ?.map((link) => new URL(link))
    .find((candidate) => candidate.pathname === path);
  if (!url) {
    throw new Error(`expected a ${path} link in the email: ${emailLogLine}`);
  }
  return { path: url.pathname + url.search, params: url.searchParams };
}

export async function readEmailLink(to: string, path: string) {
  return findEmailLink(await readLatestEmail(to), path).path;
}
