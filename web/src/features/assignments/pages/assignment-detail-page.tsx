import { useSession } from "@/app/providers/session-provider";
import { Confirm } from "@/components/feedback/confirm";
import { LoadState } from "@/components/feedback/load-state";
import { Field } from "@/components/form/field";
import { PageHeading } from "@/components/page-heading";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Status } from "@/features/assignments/components/assignment-status";
import { SubmissionHistory } from "@/features/assignments/components/submission-history";
import { UploadBox } from "@/features/attachments/components/upload-box";
import { useData } from "@/hooks/use-data";
import { perform } from "@/lib/actions";
import { api, json } from "@/lib/api/client";
import { fileURL } from "@/lib/api/urls";
import { dateText, sizeText } from "@/lib/format";
import type { Assignment, Attachment, Submission } from "@/types/models";
import { ArrowLeft, Clock, Paperclip, Send } from "lucide-react";
import { useState } from "react";
import ReactMarkdown from "react-markdown";
import { Link, useParams } from "react-router-dom";
import remarkGfm from "remark-gfm";
export function AssignmentDetail() {
  const { id } = useParams();
  const q = useData(
    () =>
      api<{ assignment: Assignment; attachments: Attachment[] }>(`/assignments/${id}`),
    [id],
  );
  const mine = useData(() => api<Submission[]>(`/assignments/${id}/submissions`), [id]);
  const [content, setContent] = useState("");
  const [files, setFiles] = useState<Attachment[]>([]);
  const [busy, setBusy] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [version, setVersion] = useState(0);
  const [withdraw, setWithdraw] = useState<Submission | null>(null);
  const { user } = useSession();
  const a = q.data?.assignment;
  const own = (mine.data || []).filter((s) => s.user_id === user?.id);
  const active = own.filter((s) => !s.withdrawn);
  const missingRealName = !!a?.require_real_name && !user?.student_no;
  const missingQQ = !!a?.require_qq && !user?.qq;
  const canSubmit =
    a?.status === "published" &&
    (!a.deadline || new Date(a.deadline).getTime() > Date.now()) &&
    (a.multiple || active.length === 0);
  return (
    <>
      <PageHeading
        title={a?.name || "作业详情"}
        description={a ? a.summary || "暂无作业简介" : ""}
      >
        <Link to="/">
          <Button variant="outline">
            <ArrowLeft size={15} />
            返回作业
          </Button>
        </Link>
      </PageHeading>
      <LoadState {...q} retry={q.reload}>
        {a && (
          <div className="assignment-detail-layout">
            <Panel className="assignment-detail-description">
              <div className="row" style={{ marginBottom: 22 }}>
                <Status status={a.status} />
                <span className="muted small row">
                  <Clock size={14} />
                  {dateText(a.deadline)}
                </span>
                <span className="muted small">{a.submitters} 人已提交</span>
              </div>
              <div className="markdown">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {a.description || "本作业暂无补充说明。"}
                </ReactMarkdown>
              </div>
              {q.data!.attachments.length > 0 && (
                <div className="stack" style={{ marginTop: 25, gap: 10 }}>
                  {q.data!.attachments.map((f) => (
                    <a className="upload-file" key={f.id} href={fileURL(f.id)}>
                      <Paperclip size={15} />
                      <span className="link">{f.name}</span>
                      <small className="muted">{sizeText(f.size)}</small>
                    </a>
                  ))}
                </div>
              )}
            </Panel>
            <div className="assignment-detail-sidebar stack">
              <Panel
                title="提交作业"
                action={
                  <span className="muted small">
                    {a.multiple ? "可多次提交" : "仅允许一次有效提交"}
                  </span>
                }
              >
                {(a.require_real_name || a.require_qq) && (
                  <p className="muted small" style={{ marginBottom: 16 }}>
                    提交条件：
                    {[
                      a.require_real_name ? "完成实名认证" : "",
                      a.require_qq ? "绑定 QQ" : "",
                    ]
                      .filter(Boolean)
                      .join("、")}
                    {a.require_real_name && a.require_qq ? "（需全部满足）" : ""}
                  </p>
                )}
                {canSubmit && (missingRealName || missingQQ) ? (
                  <div className="note">
                    请先
                    {[missingRealName ? "完成实名认证" : "", missingQQ ? "绑定 QQ" : ""]
                      .filter(Boolean)
                      .join("并")}
                    后再提交作业。
                    <Link className="link" to="/profile">
                      前往个人信息
                    </Link>
                  </div>
                ) : canSubmit ? (
                  <form
                    className="stack"
                    onSubmit={(e) => {
                      e.preventDefault();
                      setBusy(true);
                      perform(async () => {
                        await api(
                          `/assignments/${id}/submissions`,
                          json("POST", {
                            content,
                            attachment_ids: files.map((f) => f.id),
                          }),
                        );
                        setContent("");
                        setFiles([]);
                        q.reload();
                        mine.reload();
                        setVersion((v) => v + 1);
                      }, "作业提交成功").finally(() => setBusy(false));
                    }}
                  >
                    <Field label="作业内容">
                      <Textarea
                        rows={7}
                        value={content}
                        onChange={(e) => setContent(e.target.value)}
                        placeholder="在这里填写你的作业内容、思路或补充说明…"
                      />
                    </Field>
                    {a.allow_attachments && (
                      <Field
                        label="作业附件"
                        hint={`允许的后缀：${a.extensions || "全部"}；单个附件最大 ${sizeText(a.max_size)}`}
                      >
                        <UploadBox
                          files={files}
                          onChange={setFiles}
                          assignmentID={a.id}
                          maxSize={a.max_size}
                          extensions={a.extensions}
                          onBusy={setUploading}
                        />
                      </Field>
                    )}
                    <div className="row" style={{ justifyContent: "space-between" }}>
                      <span className="small muted">
                        {a.withdraw ? "提交后可以撤回" : "提交后不允许撤回"}
                      </span>
                      <Button
                        type="submit"
                        disabled={busy || uploading || (!content.trim() && !files.length)}
                      >
                        <Send size={15} />
                        {busy ? "正在提交…" : "提交作业"}
                      </Button>
                    </div>
                  </form>
                ) : (
                  <div className="note">
                    {a.status === "closed" ||
                    (a.deadline && new Date(a.deadline).getTime() <= Date.now())
                      ? "此作业已截止。"
                      : a.status === "hidden"
                        ? "此作业未发布。"
                        : "你已完成提交，此作业不允许重复提交。"}
                  </div>
                )}
              </Panel>
              {a.withdraw && active.length > 0 && (
                <Panel title="撤回提交">
                  <div className="stack" style={{ gap: 12 }}>
                    {active.map((s) => (
                      <div
                        className="row"
                        style={{ justifyContent: "space-between" }}
                        key={s.id}
                      >
                        <span className="small muted">
                          提交 #{s.id} · {dateText(s.created_at)}
                        </span>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setWithdraw(s)}
                        >
                          撤回这次提交
                        </Button>
                      </div>
                    ))}
                  </div>
                </Panel>
              )}
              <SubmissionHistory key={version} assignmentID={a.id} admin={false} />
            </div>
          </div>
        )}
      </LoadState>
      <Confirm
        open={!!withdraw}
        onClose={() => setWithdraw(null)}
        title="撤回这次提交？"
        description="提交记录仍会保留在历史中，状态变为已撤回。是否能再次提交取决于作业开放状态。"
        onConfirm={async () => {
          await api(`/submissions/${withdraw?.id}/withdraw`, json("POST"));
          q.reload();
          mine.reload();
          setVersion((v) => v + 1);
        }}
      />
    </>
  );
}
