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
import { perform } from "@/lib/actions";
import { api, json } from "@/lib/api/client";
import { fileURL } from "@/lib/api/urls";
import { dateText, sizeText } from "@/lib/format";
import type { Submission } from "@/types/models";
import { Download, Trash2 } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
export function SubmissionHistory({
  assignmentID,
  admin = true,
}: {
  assignmentID: number;
  admin?: boolean;
}) {
  const q = useData(
    () => api<Submission[]>(`/assignments/${assignmentID}/submissions`),
    [assignmentID],
  );
  const [selected, setSelected] = useState<Submission | null>(null);
  const [remove, setRemove] = useState<Submission | null>(null);
  const [deleteFiles, setDeleteFiles] = useState(false);
  return (
    <Panel
      title={admin ? "全部历史提交" : "我的历史提交"}
      action={
        <Button variant="ghost" size="sm" onClick={q.reload}>
          刷新
        </Button>
      }
    >
      <LoadState {...q} retry={q.reload}>
        {q.data?.length ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>提交 ID</th>
                  {admin && <th>用户</th>}
                  <th>提交时间</th>
                  <th>状态</th>
                  <th>附件</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {q.data.map((s) => (
                  <tr key={s.id}>
                    <td>#{s.id}</td>
                    {admin && (
                      <td>
                        <Link to={`/admin/users/${s.user_id}`} className="link">
                          {s.username}
                        </Link>
                      </td>
                    )}
                    <td className="muted">{dateText(s.created_at)}</td>
                    <td>
                      <span className={`badge ${s.withdrawn ? "gray" : "green"}`}>
                        {s.withdrawn ? "已撤回" : "已提交"}
                      </span>
                    </td>
                    <td>{s.attachments?.length || 0} 个</td>
                    <td>
                      <div className="row">
                        <Button size="sm" variant="ghost" onClick={() => setSelected(s)}>
                          查看
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
            <a className="upload-file link" key={f.id} href={fileURL(f.id)}>
              <Download size={15} />
              {f.name}
              <span>{sizeText(f.size)}</span>
            </a>
          ))}
        </DialogContent>
      </Dialog>
      <Dialog open={!!remove} onOpenChange={(v) => !v && setRemove(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>删除提交 #{remove?.id}？</DialogTitle>
            <DialogDescription>
              此记录将永久删除。共用附件仅在没有其他关联时清理。
            </DialogDescription>
          </DialogHeader>
          <CheckField
            label="同时删除不再使用的关联附件，释放空间"
            checked={deleteFiles}
            onChange={setDeleteFiles}
          />
          <Button
            variant="destructive"
            onClick={() =>
              perform(async () => {
                await api(
                  `/admin/submissions/${remove?.id}?files=${deleteFiles}`,
                  json("DELETE"),
                );
                setRemove(null);
                q.reload();
              }, "提交已删除")
            }
          >
            确认删除
          </Button>
        </DialogContent>
      </Dialog>
    </Panel>
  );
}
