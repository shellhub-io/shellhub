const releaseVersion = import.meta.env.PUBLIC_SHELLHUB_VERSION || null;

/** The tag to put in a command. `vX.Y.Z` stands in on a dev server, and nowhere that is served. */
export const releaseTag = releaseVersion ?? "vX.Y.Z";
