import { flattenItems, type SidebarSection } from "@/data/sidebar";
import { SITE, markdownPath, normalizePath } from "../seo";
import type { RenderedPage } from "./markdown";

const SUMMARY = `# ShellHub

> ShellHub is an open-source, centralized SSH gateway for reaching Linux devices that have no
> public address. A small agent on each device dials out to the ShellHub server, so a device behind
> NAT, a firewall or a cellular network is reachable from any standard SSH client, with no inbound
> port and no VPN.

- Connect with the device's SSHID as the SSH username: \`ssh <username>@<namespace>.<device>@<server>\`.
- Editions: Community (open source, self-hosted), Cloud (hosted at cloud.shellhub.io) and
  Enterprise (a dedicated hosted instance). Pages say when a feature needs Cloud or Enterprise.
- Every docs page has a Markdown copy at its address plus \`.md\`, which is what the links below
  point to.
`;

const API_SPECS = [
  ["Community", "/openapi/community-openapi.yaml"],
  ["Cloud", "/openapi/cloud-openapi.yaml"],
  ["Enterprise", "/openapi/enterprise-openapi.yaml"],
] as const;

const entry = (label: string, href: string, description: string): string =>
  `- [${label}](${SITE}${href})${description ? `: ${description}` : ""}`;

const pagesInOrder = (
  sections: SidebarSection[],
  pages: Map<string, RenderedPage>,
): RenderedPage[] =>
  sections
    .flatMap((section) => flattenItems(section.items))
    .map((item) => pages.get(normalizePath(item.href)))
    .filter((page): page is RenderedPage => page !== undefined);

/**
 * The `/llms.txt` index (https://llmstxt.org): a summary of the product, then every page the
 * navigation lists, in navigation order and grouped by section, linked to its Markdown copy. A
 * navigation entry with no rendered page is left out.
 */
export const llmsTxt = (
  sections: SidebarSection[],
  pages: Map<string, RenderedPage>,
): string => {
  const body = sections.map((section) => {
    const lines = flattenItems(section.items).flatMap((item) => {
      const page = pages.get(normalizePath(item.href));
      return page
        ? [entry(page.title, markdownPath(item.href), page.description)]
        : [];
    });

    return `## ${section.label}\n\n${section.description}\n\n${lines.join("\n")}\n`;
  });

  const specs = API_SPECS.map(([edition, href]) =>
    entry(
      `OpenAPI specification, ${edition}`,
      href,
      `Every HTTP API route in ${edition}.`,
    ),
  );

  const optional = entry(
    "Complete documentation",
    "/llms-full.txt",
    "Every page above in one Markdown file, in the same order.",
  );

  return [
    SUMMARY,
    ...body,
    `## API specifications\n\n${specs.join("\n")}\n`,
    `## Optional\n\n${optional}\n`,
  ].join("\n");
};

/**
 * The `/llms-full.txt` companion: the summary, then the full Markdown of every page the navigation
 * lists, each opened by its source address so a quoted passage can be traced back to its page. A
 * navigation entry with no rendered page is left out.
 */
export const llmsFullTxt = (
  sections: SidebarSection[],
  pages: Map<string, RenderedPage>,
): string =>
  [
    SUMMARY,
    ...pagesInOrder(sections, pages).map(
      (page) => `---\n\nSource: ${page.url}\n\n${page.markdown}`,
    ),
  ].join("\n");
