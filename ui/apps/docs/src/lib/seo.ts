import type { SidebarItem, SidebarSection } from "../data/sidebar";

/** The address the released documentation is served from, and the origin every canonical uses. */
export const SITE = "https://docs.shellhub.io";

const PRODUCT_SITE = "https://shellhub.io";

/** The suffix every page title carries, so a search result names the product. */
export const TITLE_SUFFIX = "ShellHub Docs";

/** The card a page shares on social sites and chat previews, served from `public/`. */
export const SOCIAL_IMAGE = "/og.png";

/**
 * Whether this build describes a release. Only a stable build is indexed: the unreleased channel
 * mirrors the same pages at another address, and a search engine would otherwise split ranking
 * between the two copies.
 */
export const isStableChannel = (channel: string | undefined): boolean =>
  channel === "stable";

/**
 * The address a page is served at: no trailing slash, no `.html`, no `index`. The build renders
 * pages as files, so `Astro.url.pathname` reads `/introduction.html` while the page is served, and
 * linked, as `/introduction`.
 */
export const normalizePath = (pathname: string): string =>
  pathname
    .replace(/\.html$/, "")
    .replace(/(^|\/)index$/, "")
    .replace(/\/+$/, "") || "/";

const UNINDEXED = new Set(["/404"]);

/**
 * Whether a page may be published; false only for the 404 page. A true page is listed in the
 * sitemap and is indexed on a stable build. It gets a Markdown copy only if it uses the docs layout
 * (not the home or API reference), and an `llms.txt` entry only if the navigation also lists it.
 */
export const isIndexable = (pathname: string): boolean => !UNINDEXED.has(normalizePath(pathname));

/** The absolute address of a page on the released site. */
export const absoluteUrl = (pathname: string): string => {
  const path = normalizePath(pathname);
  return path === "/" ? `${SITE}/` : `${SITE}${path}`;
};

/** The Markdown copy of a page that `agents()` writes beside its HTML, for agents to read. */
export const markdownPath = (pathname: string): string => {
  const path = normalizePath(pathname);
  return path === "/" ? "/index.md" : `${path}.md`;
};

/** The full `<title>` of a page: its own title and the site's. */
export const pageTitle = (title: string): string =>
  `${title} | ${TITLE_SUFFIX}`;

interface Crumb {
  name: string;
  path: string;
}

const trailTo = (
  items: SidebarItem[],
  path: string,
  trail: Crumb[],
): Crumb[] | undefined => {
  for (const item of items) {
    const here = item.href
      ? [...trail, { name: item.label, path: item.href }]
      : trail;
    if (item.href && normalizePath(item.href) === path) return here;

    const below = trailTo(item.items ?? [], path, here);
    if (below) return below;
  }

  return undefined;
};

/**
 * The navigation path to a page in `sections`, from the docs home through every group heading that
 * is itself a page. A page the navigation does not list gets only the home and itself.
 */
export const breadcrumbsFor = (
  sections: SidebarSection[],
  pathname: string,
  title: string,
): Crumb[] => {
  const path = normalizePath(pathname);
  const home = { name: "Docs", path: "/" };

  for (const section of sections) {
    const trail = trailTo(section.items, path, []);
    if (trail) return [home, ...trail];
  }

  return path === "/" ? [home] : [home, { name: title, path }];
};

const publisher = {
  "@type": "Organization",
  "@id": `${PRODUCT_SITE}/#organization`,
  name: "ShellHub",
  url: PRODUCT_SITE,
  logo: `${SITE}/favicon.ico`,
  sameAs: ["https://github.com/shellhub-io/shellhub"],
};

const software = {
  "@type": "SoftwareApplication",
  "@id": `${PRODUCT_SITE}/#software`,
  name: "ShellHub",
  url: PRODUCT_SITE,
  applicationCategory: "DeveloperApplication",
  operatingSystem: "Linux",
  description:
    "Centralized SSH gateway that reaches Linux devices behind NAT and firewalls through an " +
    "outbound agent connection.",
  publisher: { "@id": publisher["@id"] },
};

const website = {
  "@type": "WebSite",
  "@id": `${SITE}/#website`,
  name: TITLE_SUFFIX,
  url: `${SITE}/`,
  inLanguage: "en",
  publisher: { "@id": publisher["@id"] },
  about: { "@id": software["@id"] },
};

/** A question and its answer, as `extractFaq` reads them from a built page. */
export interface Faq {
  question: string;
  answer: string;
}

interface PageData {
  pathname: string;
  title: string;
  description?: string;
  sections: SidebarSection[];
}

/** The schema.org graph for the docs home: the site, its publisher and the product. */
export const homeGraph = (): object => ({
  "@context": "https://schema.org",
  "@graph": [website, publisher, software],
});

/**
 * The schema.org graph for a documentation page: the article and where it sits in `sections`.
 */
export const pageGraph = ({
  pathname,
  title,
  description,
  sections,
}: PageData): object => {
  const url = absoluteUrl(pathname);
  const crumbs = breadcrumbsFor(sections, pathname, title);

  const article = {
    "@type": "TechArticle",
    "@id": `${url}#article`,
    headline: title,
    ...(description ? { description } : {}),
    url,
    mainEntityOfPage: url,
    inLanguage: "en",
    image: `${SITE}${SOCIAL_IMAGE}`,
    isPartOf: { "@id": website["@id"] },
    publisher: { "@id": publisher["@id"] },
    about: { "@id": software["@id"] },
  };

  const breadcrumbs = {
    "@type": "BreadcrumbList",
    itemListElement: crumbs.map((crumb, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: crumb.name,
      item: absoluteUrl(crumb.path),
    })),
  };

  return {
    "@context": "https://schema.org",
    "@graph": [article, breadcrumbs, website, publisher, software],
  };
};

/** The schema.org FAQPage for a page's questions, addressed as a fragment of the page. */
export const faqGraph = (pathname: string, faq: Faq[]): object => ({
  "@context": "https://schema.org",
  "@type": "FAQPage",
  "@id": `${absoluteUrl(pathname)}#faq`,
  url: absoluteUrl(pathname),
  mainEntity: faq.map(({ question, answer }) => ({
    "@type": "Question",
    name: question,
    acceptedAnswer: { "@type": "Answer", text: answer },
  })),
});

/**
 * Serializes structured data for a `<script type="application/ld+json">` element. `<` is escaped
 * so a value containing `</script>` cannot close the element early.
 */
export const jsonLd = (data: object): string =>
  JSON.stringify(data).replace(/</g, "\\u003c");
