import { LoadState } from "@/components/feedback/load-state";
import { SearchBox } from "@/components/form/search-box";
import { PageHeading } from "@/components/page-heading";
import { Pager } from "@/components/pagination";
import { Panel } from "@/components/panel";
import { AttachmentTable } from "@/features/attachments/components/attachment-table";
import { useData } from "@/hooks/use-data";
import { api } from "@/lib/api/client";
import type { Attachment } from "@/types/models";
import { useState } from "react";

export function MyAttachments() {
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(20);
  const [status, setStatus] = useState("");
  const [search, setSearch] = useState("");
  const q = useData(() => api<Attachment[]>("/attachments"));
  const term = search.trim().toLowerCase();
  const files = (q.data || []).filter(
    (file) =>
      (!status || file.status === status) &&
      (!term || file.name.toLowerCase().includes(term) || file.sha256.includes(term)),
  );
  const currentPage = Math.min(page, Math.max(1, Math.ceil(files.length / size)));

  return (
    <>
      <PageHeading title="附件管理" description="查看、下载和删除自己上传的附件。" />
      <Panel>
        <div className="toolbar">
          <SearchBox
            placeholder="搜索文件名或 SHA256…"
            value={search}
            onChange={(value) => {
              setSearch(value);
              setPage(1);
            }}
          />
          <select
            aria-label="附件状态筛选"
            value={status}
            onChange={(event) => {
              setStatus(event.target.value);
              setPage(1);
            }}
          >
            <option value="">全部上传状态</option>
            <option value="complete">已上传完成</option>
            <option value="uploading">未上传完成</option>
          </select>
        </div>
        <LoadState {...q} retry={q.reload}>
          <AttachmentTable
            admin={false}
            files={files.slice((currentPage - 1) * size, currentPage * size)}
            reload={q.reload}
          />
        </LoadState>
        <Pager
          page={currentPage}
          size={size}
          total={files.length}
          onPage={setPage}
          onSize={setSize}
        />
        <p className="small muted" style={{ marginTop: 18, lineHeight: 1.8 }}>
          未完成的附件暂不可下载，可回到作业页重新选择原文件继续上传。
          删除正在使用的附件会移除对应的提交附件或头像，请确认后操作。
        </p>
      </Panel>
    </>
  );
}
