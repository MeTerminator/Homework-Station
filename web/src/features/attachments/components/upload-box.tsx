import { Button } from "@/components/ui/button";
import { upload } from "@/features/attachments/api/upload";
import { fileURL } from "@/lib/api/urls";
import { sizeText } from "@/lib/format";
import type { Attachment } from "@/types/models";
import { FileText, Upload, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
export function UploadBox({
  files,
  onChange,
  assignmentID,
  maxSize,
  extensions,
  onBusy,
}: {
  files: Attachment[];
  onChange: (f: Attachment[]) => void;
  assignmentID?: number;
  maxSize?: number;
  extensions?: string;
  onBusy?: (b: boolean) => void;
}) {
  const [progress, setProgress] = useState(0);
  const [label, setLabel] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const controller = useRef<AbortController | null>(null);
  const liveFiles = useRef(files);
  liveFiles.current = files;
  useEffect(() => () => controller.current?.abort(), []);
  const start = async (file?: File) => {
    if (!file) return;
    if (maxSize && file.size > maxSize) {
      toast.error(`文件不能超过 ${sizeText(maxSize)}`);
      return;
    }
    if (extensions?.trim()) {
      const allowed = extensions
        .split(/[,;，\s]+/)
        .filter(Boolean)
        .some((s) =>
          file.name.toLowerCase().endsWith("." + s.toLowerCase().replace(/^\./, "")),
        );
      if (!allowed) {
        toast.error("文件后缀不符合要求");
        return;
      }
    }
    setBusy(true);
    onBusy?.(true);
    setError("");
    setProgress(0);
    controller.current = new AbortController();
    try {
      const f = await upload(
        file,
        assignmentID,
        (p, l) => {
          setProgress(p);
          setLabel(l);
        },
        controller.current.signal,
      );
      if (!liveFiles.current.some((x) => x.id === f.id))
        onChange([...liveFiles.current, f]);
      toast.success("附件上传完成");
    } catch (e) {
      setError(
        controller.current.signal.aborted
          ? "上传已暂停，重新选择同一文件即可续传"
          : e instanceof Error
            ? e.message
            : "上传失败",
      );
    } finally {
      setBusy(false);
      onBusy?.(false);
    }
  };
  return (
    <div className="stack" style={{ gap: 10 }}>
      <div
        className="upload-zone"
        onDragOver={(e) => e.preventDefault()}
        onDrop={(e) => {
          e.preventDefault();
          if (!busy) void start(e.dataTransfer.files[0]);
        }}
      >
        <Upload size={24} style={{ margin: "0 auto 10px", color: "var(--primary)" }} />
        <p>
          拖拽文件到这里，或{" "}
          <label className="link" style={{ cursor: busy ? "wait" : "pointer" }}>
            选择文件
            <input
              aria-label="选择上传附件"
              type="file"
              hidden
              disabled={busy}
              onChange={(e) => {
                void start(e.target.files?.[0]);
                e.target.value = "";
              }}
            />
          </label>
        </p>
        <p style={{ marginTop: 7, fontSize: 11 }}>
          支持分片上传与断点续传
          {maxSize ? ` · 单个附件不超过 ${sizeText(maxSize)}` : ""}
          {extensions ? ` · ${extensions}` : ""}
        </p>
        {busy && (
          <>
            <div className="progress">
              <div style={{ width: `${progress}%` }} />
            </div>
            <div className="row" style={{ justifyContent: "center", marginTop: 10 }}>
              {label} · {Math.round(progress)}%
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => controller.current?.abort()}
              >
                暂停上传
              </Button>
            </div>
          </>
        )}
        {error && <p style={{ marginTop: 10, color: "#dc6472" }}>{error}</p>}
      </div>
      {files.map((f) => (
        <div className="upload-file" key={f.id}>
          <FileText size={17} className="muted" />
          <span>
            <a href={fileURL(f.id)} className="link">
              {f.name}
            </a>
            <small className="muted" style={{ marginLeft: 10 }}>
              {sizeText(f.size)}
            </small>
          </span>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => onChange(files.filter((x) => x.id !== f.id))}
            aria-label={`移除 ${f.name}`}
          >
            <X size={15} />
          </Button>
        </div>
      ))}
    </div>
  );
}
