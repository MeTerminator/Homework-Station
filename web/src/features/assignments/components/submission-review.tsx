import { useSession } from "@/app/providers/session-provider";
import { LoadState } from "@/components/feedback/load-state";
import { MarkdownEditor } from "@/components/markdown-editor";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { UploadBox } from "@/features/attachments/components/upload-box";
import { AttachmentPreview } from "@/features/attachments/components/attachment-preview";
import { useData } from "@/hooks/use-data";
import { perform } from "@/lib/actions";
import { api, json } from "@/lib/api/client";
import { fileURL } from "@/lib/api/urls";
import { dateText } from "@/lib/format";
import type { Attachment, ReviewMessage, Submission } from "@/types/models";
import { useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ReviewStatus } from "./assignment-status";

export function SubmissionReview({
  submission,
  onClose,
  onChange,
}: {
  submission: Submission;
  onClose: () => void;
  onChange: () => void;
}) {
  const { user } = useSession();
  const q = useData(
    () => api<ReviewMessage[]>(`/submissions/${submission.id}/reviews`),
    [submission.id],
  );
  const [content, setContent] = useState("");
  const [files, setFiles] = useState<Attachment[]>([]);
  const [busy, setBusy] = useState(false);
  const [uploading, setUploading] = useState(false);
  const last = q.data?.at(-1);
  const status = last ? (last.admin ? "reviewed" : "pending") : submission.review_status;
  const canWrite = !submission.withdrawn && (user?.admin || !!q.data?.length);
  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !busy && !uploading) onClose();
      }}
    >
      <DialogContent
        className="sm:max-w-4xl"
        style={{ maxHeight: "90vh", overflow: "auto" }}
      >
        <DialogHeader>
          <DialogTitle>提交 #{submission.id} · 评价交流</DialogTitle>
          <DialogDescription>
            {submission.username} · {dateText(submission.created_at)} ·
            多位管理员共同参与，按发送顺序展示。
          </DialogDescription>
        </DialogHeader>
        <div className="row" style={{ justifyContent: "space-between" }}>
          <ReviewStatus status={status} withdrawn={submission.withdrawn} />
          <Button
            variant="ghost"
            size="sm"
            disabled={busy || uploading}
            onClick={q.reload}
          >
            刷新对话
          </Button>
        </div>
        <details>
          <summary className="link">查看原始提交</summary>
          <p style={{ whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}>
            {submission.content || "无文本内容"}
          </p>
          {submission.attachments?.map((f) => (
            <AttachmentPreview key={f.id} file={f} />
          ))}
        </details>
        <LoadState {...q} retry={q.reload}>
          <div className="stack" aria-live="polite">
            {!q.data?.length && <p className="muted">暂无评价，等待管理员开始交流。</p>}
            {q.data?.map((m) => (
              <article key={m.id} className="review-message">
                <div className="row">
                  {m.avatar_id ? (
                    <img
                      className="avatar"
                      src={fileURL(m.avatar_id, true)}
                      alt={`${m.username} 的头像`}
                    />
                  ) : (
                    <span className="avatar">{m.username.slice(0, 1).toUpperCase()}</span>
                  )}
                  <strong>{m.username || "已删除用户"}</strong>
                  <span className={`badge ${m.admin ? "green" : "gray"}`}>
                    {m.admin ? "管理员" : "提交者"}
                  </span>
                  <time className="muted small">{dateText(m.created_at)}</time>
                </div>
                {user?.admin && m.admin && (
                  <div className="row small" style={{ marginTop: 8 }}>
                    <span className="muted">
                      最后通知：
                      {m.last_notified_at ? dateText(m.last_notified_at) : "尚未发送"}
                      {m.notification_pending
                        ? " · 排队中"
                        : m.notification_status.startsWith("skipped_")
                          ? " · 已跳过"
                          : ""}
                    </span>
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={busy}
                      onClick={() => {
                        setBusy(true);
                        void perform(async () => {
                          await api(
                            `/admin/reviews/${m.id}/notification/resend`,
                            json("POST", {}),
                          );
                          q.reload();
                        }, "通知已加入发送队列").finally(() => setBusy(false));
                      }}
                    >
                      重新通知
                    </Button>
                  </div>
                )}
                <div className="markdown">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>{m.content}</ReactMarkdown>
                </div>
                {m.attachments.map((f) => (
                  <AttachmentPreview key={f.id} file={f} />
                ))}
              </article>
            ))}
          </div>
          {canWrite ? (
            <form
              className="stack"
              onSubmit={(e) => {
                e.preventDefault();
                if (busy || uploading) return;
                setBusy(true);
                void perform(async () => {
                  await api(
                    `/submissions/${submission.id}/reviews`,
                    json("POST", { content, attachment_ids: files.map((f) => f.id) }),
                  );
                  setContent("");
                  setFiles([]);
                  q.reload();
                  onChange();
                }, "发送成功").finally(() => setBusy(false));
              }}
            >
              <MarkdownEditor
                value={content}
                onChange={setContent}
                label={user?.admin ? "评价内容" : "回复内容"}
                placeholder="支持 Markdown，可补充附件…"
              />
              <UploadBox files={files} onChange={setFiles} onBusy={setUploading} />
              <p className="muted small">
                交流附件使用站点上传限制；管理员评价后为“已评价”，提交者回复后为“待回复”。
              </p>
              <Button
                type="submit"
                disabled={busy || uploading || (!content.trim() && !files.length)}
              >
                {busy ? "正在发送…" : user?.admin ? "发送评价" : "发送回复"}
              </Button>
            </form>
          ) : (
            <p className="muted small">
              {submission.withdrawn
                ? "此提交已撤回，对话仅供查看。"
                : "管理员评价后，你可以在这里回复。"}
            </p>
          )}
        </LoadState>
      </DialogContent>
    </Dialog>
  );
}
