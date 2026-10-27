import { useSession } from "@/app/providers/session-provider";
import { Empty } from "@/components/feedback/empty";
import { LoadState } from "@/components/feedback/load-state";
import { CheckField } from "@/components/form/check-field";
import { SearchBox } from "@/components/form/search-box";
import { PageHeading } from "@/components/page-heading";
import { Panel } from "@/components/panel";
import { AssignmentCategorySection } from "@/features/assignments/components/assignment-category-section";
import { CategoryFilter } from "@/features/categories/components/category-filter";
import {
  categoryTree,
  descendantIDs,
  flattenCategories,
  groupAssignments,
} from "@/features/categories/lib/category-tree";
import { selectedCategories } from "@/features/categories/lib/preferences";
import { useData } from "@/hooks/use-data";
import { perform } from "@/lib/actions";
import { api, json } from "@/lib/api/client";
import type { Assignment, Category, Page, User } from "@/types/models";
import { GraduationCap } from "lucide-react";
import { useEffect, useState } from "react";
export function MyAssignments() {
  const { user, setUser } = useSession();
  const [search, setSearch] = useState("");
  const [onlyPending, setOnlyPending] = useState(false);
  const cats = useData(() => api<Category[]>("/categories"));
  const q = useData(
    () => api<Page<Assignment>>(`/assignments?search=${encodeURIComponent(search)}`),
    [search],
  );
  const [selected, setSelected] = useState<number[]>(() => selectedCategories(user));
  useEffect(() => setSelected(selectedCategories(user)), [user?.id]);
  const toggle = (id: number) => {
    const next = selected.includes(id)
      ? selected.filter((n) => n !== id)
      : [...selected, id];
    perform(async () => {
      const u = await api<User>("/me", json("PATCH", { category_ids: next }));
      setSelected(next);
      setUser(u);
    });
  };
  const tree = categoryTree(cats.data || []);
  const nodes = flattenCategories(tree);
  const selectedIDs = new Set(
    selected.flatMap((id) => {
      const node = nodes.find((n) => n.category.id === id);
      return node ? descendantIDs([node]) : [];
    }),
  );
  const all = q.data?.items || [];
  const visible = all.filter(
    (a) =>
      (selected.length === 0 || selectedIDs.has(a.category_id || 0)) &&
      (!onlyPending || !a.submitted),
  );
  const groups = groupAssignments(cats.data || [], visible);
  const pending = all.filter((a) => !a.submitted && a.status === "published").length;
  return (
    <>
      <PageHeading title="我的作业" description="查看作业要求与提交状态。" />
      <div className="hero-strip">
        <div>
          <h2>
            你好，{user?.real_name || user?.username}{" "}
            <span style={{ fontSize: 19 }}>☀</span>
          </h2>
          <p>
            你有 <strong style={{ color: "var(--primary)" }}>{pending}</strong>{" "}
            项可提交的作业等待完成。
          </p>
        </div>
        <GraduationCap
          size={56}
          style={{ color: "var(--primary)", opacity: 0.5, flexShrink: 0 }}
        />
      </div>

      <div className="assignment-browser">
        <div className="assignment-results">
          <LoadState
            loading={q.loading || cats.loading}
            error={q.error || cats.error}
            retry={() => {
              q.reload();
              cats.reload();
            }}
          >
            {groups.length ? (
              groups.map((group) => (
                <AssignmentCategorySection key={group.id} group={group} />
              ))
            ) : (
              <Panel className="mt-6">
                <Empty
                  title={all.length ? "没有符合条件的作业" : "暂时还没有作业"}
                  description={
                    all.length
                      ? "调整搜索内容或分类，看看其他作业。"
                      : "发布后的作业会出现在这里。"
                  }
                />
              </Panel>
            )}
          </LoadState>
        </div>
        <aside className="assignment-filters" aria-label="作业筛选">
          <Panel>
            <div className="toolbar">
              <SearchBox
                value={search}
                onChange={setSearch}
                placeholder="搜索作业名称…"
              />
              <CheckField
                label="仅查看未提交作业"
                checked={onlyPending}
                onChange={setOnlyPending}
              />
            </div>
            <div className="row" style={{ alignItems: "flex-start" }}>
              <span className="muted small" style={{ paddingTop: 8, minWidth: 55 }}>
                学习方向
              </span>
              <div className="category-chips">
                <button
                  className={`chip ${selected.length === 0 ? "selected" : ""}`}
                  onClick={() =>
                    perform(async () => {
                      const u = await api<User>(
                        "/me",
                        json("PATCH", { category_ids: [] }),
                      );
                      setSelected([]);
                      setUser(u);
                    })
                  }
                >
                  全部方向
                </button>
                <div className="category-filter-tree">
                  {tree.map((node) => (
                    <CategoryFilter
                      key={node.category.id}
                      node={node}
                      selected={selected}
                      onToggle={toggle}
                      cats={cats.data || []}
                    />
                  ))}
                </div>
              </div>
            </div>
            <p className="small muted" style={{ marginTop: 15, fontSize: 11 }}>
              可多选，偏好会保存到你的账号；选择父分类会同时展示其子分类作业。
            </p>
            {cats.error && (
              <p className="small" style={{ color: "#dc6472", marginTop: 10 }}>
                {cats.error}
              </p>
            )}
          </Panel>
        </aside>
      </div>
    </>
  );
}
