import { Empty } from "@/components/feedback/empty";
import { LoadState } from "@/components/feedback/load-state";
import { SearchBox } from "@/components/form/search-box";
import { PageHeading } from "@/components/page-heading";
import { Pager } from "@/components/pagination";
import { Panel } from "@/components/panel";
import { Avatar } from "@/features/profile/components/avatar";
import { useData } from "@/hooks/use-data";
import { api } from "@/lib/api/client";
import { dateText } from "@/lib/format";
import type { Page, User } from "@/types/models";
import { ChevronRight } from "lucide-react";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
export function UserList() {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(20);
  const q = useData(
    () =>
      api<Page<User>>(
        `/admin/users?search=${encodeURIComponent(search)}&page=${page}&size=${size}`,
      ),
    [search, page, size],
  );
  const nav = useNavigate();
  return (
    <>
      <PageHeading title="用户管理" description="了解用户信息，管理权限与身份绑定。" />
      <Panel>
        <div className="toolbar">
          <SearchBox
            placeholder="搜索用户名…"
            value={search}
            onChange={(s) => {
              setSearch(s);
              setPage(1);
            }}
          />
          <span className="muted small">共 {q.data?.total || 0} 位用户</span>
        </div>
        <LoadState {...q} retry={q.reload}>
          {q.data?.items.length ? (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    {["用户", "ID", "权限", "实名认证", "QQ", "注册时间", ""].map(
                      (s, i) => (
                        <th key={i}>{s}</th>
                      ),
                    )}
                  </tr>
                </thead>
                <tbody>
                  {q.data.items.map((u) => (
                    <tr
                      key={u.id}
                      className="click-row"
                      onClick={() => nav(`/admin/users/${u.id}`)}
                    >
                      <td>
                        <span className="row">
                          <Avatar user={u} />
                          <strong>
                            <Link to={`/admin/users/${u.id}`}>{u.username}</Link>
                          </strong>
                        </span>
                      </td>
                      <td className="muted">#{u.id}</td>
                      <td>
                        <span className={`badge ${u.admin ? "" : "gray"}`}>
                          {u.admin ? "管理员" : "用户"}
                        </span>
                      </td>
                      <td>{u.student_no ? "已认证" : "未认证"}</td>
                      <td className="muted">{u.qq || "未绑定"}</td>
                      <td className="muted">{dateText(u.created_at)}</td>
                      <td>
                        <ChevronRight size={15} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <Empty title="没有找到用户" description="尝试搜索其他用户名。" />
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
