/**
 * URL of the public website used by the docs app.
 * Uses a local development host when running in DEV mode, otherwise the production URL.
 */
export const websiteUrl = import.meta.env.DEV ? "http://website.localhost" : "https://shellhub.io";

/**
 * URL of this documentation site. Same rule as the website: a local host in development, the
 * published address otherwise.
 */
export const docsUrl = import.meta.env.DEV ? "http://docs.localhost" : "https://docs.shellhub.io";
