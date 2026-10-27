import { LoadState } from "@/components/feedback/load-state";
import { SearchBox } from "@/components/form/search-box";
import { PageHeading } from "@/components/page-heading";
import { Pager } from "@/components/pagination";
import { Panel } from "@/components/panel";
import { AttachmentTable } from "@/features/admin/attachments/components/attachment-table";
import { useData } from "@/hooks/use-data";
import { api } from "@/lib/api/client";
import type { Attachment, Page } from "@/types/models";
import { useState } from "react";
export function AttachmentManager() {
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(20);
  const [status, setStatus] = useState("");
  const [search, setSearch] = useState("");
  const q = useData(
    () =>
      api<Page<Attachment>>(
        `/admin/attachments?page=${page}&size=${size}&status=${status}&search=${encodeURIComponent(search)}`,
      ),
    [page, size, status, search],
  );
  return (
    <>
      <PageHeading
        title="附件管理"
        description="统一管理附件与上传分片，相同文件只占用一份存储空间。"
      />
      <Panel>
        <div className="toolbar">
          <SearchBox
            placeholder="搜索文件名或 SHA256…"
            value={search}
            onChange={(s) => {
              setSearch(s);
              setPage(1);
            }}
          />
          <select
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
            aria-label="附件状态筛选"
          >
            <option value="">全部上传状态</option>
            <option value="complete">已上传完成</option>
            <option value="uploading">未上传完成</option>
          </select>
        </div>
        <LoadState {...q} retry={q.reload}>
          <AttachmentTable files={q.data?.items || []} reload={q.reload} />
        </LoadState>
        <Pager
          page={page}
          size={size}
          total={q.data?.total || 0}
          onPage={setPage}
          onSize={setSize}
        />
        <p className="small muted" style={{ marginTop: 18, lineHeight: 1.8 }}>
          未完成附件可下载为 ZIP，包含已接收的分片及
          manifest.json；重新选择原文件可继续上传。
        </p>
      </Panel>
    </>
  );
}
