import type { Code, Image, InlineCode, Link, Nodes, Root, Table, TableCell, Text } from "mdast";
import { describe, expect, it } from "vitest";
import { renderPage } from "@/lib/agents/markdown";
import { nodesOf, parse } from "./mdast";

const PAGE_URL = "https://docs.shellhub.io/connect";

const page = (article: string, head = ""): string =>
  `<html><head><title>Page | ShellHub Docs</title>${head}</head><body>` +
  `<nav><a href="/elsewhere">Sidebar</a></nav><article class="prose">${article}</article>` +
  "<script>tracking()</script></body></html>";

const markdownOf = (article: string): Root =>
  parse(renderPage(page(article), PAGE_URL)?.markdown ?? "");

const textsOf = (tree: Nodes): string[] => nodesOf<Text>(tree, "text").map((node) => node.value);

const cellText = (cell: TableCell): string =>
  [...nodesOf<Text>(cell, "text"), ...nodesOf<InlineCode>(cell, "inlineCode")]
    .map((node) => node.value)
    .join("");

describe("renderPage", () => {
  it("returns nothing for a page outside the documentation layout", () => {
    expect(renderPage("<html><body><main>Home</main></body></html>", PAGE_URL)).toBeUndefined();
  });

  it("reads the title from the article heading and the description from the meta tag", () => {
    const rendered = renderPage(
      page("<h1>Connecting</h1><p>Body.</p>", '<meta name="description" content="How to connect.">'),
      PAGE_URL,
    );

    expect(rendered).toMatchObject({
      title: "Connecting",
      description: "How to connect.",
      url: PAGE_URL,
    });
  });

  it("keeps only the article, with links and images made absolute", () => {
    const tree = markdownOf(
      '<h1>T</h1><p>See <a href="/manage">Manage</a> and <a href="#here">here</a>.</p>' +
        '<figure><button data-shot-zoom aria-label="Enlarge"><img src="/img/a.png" alt="The panel"></button></figure>',
    );

    expect(nodesOf<Link>(tree, "link").map((node) => node.url)).toEqual([
      "https://docs.shellhub.io/manage",
      "https://docs.shellhub.io/connect#here",
    ]);
    expect(nodesOf<Image>(tree, "image").map((node) => [node.url, node.alt])).toEqual([
      ["https://docs.shellhub.io/img/a.png", "The panel"],
    ]);
    expect(textsOf(tree)).not.toContain("Sidebar");
    expect(textsOf(tree)).not.toContain("tracking()");
  });

  it("links a recorded clip at its absolute address, labelled with its description", () => {
    const tree = markdownOf(
      '<h1>T</h1><figure><video autoplay loop muted aria-label="Enabling MFA">' +
        '<source src="/img/manual/mfa.webm" type="video/webm">' +
        '<source src="/img/manual/mfa.mp4" type="video/mp4"></video></figure>',
    );

    expect(nodesOf<Link>(tree, "link").map((node) => [textsOf(node), node.url])).toEqual([
      [["Enabling MFA"], "https://docs.shellhub.io/img/manual/mfa.webm"],
    ]);
  });

  it("labels a highlighted code block with its language", () => {
    const [code] = nodesOf<Code>(
      markdownOf(
        '<h1>T</h1><pre class="astro-code" data-language="bash"><code><span class="line">ssh root@dev</span></code></pre>',
      ),
      "code",
    );

    expect(code).toMatchObject({ lang: "bash", value: "ssh root@dev" });
  });

  it("replaces a diagram with its accessible description", () => {
    const tree = markdownOf(
      '<h1>T</h1><div role="img" aria-label="The agent dials out."><svg><text>Agent</text></svg></div>',
    );

    expect(textsOf(tree)).toEqual(["T", "The agent dials out."]);
  });

  it("keeps a session replay's description and drops its player controls", () => {
    const tree = markdownOf(
      '<h1>T</h1><figure class="replay"><div class="replay-frame">' +
        '<div class="replay-term" role="img" aria-label="A recorded session."></div>' +
        '<div class="replay-bar"><button aria-label="Play">Play</button>' +
        '<span class="replay-time"><span data-elapsed>00:00</span> / <span data-total>00:00</span></span>' +
        '<input type="range"><select><option>1x</option></select></div></div></figure>',
    );

    expect(textsOf(tree)).toEqual(["T", "A recorded session."]);
  });

  it("turns a callout into a quote led by its heading", () => {
    const tree = markdownOf(
      '<h1>T</h1><aside class="callout"><p class="callout-head"><svg></svg>Careful</p><p>It breaks.</p></aside>',
    );
    const [quote] = nodesOf(tree, "blockquote");
    const [strong] = nodesOf(quote, "strong");

    expect(textsOf(quote)).toEqual(["Careful", "It breaks."]);
    expect(textsOf(strong)).toEqual(["Careful"]);
  });

  it("keeps tables as tables", () => {
    const [table] = nodesOf<Table>(
      markdownOf(
        "<h1>T</h1><table><thead><tr><th>Part</th><th>Meaning</th></tr></thead>" +
          "<tbody><tr><td><code>root</code></td><td>The unix user</td></tr></tbody></table>",
      ),
      "table",
    );

    expect(table.children.map((row) => row.children.map(cellText))).toEqual([
      ["Part", "Meaning"],
      ["root", "The unix user"],
    ]);
  });
});
