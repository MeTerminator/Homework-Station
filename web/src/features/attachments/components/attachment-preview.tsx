import { useState, lazy, Suspense } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { fileURL } from "@/lib/api/urls";
import { sizeText } from "@/lib/format";
import type { Attachment } from "@/types/models";
import { Download, Eye } from "lucide-react";
const PreviewContent = lazy(() =>
  import("./preview-content").then((m) => ({ default: m.PreviewContent })),
);

export function AttachmentPreview({ file }: { file: Attachment }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button
        type="button"
        variant="ghost"
        className="upload-file link"
        onClick={() => setOpen(true)}
        style={{
          height: "auto",
          whiteSpace: "normal",
          justifyContent: "flex-start",
          textAlign: "left",
        }}
      >
        <Eye size={15} />
        <span style={{ overflowWrap: "anywhere" }}>{file.name}</span>
        <span className="muted">{sizeText(file.size)}</span>
      </Button>
      <AttachmentPreviewDialog file={file} open={open} onOpenChange={setOpen} />
    </>
  );
}

export function AttachmentPreviewDialog({
  file,
  open,
  onOpenChange,
}: {
  file: Attachment;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="sm:max-w-5xl"
        style={{ maxHeight: "90vh", overflow: "auto" }}
      >
        <DialogHeader>
          <DialogTitle className="overflow-name">{file.name}</DialogTitle>
          <DialogDescription>附件预览 · {sizeText(file.size)}</DialogDescription>
        </DialogHeader>
        {open && (
          <Suspense fallback={<p role="status">正在加载预览…</p>}>
            <PreviewContent file={file} />
          </Suspense>
        )}
        <div className="row">
          <Button asChild>
            <a href={fileURL(file.id)}>
              <Download size={15} />
              下载原文件
            </a>
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
