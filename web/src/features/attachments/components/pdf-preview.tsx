import { useEffect, useRef, useState } from "react";
import {
  getDocument,
  GlobalWorkerOptions,
  type PDFDocumentProxy,
  type RenderTask,
} from "pdfjs-dist";
import workerURL from "pdfjs-dist/build/pdf.worker.min.mjs?url";
import { Button } from "@/components/ui/button";
import { fileURL } from "@/lib/api/urls";
GlobalWorkerOptions.workerSrc = workerURL;

export function PDFPreview({ id, name }: { id: number; name: string }) {
  const [doc, setDoc] = useState<PDFDocumentProxy>();
  const [page, setPage] = useState(1);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const canvas = useRef<HTMLCanvasElement>(null);
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
        if (active) {
          setError("PDF 无法预览，文件可能已损坏或需要密码，请下载查看。");
          setLoading(false);
        }
      });
    return () => {
      active = false;
      void task.destroy();
    };
  }, [id]);
  useEffect(() => {
    if (!doc || !canvas.current) return;
    let active = true;
    let render: RenderTask | undefined;
    const target = canvas.current;
    setLoading(true);
    setError("");
    doc
      .getPage(page)
      .then(async (pdfPage) => {
        if (!active) return;
        const base = pdfPage.getViewport({ scale: 1 });
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
  }, [doc, page]);
  return (
    <div className="stack">
      {loading && <p role="status">正在渲染 PDF…</p>}
      {error && <p role="alert">{error}</p>}
      <div
        style={{
          overflow: "auto",
          maxHeight: "60vh",
          background: "#e5e7eb",
          padding: 12,
        }}
      >
        <canvas
          ref={canvas}
          aria-label={`${name} 第 ${page} 页`}
          role="img"
          style={{
            display: error || loading ? "none" : "block",
            maxWidth: "100%",
            height: "auto",
            margin: "auto",
          }}
        />
      </div>
      {doc && (
        <div className="row" style={{ justifyContent: "center" }}>
          <Button
            variant="outline"
            disabled={loading || page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            上一页
          </Button>
          <span>
            第 {page} / {doc.numPages} 页
          </span>
          <Button
            variant="outline"
            disabled={loading || page >= doc.numPages}
            onClick={() => setPage((p) => p + 1)}
          >
            下一页
          </Button>
        </div>
      )}
    </div>
  );
}
