import { useEffect, useRef, useState } from "react";
import { Popover } from "radix-ui";
import { Paperclip, Eye } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AttachmentPreviewDialog } from "@/features/attachments/components/attachment-preview";
import { sizeText } from "@/lib/format";
import type { Attachment } from "@/types/models";

export function SubmissionAttachments({ files }: { files: Attachment[] }) {
  const [open, setOpen] = useState(false);
  const [preview, setPreview] = useState<Attachment | null>(null);
  const closeTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  useEffect(() => () => clearTimeout(closeTimer.current), []);
  function show() {
    clearTimeout(closeTimer.current);
    setOpen(true);
  }
  function scheduleClose() {
    clearTimeout(closeTimer.current);
    closeTimer.current = setTimeout(() => setOpen(false), 200);
  }
  if (!files.length) return <span className="muted">无附件</span>;
  return (
    <>
      <Popover.Root open={open} onOpenChange={setOpen}>
        <Popover.Trigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="submission-attachments-trigger"
            aria-label={`查看全部 ${files.length} 个附件：${files[0].name}`}
            onMouseEnter={show}
            onMouseLeave={scheduleClose}
          >
            <Paperclip size={14} />
            <span className="submission-attachment-name" title={files[0].name}>
              {files[0].name}
            </span>
            {files.length > 1 && <span className="muted">+{files.length - 1}</span>}
          </Button>
        </Popover.Trigger>
        <Popover.Portal>
          <Popover.Content
            className="submission-attachments-popover"
            sideOffset={6}
            collisionPadding={12}
            align="start"
            aria-label="提交附件"
            onOpenAutoFocus={(e) => e.preventDefault()}
            onCloseAutoFocus={(e) => e.preventDefault()}
            onMouseEnter={show}
            onMouseLeave={scheduleClose}
          >
            <p className="small muted">全部附件 · {files.length} 个</p>
            <div className="submission-attachments-list">
              {files.map((file) => (
                <Button
                  type="button"
                  variant="ghost"
                  className="submission-attachment-item"
                  key={file.id}
                  onClick={() => {
                    clearTimeout(closeTimer.current);
                    setOpen(false);
                    setPreview(file);
                  }}
                >
                  <Eye size={15} />
                  <span className="submission-attachment-name" title={file.name}>
                    {file.name}
                  </span>
                  <small className="muted">{sizeText(file.size)}</small>
                </Button>
              ))}
            </div>
          </Popover.Content>
        </Popover.Portal>
      </Popover.Root>
      {preview && (
        <AttachmentPreviewDialog
          file={preview}
          open
          onOpenChange={(value) => {
            if (!value) setPreview(null);
          }}
        />
      )}
    </>
  );
}
