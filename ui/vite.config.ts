/* eslint-disable */
import { defineConfig } from "vite";
import Vue from "@vitejs/plugin-vue";
import VuetifyPlugin from "vite-plugin-vuetify";
import { nodePolyfills } from "vite-plugin-node-polyfills";
import Markdown from "unplugin-vue-markdown/vite";
import { fileURLToPath } from "node:url";
import * as path from "node:path";
import * as fs from "fs";

const dirname = path.dirname(fileURLToPath(import.meta.url));

// https://vitejs.dev/config/
export default defineConfig({
  build: {
    rollupOptions: {
      input: {
        main: path.resolve(dirname, "index.html"),
        admin: path.resolve(dirname, "admin/index.html"),
      },
    },
    target: "es2022",
  },
  plugins: [
    nodePolyfills(),
    Vue(),
    VuetifyPlugin({
      autoImport: true,
      styles: {
        configFile: 'src/settings.scss',
      },
    }),
    Markdown({ markdownItOptions: {
      html: true,
      typographer: true,
    } }),
    {
      name: "admin-handler",
      configureServer(server) {
        server.middlewares.use(async (req, res, next) => {
          if (!req.url?.startsWith("/admin")) return next();

          const parsedUrl = new URL(req.url, "http://localhost");
          const { pathname } = parsedUrl;

          const relativePath = pathname.replace("/admin", "");
          const filePath = path.resolve(__dirname, "admin", `.${relativePath}`);

          if (fs.existsSync(filePath) && fs.statSync(filePath).isFile()) {
            return next();
          }

          const indexHtmlPath = path.resolve(__dirname, "admin/index.html");
          if (!fs.existsSync(indexHtmlPath)) {
            res.statusCode = 404;
            res.end("admin/index.html not found");
            return null;
          }

          let html = fs.readFileSync(indexHtmlPath, "utf-8");
          html = await server.transformIndexHtml(pathname, html);

          res.statusCode = 200;
          res.setHeader("Content-Type", "text/html");
          res.end(html);

          return null;
        });
      },
    },
  ],
  server: {
    port: 8080,
    allowedHosts: true,
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: "modern",
        silenceDeprecations: ['legacy-js-api'],
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
      "@ui": path.resolve(__dirname, "src"),
      "@admin": path.resolve(__dirname, "admin/src"),
    },
  },
  define: {
    "process.env": process.env,
  },
});
