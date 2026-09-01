import { defineConfig } from "@hey-api/openapi-ts";

const input = process.env.OPENAPI_SPEC_PATH;
if (!input) {
  throw new Error(
    "OPENAPI_SPEC_PATH is not set; run `npm run generate -w @shellhub/console`.",
  );
}

export default defineConfig({
  input,
  output: "src/client",
  plugins: [
    "@hey-api/typescript",
    "@hey-api/sdk",
    {
      name: "@hey-api/client-fetch",
      runtimeConfigPath: "./src/api/fetchClient",
    },
    {
      name: "@tanstack/react-query",
      queryOptions: true,
      mutationOptions: true,
      includeInEntry: true,
    },
  ],
});
