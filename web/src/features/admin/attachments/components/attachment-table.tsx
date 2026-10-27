import { Confirm } from "@/components/feedback/confirm";
import { Empty } from "@/components/feedback/empty";
import { Button } from "@/components/ui/button";
import { api, json } from "@/lib/api/client";
import { fileURL } from "@/lib/api/urls";
import { dateText, sizeText } from "@/lib/format";
import type { Attachment } from "@/types/models";
import { Download, FileText, Trash2 } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
export function AttachmentTable({
  files,
  reload,
}: {
  files: Attachment[];
  reload: () => void;
}) {
  const [remove, setRemove] = useState<Attachment | null>(null);
  return (
    <>
      {files.length ? (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                {[
                  "文件名",
                  "用户",
                  "SHA256",
                  "大小",
                  "状态",
                  "上传时间",
                  "使用位置",
                  "操作",
                ].map((s) => (
                  <th key={s}>{s}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {files.map((f) => (
                <tr key={f.id}>
                  <td>
                    <span className="row overflow-name">
                      <FileText size={16} className="muted" />
                      {f.name}
                    </span>
                  </td>
                  <td>
                    <Link className="link" to={`/admin/users/${f.user_id}`}>
                      #{f.user_id}
                    </Link>
                  </td>
                  <td>
                    <div className="file-hash" title={f.sha256}>
                      {f.sha256}
                    </div>
                  </td>
                  <td>{sizeText(f.size)}</td>
                  <td>
                    <span
                      className={`badge ${f.status === "complete" ? "green" : "gray"}`}
                    >
                      {f.status === "complete"
                        ? "已完成"
                        : `上传中 ${f.received?.length || 0}/${f.total_chunks}`}
                    </span>
                  </td>
                  <td className="muted">{dateText(f.created_at, true)}</td>
                  <td
                    style={{
                      whiteSpace: "normal",
                      minWidth: 180,
                      maxWidth: 260,
                    }}
                  >
                    {f.usage?.length ? (
                      f.usage.map((s, i) => {
                        const match = s.match(/作业 #(\d+)/);
                        return (
                          <div key={i} style={{ marginBottom: 5 }}>
                            {match ? (
                              <Link
                                className="link small"
                                to={`/admin/assignments/${match[1]}`}
                              >
                                {s}
                              </Link>
                            ) : (
                              <span className="small muted">{s}</span>
                            )}
                          </div>
                        );
                      })
                    ) : (
                      <span className="muted small">未关联</span>
                    )}
                  </td>
                  <td>
                    <div className="row">
                      <a
                        href={fileURL(f.id)}
                        title={
                          f.status === "complete" ? "下载文件" : "下载已接收分片 ZIP"
                        }
                      >
                        <Button variant="ghost" size="icon" aria-label="下载附件">
                          <Download size={15} />
                        </Button>
                      </a>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setRemove(f)}
                        aria-label="删除附件"
                      >
                        <Trash2 size={15} />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <Empty
          title="暂无附件"
          description="上传中的分片和完成后的文件都会显示在这里。"
        />
      )}
      <Confirm
        open={!!remove}
        onClose={() => setRemove(null)}
        title={`删除「${remove?.name}」？`}
        description={
          remove?.usage?.length
            ? `此附件正在用于：${remove.usage.join("、")}。删除会解除这些关联；相同实体文件仅在最后一个副本删除后释放。`
            : "将永久删除附件及未完成分片。相同实体文件仅在最后一个副本删除后释放。"
        }
        onConfirm={async () => {
          await api(`/admin/attachments/${remove?.id}?force=true`, json("DELETE"));
          reload();
        }}
      />
    </>
  );
}
