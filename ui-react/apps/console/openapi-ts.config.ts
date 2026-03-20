import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input:
    process.env.OPENAPI_SPEC_PATH || "http://openapi:8080/openapi/openapi.json",
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
