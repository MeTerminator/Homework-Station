import { ReviewStatus } from "./assignment-status";
import { SubmissionReview } from "./submission-review";
import { SubmissionAttachments } from "./submission-attachments";
import { Confirm } from "@/components/feedback/confirm";
import { Pager } from "@/components/pagination";
import { AttachmentPreview } from "@/features/attachments/components/attachment-preview";
import { Empty } from "@/components/feedback/empty";
import { LoadState } from "@/components/feedback/load-state";
import { CheckField } from "@/components/form/check-field";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useData } from "@/hooks/use-data";
import { api, json } from "@/lib/api/client";
import { dateText } from "@/lib/format";
import type { Submission, Page } from "@/types/models";
import { Trash2 } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
export function SubmissionHistory({
  assignmentID,
  admin = true,
}: {
  assignmentID?: number;
  admin?: boolean;
}) {
  const recent = assignmentID === undefined;
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(20);
  const q = useData(async () => {
    if (recent)
      return api<Page<Submission>>(`/admin/submissions?page=${page}&size=${size}`);
    const items = await api<Submission[]>(`/assignments/${assignmentID}/submissions`);
    return { items, total: items.length };
  }, [assignmentID, page, size]);
  const [review, setReview] = useState<Submission | null>(null);
  const [selected, setSelected] = useState<Submission | null>(null);
  const [remove, setRemove] = useState<Submission | null>(null);
  const [deleteFiles, setDeleteFiles] = useState(false);
  return (
    <Panel
      title={recent ? "近期提交" : admin ? "全部历史提交" : "我的历史提交"}
      action={
        <Button variant="ghost" size="sm" onClick={q.reload}>
          刷新
        </Button>
      }
    >
      <LoadState {...q} retry={q.reload}>
        {q.data?.items.length ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>提交 ID</th>
                  {recent && <th>作业</th>}
                  {admin && <th>用户</th>}
                  <th>提交时间</th>
                  <th>状态</th>
                  <th>附件</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {q.data.items.map((s) => (
                  <tr key={s.id}>
                    <td>#{s.id}</td>
                    {recent && (
                      <td>
                        <Link
                          className="link"
                          to={`/admin/assignments/${s.assignment_id}`}
                        >
                          {s.assignment_name || `作业 #${s.assignment_id}`}
                        </Link>
                      </td>
                    )}
                    {admin && (
                      <td>
                        <Link to={`/admin/users/${s.user_id}`} className="link">
                          {s.username}
                        </Link>
                      </td>
                    )}
                    <td className="muted">{dateText(s.created_at)}</td>
                    <td>
                      <ReviewStatus status={s.review_status} withdrawn={s.withdrawn} />
                    </td>
                    <td>
                      <SubmissionAttachments files={s.attachments || []} />
                    </td>
                    <td>
                      <div className="row">
                        <Button size="sm" variant="ghost" onClick={() => setSelected(s)}>
                          查看
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => setReview(s)}>
                          {admin ? "评价" : "评价交流"}
                        </Button>
                        {admin && (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => {
                              setRemove(s);
                              setDeleteFiles(false);
                            }}
                            aria-label="删除提交"
                          >
                            <Trash2 size={14} />
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <Empty title="还没有提交记录" description="提交后的内容和附件会保存在这里。" />
        )}
      </LoadState>
      {review && (
        <SubmissionReview
          key={review.id}
          submission={review}
          onClose={() => setReview(null)}
          onChange={() => {
            q.reload();
          }}
        />
      )}
      <Dialog open={!!selected} onOpenChange={(v) => !v && setSelected(null)}>
        <DialogContent
          className="sm:max-w-2xl"
          style={{ maxHeight: "85vh", overflow: "auto" }}
        >
          <DialogHeader>
            <DialogTitle>提交 #{selected?.id}</DialogTitle>
            <DialogDescription>
              {selected?.username} · {dateText(selected?.created_at || null)}
            </DialogDescription>
          </DialogHeader>
          <div
            style={{
              whiteSpace: "pre-wrap",
              overflowWrap: "anywhere",
              lineHeight: 1.8,
            }}
          >
            {selected?.content || "无文本内容"}
          </div>
          {selected?.attachments.map((f) => (
            <AttachmentPreview key={f.id} file={f} />
          ))}
        </DialogContent>
      </Dialog>
      {recent && (
        <Pager
          page={page}
          size={size}
          total={q.data?.total || 0}
          onPage={setPage}
          onSize={setSize}
        />
      )}
      <Confirm
        open={!!remove}
        onClose={() => setRemove(null)}
        title={`删除提交 #${remove?.id}？`}
        description="此记录及全部评价对话将永久删除。共用附件仅在没有其他关联时清理。"
        onConfirm={async () => {
          await api(
            `/admin/submissions/${remove?.id}?files=${deleteFiles}`,
            json("DELETE"),
          );
          if (selected?.id === remove?.id) setSelected(null);
          if (recent && q.data?.items.length === 1 && page > 1) setPage(page - 1);
          else q.reload();
        }}
      >
        <CheckField
          label="同时删除不再使用的关联附件，释放空间"
          checked={deleteFiles}
          onChange={setDeleteFiles}
        />
      </Confirm>
    </Panel>
  );
}
