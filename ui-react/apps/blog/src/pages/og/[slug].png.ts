import type { APIRoute, GetStaticPaths } from "astro";
import { getCollection } from "astro:content";
import satori from "satori";
import { Resvg } from "@resvg/resvg-js";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { techIcons } from "../../lib/tech-icons";

const rootDir = process.cwd();
const fontsDir = join(rootDir, "src/fonts");

const plexSansRegular = readFileSync(join(fontsDir, "plex-sans-regular.ttf"));
const plexSansBold = readFileSync(join(fontsDir, "plex-sans-bold.ttf"));
const plexMono = readFileSync(join(fontsDir, "plex-mono-medium.ttf"));

// Logo as base64 data URL for Satori <img>
const logoPng = readFileSync(join(rootDir, "public/logo-inverted.png"));
const logoDataUrl = `data:image/png;base64,${logoPng.toString("base64")}`;

export const getStaticPaths: GetStaticPaths = async () => {
  const posts = await getCollection("blog");
  return posts.map((post) => ({
    params: { slug: post.id },
    props: { post },
  }));
};

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

/** Build a Satori SVG element from a tech icon definition */
function renderIcon(
  name: string,
  size: number,
  color: string,
  forceColor = false,
) {
  const icon = techIcons[name];
  if (!icon) return null;

  return {
    type: "svg",
    props: {
      width: String(size),
      height: String(size),
      viewBox: icon.viewBox,
      fill: "none",
      children: icon.paths.map((p) => ({
        type: "path",
        props: {
          d: p.d,
          ...(p.fill
            ? { fill: forceColor ? color : p.fill === "currentColor" ? color : p.fill }
            : {}),
          ...(p.stroke
            ? { stroke: forceColor ? color : p.stroke === "currentColor" ? color : p.stroke }
            : {}),
          ...(p.strokeWidth ? { strokeWidth: p.strokeWidth } : {}),
          ...(p.strokeLinecap ? { strokeLinecap: p.strokeLinecap } : {}),
          ...(p.strokeLinejoin ? { strokeLinejoin: p.strokeLinejoin } : {}),
        },
      })),
    },
  };
}

// Background decorative dots
const dots = [
  { x: 180, y: 90, size: 5 },
  { x: 920, y: 65, size: 4 },
  { x: 1050, y: 180, size: 6 },
  { x: 350, y: 520, size: 4 },
  { x: 780, y: 480, size: 5 },
  { x: 1100, y: 430, size: 4 },
  { x: 60, y: 350, size: 3 },
  { x: 600, y: 100, size: 3 },
];

const hLines = [
  { x: 100, y: 130, w: 220 },
  { x: 850, y: 110, w: 180 },
  { x: 200, y: 490, w: 160 },
  { x: 900, y: 460, w: 200 },
];
const vLines = [
  { x: 170, y: 80, h: 140 },
  { x: 1060, y: 160, h: 120 },
  { x: 370, y: 440, h: 100 },
  { x: 790, y: 400, h: 110 },
];

function buildBackground() {
  return [
    // Grid
    {
      type: "div",
      props: {
        style: {
          position: "absolute" as const,
          top: 0, left: 0, right: 0, bottom: 0,
          backgroundImage:
            "linear-gradient(rgba(102,122,204,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(102,122,204,0.03) 1px, transparent 1px)",
          backgroundSize: "40px 40px",
        },
      },
    },
    // Central glow
    {
      type: "div",
      props: {
        style: {
          position: "absolute" as const,
          top: "50%", left: "40%",
          width: "800px", height: "600px",
          borderRadius: "50%",
          background: "radial-gradient(ellipse, rgba(102,122,204,0.08) 0%, transparent 70%)",
          transform: "translate(-50%, -50%)",
        },
      },
    },
    // Top-right glow
    {
      type: "div",
      props: {
        style: {
          position: "absolute" as const,
          top: "-100px", right: "100px",
          width: "400px", height: "400px",
          borderRadius: "50%",
          background: "radial-gradient(circle, rgba(102,122,204,0.06) 0%, transparent 70%)",
        },
      },
    },
    // Dots
    ...dots.map((dot) => ({
      type: "div",
      props: {
        style: {
          position: "absolute" as const,
          left: `${dot.x}px`,
          top: `${dot.y}px`,
          width: `${dot.size}px`,
          height: `${dot.size}px`,
          borderRadius: "50%",
          backgroundColor: "#667ACC",
          boxShadow: "0 0 8px rgba(102,122,204,0.6), 0 0 16px rgba(102,122,204,0.3)",
        },
      },
    })),
    // H-lines
    ...hLines.map((line) => ({
      type: "div",
      props: {
        style: {
          position: "absolute" as const,
          left: `${line.x}px`,
          top: `${line.y}px`,
          width: `${line.w}px`,
          height: "1px",
          background: "linear-gradient(90deg, transparent, rgba(102,122,204,0.15), transparent)",
        },
      },
    })),
    // V-lines
    ...vLines.map((line) => ({
      type: "div",
      props: {
        style: {
          position: "absolute" as const,
          left: `${line.x}px`,
          top: `${line.y}px`,
          width: "1px",
          height: `${line.h}px`,
          background: "linear-gradient(180deg, transparent, rgba(102,122,204,0.15), transparent)",
        },
      },
    })),
  ];
}

function buildCategoryBadge(category: string) {
  return {
    type: "div",
    props: {
      style: { display: "flex" },
      children: [
        {
          type: "span",
          props: {
            style: {
              fontFamily: "IBM Plex Mono",
              fontSize: "12px",
              fontWeight: 500,
              letterSpacing: "0.12em",
              color: "#667ACC",
              backgroundColor: "rgba(102,122,204,0.08)",
              border: "1px solid rgba(102,122,204,0.15)",
              borderRadius: "9999px",
              padding: "6px 16px",
            },
            children: category,
          },
        },
      ],
    },
  };
}

function buildLogo() {
  return {
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        position: "relative" as const,
      },
      children: [
        {
          type: "img",
          props: {
            src: logoDataUrl,
            height: 28,
          },
        },
      ],
    },
  };
}

function buildAccentLine() {
  return {
    type: "div",
    props: {
      style: {
        position: "absolute" as const,
        bottom: 0,
        left: 0,
        right: 0,
        height: "2px",
        background: "linear-gradient(90deg, transparent 10%, #667ACC 50%, transparent 90%)",
      },
    },
  };
}

function buildTechBadges(iconNames: string[]) {
  return iconNames.map((name) => {
    const icon = techIcons[name];
    const label = icon?.label || name;

    return {
      type: "div",
      props: {
        style: {
          display: "flex",
          alignItems: "center",
          gap: "6px",
          fontFamily: "IBM Plex Mono",
          fontSize: "11px",
          fontWeight: 500,
          letterSpacing: "0.05em",
          color: "#667ACC",
          backgroundColor: "rgba(102,122,204,0.06)",
          border: "1px solid rgba(102,122,204,0.12)",
          borderRadius: "6px",
          padding: "5px 12px",
        },
        children: [
          // Small icon inside badge
          ...(renderIcon(name, 14, "#667ACC") ? [renderIcon(name, 14, "#667ACC")] : []),
          { type: "span", props: { children: label } },
        ],
      },
    };
  });
}

// ---------------------------------------------------------------------------
// Layout: Default
// ---------------------------------------------------------------------------

interface LayoutOptions {
  category: string;
  tagline: string;
  iconNames: string[];
  partnerIcon?: string;
}

function buildDefaultLayout(opts: LayoutOptions) {
  const { category, tagline, iconNames } = opts;

  // Determine tagline font size based on length
  const taglineFontSize = tagline.length > 40 ? "44px" : tagline.length > 25 ? "52px" : "60px";

  // Build large decorative icons (watermark-style, right side)
  const decorativeIcons = iconNames.slice(0, 3).map((name, i) => {
    const icon = renderIcon(name, 120, "rgba(102,122,204,0.12)");
    if (!icon) return null;

    const positions = [
      { right: 80, top: 140 },
      { right: 240, top: 260 },
      { right: 100, top: 380 },
    ];
    const pos = positions[i];

    return {
      type: "div",
      props: {
        style: {
          position: "absolute" as const,
          right: `${pos.right}px`,
          top: `${pos.top}px`,
          display: "flex",
        },
        children: [icon],
      },
    };
  }).filter(Boolean);

  const techBadges = buildTechBadges(iconNames);

  return [
    ...buildBackground(),
    // Large decorative icons (watermark)
    ...decorativeIcons,
    // Content: category + tagline + tech badges
    {
      type: "div",
      props: {
        style: {
          display: "flex",
          flexDirection: "column",
          gap: "20px",
          position: "relative" as const,
          maxWidth: "750px",
        },
        children: [
          buildCategoryBadge(category),
          // Tagline
          {
            type: "h1",
            props: {
              style: {
                fontFamily: "IBM Plex Sans",
                fontSize: taglineFontSize,
                fontWeight: 700,
                color: "#E1E4EA",
                lineHeight: 1.1,
                letterSpacing: "-0.03em",
                margin: 0,
              },
              children: tagline,
            },
          },
          // Tech badges row
          ...(techBadges.length > 0
            ? [
                {
                  type: "div",
                  props: {
                    style: {
                      display: "flex",
                      gap: "8px",
                      flexWrap: "wrap" as const,
                      marginTop: "4px",
                    },
                    children: techBadges,
                  },
                },
              ]
            : []),
        ],
      },
    },
    // Bottom: ShellHub logo
    buildLogo(),
    // Bottom accent line
    buildAccentLine(),
  ];
}

// ---------------------------------------------------------------------------
// Layout: Integration
// ---------------------------------------------------------------------------

function buildIntegrationLayout(opts: LayoutOptions) {
  const { category, tagline, iconNames, partnerIcon } = opts;

  // Smaller tagline for integration layout
  const taglineFontSize = tagline.length > 35 ? "36px" : tagline.length > 20 ? "42px" : "48px";

  const techBadges = buildTechBadges(iconNames);

  // Build the two large icon containers (ShellHub + partner)
  const shellhubIcon = renderIcon("shellhub", 100, "#E1E4EA", true);
  const partnerIconEl = partnerIcon ? renderIcon(partnerIcon, 100, "#E1E4EA", true) : null;

  const iconContainer = (iconEl: ReturnType<typeof renderIcon>) => ({
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        width: "160px",
        height: "160px",
        borderRadius: "32px",
        backgroundColor: "rgba(102,122,204,0.08)",
        border: "1px solid rgba(102,122,204,0.2)",
      },
      children: iconEl ? [iconEl] : [],
    },
  });

  // Connector "x" between icons
  const connector = {
    type: "div",
    props: {
      style: {
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        fontFamily: "IBM Plex Sans",
        fontSize: "28px",
        fontWeight: 700,
        color: "rgba(102,122,204,0.6)",
        padding: "0 24px",
      },
      children: "\u00d7",
    },
  };

  return [
    ...buildBackground(),
    // Top-left category badge
    {
      type: "div",
      props: {
        style: {
          display: "flex",
          position: "relative" as const,
        },
        children: [buildCategoryBadge(category)],
      },
    },
    // Center: icon pair + tagline + badges
    {
      type: "div",
      props: {
        style: {
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: "28px",
          position: "relative" as const,
          flex: 1,
          justifyContent: "center",
        },
        children: [
          // Icon pair row
          {
            type: "div",
            props: {
              style: {
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              },
              children: [
                iconContainer(shellhubIcon),
                connector,
                iconContainer(partnerIconEl),
              ],
            },
          },
          // Tagline (centered, smaller)
          {
            type: "div",
            props: {
              style: {
                fontFamily: "IBM Plex Sans",
                fontSize: taglineFontSize,
                fontWeight: 700,
                color: "#E1E4EA",
                lineHeight: 1.1,
                letterSpacing: "-0.03em",
                textAlign: "center" as const,
              },
              children: tagline,
            },
          },
          // Tech badges (centered)
          ...(techBadges.length > 0
            ? [
                {
                  type: "div",
                  props: {
                    style: {
                      display: "flex",
                      gap: "8px",
                      flexWrap: "wrap" as const,
                      justifyContent: "center",
                    },
                    children: techBadges,
                  },
                },
              ]
            : []),
        ],
      },
    },
    // Bottom: ShellHub logo
    buildLogo(),
    // Bottom accent line
    buildAccentLine(),
  ];
}

// ---------------------------------------------------------------------------
// GET handler
// ---------------------------------------------------------------------------

export const GET: APIRoute = async ({ props }) => {
  const { post } = props as { post: Awaited<ReturnType<typeof getCollection>>[number] };

  const category = post.data.categories?.[0]?.toUpperCase() || "BLOG";
  const tagline = post.data.ogTagline || post.data.title;
  const iconNames: string[] = post.data.ogIcons || [];
  const layout = post.data.ogLayout || "default";
  const partnerIcon: string | undefined = post.data.ogPartnerIcon;

  const opts: LayoutOptions = { category, tagline, iconNames, partnerIcon };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let children: any[];
  switch (layout) {
    case "integration":
      children = buildIntegrationLayout(opts);
      break;
    default:
      children = buildDefaultLayout(opts);
  }

  const svg = await satori(
    {
      type: "div",
      props: {
        style: {
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          backgroundColor: "#18191B",
          padding: "64px 80px",
          position: "relative",
          overflow: "hidden",
        },
        children,
      },
    },
    {
      width: 1200,
      height: 630,
      fonts: [
        { name: "IBM Plex Sans", data: plexSansRegular, weight: 400, style: "normal" as const },
        { name: "IBM Plex Sans", data: plexSansBold, weight: 700, style: "normal" as const },
        { name: "IBM Plex Mono", data: plexMono, weight: 500, style: "normal" as const },
      ],
    },
  );

  const resvg = new Resvg(svg, {
    fitTo: { mode: "width", value: 1200 },
  });

  const png = resvg.render().asPng();

  return new Response(png, {
    headers: {
      "Content-Type": "image/png",
      "Cache-Control": "public, max-age=31536000, immutable",
    },
  });
};
