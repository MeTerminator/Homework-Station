import { categoryLabel } from "@/features/categories/lib/category-label";
import type { CategoryNode } from "@/features/categories/lib/category-tree";
import type { Category } from "@/types/models";
export function CategoryFilter({
  node,
  selected,
  onToggle,
  cats,
}: {
  node: CategoryNode;
  selected: number[];
  onToggle: (id: number) => void;
  cats: Category[];
}) {
  return (
    <div className="category-filter-branch">
      <button
        className={`chip ${selected.includes(node.category.id) ? "selected" : ""}`}
        aria-pressed={selected.includes(node.category.id)}
        title={categoryLabel(node.category.id, cats)}
        onClick={() => onToggle(node.category.id)}
      >
        {node.category.name}
      </button>
      {node.children.length > 0 && (
        <div className="category-filter-children">
          {node.children.map((child) => (
            <CategoryFilter
              key={child.category.id}
              node={child}
              selected={selected}
              onToggle={onToggle}
              cats={cats}
            />
          ))}
        </div>
      )}
    </div>
  );
}
