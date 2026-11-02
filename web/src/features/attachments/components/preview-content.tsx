import { lazy, Suspense, useEffect, useMemo, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeSanitize from "rehype-sanitize";
import hljs from "highlight.js/lib/common";
import "highlight.js/styles/github.css";
import { fileURL } from "@/lib/api/urls";
import type { Attachment } from "@/types/models";

const PDFPreview = lazy(() =>
  import("./pdf-preview").then((m) => ({ default: m.PDFPreview })),
);

const languages: Record<string, string> = {
  js: "javascript",
  jsx: "javascript",
  mjs: "javascript",
  cjs: "javascript",
  ts: "typescript",
  tsx: "typescript",
  py: "python",
  go: "go",
  java: "java",
  c: "c",
  h: "c",
  cpp: "cpp",
  cc: "cpp",
  cxx: "cpp",
  hpp: "cpp",
  cs: "csharp",
  rs: "rust",
  rb: "ruby",
  php: "php",
  swift: "swift",
  kt: "kotlin",
  kts: "kotlin",
  sh: "bash",
  bash: "bash",
  zsh: "bash",
  json: "json",
  yml: "yaml",
  yaml: "yaml",
  xml: "xml",
  html: "xml",
  htm: "xml",
  svg: "xml",
  vue: "xml",
  svelte: "xml",
  css: "css",
  scss: "scss",
  less: "less",
  sql: "sql",
  r: "r",
  lua: "lua",
  pl: "perl",
  ini: "ini",
  toml: "ini",
  conf: "ini",
  dart: "dart",
  m: "objectivec",
  mm: "objectivec",
  txt: "plaintext",
  log: "plaintext",
  csv: "plaintext",
  tsv: "plaintext",
};
export function CodeView({ text, language }: { text: string; language: string }) {
  const html = useMemo(
    () =>
      hljs.highlight(text, {
        language: hljs.getLanguage(language) ? language : "plaintext",
        ignoreIllegals: true,
      }).value,
    [text, language],
  );
  // Only escaped output produced by highlight.js enters the HTML sink.
  return <code className="hljs" dangerouslySetInnerHTML={{ __html: html }} />;
}
export function PreviewContent({ file }: { file: Attachment }) {
  const ext = file.name.split(".").pop()?.toLowerCase() || "";
  const markdown = ["md", "markdown", "mdown", "mkd"].includes(ext);
  const language =
    languages[ext] ||
    (/^(dockerfile|makefile)$/i.test(file.name) ? file.name.toLowerCase() : "");
  const textFile = markdown || !!language;
  const [result, setResult] = useState<{ text?: string; error?: string }>({});
  const [mediaError, setMediaError] = useState(false);
  useEffect(() => {
    if (!textFile || file.size > 2 * 1024 * 1024) return;
    const controller = new AbortController();
    async function load() {
      try {
        const res = await fetch(fileURL(file.id), {
          credentials: "same-origin",
          signal: controller.signal,
        });
        if (!res.ok) throw new Error("预览加载失败，请重试或下载原文件。");
        const reader = res.body?.getReader();
        if (!reader) throw new Error("无法读取文件，请下载原文件。");
        const chunks: Uint8Array[] = [];
        let size = 0;
        while (true) {
          const { value, done } = await reader.read();
          if (done) break;
          size += value.byteLength;
          if (size > 2 * 1024 * 1024) {
            await reader.cancel();
            throw new Error("文本超过 2MB，请下载查看。");
          }
          chunks.push(value);
        }
        const bytes = new Uint8Array(size);
        let offset = 0;
        for (const chunk of chunks) {
          bytes.set(chunk, offset);
          offset += chunk.length;
        }
        const text = new TextDecoder("utf-8", { fatal: true }).decode(bytes);
        if (text.includes("\0")) throw new Error("文件不是可预览的文本，请下载查看。");
        if (!controller.signal.aborted) setResult({ text });
      } catch (e) {
        if (!controller.signal.aborted)
          setResult({ error: e instanceof Error ? e.message : "预览失败" });
      }
    }
    void load();
    return () => controller.abort();
  }, [file.id, file.size, textFile]);
  if (file.status !== "complete")
    return <p className="note">文件尚未上传完成，暂不支持预览。</p>;
  if (["png", "jpg", "jpeg", "gif", "webp", "bmp", "ico"].includes(ext))
    return mediaError ? (
      <p role="alert">图片预览失败，请下载查看。</p>
    ) : (
      <img
        src={fileURL(file.id, true)}
        alt={file.name}
        onError={() => setMediaError(true)}
        style={{
          maxWidth: "100%",
          maxHeight: "65vh",
          objectFit: "contain",
          margin: "auto",
        }}
      />
    );
  if (ext === "pdf")
    return (
      <Suspense fallback={<p role="status">正在加载 PDF 阅读器…</p>}>
        <PDFPreview id={file.id} name={file.name} />
      </Suspense>
    );
  if (!textFile) return <p className="note">此文件格式暂不支持在线预览，请下载查看。</p>;
  if (file.size > 2 * 1024 * 1024)
    return <p className="note">文本超过 2MB，请下载查看。</p>;
  if (result.error) return <p role="alert">{result.error}</p>;
  if (result.text === undefined) return <p role="status">正在读取文件…</p>;
  if (markdown)
    return (
      <div className="markdown attachment-markdown">
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          rehypePlugins={[rehypeSanitize]}
          components={{
            pre({ children }) {
              return <pre className="code-preview">{children}</pre>;
            },
            code({ children, className }) {
              const lang = /language-(\S+)/.exec(className || "")?.[1];
              return lang ? (
                <CodeView text={String(children).replace(/\n$/, "")} language={lang} />
              ) : (
                <code className={className}>{children}</code>
              );
            },
          }}
        >
          {result.text}
        </ReactMarkdown>
      </div>
    );
  return (
    <pre className="code-preview">
      <CodeView text={result.text} language={language} />
    </pre>
  );
}
