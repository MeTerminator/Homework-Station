import { Confirm } from "@/components/feedback/confirm";
import { Empty } from "@/components/feedback/empty";
import { LoadState } from "@/components/feedback/load-state";
import { Field } from "@/components/form/field";
import { PageHeading } from "@/components/page-heading";
import { Panel } from "@/components/panel";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { CategorySelect } from "@/features/categories/components/category-select";
import { categoryTree, flattenCategories } from "@/features/categories/lib/category-tree";
import { useData } from "@/hooks/use-data";
import { perform } from "@/lib/actions";
import { api, json } from "@/lib/api/client";
import type { Category } from "@/types/models";
import { FolderTree, Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
export function CategoryManager() {
  const q = useData(() => api<Category[]>("/categories"));
  const [edit, setEdit] = useState<Category | null>(null);
  const [remove, setRemove] = useState<Category | null>(null);
  const [busy, setBusy] = useState(false);
  const all = q.data || [];
  const ordered = flattenCategories(categoryTree(all));
  return (
    <>
      <PageHeading title="作业分类" description="按学习方向组织作业，支持多级分类。">
        <Button onClick={() => setEdit({ id: 0, name: "", parent_id: null, sort: 0 })}>
          <Plus size={16} />
          新建分类
        </Button>
      </PageHeading>
      <Panel
        title="全部分类"
        action={<span className="muted small">{all.length} 个分类</span>}
      >
        <LoadState {...q} retry={q.reload}>
          {ordered.length ? (
            ordered.map(({ category: c, depth }) => (
              <div className="category-row" key={c.id}>
                <span
                  className="row category-row-label"
                  style={{ paddingLeft: depth * 28 }}
                >
                  <FolderTree size={17} style={{ color: "var(--primary)" }} />
                  <span>{c.name}</span>
                  <span className="muted small">排序 {c.sort}</span>
                </span>
                <div className="row">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setEdit({ id: 0, name: "", parent_id: c.id, sort: 0 })}
                  >
                    <Plus size={14} />
                    子分类
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setEdit(c)}
                    aria-label="编辑分类"
                  >
                    <Pencil size={15} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setRemove(c)}
                    aria-label="删除分类"
                  >
                    <Trash2 size={15} />
                  </Button>
                </div>
              </div>
            ))
          ) : (
            <Empty
              title="还没有分类"
              description="试着创建前端、后端、运维或安全分类。"
            />
          )}
        </LoadState>
      </Panel>
      <Dialog open={!!edit} onOpenChange={(v) => !v && setEdit(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{edit?.id ? "编辑分类" : "新建分类"}</DialogTitle>
            <DialogDescription>
              支持多级嵌套；分类不能移动到自身或其子分类下。
            </DialogDescription>
          </DialogHeader>
          {edit && (
            <form
              className="stack"
              onSubmit={(e) => {
                e.preventDefault();
                setBusy(true);
                perform(async () => {
                  await api(
                    edit.id ? `/admin/categories/${edit.id}` : "/admin/categories",
                    json(edit.id ? "PUT" : "POST", edit),
                  );
                  setEdit(null);
                  q.reload();
                }, "分类已保存").finally(() => setBusy(false));
              }}
            >
              <Field label="分类名称">
                <Input
                  required
                  value={edit.name}
                  onChange={(e) => setEdit({ ...edit, name: e.target.value })}
                />
              </Field>
              <Field label="父级分类">
                <CategorySelect
                  exclude={edit.id}
                  cats={all}
                  value={edit.parent_id}
                  onChange={(v) => setEdit({ ...edit, parent_id: v })}
                />
              </Field>
              <Field label="排序">
                <Input
                  type="number"
                  value={edit.sort}
                  onChange={(e) => setEdit({ ...edit, sort: +e.target.value })}
                />
              </Field>
              <Button disabled={busy} type="submit">
                保存分类
              </Button>
            </form>
          )}
        </DialogContent>
      </Dialog>
      <Confirm
        open={!!remove}
        onClose={() => setRemove(null)}
        title={`删除「${remove?.name}」？`}
        description="存在子分类或作业的分类不能直接删除，请先移动它们。"
        onConfirm={async () => {
          await api(`/admin/categories/${remove?.id}`, json("DELETE"));
          q.reload();
        }}
      />
    </>
  );
}
