/**
 * The console origin. Hardcoded to the local dev URL, as the whole of this module is: the
 * website has no production build or deploy yet, so dev is the only environment it runs in.
 * Making these env-driven is a follow-up, pending a decision on the production hostnames.
 */
export const consoleUrl = "http://localhost";
/**
 * The docs origin. A cross-origin link, so it must not be routed by react-router.
 */
export const docsUrl = import.meta.env.DEV ? "http://docs.localhost" : "https://docs.shellhub.io";
/**
 * This site's own origin, which the shared footer needs to build the links that stay here.
 */
export const websiteUrl = import.meta.env.DEV ? "http://website.localhost" : "https://shellhub.io";
/**
 * The public repository, linked from the navigation and the footer.
 */
export const githubUrl = "https://github.com/shellhub-io/shellhub";
/**
 * Where every sign-in call to action points.
 */
export const loginUrl = `${consoleUrl}/login`;
/**
 * Where every sign-up call to action points.
 */
export const signupUrl = `${consoleUrl}/sign-up`;
