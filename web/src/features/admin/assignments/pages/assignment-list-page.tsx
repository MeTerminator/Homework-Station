import { Empty } from "@/components/feedback/empty";
import { LoadState } from "@/components/feedback/load-state";
import { SearchBox } from "@/components/form/search-box";
import { PageHeading } from "@/components/page-heading";
import { Pager } from "@/components/pagination";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { AssignmentTable } from "@/features/admin/assignments/components/assignment-table";
import { useData } from "@/hooks/use-data";
import { api } from "@/lib/api/client";
import type { Assignment, Category, Page } from "@/types/models";
import { Plus } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
export function AssignmentList() {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(20);
  const q = useData(
    () =>
      api<Page<Assignment>>(
        `/admin/assignments?search=${encodeURIComponent(search)}&page=${page}&size=${size}`,
      ),
    [search, page, size],
  );
  const cats = useData(() => api<Category[]>("/categories"));
  return (
    <>
      <PageHeading title="作业管理" description="创建与管理作业，查看每一份学习成果。">
        <Link to="/admin/assignments/new">
          <Button>
            <Plus size={16} />
            创建作业
          </Button>
        </Link>
      </PageHeading>
      <Panel>
        <div className="toolbar">
          <SearchBox
            placeholder="搜索作业名称…"
            value={search}
            onChange={(s) => {
              setSearch(s);
              setPage(1);
            }}
          />
          <span className="small muted">全部作业 · {q.data?.total || 0} 项</span>
        </div>
        <LoadState {...q} retry={q.reload}>
          {q.data?.items.length ? (
            <AssignmentTable items={q.data.items} cats={cats.data || []} />
          ) : (
            <Empty
              title="暂无作业"
              description={
                search ? "尝试其他作业名称。" : "点击右上角创建你的第一份作业。"
              }
            />
          )}
        </LoadState>
        <Pager
          page={page}
          size={size}
          total={q.data?.total || 0}
          onPage={setPage}
          onSize={setSize}
        />
      </Panel>
    </>
  );
}
