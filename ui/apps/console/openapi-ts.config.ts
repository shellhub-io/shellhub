import { defineConfig } from "@hey-api/openapi-ts";

const input = process.env.OPENAPI_SPEC_PATH;
if (!input) {
  // The openapi container only serves the spec for the active edition, so
  // generating from its URL produces a client missing the other editions'
  // schemas. Use `npm run generate -w @shellhub/console`, which bundles the
  // combined spec before invoking openapi-ts.
  throw new Error("OPENAPI_SPEC_PATH is not set; run `npm run generate -w @shellhub/console`.");
}

export default defineConfig({
  input,
  output: "src/client",
  plugins: [
    "@hey-api/typescript",
    "@hey-api/sdk",
    {
      name: "@hey-api/client-fetch",
      runtimeConfigPath: "../api/fetchClient",
    },
    {
      name: "@tanstack/react-query",
      queryOptions: true,
      mutationOptions: true,
    },
  ],
});
