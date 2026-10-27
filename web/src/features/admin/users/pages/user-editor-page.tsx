import { useSession } from "@/app/providers/session-provider";
import { Confirm } from "@/components/feedback/confirm";
import { Empty } from "@/components/feedback/empty";
import { LoadState } from "@/components/feedback/load-state";
import { CheckField } from "@/components/form/check-field";
import { Field } from "@/components/form/field";
import { PageHeading } from "@/components/page-heading";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { AttachmentTable } from "@/features/admin/attachments/components/attachment-table";
import { categoryLabel } from "@/features/categories/lib/category-label";
import { selectedCategories } from "@/features/categories/lib/preferences";
import { Avatar } from "@/features/profile/components/avatar";
import { useData } from "@/hooks/use-data";
import { perform } from "@/lib/actions";
import { api, json } from "@/lib/api/client";
import { dateText } from "@/lib/format";
import type { Attachment, Category, Submission, User } from "@/types/models";
import { ArrowUpRight, BookOpen } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
export function UserEditor() {
  const { id } = useParams();
  const q = useData(
    () =>
      api<{ user: User; submissions: Submission[]; attachments: Attachment[] }>(
        `/admin/users/${id}`,
      ),
    [id],
  );
  const cats = useData(() => api<Category[]>("/categories"));
  const [form, setForm] = useState({
    username: "",
    admin: false,
    password: "",
  });
  const [busy, setBusy] = useState(false);
  const [unbind, setUnbind] = useState("");
  const { user, setUser } = useSession();
  useEffect(() => {
    if (q.data)
      setForm({
        username: q.data.user.username,
        admin: q.data.user.admin,
        password: "",
      });
  }, [q.data]);
  return (
    <>
      <PageHeading title="编辑用户" description="管理用户资料、权限及绑定信息。">
        <Link to="/admin/users">
          <Button variant="outline">返回列表</Button>
        </Link>
      </PageHeading>
      <LoadState {...q} retry={q.reload}>
        {q.data && (
          <div className="stack">
            <Panel title="用户信息">
              <div className="profile-summary">
                <Avatar user={q.data.user} />
                <div>
                  <h2>{q.data.user.username}</h2>
                  <p className="small muted" style={{ marginTop: 7 }}>
                    用户 #{id} · 注册于 {dateText(q.data.user.created_at)}
                  </p>
                </div>
              </div>
              <form
                className="form-grid"
                onSubmit={(e) => {
                  e.preventDefault();
                  setBusy(true);
                  perform(async () => {
                    const v = await api<User>(`/admin/users/${id}`, json("PATCH", form));
                    if (v.id === user?.id) setUser(v);
                    q.reload();
                  }, "用户已保存").finally(() => setBusy(false));
                }}
              >
                <Field label="用户名">
                  <Input
                    value={form.username}
                    required
                    onChange={(e) => setForm({ ...form, username: e.target.value })}
                  />
                </Field>
                <Field
                  label="重置密码"
                  hint="留空则保留现有密码。修改密码后，其他设备需重新登录。"
                >
                  <Input
                    type="password"
                    autoComplete="new-password"
                    minLength={6}
                    maxLength={72}
                    value={form.password}
                    placeholder="输入新密码"
                    onChange={(e) => setForm({ ...form, password: e.target.value })}
                  />
                </Field>
                <div className="full">
                  <CheckField
                    label="设置为管理员"
                    checked={form.admin}
                    onChange={(v) => setForm({ ...form, admin: v })}
                  />
                </div>
                <div>
                  <Button type="submit" disabled={busy}>
                    保存用户
                  </Button>
                </div>
              </form>
            </Panel>
            <div className="grid-two">
              <Panel title="实名认证">
                <p>
                  {q.data.user.student_no
                    ? `${q.data.user.real_name} · ${q.data.user.student_no}`
                    : "尚未绑定实名信息"}
                </p>
                {q.data.user.student_no && (
                  <Button
                    variant="outline"
                    size="sm"
                    style={{ marginTop: 20 }}
                    onClick={() => setUnbind("real")}
                  >
                    解除实名绑定
                  </Button>
                )}
              </Panel>
              <Panel title="QQ 绑定">
                <p>{q.data.user.qq || "尚未绑定 QQ"}</p>
                {q.data.user.qq && (
                  <Button
                    variant="outline"
                    size="sm"
                    style={{ marginTop: 20 }}
                    onClick={() => setUnbind("qq")}
                  >
                    解除 QQ 绑定
                  </Button>
                )}
              </Panel>
            </div>
            <Panel title="用户关注的分类">
              <div className="category-chips">
                {selectedCategories(q.data.user).length ? (
                  selectedCategories(q.data.user).map((n) => (
                    <span className="badge" key={n}>
                      {categoryLabel(n, cats.data || [])}
                    </span>
                  ))
                ) : (
                  <span className="muted small">未筛选，查看全部分类</span>
                )}
              </div>
            </Panel>
            <Panel title="用户全部作业提交">
              {q.data.submissions.length ? (
                <div className="table-wrap">
                  <table>
                    <thead>
                      <tr>
                        <th>提交 ID</th>
                        <th>作业</th>
                        <th>提交内容</th>
                        <th>时间</th>
                        <th>状态</th>
                      </tr>
                    </thead>
                    <tbody>
                      {q.data.submissions.map((s) => (
                        <tr key={s.id}>
                          <td>#{s.id}</td>
                          <td>
                            <Link
                              className="link row"
                              to={`/admin/assignments/${s.assignment_id}`}
                            >
                              <BookOpen size={14} />
                              作业 #{s.assignment_id}
                              <ArrowUpRight size={13} />
                            </Link>
                          </td>
                          <td className="overflow-name">
                            {s.content || `${s.attachments.length} 个附件`}
                          </td>
                          <td className="muted">{dateText(s.created_at)}</td>
                          <td>
                            <span className={`badge ${s.withdrawn ? "gray" : "green"}`}>
                              {s.withdrawn ? "已撤回" : "已提交"}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <Empty title="还没有提交过作业" />
              )}
            </Panel>
            <Panel title="用户全部附件">
              <AttachmentTable files={q.data.attachments} reload={q.reload} />
            </Panel>
          </div>
        )}
      </LoadState>
      <Confirm
        open={!!unbind}
        onClose={() => setUnbind("")}
        title="解除此绑定？"
        description="解除后，用户可以重新验证并绑定身份。"
        onConfirm={async () => {
          await api(`/admin/users/${id}/unbind/${unbind}`, json("POST"));
          q.reload();
        }}
      />
    </>
  );
}
