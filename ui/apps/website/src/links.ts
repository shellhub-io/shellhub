const devDomain = import.meta.env.PUBLIC_SHELLHUB_DOMAIN || "localhost";

/**
 * The console origin. Always the dev stack's: the website has no production build or deploy
 * yet, so dev is the only environment it runs in.
 */
export const consoleUrl = `http://${devDomain}`;
/**
 * The docs origin. A cross-origin link, so it must not be routed by react-router.
 */
export const docsUrl = import.meta.env.DEV
  ? `http://docs.${devDomain}`
  : "https://docs.shellhub.io";
/**
 * This site's own origin, which the shared footer needs to build the links that stay here.
 */
export const websiteUrl = import.meta.env.DEV
  ? `http://website.${devDomain}`
  : "https://shellhub.io";
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
