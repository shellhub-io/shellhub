import { BookOpenIcon } from "@heroicons/react/24/outline";
import { describe, expect, it } from "vitest";
import type { SidebarSection } from "@/data/sidebar";
import {
  absoluteUrl,
  breadcrumbsFor,
  isIndexable,
  isStableChannel,
  jsonLd,
  markdownPath,
  normalizePath,
  pageGraph,
} from "@/lib/seo";

describe("normalizePath", () => {
  it.each([
    ["/", "/"],
    ["/index.html", "/"],
    ["/introduction", "/introduction"],
    ["/introduction/", "/introduction"],
    ["/introduction.html", "/introduction"],
    ["/manage/devices.html", "/manage/devices"],
    ["/manage/index.html", "/manage"],
  ])("serves %s at %s", (pathname, expected) => {
    expect(normalizePath(pathname)).toBe(expected);
  });
});

describe("absoluteUrl", () => {
  it("keeps the slash on the home and drops it everywhere else", () => {
    expect(absoluteUrl("/")).toBe("https://docs.shellhub.io/");
    expect(absoluteUrl("/connect.html")).toBe(
      "https://docs.shellhub.io/connect",
    );
  });
});

describe("markdownPath", () => {
  it("puts the copy beside the page", () => {
    expect(markdownPath("/connect.html")).toBe("/connect.md");
    expect(markdownPath("/")).toBe("/index.md");
  });
});

describe("isIndexable", () => {
  it.each([
    ["/404.html", false],
    ["/404", false],
    ["/connect.html", true],
    ["/", true],
  ])("publishes %s: %s", (pathname, expected) => {
    expect(isIndexable(pathname)).toBe(expected);
  });
});

describe("isStableChannel", () => {
  it.each([
    ["stable", true],
    ["unreleased", false],
    [undefined, false],
  ])("indexes %s: %s", (channel, expected) => {
    expect(isStableChannel(channel)).toBe(expected);
  });
});

const sections: SidebarSection[] = [
  {
    label: "Get Started",
    description: "Install.",
    icon: BookOpenIcon,
    items: [
      { label: "Connecting", href: "/connect" },
      {
        label: "Install the Agent",
        href: "/install",
        items: [{ label: "Yocto Project", href: "/install/yocto" }],
      },
      { label: "Platforms", items: [{ label: "FreeBSD", href: "/install/freebsd" }] },
    ],
  },
];

describe("breadcrumbsFor", () => {
  it("passes through a group heading that is itself a page", () => {
    expect(breadcrumbsFor(sections, "/install/yocto.html", "Yocto")).toEqual([
      { name: "Docs", path: "/" },
      { name: "Install the Agent", path: "/install" },
      { name: "Yocto Project", path: "/install/yocto" },
    ]);
  });

  it("leaves out a group heading that is not a page, which has no address to link", () => {
    expect(breadcrumbsFor(sections, "/install/freebsd", "FreeBSD")).toEqual([
      { name: "Docs", path: "/" },
      { name: "FreeBSD", path: "/install/freebsd" },
    ]);
  });

  it("gives a page the navigation does not list only the home and itself", () => {
    expect(breadcrumbsFor(sections, "/unlisted", "Unlisted")).toEqual([
      { name: "Docs", path: "/" },
      { name: "Unlisted", path: "/unlisted" },
    ]);
  });
});

describe("pageGraph", () => {
  it("describes the page as an article at its canonical address with its breadcrumbs", () => {
    const graph = pageGraph({
      pathname: "/connect.html",
      title: "Connecting",
      description: "How.",
      sections,
    }) as {
      "@graph": {
        "@type": string;
        url?: string;
        itemListElement?: { item: string }[];
      }[];
    };

    const article = graph["@graph"].find(
      (node) => node["@type"] === "TechArticle",
    );
    const crumbs = graph["@graph"].find(
      (node) => node["@type"] === "BreadcrumbList",
    );

    expect(article?.url).toBe("https://docs.shellhub.io/connect");
    expect(crumbs?.itemListElement?.map((crumb) => crumb.item)).toEqual([
      "https://docs.shellhub.io/",
      "https://docs.shellhub.io/connect",
    ]);
  });
});

describe("jsonLd", () => {
  it("cannot close its script element", () => {
    const serialized = jsonLd({ name: "</script><script>alert(1)</script>" });

    expect(serialized).not.toContain("</script>");
    expect(JSON.parse(serialized)).toEqual({
      name: "</script><script>alert(1)</script>",
    });
  });
});
