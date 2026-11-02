import { readFileSync, readdirSync } from "node:fs";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { fileURLToPath, URL } from "node:url";
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    {
      name: "pdfjs-resources",
      generateBundle() {
        for (const folder of ["cmaps", "standard_fonts", "wasm"]) {
          const dir = new URL(`./node_modules/pdfjs-dist/${folder}/`, import.meta.url);
          for (const name of readdirSync(dir)) {
            this.emitFile({
              type: "asset",
              fileName: `pdfjs/${folder}/${name}`,
              source: readFileSync(new URL(name, dir)),
            });
          }
        }
      },
      configureServer(server) {
        server.middlewares.use((req, res, next) => {
          const match = /^\/pdfjs\/(cmaps|standard_fonts|wasm)\/([\w.-]+)$/.exec(
            (req.url || "").split("?")[0],
          );
          if (!match) return next();
          try {
            const bytes = readFileSync(
              new URL(
                `./node_modules/pdfjs-dist/${match[1]}/${match[2]}`,
                import.meta.url,
              ),
            );
            res.setHeader(
              "Content-Type",
              match[2].endsWith(".wasm")
                ? "application/wasm"
                : "application/octet-stream",
            );
            res.end(bytes);
          } catch {
            next();
          }
        });
      },
    },
  ],
  resolve: { alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) } },
  build: {
    outDir: "../backend/internal/frontend/dist",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: {
          charts: ["recharts"],
          markdown: ["react-markdown", "remark-gfm"],
          hash: ["hash-wasm"],
        },
      },
    },
  },
  server: { proxy: { "/api": "http://127.0.0.1:8080" } },
});
