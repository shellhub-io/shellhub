import { defineConfig } from "orval";

const input = process.env.OPENAPI_SPEC_PATH;
if (!input) {
  throw new Error(
    "OPENAPI_SPEC_PATH is not set; run `npm run generate -w @shellhub/console`.",
  );
}

export default defineConfig({
  shellhub: {
    input: {
      target: input,
    },
    output: {
      target: "./src/client/api.ts",
      schemas: "./src/client/model",
      client: "react-query",
      httpClient: "fetch",
      mode: "single",
      clean: true,
      override: {
        mutator: {
          path: "./src/api/customInstance.ts",
          name: "customInstance",
        },
        fetch: {
          includeHttpResponseReturnType: false,
        },
        query: {
          signal: true,
        },
      },
      mock: {
        generators: [{ type: "msw" }, { type: "faker" }],
      },
    },
  },
});
