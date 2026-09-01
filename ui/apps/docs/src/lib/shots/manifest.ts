import { isDeepStrictEqual } from "node:util";

/** Which edition a shot needs the instance to be running. */
export type ShotEdition = "ce" | "enterprise";

/** How the capture finds an element: by role and name, text, or test id. */
export interface ShotSelector {
  role?: string;
  name?: string;
  text?: string;
  testId?: string;
}

/** A step the capture performs before the picture is taken. */
export interface ShotInteraction {
  click?: ShotSelector;
}

/** The browser size a shot is taken at. */
export interface ShotViewport {
  width: number;
  height: number;
}

/** A single `<Shot>` tag, as recorded while the page that contains it renders. */
export interface ShotDeclaration {
  id: string;
  route: string;
  page: string;
  of?: ShotSelector;
  viewport?: ShotViewport;
  edition?: ShotEdition;
  before?: ShotInteraction[];
}

/** One shot in the manifest, with the pages that declared it. */
export interface ManifestShot {
  id: string;
  route: string;
  viewport: ShotViewport;
  edition: ShotEdition;
  of?: ShotSelector;
  before?: ShotInteraction[];
  usedBy: string[];
}

/** The whole shot list, as the capture tool reads it. */
export interface ShotManifest {
  shots: ManifestShot[];
}

/** The viewport a shot is taken at unless it asks for another. */
export const DEFAULT_VIEWPORT: ShotViewport = { width: 1440, height: 900 };
/** The edition assumed when a shot does not name one. */
export const DEFAULT_EDITION: ShotEdition = "ce";

type Capture = Omit<ManifestShot, "usedBy">;

function toCapture(declaration: ShotDeclaration): Capture {
  const capture: Capture = {
    id: declaration.id,
    route: declaration.route,
    viewport: declaration.viewport ?? DEFAULT_VIEWPORT,
    edition: declaration.edition ?? DEFAULT_EDITION,
  };

  if (declaration.of) capture.of = declaration.of;
  if (declaration.before) capture.before = declaration.before;

  return capture;
}

/**
 * Collapse the declarations gathered during a build into the manifest the
 * capture tool consumes.
 *
 * Ordering is fixed rather than inherited from the build: Astro renders pages
 * concurrently, so input order varies between runs that changed no content.
 */
export function buildManifest(declarations: ShotDeclaration[]): ShotManifest {
  const collected = new Map<string, { capture: Capture; pages: Set<string> }>();

  for (const declaration of declarations) {
    const capture = toCapture(declaration);
    const existing = collected.get(capture.id);

    if (!existing) {
      collected.set(capture.id, {
        capture,
        pages: new Set([declaration.page]),
      });
      continue;
    }

    if (!isDeepStrictEqual(existing.capture, capture)) {
      throw new Error(
        `shot "${capture.id}" is declared twice with different capture coordinates; ` +
          "give one of them a new id, or make the two tags identical",
      );
    }

    existing.pages.add(declaration.page);
  }

  const shots = [...collected.values()]
    .map(({ capture, pages }) => ({ ...capture, usedBy: [...pages].sort() }))
    .sort((a, b) => (a.id < b.id ? -1 : a.id > b.id ? 1 : 0));

  return { shots };
}
