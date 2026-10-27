import type { Category } from "@/types/models";
export function categoryLabel(id: number | null, cats: Category[]): string {
  if (id === null) return "未分类";
  const c = cats.find((c) => c.id === id);
  if (!c) return "未分类";
  const visited = new Set<number>();
  const names = [c.name];
  let parent = c.parent_id;
  while (parent && !visited.has(parent)) {
    visited.add(parent);
    const p = cats.find((x) => x.id === parent);
    if (!p) break;
    names.unshift(p.name);
    parent = p.parent_id;
  }
  return names.join(" / ");
}
