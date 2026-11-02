import { useEffect, useRef, useState, type RefObject } from "react";
import {
  getDocument,
  GlobalWorkerOptions,
  type PDFDocumentProxy,
  type RenderTask,
} from "pdfjs-dist";
import workerURL from "pdfjs-dist/build/pdf.worker.min.mjs?url";
import { fileURL } from "@/lib/api/urls";
GlobalWorkerOptions.workerSrc = workerURL;

function PDFPage({
  doc,
  page,
  name,
  scrollRoot,
}: {
  doc: PDFDocumentProxy;
  page: number;
  name: string;
  scrollRoot: RefObject<HTMLDivElement | null>;
}) {
  const container = useRef<HTMLDivElement>(null);
  const canvas = useRef<HTMLCanvasElement>(null);
  const [visible, setVisible] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [aspectRatio, setAspectRatio] = useState(1 / Math.SQRT2);

  useEffect(() => {
    const target = container.current;
    if (!target) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          setVisible(true);
          observer.disconnect();
        }
      },
      { root: scrollRoot.current, rootMargin: "600px" },
    );
    observer.observe(target);
    return () => observer.disconnect();
  }, [scrollRoot]);

  useEffect(() => {
    if (!visible || !canvas.current) return;
    let active = true;
    let render: RenderTask | undefined;
    const target = canvas.current;
    doc
      .getPage(page)
      .then(async (pdfPage) => {
        if (!active) return;
        const base = pdfPage.getViewport({ scale: 1 });
        setAspectRatio(base.width / base.height);
        const scale = Math.min(2, 1400 / Math.max(base.width, base.height));
        const viewport = pdfPage.getViewport({ scale });
        target.width = viewport.width;
        target.height = viewport.height;
        render = pdfPage.render({ canvas: target, viewport });
        await render.promise;
        if (active) setLoading(false);
      })
      .catch(() => {
        if (active) {
          setError("PDF 页面渲染失败，请下载查看。");
          setLoading(false);
        }
      });
    return () => {
      active = false;
      render?.cancel();
    };
  }, [doc, page, visible]);

  return (
    <div ref={container} className="stack">
      <div
        style={{
          position: "relative",
          aspectRatio,
          background: "white",
        }}
      >
        {loading && (
          <p role="status" style={{ position: "absolute", inset: 12 }}>
            正在渲染第 {page} 页…
          </p>
        )}
        {error && <p role="alert">{error}</p>}
        <canvas
          ref={canvas}
          aria-label={`${name} 第 ${page} 页`}
          role="img"
          style={{
            display: error || loading ? "none" : "block",
            width: "100%",
            height: "auto",
          }}
        />
      </div>
      <p className="muted" style={{ textAlign: "center", margin: 0 }}>
        第 {page} / {doc.numPages} 页
      </p>
    </div>
  );
}

function PDFDocumentPreview({ id, name }: { id: number; name: string }) {
  const [doc, setDoc] = useState<PDFDocumentProxy>();
  const [error, setError] = useState("");
  const scrollRoot = useRef<HTMLDivElement>(null);
  useEffect(() => {
    let active = true;
    const task = getDocument({
      url: fileURL(id),
      cMapUrl: "/pdfjs/cmaps/",
      cMapPacked: true,
      standardFontDataUrl: "/pdfjs/standard_fonts/",
      wasmUrl: "/pdfjs/wasm/",
    });
    task.promise
      .then((pdf) => {
        if (active) setDoc(pdf);
      })
      .catch(() => {
        if (active) setError("PDF 无法预览，文件可能已损坏或需要密码，请下载查看。");
      });
    return () => {
      active = false;
      void task.destroy();
    };
  }, [id]);

  return (
    <div className="stack">
      {!doc && !error && <p role="status">正在加载 PDF…</p>}
      {error && <p role="alert">{error}</p>}
      <div
        ref={scrollRoot}
        role="region"
        aria-label={`${name} PDF 连续预览`}
        tabIndex={0}
        className="stack"
        style={{
          overflow: "auto",
          maxHeight: "60vh",
          background: "#e5e7eb",
          padding: 12,
        }}
      >
        {doc &&
          Array.from({ length: doc.numPages }, (_, index) => (
            <PDFPage
              key={index + 1}
              doc={doc}
              page={index + 1}
              name={name}
              scrollRoot={scrollRoot}
            />
          ))}
      </div>
    </div>
  );
}

export function PDFPreview({ id, name }: { id: number; name: string }) {
  return <PDFDocumentPreview key={id} id={id} name={name} />;
}
