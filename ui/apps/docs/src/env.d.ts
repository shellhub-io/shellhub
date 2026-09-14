/// <reference types="astro/client" />

interface ImportMetaEnv {
  /** `stable` for the build a release publishes, anything else for a build off master. */
  readonly PUBLIC_DOCS_CHANNEL?: string;
  /** The stack's `SHELLHUB_VERSION`, carried in by the workflow that builds the site. */
  readonly PUBLIC_SHELLHUB_VERSION?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
