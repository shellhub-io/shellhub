/**
 * The ShellHub version the pages name: this release when a release built the site, the newest tag
 * there is otherwise — a release candidate included, since a build off master describes what is
 * coming. Null only where nothing supplied it, which is a dev server.
 */
export const releaseVersion = import.meta.env.PUBLIC_SHELLHUB_VERSION || null;

/** The tag to put in a command. `vX.Y.Z` stands in on a dev server, and nowhere that is served. */
export const releaseTag = releaseVersion ?? "vX.Y.Z";
