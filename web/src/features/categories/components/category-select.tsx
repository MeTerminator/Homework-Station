import {
  categoryTree,
  flattenCategories,
  parentCategoryOptions,
} from "@/features/categories/lib/category-tree";
import type { Category } from "@/types/models";
export function CategorySelect({
  value,
  onChange,
  cats,
  exclude,
}: {
  value: number | null;
  onChange: (v: number | null) => void;
  cats: Category[];
  exclude?: number;
}) {
  return (
    <select
      value={value ?? ""}
      onChange={(e) => onChange(e.target.value ? +e.target.value : null)}
      aria-label="选择分类"
    >
      <option value="">未分类 / 无</option>
      {(exclude === undefined
        ? flattenCategories(categoryTree(cats))
        : parentCategoryOptions(cats, exclude)
      ).map((node) => (
        <option key={node.category.id} value={node.category.id}>
          {"　".repeat(node.depth)}
          {node.depth > 0 ? "└ " : ""}
          {node.category.name}
        </option>
      ))}
    </select>
  );
}
