import { BookOpenIcon } from "@heroicons/react/24/outline";
import type { Heading, Link, Nodes, Paragraph, Root, Text } from "mdast";
import { describe, expect, it } from "vitest";
import type { SidebarSection } from "@/data/sidebar";
import { llmsFullTxt, llmsTxt } from "@/lib/agents/llms";
import type { RenderedPage } from "@/lib/agents/markdown";
import { nodesOf, parse } from "./mdast";

const sections: SidebarSection[] = [
  {
    label: "Connect",
    description: "Open a terminal.",
    icon: BookOpenIcon,
    items: [
      { label: "Connecting", href: "/connect" },
      { label: "Group", href: "/group", items: [{ label: "Child", href: "/group/child" }] },
      { label: "Missing", href: "/missing" },
    ],
  },
];

const rendered = (path: string, title: string, description = ""): [string, RenderedPage] => [
  path,
  {
    title,
    description,
    url: `https://docs.shellhub.io${path}`,
    markdown: `# ${title}\n\nBody of ${title}.\n`,
  },
];

const pages = new Map([
  rendered("/connect", "Connecting to a Device", "Open a shell."),
  rendered("/group", "Group Overview"),
  rendered("/group/child", "Child Page", "Below the group."),
]);

const textOf = (tree: Nodes): string =>
  nodesOf<Text>(tree, "text").map((node) => node.value).join("");

const headings = (tree: Root, depth: number): string[] =>
  nodesOf<Heading>(tree, "heading")
    .filter((node) => node.depth === depth)
    .map((node) => textOf(node));

const links = (tree: Root): [string, string][] =>
  nodesOf<Link>(tree, "link").map((node) => [textOf(node), node.url]);

describe("llmsTxt", () => {
  const tree = parse(llmsTxt(sections, pages));

  it("opens with the product name followed by its summary as a quote", () => {
    const [first, second] = tree.children;

    expect(first).toMatchObject({ type: "heading", depth: 1 });
    expect(textOf(first)).toBe("ShellHub");
    expect(second.type).toBe("blockquote");
    expect(textOf(second)).toMatch(/^ShellHub is an open-source, centralized SSH gateway/);
  });

  it("groups pages under their section, then the API specifications, then the optional file", () => {
    expect(headings(tree, 2)).toEqual(["Connect", "API specifications", "Optional"]);
  });

  it("links every rendered page in navigation order to its Markdown copy, and nothing else", () => {
    expect(links(tree)).toEqual([
      ["Connecting to a Device", "https://docs.shellhub.io/connect.md"],
      ["Group Overview", "https://docs.shellhub.io/group.md"],
      ["Child Page", "https://docs.shellhub.io/group/child.md"],
      ["OpenAPI specification, Community", "https://docs.shellhub.io/openapi/community-openapi.yaml"],
      ["OpenAPI specification, Cloud", "https://docs.shellhub.io/openapi/cloud-openapi.yaml"],
      ["OpenAPI specification, Enterprise", "https://docs.shellhub.io/openapi/enterprise-openapi.yaml"],
      ["Complete documentation", "https://docs.shellhub.io/llms-full.txt"],
    ]);
  });
});

describe("llmsFullTxt", () => {
  it("carries every page in navigation order, each under its source address", () => {
    const tree = parse(llmsFullTxt(sections, pages));
    const sources = nodesOf<Paragraph>(tree, "paragraph")
      .map((node) => textOf(node))
      .filter((text) => text.startsWith("Source: "));

    expect(headings(tree, 1)).toEqual(["ShellHub", "Connecting to a Device", "Group Overview", "Child Page"]);
    expect(sources).toEqual([
      "Source: https://docs.shellhub.io/connect",
      "Source: https://docs.shellhub.io/group",
      "Source: https://docs.shellhub.io/group/child",
    ]);
  });
});
